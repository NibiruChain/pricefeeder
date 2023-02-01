package sources

import (
	"testing"

	"github.com/NibiruChain/price-feeder/types"
	"github.com/stretchr/testify/require"
)

func TestBinanceSource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rawPrices, err := BinancePriceUpdate([]types.Symbol{"BTCUSD", "ETHUSD"})
		require.NoError(t, err)
		require.Equal(t, 2, len(rawPrices))
		require.NotZero(t, rawPrices["BTCUSD"])
		require.NotZero(t, rawPrices["ETHUSD"])
	})

}
