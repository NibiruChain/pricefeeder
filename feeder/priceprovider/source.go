package priceprovider

import (
	"time"

	"github.com/rs/zerolog"
)

var (
	// PriceTimeout defines after how much time a price is considered expired.
	PriceTimeout = 15 * time.Second
	// UpdateTick defines the wait time between price updates.
	UpdateTick = 3 * time.Second
)

const (
	Bitfinex = "bitfinex"
)

// Source defines a source for price provision.
// This source has no knowledge of nibiru internals
// and mappings across common.AssetPair and the Source
// symbols.
type Source interface {
	// PriceUpdates is a readonly channel which provides
	// the latest prices update. Updates can be provided
	// for one asset only or in batches, hence the map.
	PriceUpdates() <-chan map[string]PriceUpdate
	// Close closes the Source.
	Close()
}

// PriceUpdate defines an update for a symbol for Source implementers.
type PriceUpdate struct {
	Price      float64
	UpdateTime time.Time
}

// FetchPricesFunc is the function used to fetch updated prices.
// The symbols passed are the symbols we require prices for.
// The returned map must map symbol to its float64 price, or an error.
// If there's a failure in updating only one price then the map can be returned
// without the provided symbol.
type FetchPricesFunc func(symbols []string) (map[string]float64, error)

// NewTickSource instantiates a new TickSource instance, given the symbols and a price updater function
// which returns the latest prices for the provided symbols.
func NewTickSource(symbols []string, fetchPricesFunc FetchPricesFunc, logger zerolog.Logger) *TickSource {
	ts := &TickSource{
		logger:             logger,
		stopSignal:         make(chan struct{}),
		done:               make(chan struct{}),
		tick:               time.NewTicker(UpdateTick),
		symbols:            symbols,
		fetchPrices:        fetchPricesFunc,
		priceUpdateChannel: make(chan map[string]PriceUpdate),
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
	symbols            []string
	fetchPrices        func(symbols []string) (map[string]float64, error)
	priceUpdateChannel chan map[string]PriceUpdate
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

			prices, err := s.fetchPrices(s.symbols)
			if err != nil {
				s.logger.Err(err).Msg("failed to update prices")
				break // breaks the current select case, not the for cycle
			}

			priceUpdate := make(map[string]PriceUpdate, len(prices))
			for symbol, price := range prices {
				priceUpdate[symbol] = PriceUpdate{
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

func (s *TickSource) PriceUpdates() <-chan map[string]PriceUpdate {
	return s.priceUpdateChannel
}

func (s *TickSource) Close() {
	close(s.stopSignal)
	<-s.done
}
