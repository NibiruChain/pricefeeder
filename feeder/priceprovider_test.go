package feeder

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/v2/x/common/asset"
	"github.com/NibiruChain/nibiru/v2/x/common/denoms"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/sources"
	"github.com/NibiruChain/pricefeeder/types"
)

// -------------------------------------------------
// PriceProvider
// -------------------------------------------------

var _ types.Source = (*testAsyncSource)(nil)

type testAsyncSource struct {
	closeFn       func()
	priceUpdatesC chan map[types.Symbol]types.RawPrice
}

func (t testAsyncSource) Close() { t.closeFn() }
func (t testAsyncSource) PriceUpdates() <-chan map[types.Symbol]types.RawPrice {
	return t.priceUpdatesC
}

func TestPriceProvider(t *testing.T) {
	// Speed up tests by using a much shorter tick duration
	// Lock for the entire test to serialize tests that modify UpdateTick
	sources.UpdateTickTestLock.Lock()
	defer sources.UpdateTickTestLock.Unlock()
	originalUpdateTick := sources.UpdateTick
	sources.UpdateTick = 10 * time.Millisecond
	t.Cleanup(func() {
		sources.UpdateTickTestLock.Lock()
		defer sources.UpdateTickTestLock.Unlock()
		sources.UpdateTick = originalUpdateTick
	})

	t.Run("bitfinex success", func(t *testing.T) {
		pp := NewPriceProvider(
			sources.SourceBitfinex,
			map[asset.Pair]types.Symbol{asset.Registry.Pair(denoms.BTC, denoms.NUSD): "tBTCUSD"},
			json.RawMessage{},
			zerolog.New(io.Discard),
		)
		defer pp.Close()
		// Wait for HTTP call + tick, but much shorter now
		<-time.After(sources.UpdateTick + 2*time.Second)

		price := pp.GetPrice(asset.Registry.Pair(denoms.BTC, denoms.NUSD))
		require.True(t, price.Valid)
		require.Equal(t, asset.Registry.Pair(denoms.BTC, denoms.NUSD), price.Pair)
		require.Equal(t, sources.SourceBitfinex, price.SourceName)
	})

	t.Run("eris protocol success", func(t *testing.T) {
		t.Setenv("GRPC_READ_ENDPOINT", "grpc.nibiru.fi:443")
		pp := NewPriceProvider(
			sources.SourceErisProtocol,
			map[asset.Pair]types.Symbol{asset.NewPair("ustnibi", denoms.NIBI): "ustnibi:unibi"},
			json.RawMessage{},
			zerolog.New(io.Discard),
		)
		defer pp.Close()
		// Wait for HTTP call + tick, but much shorter now
		<-time.After(sources.UpdateTick + 2*time.Second)

		price := pp.GetPrice(asset.NewPair("ustnibi", denoms.NIBI))
		require.True(t, price.Valid)
		require.Equal(t, asset.NewPair("ustnibi", denoms.NIBI), price.Pair)
		require.Equal(t, sources.SourceErisProtocol, price.SourceName)
	})

	t.Run("panics on unknown price source", func(t *testing.T) {
		require.Panics(t, func() {
			NewPriceProvider(
				"unknown",
				nil,
				nil,
				zerolog.New(io.Discard),
			)
		})
	})

	t.Run("returns invalid price on unknown AssetPair", func(t *testing.T) {
		pp := newPriceProvider(testAsyncSource{}, "test", map[asset.Pair]types.Symbol{}, zerolog.New(io.Discard))
		price := pp.GetPrice(asset.Registry.Pair(denoms.BTC, denoms.NUSD))
		require.False(t, price.Valid)
		require.Equal(t, float64(-1), price.Price)
		require.Equal(t, asset.Registry.Pair(denoms.BTC, denoms.NUSD), price.Pair)
	})

	t.Run("returns correct price", func(t *testing.T) {
		priceUpdatesC := make(chan map[types.Symbol]types.RawPrice)
		source := testAsyncSource{
			priceUpdatesC: priceUpdatesC,
			closeFn:       func() { close(priceUpdatesC) },
		}
		pp := newPriceProvider(source, "test", map[asset.Pair]types.Symbol{asset.Registry.Pair(denoms.BTC, denoms.NUSD): "BTC:NUSD"}, zerolog.New(io.Discard))

		priceUpdatesC <- map[types.Symbol]types.RawPrice{"BTC:NUSD": {Price: 10, UpdateTime: time.Now()}}
		price := pp.GetPrice(asset.Registry.Pair(denoms.BTC, denoms.NUSD))

		require.True(t, price.Valid)
		require.Equal(t, float64(10), price.Price)
		require.Equal(t, asset.Registry.Pair(denoms.BTC, denoms.NUSD), price.Pair)
		require.Equal(t, "test", price.SourceName)
	})

	t.Run("Close assertions", func(t *testing.T) {
		closed := false
		pp := newPriceProvider(testAsyncSource{
			closeFn: func() {
				closed = true
			},
		}, "test", map[asset.Pair]types.Symbol{}, zerolog.New(io.Discard))

		pp.Close()
		require.True(t, closed)
	})
}

