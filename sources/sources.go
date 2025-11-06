package sources

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/types"
)

type SourceFactory func(
	symbols set.Set[types.Symbol],
	cfg json.RawMessage,
	logger zerolog.Logger,
) types.Source

var (
	muSource       sync.RWMutex
	sourceRegistry = map[string]SourceFactory{}
)

type NamedSource struct {
	Name string
	F    SourceFactory
}

// Register adds a [NamedSource] to the price feeder application.
func Register(ns NamedSource) {
	muSource.Lock()
	sourceRegistry[ns.Name] = ns.F
	muSource.Unlock()
}

func GetRegisteredSource(
	name string,
	symbols set.Set[types.Symbol],
	cfg json.RawMessage,
	logger zerolog.Logger,
) (types.Source, error) {
	muSource.RLock()
	sourceFactory, ok := sourceRegistry[name]
	muSource.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown data provider source name: %s", name)
	} else if sourceFactory == nil {
		return nil, fmt.Errorf("source name %s registered without a SourceFactory", name)
	}
	return sourceFactory(symbols, cfg, logger), nil
}

func init() {
	for _, namedSource := range allSources {
		Register(namedSource)
	}
}

var allSources = []NamedSource{
	{
		Name: SourceNameBinance,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, BinancePriceUpdate, logger)
		},
	},
	{
		Name: SourceNameCoingecko,
		F: func(
			symbols set.Set[types.Symbol],
			cfg json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, CoingeckoPriceUpdate(cfg), logger)
		},
	},
	{
		Name: SourceNameBitfinex,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, BitfinexPriceUpdate, logger)
		},
	},
	{
		Name: SourceNameOkex,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, OkexPriceUpdate, logger)
		},
	},
	{
		Name: SourceNameGateIo,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, GateIoPriceUpdate, logger)
		},
	},
	{
		Name: SourceNameCoinMarketCap,
		F: func(
			symbols set.Set[types.Symbol],
			cfg json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, CoinmarketcapPriceUpdate(cfg), logger)
		},
	},
	{
		Name: SourceNameBybit,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, BybitPriceUpdate, logger)
		},
	},
	{
		Name: SourceNameErisProtocol,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, ErisProtocolPriceUpdate, logger)
		},
	},
	{
		Name: SourceNameUniswapV3,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, UniswapV3PriceUpdate, logger)
		},
	},
	{
		Name: SourceNameAvalon,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, AvalonPriceUpdate, logger)
		},
	},
	{
		Name: SourceNameChainLink,
		F: func(
			symbols set.Set[types.Symbol],
			_ json.RawMessage,
			logger zerolog.Logger,
		) types.Source {
			return NewTickSource(symbols, ChainlinkPriceUpdate, logger)
		},
	},
}
