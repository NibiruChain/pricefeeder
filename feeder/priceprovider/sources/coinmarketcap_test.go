package sources

import (
	"encoding/json"
	"testing"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestCoinmarketcapPriceUpdate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("success", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET", Link+"?slug=bitcoin%2Cethereum",
			httpmock.NewStringResponder(200, "{\"status\": {\"error_code\":0},\"data\":{\"1\":{\"slug\":\"bitcoin\",\"quote\":{\"USD\":{\"price\":23829}}}, \"100\":{\"slug\":\"ethereum\",\"quote\":{\"USD\":{\"price\":1676.85}}}}}"),
		)
		rawPrices, err := CoinmarketcapPriceUpdate(json.RawMessage{})(
			set.New[types.Symbol](
				"bitcoin",
				"ethereum",
			),
		)
		require.NoError(t, err)

		require.Equal(t, 2, len(rawPrices))
		require.Equal(t, rawPrices["bitcoin"], 23829.0)
		require.Equal(t, rawPrices["ethereum"], 1676.85)
	})
}
