package main

import (
	"errors"
	"flag"
	"os"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/price-feeder/config"
	"github.com/NibiruChain/price-feeder/feeder"
	"github.com/NibiruChain/price-feeder/feeder/eventstream"
	"github.com/NibiruChain/price-feeder/feeder/priceposter"
	"github.com/NibiruChain/price-feeder/feeder/priceprovider"
	"github.com/rs/zerolog"
)

func setupLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()
	// Default level is INFO, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return zerolog.New(os.Stderr).With().Timestamp().Logger()
}

func main() {
	logger := setupLogger()
	app.SetPrefixes(app.AccountAddressPrefix)

	c := config.MustGet()
	if c == nil {
		panic(errors.New("invalid config"))
	}

	eventStream := eventstream.Dial(c.WebsocketEndpoint, c.GRPCEndpoint, logger)
	priceProvider := priceprovider.NewAggregatePriceProvider(c.ExchangesToPairToSymbolMap, logger)
	kb, valAddr, feederAddr := config.GetAuth(c.FeederMnemonic)
	pricePoster := priceposter.Dial(c.GRPCEndpoint, c.ChainID, kb, valAddr, feederAddr, logger)

	f := feeder.NewFeeder(eventStream, priceProvider, pricePoster, logger)
	f.Run()
	defer f.Close()

	select {}
}
