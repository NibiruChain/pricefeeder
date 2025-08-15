package priceprovider

import (
	"encoding/json"

	"github.com/NibiruChain/nibiru/v2/x/common/asset"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
)

var _ types.PriceProvider = (*AggregatePriceProvider)(nil)

// AggregatePriceProvider aggregates multiple price providers
// and queries them for prices.
type AggregatePriceProvider struct {
	logger    zerolog.Logger
	providers map[int]types.PriceProvider // we use a map here to provide random ranging (since golang's map range is unordered)
}

// NewAggregatePriceProvider instantiates a new AggregatePriceProvider instance
// given multiple PriceProvider.
func NewAggregatePriceProvider(
	sourcesToPairSymbolMap map[string]map[asset.Pair]types.Symbol,
	sourceConfigMap map[string]json.RawMessage,
	logger zerolog.Logger,
) types.PriceProvider {
	providers := make(map[int]types.PriceProvider, len(sourcesToPairSymbolMap))
	i := 0
	for sourceName, pairToSymbolMap := range sourcesToPairSymbolMap {
		providers[i] = NewPriceProvider(sourceName, pairToSymbolMap, sourceConfigMap[sourceName], logger)
		i++
	}

	return AggregatePriceProvider{
		logger:    logger.With().Str("component", "aggregate-price-provider").Logger(),
		providers: providers,
	}
}

var aggregatePriceProvider = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: metrics.PrometheusNamespace,
	Name:      "aggregate_prices_total",
	Help:      "The total number prices provided by the aggregate price provider, by pair, source, and success status",
}, []string{"pair", "source", "success"})

// GetPrice fetches the first available and correct price from the wrapped PriceProviders.
// Iteration is exhaustive and random.
// If no correct PriceResponse is found, then an invalid PriceResponse is returned.
func (a AggregatePriceProvider) GetPrice(pair asset.Pair) types.Price {
	// SPECIAL CASE FOR stNIBI
	// fetch unibi:uusd first to calculate the ustnibi:unibi price
	if pair.String() == "ustnibi:uusd" {
		unibiUusdPrice := -1.0 // default to -1 to indicate we haven't found a valid price yet
		for _, p := range a.providers {
			price := p.GetPrice("unibi:uusd")
			if !price.Valid {
				continue
			}

			unibiUusdPrice = price.Price
			break
		}

		if unibiUusdPrice <= 0 {
			// if we can't find a valid unibi:uusd price, return an invalid price
			a.logger.Warn().Str("pair", "ustnibi:uusd").Msg("no valid price found for unibi:uusd")
			aggregatePriceProvider.WithLabelValues("ustnibi:uusd", "missing", "false").Inc()
			return types.Price{
				SourceName: "missing",
				Pair:       pair,
				Price:      0,
				Valid:      false,
			}
		}

		// now we can calculate the ustnibi:unibi price
		for _, p := range a.providers {
			price := p.GetPrice("ustnibi:unibi")
			if !price.Valid {
				continue
			}

			return types.Price{
				Pair:       pair,
				Price:      price.Price * unibiUusdPrice, // ustnibi:uusd = ustnibi:unibi * unibi:uusd
				SourceName: price.SourceName,             // use the source of ustnibi
				Valid:      true,
			}
		}
	}

	// for all other price pairs, iterate randomly, if we find a valid price, we return it
	// otherwise we go onto the next PriceProvider to ask for prices.
	for _, p := range a.providers {
		price := p.GetPrice(pair)
		if price.Valid {
			aggregatePriceProvider.WithLabelValues(pair.String(), price.SourceName, "true").Inc()
			return price
		}
	}

	// if we reach here no valid symbols were found
	a.logger.Warn().Str("pair", pair.String()).Msg("no valid price found")
	aggregatePriceProvider.WithLabelValues(pair.String(), "missing", "false").Inc()
	return types.Price{
		SourceName: "missing",
		Pair:       pair,
		Price:      0,
		Valid:      false,
	}
}

func (a AggregatePriceProvider) Close() {
	for _, p := range a.providers {
		p.Close()
	}
}
