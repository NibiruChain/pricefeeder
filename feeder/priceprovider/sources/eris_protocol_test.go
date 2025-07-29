package sources

import (
	"io"
	"testing"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestErisProtocolPriceUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rawPrices, err := ErisProtocolPriceUpdate(set.New[types.Symbol](), zerolog.New(io.Discard))
		require.NoError(t, err)
		require.Equal(t, 1, len(rawPrices))
		require.NotZero(t, rawPrices["ustnibi:unibi"])
	})
}
