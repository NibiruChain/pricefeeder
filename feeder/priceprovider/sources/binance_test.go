package sources

import (
	"testing"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/stretchr/testify/require"
)

func TestBinanceSource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rawPrices, err := BinancePriceUpdate(set.New[types.Symbol]("BTCUSD", "ETHUSD"))
		require.NoError(t, err)
		require.Equal(t, 2, len(rawPrices))
		require.NotZero(t, rawPrices["BTCUSD"])
		require.NotZero(t, rawPrices["ETHUSD"])
	})

}
