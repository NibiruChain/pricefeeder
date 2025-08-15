package sources

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/jarcoal/httpmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/types"
)

func TestCoingeckoPriceUpdate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("success", func(t *testing.T) {
		httpmock.Reset()
		httpmock.RegisterResponder(
			"GET", FreeLink+"simple/price?ids=bitcoin%2Cethereum&vs_currencies=usd",
			httpmock.NewStringResponder(200, "{\"bitcoin\":{\"usd\":23829},\"ethereum\":{\"usd\":1676.85}}"),
		)
		rawPrices, err := CoingeckoPriceUpdate(json.RawMessage{})(
			set.New[types.Symbol](
				"bitcoin",
				"ethereum",
			),
			zerolog.New(io.Discard),
		)
		require.NoError(t, err)

		require.Equal(t, 2, len(rawPrices))
		require.Equal(t, rawPrices["bitcoin"], 23829.0)
		require.Equal(t, rawPrices["ethereum"], 1676.85)
	})
}

func TestCoingeckoWithConfig(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("providing valid config", func(t *testing.T) {
		httpmock.Reset()
		httpmock.RegisterResponder(
			"GET", PaidLink+"simple/price?ids=bitcoin%2Cethereum&vs_currencies=usd&"+ApiKeyParam+"=1234567890",
			httpmock.NewStringResponder(200, "{\"bitcoin\":{\"usd\":23829},\"ethereum\":{\"usd\":1676.85}}"),
		)

		// TODO(k-yang): set iteration is non-deterministic, so we need to account for both orderings of the coin ids
		httpmock.RegisterResponder(
			"GET", PaidLink+"simple/price?ids=ethereum%2Cbitcoin&vs_currencies=usd&"+ApiKeyParam+"=1234567890",
			httpmock.NewStringResponder(200, "{\"bitcoin\":{\"usd\":23829},\"ethereum\":{\"usd\":1676.85}}"),
		)

		options := "{\"api_key\": \"1234567890\"}"
		jsonOptions := json.RawMessage{}
		err := json.Unmarshal([]byte(options), &jsonOptions)
		require.NoError(t, err)

		rawPrices, err := CoingeckoPriceUpdate(jsonOptions)(
			set.New[types.Symbol](
				"bitcoin",
				"ethereum",
			),
			zerolog.New(io.Discard),
		)
		require.NoError(t, err)

		require.Equal(t, 2, len(rawPrices))
		require.Equal(t, rawPrices["bitcoin"], 23829.0)
		require.Equal(t, rawPrices["ethereum"], 1676.85)
	})

	t.Run("providing config without api_key ignores and calls free endpoint", func(t *testing.T) {
		httpmock.Reset()
		httpmock.RegisterResponder(
			"GET", FreeLink+"simple/price?ids=bitcoin%2Cethereum&vs_currencies=usd",
			httpmock.NewStringResponder(200, "{\"bitcoin\":{\"usd\":23829},\"ethereum\":{\"usd\":1676.85}}"),
		)

		options := "{\"not_api_key\": \"1234567890\"}"
		jsonOptions := json.RawMessage{}
		err := json.Unmarshal([]byte(options), &jsonOptions)
		require.NoError(t, err)

		rawPrices, err := CoingeckoPriceUpdate(jsonOptions)(
			set.New[types.Symbol](
				"bitcoin",
				"ethereum",
			),
			zerolog.New(io.Discard),
		)
		require.NoError(t, err)

		require.Equal(t, 2, len(rawPrices))
		require.Equal(t, rawPrices["bitcoin"], 23829.0)
		require.Equal(t, rawPrices["ethereum"], 1676.85)
	})
}
