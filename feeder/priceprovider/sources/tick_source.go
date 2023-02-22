package sources

import (
	"time"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
)

var (
	// UpdateTick defines the wait time between price updates.
	UpdateTick = 8 * time.Second
)

var _ types.Source = (*TickSource)(nil)

// NewTickSource instantiates a new TickSource instance, given the symbols and a price updater function
// which returns the latest prices for the provided symbols.
func NewTickSource(symbols set.Set[types.Symbol], fetchPricesFunc types.FetchPricesFunc, logger zerolog.Logger) *TickSource {
	ts := &TickSource{
		logger:             logger,
		stopSignal:         make(chan struct{}),
		done:               make(chan struct{}),
		tick:               time.NewTicker(UpdateTick),
		symbols:            symbols,
		fetchPrices:        fetchPricesFunc,
		priceUpdateChannel: make(chan map[types.Symbol]types.RawPrice),
	}

	go ts.loop()

	return ts
}

// TickSource is a Source which updates prices
// every x time.Duration.
type TickSource struct {
	logger             zerolog.Logger
	stopSignal         chan struct{} // external signal to stop the loop
	done               chan struct{} // internal signal to wait for shutdown operations
	tick               *time.Ticker
	symbols            set.Set[types.Symbol] // symbols as named on the third party data source
	fetchPrices        func(symbols set.Set[types.Symbol]) (map[types.Symbol]float64, error)
	priceUpdateChannel chan map[types.Symbol]types.RawPrice
}

func (s *TickSource) loop() {
	defer s.tick.Stop()
	defer close(s.done)

	for {
		select {
		case <-s.stopSignal:
			return
		case <-s.tick.C:
			s.logger.Debug().Msg("received tick, updating prices")

			rawPrices, err := s.fetchPrices(s.symbols)
			if err != nil {
				s.logger.Err(err).Msg("failed to update prices")
				break // breaks the current select case, not the for cycle
			}

			priceUpdate := make(map[types.Symbol]types.RawPrice, len(rawPrices))
			for symbol, price := range rawPrices {
				priceUpdate[symbol] = types.RawPrice{
					Price:      price,
					UpdateTime: time.Now(),
				}
			}

			s.logger.Debug().Msg("sending price update")
			select {
			case s.priceUpdateChannel <- priceUpdate:
				s.logger.Debug().Msg("sent price update")
			case <-s.stopSignal:
				s.logger.Warn().Msg("dropped price update due to shutdown")
				return
			}
		}
	}
}

func (s *TickSource) PriceUpdates() <-chan map[types.Symbol]types.RawPrice {
	return s.priceUpdateChannel
}

func (s *TickSource) Close() {
	close(s.stopSignal)
	<-s.done
}
