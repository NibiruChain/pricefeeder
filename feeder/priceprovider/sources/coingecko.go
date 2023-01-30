package sources

import (
	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/price-feeder/types"
)

type coingecko struct {
}

func (c coingecko) GetPrice(pair common.AssetPair) types.Price {
	//TODO implement me
	panic("implement me")
}

func (c coingecko) Close() {
	//TODO implement me
	panic("implement me")
}

func NewCoingecko() types.PriceProvider {
	return &coingecko{}
}
