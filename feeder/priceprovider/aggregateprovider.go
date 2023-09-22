package priceprovider

import (
	"encoding/json"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/pricefeeder/types"
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

// GetPrice fetches the first available and correct price from the wrapped PriceProviders.
// Iteration is exhaustive and random.
// If no correct PriceResponse is found, then an invalid PriceResponse is returned.
func (a AggregatePriceProvider) GetPrice(pair asset.Pair) types.Price {
	// Temporarily treat NUSD as perfectly pegged to the US fiat dollar
	// TODO(k-yang): add the NUSD pricefeed once it's available on exchanges
	if pair.Equal(asset.Registry.Pair(denoms.NUSD, denoms.USD)) {
		return types.Price{
			SourceName: "temporarily-hardcoded",
			Pair:       pair,
			Price:      1,
			Valid:      true,
		}
	}

	// Temporarily provide NIBI price based on last raise, 1.5 billion token supply @ $100M
	// TODO(k-yang): add the NUSD pricefeed once it's available on exchanges
	if pair.Equal(asset.Registry.Pair(denoms.NIBI, denoms.USD)) ||
		pair.Equal(asset.Registry.Pair(denoms.NIBI, denoms.NUSD)) {
		return types.Price{
			SourceName: "temporarily-hardcoded",
			Pair:       pair,
			Price:      0.06666666666666667,
			Valid:      true,
		}
	}

	// iterate randomly, if we find a valid price, we return it
	// otherwise we go onto the next PriceProvider to ask for prices.
	for _, p := range a.providers {
		price := p.GetPrice(pair)
		if price.Valid {
			return price
		}
	}

	// if we reach here no valid symbols were found
	a.logger.Warn().Str("pair", pair.String()).Msg("no valid price found")
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