func TestIsValid(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		require.True(t, isValid(types.RawPrice{
			Price:      10,
			UpdateTime: time.Now(),
		}, true))
	})

	t.Run("price not found", func(t *testing.T) {
		require.False(t, isValid(types.RawPrice{
			Price:      10,
			UpdateTime: time.Now(),
		}, false))
	})

	t.Run("price expired", func(t *testing.T) {
		require.False(t, isValid(types.RawPrice{
			Price:      20,
			UpdateTime: time.Now().Add(-1 - 1*types.PriceTimeout),
		}, true))
	})
}

// -------------------------------------------------
// AggregatePriceProvider
// -------------------------------------------------

func TestAggregatePriceProvider(t *testing.T) {
	t.Run("eris protocol success", func(t *testing.T) {
		t.Setenv("GRPC_READ_ENDPOINT", "grpc.nibiru.fi:443")
		pp := NewAggregatePriceProvider(
			map[string]map[asset.Pair]types.Symbol{
				sources.SourceErisProtocol: {
					asset.NewPair("ustnibi", denoms.NIBI): "ustnibi:unibi",
				},
				sources.SourceGateIo: {
					asset.NewPair("unibi", denoms.USD): "NIBI_USDT",
				},
			},
			map[string]json.RawMessage{},
			zerolog.New(io.Discard),
		)
		defer pp.Close()
		<-time.After(sources.UpdateTick + 5*time.Second)

		price := pp.GetPrice("ustnibi:uusd")
		require.True(t, price.Valid)
		require.Equal(t, asset.NewPair("ustnibi", denoms.USD), price.Pair)
		require.Equal(t, sources.SourceErisProtocol, price.SourceName)
	})

	t.Run(sources.SourceAvalon, func(t *testing.T) {
		pp := NewAggregatePriceProvider(
			map[string]map[asset.Pair]types.Symbol{
				sources.SourceAvalon: {
					"susda:usda": sources.Symbol_sUSDaUSDa,
				},
				sources.SourceUniswapV3: {
					"usda:usd": sources.Symbol_UniswapV3_USDaUSD,
				},
			},
			map[string]json.RawMessage{},
			zerolog.New(io.Discard),
		)
		defer pp.Close()
		<-time.After(sources.UpdateTick + 5*time.Second)

		pair := asset.Pair("susda:usda")
		price := pp.GetPrice(pair)
		assert.Truef(t, price.Valid, "invalid price for %s", price.Pair)
		assert.Equal(t, pair, price.Pair)
		assert.Equal(t, sources.SourceAvalon, price.SourceName)

		pair = asset.Pair("usda:usd")
		price = pp.GetPrice(pair)
		assert.Truef(t, price.Valid, "invalid price for %s", price.Pair)
		assert.EqualValues(t, pair, price.Pair)
		assert.Equal(t, sources.SourceUniswapV3, price.SourceName)

		pair = asset.Pair("susda:usd")
		price = pp.GetPrice(pair)
		assert.Truef(t, price.Valid, "invalid price for %s", price.Pair)
		assert.EqualValues(t, pair, price.Pair)
		assert.Equal(t, sources.SourceAvalon, price.SourceName)
	})
}
