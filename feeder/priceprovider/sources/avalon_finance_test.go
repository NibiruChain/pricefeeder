package sources

import (
	"io"
	"testing"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/types"
)

func TestAvalonPriceUpdate(t *testing.T) {
	rawPrices, err := AvalonPriceUpdate(set.New[types.Symbol](), zerolog.New(io.Discard))
	require.NoError(t, err)
	require.Equal(t, 1, len(rawPrices))
	require.GreaterOrEqual(t, rawPrices[Symbol_sUSDaUSDa], 0.5)
}
