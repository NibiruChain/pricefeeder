package priceprovider

import (
	"encoding/json"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
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

		// calculate the ustnibi:unibi price
		var ustnibiUnibiPrice float64 = 0 // default to 0 to indicate we haven't found a valid price yet
		for _, p := range a.providers {
			price := p.GetPrice("ustnibi:unibi")
			if !price.Valid {
				continue
			}

			ustnibiUnibiPrice = price.Price
			break
		}

		// calculate the unibi:uusd price
		var unibiUusdPrice float64 = 0 // default to 0 to indicate we haven't found a valid price yet
		for _, p := range a.providers {
			price := p.GetPrice("unibi:uusd")
			if !price.Valid {
				continue
			}

			unibiUusdPrice = price.Price
			break
		}

		if ustnibiUnibiPrice <= 0 || unibiUusdPrice <= 0 {
			a.logger.Warn().Float64("ustnibiUnibiPrice", ustnibiUnibiPrice).Float64("unibiUusdPrice", unibiUusdPrice).Msg("ustnibiUnibiPrice and unibiUusdPrice")
			// if we can't find a valid unibi:uusd price, return an invalid price
			a.logger.Warn().Str("pair", "ustnibi:uusd").Msg("no valid price found for ustnibi:unibi or unibi:uusd")
			aggregatePriceProvider.WithLabelValues("ustnibi:uusd", "missing", "false").Inc()
			return types.Price{
				SourceName: "missing",
				Pair:       pair,
				Price:      0,
				Valid:      false,
			}
		}

		return types.Price{
			SourceName: "ustnibi:unibi",
			Pair:       pair,
			Price:      ustnibiUnibiPrice * unibiUusdPrice,
			Valid:      true,
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
