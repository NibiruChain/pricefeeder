package config

import (
	"encoding/json"
	"github.com/NibiruChain/nibiru/x/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_JSON(t *testing.T) {
	wantConf := &Config{
		ExchangesToPairToSymbolMap: map[string]map[common.AssetPair]string{
			"BITFINEX": {
				common.Pair_BTC_NUSD: "tBTCNUSD",
				common.Pair_ETH_NUSD: "tETHNUSD",
			},
			"BINANCE": {
				common.Pair_BTC_NUSD: "btcnusd",
				common.Pair_ETH_NUSD: "ethnusd",
			},
		},
		GRPCEndpoint:        "somegrpcendpoint:440",
		TendermintEndpoint:  "sometendermintendpoint:900",
		FeederPrivateKeyHex: "somehexkey",
		ChainID:             "chainid",
	}

	b, err := json.Marshal(wantConf)
	require.NoError(t, err)

	gotConf := new(Config)
	err = json.Unmarshal(b, gotConf)
	require.NoError(t, err)

	require.Equal(t, wantConf, gotConf)
}
