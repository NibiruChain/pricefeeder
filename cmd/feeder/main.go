package main

import (
	"flag"
	"os"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/price-feeder/config"
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
	conf := config.MustGet()
	f := conf.Feeder(logger)

	defer f.Close()
	select {}
}
