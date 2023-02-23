package config

import (
	"os"
	"testing"

	"github.com/NibiruChain/nibiru/app"
	"github.com/stretchr/testify/require"
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
	app.SetPrefixes(app.AccountAddressPrefix)
	os.Setenv("VALIDATOR_ADDRESS", "nibivaloper1d7zygazerfwx4l362tnpcp0ramzm97xvv9ryxr")
	_, err := Get()
	require.NoError(t, err)
}
