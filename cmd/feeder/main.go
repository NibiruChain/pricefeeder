package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/NibiruChain/nibiru/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/price-feeder/config"
	"github.com/NibiruChain/price-feeder/feeder"
	"github.com/NibiruChain/price-feeder/feeder/eventstream"
	"github.com/NibiruChain/price-feeder/feeder/priceposter"
	"github.com/NibiruChain/price-feeder/feeder/priceprovider"
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

	eventStream := eventstream.Dial(c.WebsocketEndpoint, c.GRPCEndpoint, logger)
	priceProvider := priceprovider.NewAggregatePriceProvider(c.ExchangesToPairToSymbolMap, c.ExchangesToConfigMap, logger)
	kb, valAddr, feederAddr := config.GetAuth(c.FeederMnemonic)

	if c.ValidatorAddress != "" {
		v, err := sdk.ValAddressFromBech32(c.ValidatorAddress)
		if err != nil {
			panic("invalid validator address")
		}
		valAddr = v
	}

	pricePoster := priceposter.Dial(c.GRPCEndpoint, c.ChainID, kb, valAddr, feederAddr, logger)

	f := feeder.NewFeeder(eventStream, priceProvider, pricePoster, logger)
	f.Run()
	defer f.Close()

	handleInterrupt(logger, f)

	select {}
}

// handleInterrupt listens for SIGINT and gracefully shuts down the feeder.
func handleInterrupt(logger zerolog.Logger, f *feeder.Feeder) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		logger.Info().Msg("shutting down gracefully")
		f.Close()
		os.Exit(1)
	}()
}
