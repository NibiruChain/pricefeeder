package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/NibiruChain/nibiru/v2/x/common/asset"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/joho/godotenv"

	"github.com/NibiruChain/pricefeeder/sources"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	defaultGrpcEndpoint      = "localhost:9090"
	defaultWebsocketEndpoint = "ws://localhost:26657/websocket"
)

var defaultExchangeSymbolsMap = map[string]map[asset.Pair]types.Symbol{
	// https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=100&page=1
	// k-yang: default disable Coingecko because they have aggressive rate limiting
	// sources.Coingecko: {
	// 	"ubtc:uusd":  "bitcoin",
	// 	"ueth:uusd":  "ethereum",
	// 	"uusdt:uusd": "tether",
	// 	"uusdc:uusd": "usd-coin",
	// 	"uatom:uusd": "cosmos",
	// },

	// https://api-pub.bitfinex.com/v2/conf/pub:list:pair:exchange
	sources.SourceNameBitfinex: {
		"ubtc:uusd":  "tBTCUSD",
		"ueth:uusd":  "tETHUSD",
		"uusdc:uusd": "tUDCUSD",
		"uusdt:uusd": "tUSTUSD",
		"uatom:uusd": "tATOUSD",
		"usol:uusd":  "tSOLUSD",
	},

	// https://api.gateio.ws/api/v4/spot/currency_pairs
	sources.SourceNameGateIo: {
		"ubtc:uusd":  "BTC_USDT",
		"ueth:uusd":  "ETH_USDT",
		"uusdc:uusd": "USDC_USDT",
		"uusdt:uusd": "USDT_USD",
		"uatom:uusd": "ATOM_USDT",
		"unibi:uusd": "NIBI_USDT",
		"usol:uusd":  "SOL_USDT",
	},

	// https://www.okx.com/api/v5/market/tickers?instType=SPOT
	sources.SourceNameOkex: {
		"ubtc:uusd":  "BTC-USDT",
		"ueth:uusd":  "ETH-USDT",
		"uusdc:uusd": "USDC-USDT",
		"uusdt:uusd": "USDT-USDC",
		"uatom:uusd": "ATOM-USDT",
		"usol:uusd":  "SOL-USD",
	},

	// https://api.bybit.com/v5/market/tickers?category=spot
	sources.SourceNameBybit: {
		"ubtc:uusd":  "BTCUSDT",
		"ueth:uusd":  "ETHUSDT",
		"uusdc:uusd": "USDCUSDT",
		"uatom:uusd": "ATOMUSDT",
		"unibi:uusd": "NIBIUSDT",
		"usol:uusd":  "SOLUSDT",
	},

	sources.SourceNameErisProtocol: {
		"ustnibi:unibi": "ustnibi:unibi", // this is the only pair supported by the Eris Protocol smart contract
	},

	sources.SourceNameUniswapV3: {
		"usda:usd": "USDa:USDT",
	},

	sources.SourceNameChainLink: {
		"b2btc:btc": "uBTC/BTC",
	},

	sources.SourceNameAvalon: {
		"susda:usda": "susda:usda",
	},
}

func MustGet() *Config {
	conf, err := Get()
	if err != nil {
		panic(fmt.Sprintf("config error! check the environment: %v", err))
	}

	if conf == nil {
		panic(errors.New("invalid config"))
	}

	return conf
}

// Get loads the configuration from the .env file and returns a Config struct or an error
// if the configuration is invalid.
func Get() (*Config, error) {
	_ = godotenv.Load() // .env is optional

	conf := new(Config)
	conf.ChainID = os.Getenv("CHAIN_ID")
	conf.GRPCEndpoint = os.Getenv("GRPC_ENDPOINT")
	conf.WebsocketEndpoint = os.Getenv("WEBSOCKET_ENDPOINT")
	conf.FeederMnemonic = os.Getenv("FEEDER_MNEMONIC")
	conf.EnableTLS = os.Getenv("ENABLE_TLS") == "true"
	conf.ExchangesToPairToSymbolMap = defaultExchangeSymbolsMap

	if conf.GRPCEndpoint == "" {
		conf.GRPCEndpoint = defaultGrpcEndpoint
	}
	if conf.WebsocketEndpoint == "" {
		conf.WebsocketEndpoint = defaultWebsocketEndpoint
	}

	overrideExchangeSymbolsMapJson := os.Getenv("EXCHANGE_SYMBOLS_MAP")
	if overrideExchangeSymbolsMapJson != "" {
		overrideExchangeSymbolsMap := map[string]map[string]string{}
		err := json.Unmarshal([]byte(overrideExchangeSymbolsMapJson), &overrideExchangeSymbolsMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse EXCHANGE_SYMBOLS_MAP: %w", err)
		}
		for exchange, symbolMap := range overrideExchangeSymbolsMap {
			conf.ExchangesToPairToSymbolMap[exchange] = map[asset.Pair]types.Symbol{}
			for nibiAssetPair, tickerSymbol := range symbolMap {
				conf.ExchangesToPairToSymbolMap[exchange][asset.MustNewPair(nibiAssetPair)] = types.Symbol(tickerSymbol)
			}
		}
	}

	// datasource config map
	datasourceConfigMapJson := os.Getenv("DATASOURCE_CONFIG_MAP")
	datasourceConfigMap := map[string]json.RawMessage{}

	if datasourceConfigMapJson != "" {
		err := json.Unmarshal([]byte(datasourceConfigMapJson), &datasourceConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DATASOURCE_CONFIG_MAP: invalid json")
		}
	}
	conf.DataSourceConfigMap = datasourceConfigMap

	// optional validator address (for delegated feeders)
	valAddrStr := os.Getenv("VALIDATOR_ADDRESS")
	if valAddrStr != "" {
		valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
		if err == nil {
			conf.ValidatorAddr = &valAddr
		}
	}

	return conf, conf.Validate()
}

type Config struct {
	ExchangesToPairToSymbolMap map[string]map[asset.Pair]types.Symbol
	DataSourceConfigMap        map[string]json.RawMessage
	GRPCEndpoint               string
	WebsocketEndpoint          string
	FeederMnemonic             string
	ChainID                    string
	ValidatorAddr              *sdk.ValAddress
	EnableTLS                  bool
}

func (c *Config) Validate() error {
	if c.ChainID == "" {
		return fmt.Errorf("no chain id")
	}
	if c.FeederMnemonic == "" {
		return fmt.Errorf("no feeder mnemonic")
	}
	if c.WebsocketEndpoint == "" {
		return fmt.Errorf("no websocket endpoint")
	}
	if c.GRPCEndpoint == "" {
		return fmt.Errorf("no grpc endpoint")
	}
	return nil
}
