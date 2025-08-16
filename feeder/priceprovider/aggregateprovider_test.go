package priceprovider

import (
	"testing"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// mockProvider implements PriceProvider for testing.
type mockProvider struct {
	prices map[asset.Pair]types.Price
}

func (m mockProvider) GetPrice(pair asset.Pair) types.Price {
	if price, ok := m.prices[pair]; ok {
		return price
	}
	return types.Price{Price: -1, Valid: false}
}
func (m mockProvider) Close() {}

// TestAggregateNoValidPrices ensures we return an invalid price if no providers return valid data.
func TestAggregateNoValidPrices(t *testing.T) {
	agg := AggregatePriceProvider{
		logger:    zerolog.Nop(),
		providers: map[int]types.PriceProvider{0: mockProvider{prices: map[asset.Pair]types.Price{}}},
	}
	price := agg.GetPrice(asset.MustNewPair("BTC:USD"))
	require.False(t, price.Valid)
}

// TestAggregateSinglePrice ensures that if exactly one provider returns a valid price, we get that price.
func TestAggregateSinglePrice(t *testing.T) {
	btcPair := asset.MustNewPair("BTC:USD")
	agg := AggregatePriceProvider{
		logger: zerolog.Nop(),
		providers: map[int]types.PriceProvider{
			0: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 1000.0, Valid: true},
			}},
			1: mockProvider{prices: map[asset.Pair]types.Price{}},
		},
	}
	price := agg.GetPrice(btcPair)
	require.True(t, price.Valid)
	require.Equal(t, 1000.0, price.Price)
}

// TestAggregateTwoPrices ensures we average two valid prices.
func TestAggregateTwoPrices(t *testing.T) {
	btcPair := asset.MustNewPair("BTC:USD")
	agg := AggregatePriceProvider{
		logger: zerolog.Nop(),
		providers: map[int]types.PriceProvider{
			0: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 1000.0, Valid: true},
			}},
			1: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 2000.0, Valid: true},
			}},
		},
	}
	price := agg.GetPrice(btcPair)
	require.True(t, price.Valid)
	require.Equal(t, 1500.0, price.Price)
}

// TestAggregateThreePrices checks median after removing outliers.
func TestAggregateThreePrices(t *testing.T) {
	btcPair := asset.MustNewPair("BTC:USD")
	agg := AggregatePriceProvider{
		logger: zerolog.Nop(),
		providers: map[int]types.PriceProvider{
			0: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 1000.0, Valid: true, SourceName: "mock1"},
			}},
			1: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 2000.0, Valid: true, SourceName: "mock2"},
			}},
			2: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 10000.0, Valid: true, SourceName: "mock3"},
			}},
		},
	}
	price := agg.GetPrice(btcPair)
	require.True(t, price.Valid)
	// Outlier (100000) removed, median of {1000, 2000} is 1500
	require.Equal(t, 1500.0, price.Price)

	agg = AggregatePriceProvider{
		logger: zerolog.Nop(),
		providers: map[int]types.PriceProvider{
			0: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 1000.0, Valid: true, SourceName: "mock1"},
			}},
			1: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 2000.0, Valid: true, SourceName: "mock2"},
			}},
			2: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 10000.0, Valid: true, SourceName: "mock3"},
			}},
			3: mockProvider{prices: map[asset.Pair]types.Price{
				btcPair: {Price: 2000.0, Valid: true, SourceName: "mock4"},
			}},
		},
	}

	price = agg.GetPrice(btcPair)
	require.True(t, price.Valid)
	// Outlier (100000) removed, median of {1000, 2000, 2000} is 2000
	require.Equal(t, 2000.0, price.Price)
}
