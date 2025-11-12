package sources

import (
	"fmt"
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
	symbols.Add("ynETHx/ETH")
	symbols.Add("foo:bar")

	prices, err := ChainlinkPriceUpdate(symbols, logger)

	require.NoError(t, err)
	require.Len(t, prices, 2)

	price := prices["uBTC/BTC"]
	assert.Greater(t, price, 0.0)
	assert.Greater(t, prices["ynETHx/ETH"], 0.0)

	_, unknownExists := prices["foo/bar"]
	assert.False(t, unknownExists)
}

func TestConvertChainlinkPrice(t *testing.T) {
	for idx, tc := range []struct {
		answer   *big.Int
		decimals uint8
		want     float64
	}{
		{
			answer:   big.NewInt(5_000_000_000),
			decimals: 8,
			want:     50.0,
		},
		{
			answer: new(big.Int).Mul(
				big.NewInt(420_690),
				new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil),
			),
			decimals: 18,
			want:     420.69,
		},
	} {
		t.Run(fmt.Sprintf("tc %d - %+v", idx, tc), func(t *testing.T) {})

		gotPrice, err := convertChainlinkPrice(tc.answer, tc.decimals)
		assert.NoError(t, err)
		assert.Equal(t, tc.want, gotPrice)
	}
}
