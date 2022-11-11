package priceprovider

import (
	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/price-feeder/feeder/types"
	"github.com/rs/zerolog"
)

func NewAggregatePriceProvider(exchangesToPairToSymbolMap map[string]map[common.AssetPair]string, log zerolog.Logger) types.PriceProvider {
	log = log.With().Str("component", "aggregate-price-provider").Logger()
	pps := make([]types.PriceProvider, 0, len(exchangesToPairToSymbolMap))
	for exchangeName, parToSymbolMap := range exchangesToPairToSymbolMap {
		pps = append(pps, NewPriceProvider(exchangeName, parToSymbolMap, log.With().Str("exchange", exchangeName).Logger()))
	}
	return newAggregatePriceProvider(pps, log)
}

// NewAggregatePriceProvider instantiates a new AggregatePriceProvider instance
// given multiple PriceProvider.
func newAggregatePriceProvider(pps []types.PriceProvider, log zerolog.Logger) types.PriceProvider {
	a := AggregatePriceProvider{log, make(map[int]types.PriceProvider, len(pps))}
	for i, pp := range pps {
		a.pps[i] = pp
	}
	return a
}

// AggregatePriceProvider aggregates multiple price providers
// and queries them for prices.
type AggregatePriceProvider struct {
	log zerolog.Logger
	pps map[int]types.PriceProvider // we use a map here to provide random ranging (since golang's map range is unordered)
}

// GetPrice fetches the first available and correct price from the wrapped PriceProviders.
// Iteration is exhaustive and random.
// If no correct PriceResponse is found, then an invalid PriceResponse is returned.
func (a AggregatePriceProvider) GetPrice(pair common.AssetPair) types.Price {
	// iterate randomly, if we find a valid price, we return it
	// otherwise we go onto the next PriceProvider to ask for prices.
	for _, pp := range a.pps {
		price := pp.GetPrice(pair)
		if price.Valid {
			return price
		}
	}

	// if we reach here no valid symbols were found
	a.log.Warn().Str("pair", pair.String()).Msg("no valid prices")
	return types.Price{
		Pair:   pair,
		Price:  0,
		Source: "aggregate",
		Valid:  false,
	}
}

func (a AggregatePriceProvider) Close() {
	for _, pp := range a.pps {
		pp.Close()
	}
}
