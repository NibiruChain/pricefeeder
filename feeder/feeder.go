package feeder

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/NibiruChain/price-feeder/types"
)

var (
	InitTimeout = 15 * time.Second
)

// Feeder is the price feeder.
type Feeder struct {
	logger zerolog.Logger

	stop chan struct{}
	done chan struct{}

	params types.Params

	eventStream   types.EventStream
	pricePoster   types.PricePoster
	priceProvider types.PriceProvider
}

func NewFeeder(eventStream types.EventStream, priceProvider types.PriceProvider, pricePoster types.PricePoster, logger zerolog.Logger) *Feeder {
	f := &Feeder{
		logger:        logger,
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
		params:        types.Params{},
		eventStream:   eventStream,
		pricePoster:   pricePoster,
		priceProvider: priceProvider,
	}

	return f
}

// Run instantiates a new Feeder instance.
func (f *Feeder) Run() {
	// init params
	select {
	case initParams := <-f.eventStream.ParamsUpdate():
		f.handleParamsUpdate(initParams)
	case <-time.After(InitTimeout):
		panic("init timeout deadline exceeded")
	}

	go f.loop()
}

func (f *Feeder) loop() {
	defer close(f.done)
	defer f.eventStream.Close()
	defer f.pricePoster.Close()
	defer f.priceProvider.Close()
	defer f.endLastVotingPeriod()
	for {
		select {
		case <-f.stop:
			return
		case params := <-f.eventStream.ParamsUpdate():
			f.logger.Info().Interface("changes", params).Msg("params changed")
			f.handleParamsUpdate(params)
		case vp := <-f.eventStream.VotingPeriodStarted():
			f.logger.Info().Interface("voting-period", vp).Msg("new voting period")
			f.handleVotingPeriod(vp)
		}
	}
}

func (f *Feeder) handleParamsUpdate(params types.Params) {
	f.params = params
}

func (f *Feeder) handleVotingPeriod(vp types.VotingPeriod) {
	f.startNewVotingPeriod(vp)
}

func (f *Feeder) startNewVotingPeriod(vp types.VotingPeriod) {
	// gather prices
	prices := make([]types.Price, len(f.params.Pairs))
	for i, p := range f.params.Pairs {
		price := f.priceProvider.GetPrice(p)
		if !price.Valid {
			f.logger.Err(fmt.Errorf("no valid price")).Str("asset", p.String()).Str("source", price.SourceName)
			price.Price = 0
		}
		prices[i] = price
	}

	// send prices
	f.pricePoster.SendPrices(vp, prices)
}

func (f *Feeder) Close() {
	close(f.stop)
	<-f.done
}
