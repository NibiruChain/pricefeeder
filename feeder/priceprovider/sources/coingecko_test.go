package sources

import (
	"testing"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/price-feeder/types"
	"github.com/stretchr/testify/require"
)

func TestCoingeckoPriceUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rawPrices, err := CoingeckoPriceUpdate(set.New[types.Symbol]("bitcoin", "ethereum"))
		require.NoError(t, err)
		require.Equal(t, 2, len(rawPrices))
		require.NotZero(t, rawPrices["bitcoin"])
		require.NotZero(t, rawPrices["ethereum"])
	})
}
