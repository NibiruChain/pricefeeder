package sources

import (
	"io"
	"testing"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestGateIoSource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rawPrices, err := GateIoPriceUpdate(set.New[types.Symbol]("BTC_USDT", "ETH_USDT"), zerolog.New(io.Discard))
		require.NoError(t, err)
		require.Equal(t, 2, len(rawPrices))
		require.NotZero(t, rawPrices["BTC_USDT"])
		require.NotZero(t, rawPrices["ETH_USDT"])
	})
}
