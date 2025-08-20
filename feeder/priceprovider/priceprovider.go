package priceprovider

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/NibiruChain/nibiru/v2/x/common/asset"
	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/feeder/priceprovider/sources"
	"github.com/NibiruChain/pricefeeder/types"
)

var _ types.PriceProvider = (*PriceProvider)(nil)

// PriceProvider implements the types.PriceProvider interface.
// it wraps a Source and handles conversions between
// nibiru asset pair to exchange symbols.
type PriceProvider struct {
	logger              zerolog.Logger
	stopSignal          chan struct{}
	done                chan struct{}
	source              types.Source
	sourceName          string
	pairToSymbolMapping map[asset.Pair]types.Symbol
	lastPricesMutex     sync.Mutex
	lastPrices          map[types.Symbol]types.RawPrice
}

// NewPriceProvider returns a types.PriceProvider given the price source we want
// to gather prices from, the mapping between nibiru asset.Pair and the source's
// symbols, and a zerolog.Logger instance.
func NewPriceProvider(
	sourceName string,
	pairToSymbolMap map[asset.Pair]types.Symbol,
	config json.RawMessage,
	logger zerolog.Logger,
) types.PriceProvider {
	var source types.Source

	symbols := set.New[types.Symbol]()
	for _, s := range pairToSymbolMap {
		symbols.Add(s)
	}

	switch sourceName {
	case sources.Bitfinex:
		source = sources.NewTickSource(symbols, sources.BitfinexPriceUpdate, logger)
	case sources.Binance:
		source = sources.NewTickSource(symbols, sources.BinancePriceUpdate, logger)
	case sources.Coingecko:
		source = sources.NewTickSource(symbols, sources.CoingeckoPriceUpdate(config), logger)
	case sources.Okex:
		source = sources.NewTickSource(symbols, sources.OkexPriceUpdate, logger)
	case sources.GateIo:
		source = sources.NewTickSource(symbols, sources.GateIoPriceUpdate, logger)
	case sources.CoinMarketCap:
		source = sources.NewTickSource(symbols, sources.CoinmarketcapPriceUpdate(config), logger)
	case sources.Bybit:
		source = sources.NewTickSource(symbols, sources.BybitPriceUpdate, logger)
	case sources.ErisProtocol:
		source = sources.NewTickSource(symbols, sources.ErisProtocolPriceUpdate, logger)
	case sources.UniswapV3:
		source = sources.NewTickSource(symbols, sources.UniswapV3PriceUpdate, logger)
	default:
		panic("unknown price provider: " + sourceName)
	}

	return newPriceProvider(source, sourceName, pairToSymbolMap, logger)
}

// newPriceProvider returns a raw *PriceProvider given a Source implementer, the source name and the
// map of nibiru asset.Pair to Source's symbols, plus the zerolog.Logger instance.
// Exists for testing purposes.
func newPriceProvider(source types.Source, sourceName string, pairToSymbolsMap map[asset.Pair]types.Symbol, logger zerolog.Logger) *PriceProvider {
	pp := &PriceProvider{
		logger:              logger.With().Str("component", "price-provider").Str("source", sourceName).Logger(),
		stopSignal:          make(chan struct{}),
		done:                make(chan struct{}),
		source:              source,
		sourceName:          sourceName,
		pairToSymbolMapping: pairToSymbolsMap,
		lastPricesMutex:     sync.Mutex{},
		lastPrices:          map[types.Symbol]types.RawPrice{},
	}

	go pp.loop()

	return pp
}

func (p *PriceProvider) loop() {
	defer close(p.done)
	defer p.source.Close()

	for {
		select {
		case <-p.stopSignal:
			return
		case updates := <-p.source.PriceUpdates():
			p.lastPricesMutex.Lock()
			for symbol, price := range updates {
				p.lastPrices[symbol] = price
			}
			p.lastPricesMutex.Unlock()
		}
	}
}

// GetPrice returns the types.Price for the given asset.Pair
// in case price has expired, or for some reason it's impossible to
// get the last available price, then an invalid types.Price is returned.
func (p *PriceProvider) GetPrice(pair asset.Pair) types.Price {
	symbol, symbolExists := p.pairToSymbolMapping[pair]
	// in case this is an unknown symbol, which might happen
	// when for example we have a param update, then we return
	// an abstain vote on the provided asset pair.
	if !symbolExists {
		p.logger.Debug().Str("pair", pair.String()).Msg("pair not configured for this pricefeeder")
		return types.Price{
			Pair:       pair,
			Price:      -1, // abstain
			SourceName: p.sourceName,
			Valid:      false,
		}
	}

	p.lastPricesMutex.Lock()
	price, priceExists := p.lastPrices[symbol]
	p.lastPricesMutex.Unlock()

	return types.Price{
		Pair:       pair,
		Price:      price.Price,
		SourceName: p.sourceName,
		Valid:      isValid(price, priceExists),
	}
}

func (p *PriceProvider) Close() {
	close(p.stopSignal)
	<-p.done
}

// isValid is a helper function which asserts if a price is valid given
// if it was found and the time at which it was last updated.
func isValid(price types.RawPrice, found bool) bool {
	return found && time.Since(price.UpdateTime) < types.PriceTimeout
}
