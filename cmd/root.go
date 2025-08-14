package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/NibiruChain/nibiru/v2/gosdk"
	"github.com/NibiruChain/pricefeeder/config"
	"github.com/NibiruChain/pricefeeder/feeder"
	"github.com/NibiruChain/pricefeeder/feeder/eventstream"
	"github.com/NibiruChain/pricefeeder/feeder/priceposter"
	"github.com/NibiruChain/pricefeeder/feeder/priceprovider"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
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

var rootCmd = &cobra.Command{
	Use:   "pricefeeder",
	Short: "Pricefeeder daemon for posting prices to Nibiru Chain",
	Run: func(cmd *cobra.Command, args []string) {
		logger := setupLogger()
		gosdk.EnsureNibiruPrefix()

		c := config.MustGet()

		eventStream := eventstream.Dial(c.WebsocketEndpoint, c.GRPCEndpoint, c.EnableTLS, logger)
		priceProvider := priceprovider.NewAggregatePriceProvider(c.ExchangesToPairToSymbolMap, c.DataSourceConfigMap, logger)
		kb, valAddr, feederAddr := config.GetAuth(c.FeederMnemonic)

		if c.ValidatorAddr != nil {
			valAddr = *c.ValidatorAddr
		}
		pricePoster := priceposter.Dial(c.GRPCEndpoint, c.ChainID, c.EnableTLS, kb, valAddr, feederAddr, logger)

		f := feeder.NewFeeder(eventStream, priceProvider, pricePoster, logger)
		f.Run()
		defer f.Close()

		handleInterrupt(logger, f)

		metricsPort := os.Getenv("METRICS_PORT")
		if metricsPort == "" {
			metricsPort = "8080"
		}
		logger.Info().Msgf("Starting metrics server on port %s", metricsPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			logger.Error().Err(err).Msgf("Failed to start metrics server on port %s", metricsPort)
			os.Exit(1)
		}
		select {}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
