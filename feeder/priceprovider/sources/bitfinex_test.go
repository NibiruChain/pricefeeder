package sources

import (
	"testing"

	"github.com/NibiruChain/price-feeder/types"
	"github.com/stretchr/testify/require"
)

func TestBitfinexSource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rawPrices, err := BitfinexPriceUpdate([]types.Symbol{"tBTCUSD", "tETHUSD"})
		require.NoError(t, err)
		require.Equal(t, 2, len(rawPrices))
		require.NotZero(t, rawPrices["tBTCUSD"])
		require.NotZero(t, rawPrices["tETHUSD"])
	})

}
