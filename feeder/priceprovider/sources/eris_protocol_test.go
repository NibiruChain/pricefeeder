package sources

import (
	"io"
	"testing"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/types"
)

func TestErisProtocolPriceUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Setenv("GRPC_READ_ENDPOINT", "grpc.nibiru.fi:443")
		rawPrices, err := ErisProtocolPriceUpdate(set.New[types.Symbol](), zerolog.New(io.Discard))
		require.NoError(t, err)
		require.Equal(t, 1, len(rawPrices))
		require.NotZero(t, rawPrices["ustnibi:unibi"])
	})
}
