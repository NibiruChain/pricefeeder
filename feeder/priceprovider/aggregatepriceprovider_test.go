package priceprovider

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

	"github.com/NibiruChain/pricefeeder/feeder/priceprovider/sources"
	"github.com/NibiruChain/pricefeeder/types"
)

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
