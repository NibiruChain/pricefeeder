package priceprovider

import (
	"io"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/price-feeder/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var _ types.Source = (*testAsyncSource)(nil)

type testAsyncSource struct {
	closeFn       func()
	priceUpdatesC chan map[types.Symbol]types.Price
}

func (t testAsyncSource) Close() { t.closeFn() }
func (t testAsyncSource) PriceUpdates() <-chan map[types.Symbol]types.Price {
	return t.priceUpdatesC
}

func TestPriceProvider(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pp := NewPriceProvider(Bitfinex, map[common.AssetPair]types.Symbol{common.Pair_BTC_NUSD: "tBTCUSD"}, zerolog.New(io.Discard))
		defer pp.Close()
		<-time.After(UpdateTick + 2*time.Second)

		price := pp.GetPrice(common.Pair_BTC_NUSD)
		require.True(t, price.Valid)
	})

	t.Run("panics on unknown price source", func(t *testing.T) {
		require.Panics(t, func() {
			NewPriceProvider("unknown", nil, zerolog.New(io.Discard))
		})
	})

	t.Run("returns invalid price on unknown AssetPair", func(t *testing.T) {
		pp := newPriceProvider(testAsyncSource{}, "test", map[common.AssetPair]types.Symbol{}, zerolog.New(io.Discard))
		price := pp.GetPrice(common.Pair_BTC_NUSD)
		require.False(t, price.Valid)
		require.Zero(t, price.Price)
		require.Equal(t, common.Pair_BTC_NUSD, price.Pair)
	})

	t.Run("Close assertions", func(t *testing.T) {
		var closed bool
		pp := newPriceProvider(testAsyncSource{
			closeFn: func() {
				closed = true
			},
		}, "test", map[common.AssetPair]types.Symbol{}, zerolog.New(io.Discard))

		pp.Close()
		require.True(t, closed)
	})
}

func Test_isValid(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		require.True(t, isValid(types.Price{
			Price:      10,
			UpdateTime: time.Now(),
		}, true))
	})

	t.Run("price not found", func(t *testing.T) {
		require.False(t, isValid(types.Price{
			Price:      10,
			UpdateTime: time.Now(),
		}, false))
	})

	t.Run("price expired", func(t *testing.T) {
		require.False(t, isValid(types.Price{
			Price:      20,
			UpdateTime: time.Now().Add(-1 - 1*types.PriceTimeout),
		}, true))
	})
}
