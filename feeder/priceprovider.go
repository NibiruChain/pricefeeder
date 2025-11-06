package feeder

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/NibiruChain/nibiru/v2/x/common/asset"
	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/sources"
	"github.com/NibiruChain/pricefeeder/types"
)

var (
	_ types.PriceProvider = (*PriceProvider)(nil)
	_ types.PriceProvider = (*AggregatePriceProvider)(nil)
)

// -------------------------------------------------
// PriceProvider
// -------------------------------------------------

// PriceProvider implements the types.PriceProvider interface.
// it wraps a Source and handles conversions between
// nibiru asset pair to exchange symbols.
type PriceProvider struct {
	logger              zerolog.Logger
	stopSignal          chan struct{}
	done                chan struct{}
	source              types.Source
	sourceName          string
	pairToSymbolMapping map[asset.Pair]types.Symbol
	lastPricesMutex     sync.Mutex
	lastPrices          map[types.Symbol]types.RawPrice
}

// NewPriceProvider returns a types.PriceProvider given the price source we want
// to gather prices from, the mapping between nibiru asset.Pair and the source's
// NewPriceProvider constructs a PriceProvider for the named source using the provided
// pair-to-symbol mapping, configuration, and logger. It builds the set of symbols from
// the mapping and attempts to obtain a registered source; on error it logs a warning
// and returns types.NullPriceProvider, otherwise it wraps the source with newPriceProvider.
func NewPriceProvider(
	sourceName string,
	pairToSymbolMap map[asset.Pair]types.Symbol,
	config json.RawMessage,
	logger zerolog.Logger,
) types.PriceProvider {
	var source types.Source

	symbols := set.New[types.Symbol]()
	for _, s := range pairToSymbolMap {
		symbols.Add(s)
	}

	source, err := sources.GetRegisteredSource(sourceName, symbols, config, logger)
	if err != nil {
		logger.
			Warn().
			Msg(err.Error())
		return types.NullPriceProvider{}
	}

	return newPriceProvider(source, sourceName, pairToSymbolMap, logger)
}

// newPriceProvider returns a raw *PriceProvider given a Source implementer, the source name and the
// map of nibiru asset.Pair to Source's symbols, plus the zerolog.Logger instance.
// Exists for testing purposes.
func newPriceProvider(source types.Source, sourceName string, pairToSymbolsMap map[asset.Pair]types.Symbol, logger zerolog.Logger) *PriceProvider {
	pp := &PriceProvider{
		logger:              logger.With().Str("component", "price-provider").Str("source", sourceName).Logger(),
		stopSignal:          make(chan struct{}),
		done:                make(chan struct{}),
		source:              source,
		sourceName:          sourceName,
		pairToSymbolMapping: pairToSymbolsMap,
		lastPricesMutex:     sync.Mutex{},
		lastPrices:          map[types.Symbol]types.RawPrice{},
	}

	go pp.loop()

	return pp
}

func (p *PriceProvider) loop() {
	defer close(p.done)
	defer p.source.Close()

	for {
		select {
		case <-p.stopSignal:
			return
		case updates := <-p.source.PriceUpdates():
			p.lastPricesMutex.Lock()
			for symbol, price := range updates {
				p.lastPrices[symbol] = price
			}
			p.lastPricesMutex.Unlock()
		}
	}
}

// GetPrice returns the types.Price for the given asset.Pair
// in case price has expired, or for some reason it's impossible to
// get the last available price, then an invalid types.Price is returned.
func (p *PriceProvider) GetPrice(pair asset.Pair) types.Price {
	symbol, symbolExists := p.pairToSymbolMapping[pair]
	// in case this is an unknown symbol, which might happen
	// when for example we have a param update, then we return
	// an abstain vote on the provided asset pair.
	if !symbolExists {
		p.logger.Debug().Str("pair", pair.String()).Msg("pair not configured for this pricefeeder")
		return types.Price{
			Pair:       pair,
			Price:      types.PriceAbstain,
			SourceName: p.sourceName,
			Valid:      false,
		}
	}

	p.lastPricesMutex.Lock()
	price, priceExists := p.lastPrices[symbol]
	p.lastPricesMutex.Unlock()

	return types.Price{
		Pair:       pair,
		Price:      price.Price,
		SourceName: p.sourceName,
		Valid:      isValid(price, priceExists),
	}
}

func (p *PriceProvider) Close() {
	close(p.stopSignal)
	<-p.done
}

// isValid is a helper function which asserts if a price is valid given
// if it was found and the time at which it was last updated.
func isValid(price types.RawPrice, found bool) bool {
	return found && time.Since(price.UpdateTime) < types.PriceTimeout
}

// -------------------------------------------------
// AggregatePriceProvider
// -------------------------------------------------

