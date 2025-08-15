package sources

import (
	"io"
	"testing"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/types"
)

func TestOKexPriceUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rawPrices, err := OkexPriceUpdate(set.New[types.Symbol]("BTC-USDT", "ETH-USDT"), zerolog.New(io.Discard))
		require.NoError(t, err)
		require.Equal(t, 2, len(rawPrices))
		require.NotZero(t, rawPrices["BTC-USDT"])
		require.NotZero(t, rawPrices["ETH-USDT"])
	})
}
