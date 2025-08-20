package sources

import (
	"math/big"
	"testing"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/types"
)

func TestChainlinkPriceUpdate(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))

	symbols := set.New[types.Symbol]()
	symbols.Add("uBTC/BTC")
	symbols.Add("foo:bar")

	prices, err := ChainlinkPriceUpdate(symbols, logger)

	require.NoError(t, err)
	require.Len(t, prices, 1)

	price := prices["uBTC/BTC"]
	assert.Greater(t, price, 0.0)

	_, unknownExists := prices["foo/bar"]
	assert.False(t, unknownExists)
}

func TestConvertChainlinkPrice(t *testing.T) {
	answer := big.NewInt(5000000000) // 50.00000000
	decimals := uint8(8)

	price, err := convertChainlinkPrice(answer, decimals)

	require.NoError(t, err)
	assert.Equal(t, 50.0, price)
}