// AggregatePriceProvider aggregates multiple price providers
// and queries them for prices.
type AggregatePriceProvider struct {
	// providers: A set of providers implemented using a map to zero size
	// empty struct. Using a map gives us random order of iteration, the
	// intended behavior (since golang's map range is unordered)
	providers map[types.PriceProvider]struct{}
	logger    zerolog.Logger
}

// NewAggregatePriceProvider instantiates a new AggregatePriceProvider instance
// NewAggregatePriceProvider creates an AggregatePriceProvider that aggregates prices from multiple configured sources.
// It instantiates a price provider for each entry in sourcesToPairSymbolMap using the corresponding config from sourceConfigMap,
// excludes sources that fail to initialize, logs a warning when some configured sources are invalid and an error when none are available,
// and returns an AggregatePriceProvider containing the successfully created providers with a scoped logger.
func NewAggregatePriceProvider(
	sourcesToPairSymbolMap map[string]map[asset.Pair]types.Symbol,
	sourceConfigMap map[string]json.RawMessage,
	logger zerolog.Logger,
) types.PriceProvider {
	providers := make(map[types.PriceProvider]struct{})
	invalidSources := []string{}
	for sourceName, pairToSymbolMap := range sourcesToPairSymbolMap {
		pp := NewPriceProvider(sourceName, pairToSymbolMap, sourceConfigMap[sourceName], logger)
		if _, isNull := pp.(types.NullPriceProvider); isNull {
			invalidSources = append(invalidSources, sourceName)
			continue
		}
		providers[pp] = struct{}{}
	}

	if len(providers) != len(sourcesToPairSymbolMap) {
		logger.Warn().
			Msg(fmt.Sprintf("invalid source names given as key in configuration: { invalidSources: %#v }", invalidSources))
	}
	if len(providers) == 0 {
		logger.Error().
			Msg(fmt.Sprintf("no price providers available: { invalidSources: %#v }", invalidSources))
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
	switch pairStr := pair.String(); pairStr {
	// SPECIAL CASE - stNIBI
	case "ustnibi:uusd":
		// fetch unibi:uusd first to calculate the ustnibi:unibi price

		unibiUusdPrice := types.PriceAbstain // default to -1 to indicate we haven't found a valid price yet
		for p := range a.providers {
			price := p.GetPrice("unibi:uusd")
			if !price.Valid {
				continue
			}

			unibiUusdPrice = price.Price
			break
		}

		if unibiUusdPrice <= 0 {
			// if we can't find a valid unibi:uusd price, return an invalid price
			a.logger.Warn().Str("pair", pair.String()).Msg("no valid price found")
			aggregatePriceProvider.WithLabelValues("ustnibi:uusd", "missing", "false").Inc()
			return types.Price{
				SourceName: "missing",
				Pair:       pair,
				Price:      types.PriceAbstain,
				Valid:      false,
			}
		}

		// now we can calculate the ustnibi:unibi price
		for p := range a.providers {
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

	case "susda:usd":
		priceSusdaUsda := a.GetPrice("susda:usda")
		if !priceSusdaUsda.Valid || priceSusdaUsda.Price <= 0 {
			// if we can't find a valid susda:usda price, return an invalid price
			a.logger.Warn().Str("pair", pairStr).Msg("no valid price found")
			aggregatePriceProvider.WithLabelValues(pairStr, "missing", "false").Inc()
			return types.Price{
				SourceName: "missing",
				Pair:       pair,
				Price:      types.PriceAbstain,
				Valid:      false,
			}
		}

		priceUsdaUsd := a.GetPrice("usda:usd")
		if priceUsdaUsd.Valid && priceUsdaUsd.Price > 0 {
			return types.Price{
				Pair:       pair,
				Price:      priceSusdaUsda.Price * priceUsdaUsd.Price,
				SourceName: priceSusdaUsda.SourceName,
				Valid:      true,
			}
		}

		// if either price is invalid, return an invalid price
		a.logger.Warn().Str("pair", pairStr).Msg("no valid price found")
		aggregatePriceProvider.WithLabelValues(pairStr, "missing", "false").Inc()
		return types.Price{
			SourceName: "missing",
			Pair:       pair,
			Price:      types.PriceAbstain,
			Valid:      false,
		}

	default:
		// for all other price pairs, iterate randomly, if we find a valid price, we return it
		// otherwise we go onto the next PriceProvider to ask for prices.
		for p := range a.providers {
			price := p.GetPrice(pair)
			if price.Valid {
				aggregatePriceProvider.WithLabelValues(pair.String(), price.SourceName, "true").Inc()
				return price
			}
		}
	}

	// if we reach here no valid symbols were found
	a.logger.Warn().Str("pair", pair.String()).Msg("no valid price found")
	aggregatePriceProvider.WithLabelValues(pair.String(), "missing", "false").Inc()
	return types.Price{
		SourceName: "missing",
		Pair:       pair,
		Price:      types.PriceAbstain,
		Valid:      false,
	}
}

func (a AggregatePriceProvider) Close() {
	for p := range a.providers {
		p.Close()
	}
}