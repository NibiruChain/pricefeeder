package sources

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/types"
)

// SourceFactory is a function type that creates a [types.Source] instance for a given
// set of symbols, configuration, and logger. Use [SourceFactory] to register
// new price data sources with the application registry.
type SourceFactory func(
	symbols set.Set[types.Symbol],
	cfg json.RawMessage,
	logger zerolog.Logger,
) types.Source

var (
	// The [muSource] mutex protects [sourceRegistry] from concurrent access during
	// registration and retrieval operations. Use RLock for reads and Lock for writes.
	muSource sync.RWMutex

	// The [sourceRegistry] maps source names (e.g., "binance", "bitfinex") to their
	// factory functions. Sources are registered during package initialization
	// via the init function.
	sourceRegistry = map[string]SourceFactory{}
)

// NamedSource pairs a source name with its factory function for registration
// in the price feeder source registry. Use [NamedSource] with [Register] to make
// a source available to the application.
type NamedSource struct {
	Name string        // For example, "binance", "bitfinex"
	F    SourceFactory // The factory function that creates instances of this source
}

// Register adds a [NamedSource] to the price feeder application.
func Register(ns NamedSource) {
	muSource.Lock()
	defer muSource.Unlock()

	if len(ns.Name) == 0 || ns.F == nil {
		return
	}
	sourceRegistry[ns.Name] = ns.F
}

// GetRegisteredSource retrieves a registered source by name from the source
// registry and instantiates it with the provided symbols, configuration, and logger.
//   - [GetRegisteredSource] is safe for concurrent use.
//   - Returns an error if the source name is not registered or if the registered
//     factory function is nil.
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
