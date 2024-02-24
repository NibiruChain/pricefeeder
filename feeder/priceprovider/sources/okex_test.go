package sources

import (
	"testing"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/stretchr/testify/require"
)

func TestOKexPriceUpdate(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		rawPrices, err := OkexPriceUpdate(set.New[types.Symbol]("BTCUSDT", "ETHUSDT"))
		require.NoError(t, err)
		require.Equal(t, 2, len(rawPrices))
		require.NotZero(t, rawPrices["BTCUSDT"])
		require.NotZero(t, rawPrices["ETHUSDT"])
	})
}
