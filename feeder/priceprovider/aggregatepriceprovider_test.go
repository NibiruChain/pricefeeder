package priceprovider

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/pricefeeder/feeder/priceprovider/sources"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestAggregatePriceProvider(t *testing.T) {
	t.Run("eris protocol success", func(t *testing.T) {
		t.Setenv("GRPC_READ_ENDPOINT", "grpc.nibiru.fi:443")
		pp := NewAggregatePriceProvider(
			map[string]map[asset.Pair]types.Symbol{
				sources.ErisProtocol: {
					asset.NewPair("ustnibi", denoms.NIBI): "ustnibi:unibi",
				},
				sources.GateIo: {
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
		require.Equal(t, sources.ErisProtocol, price.SourceName)
	})
}
