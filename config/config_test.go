package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConfig_Get(t *testing.T) {

	os.Setenv("CHAIN_ID", "nibiru-localnet-0")
	os.Setenv("GRPC_ENDPOINT", "localhost:9090")
	os.Setenv("WEBSOCKET_ENDPOINT", "ws://localhost:26657/websocket")
	os.Setenv("FEEDER_MNEMONIC", "earth wash broom grow recall fitness")
	os.Setenv(
		"EXCHANGE_SYMBOLS_MAP",
		"{\"bitfinex\": {\"ubtc:unusd\": \"tBTCUSD\", \"ueth:unusd\": \"tETHUSD\", \"uusd:unusd\": \"tUSTUSD\"}}",
	)
	_, err := Get()
	require.NoError(t, err)
}
