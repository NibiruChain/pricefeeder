package sources

import (
	"io"
	"testing"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/types"
)

func TestBybitPriceUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rawPrices, err := BybitPriceUpdate(set.New[types.Symbol]("BTCUSDT", "ETHUSDT"), zerolog.New(io.Discard))
		if err != nil {
			require.ErrorContains(t, err, ErrBybitBlockAccess)
			return
		}
		require.NoError(t, err)
		require.Equal(t, 2, len(rawPrices))
		require.NotZero(t, rawPrices["BTCUSDT"])
		require.NotZero(t, rawPrices["ETHUSDT"])
	})
}
