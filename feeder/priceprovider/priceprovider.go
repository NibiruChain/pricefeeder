package priceprovider

import (
	"sync"
	"time"

	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/price-feeder/types"
	"github.com/rs/zerolog"
)

// NewPriceProvider returns a types.PriceProvider given the price source we want to gather prices from,
// the mapping between nibiru common.AssetPair and the source's symbols, and a zerolog.Logger instance.
func NewPriceProvider(exchangeName string, pairToSymbolMap map[common.AssetPair]string, logger zerolog.Logger) types.PriceProvider {
	var source Source
	switch exchangeName {
	case Bitfinex:
		source = NewTickSource(symbolsFromPairsToSymbolsMap(pairToSymbolMap), BitfinexPriceUpdate, logger)
	default:
		panic("unknown price provider: " + exchangeName)
	}
	return newPriceProvider(source, exchangeName, pairToSymbolMap, logger)
}

// newPriceProvider returns a raw *PriceProvider given a Source implementer, the source name and the
// map of nibiru common.AssetPair to Source's symbols, plus the zerolog.Logger instance.
// Exists for testing purposes.
func newPriceProvider(source Source, exchangeName string, pairToSymbolsMap map[common.AssetPair]string, logger zerolog.Logger) *PriceProvider {
	pp := &PriceProvider{
		logger:          logger.With().Str("component", "price-provider").Str("exchange", exchangeName).Logger(),
		stopSignal:      make(chan struct{}),
		done:            make(chan struct{}),
		source:          source,
		exchangeName:    exchangeName,
		pairToSymbol:    pairToSymbolsMap,
		lastPricesMutex: sync.Mutex{},
		lastPrices:      map[string]PriceUpdate{},
	}
	go pp.loop()
	return pp
}

// PriceProvider implements the types.PriceProvider interface.
// it wraps a Source and handles conversions between
// nibiru asset pair to exchange symbols.
type PriceProvider struct {
	logger          zerolog.Logger
	stopSignal      chan struct{}
	done            chan struct{}
	source          Source
	exchangeName    string
	pairToSymbol    map[common.AssetPair]string
	lastPricesMutex sync.Mutex
	lastPrices      map[string]PriceUpdate
}

// GetPrice returns the types.Price for the given common.AssetPair
// in case price has expired, or for some reason it's impossible to
// get the last available price, then an invalid types.Price is returned.
func (p *PriceProvider) GetPrice(pair common.AssetPair) types.Price {
	symbol, ok := p.pairToSymbol[pair]
	// in case this is an unknown symbol, which might happen
	// when for example we have a param update, then we return
	// an abstain vote on the provided asset pair.
	if !ok {
		p.logger.Warn().Str("pair", pair.String()).Msg("unknown nibiru pair")
		return types.Price{
			Pair:         pair,
			Price:        0, // TODO(heisenberg): return -1 instead for abstain vote
			ExchangeName: p.exchangeName,
			Valid:        false,
		}
	}
	p.lastPricesMutex.Lock()
	price, ok := p.lastPrices[symbol]
	p.lastPricesMutex.Unlock()
	return types.Price{
		Pair:         pair,
		Price:        price.Price,
		ExchangeName: p.exchangeName,
		Valid:        isValid(price, ok),
	}
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

func (p *PriceProvider) Close() {
	close(p.stopSignal)
	<-p.done
}

// symbolsFromPairsToSymbolsMap returns the symbols list
// given the map which maps nibiru chain pairs to exchange symbols.
func symbolsFromPairsToSymbolsMap(m map[common.AssetPair]string) []string {
	symbols := make([]string, 0, len(m))
	for _, v := range m {
		symbols = append(symbols, v)
	}
	return symbols
}

// isValid is a helper function which asserts if a price is valid given
// if it was found and the time at which it was last updated.
func isValid(price PriceUpdate, found bool) bool {
	return found && time.Since(price.UpdateTime) < PriceTimeout
}
