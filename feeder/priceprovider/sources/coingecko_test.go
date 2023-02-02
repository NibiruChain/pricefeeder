package sources

import (
	"encoding/json"
	"testing"

	"github.com/NibiruChain/price-feeder/types"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestCoingeckoPriceUpdate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("success", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET", "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin%2Cethereum&vs_currencies=usd",
			httpmock.NewStringResponder(200, "{\"bitcoin\":{\"usd\":23829},\"ethereum\":{\"usd\":1676.85}}"),
		)
		rawPrices, err := CoingeckoPriceUpdate(json.RawMessage{})([]types.Symbol{
			"bitcoin",
			"ethereum",
		})
		require.NoError(t, err)

		require.Equal(t, 2, len(rawPrices))
		require.Equal(t, rawPrices["bitcoin"], 23829.0)
		require.Equal(t, rawPrices["ethereum"], 1676.85)
	})
}

func TestCoingeckoWithConfig(t *testing.T) {

}
