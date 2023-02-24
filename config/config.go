package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/pricefeeder/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/joho/godotenv"
)

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
	exchangeSymbolsMapJson := os.Getenv("EXCHANGE_SYMBOLS_MAP")
	exchangeSymbolsMap := map[string]map[string]string{}
	err := json.Unmarshal([]byte(exchangeSymbolsMapJson), &exchangeSymbolsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EXCHANGE_SYMBOLS_MAP: invalid json")
	}

	conf.ExchangesToPairToSymbolMap = map[string]map[asset.Pair]types.Symbol{}
	for exchange, symbolMap := range exchangeSymbolsMap {
		conf.ExchangesToPairToSymbolMap[exchange] = map[asset.Pair]types.Symbol{}
		for nibiAssetPair, tickerSymbol := range symbolMap {
			conf.ExchangesToPairToSymbolMap[exchange][asset.MustNewPair(nibiAssetPair)] = types.Symbol(tickerSymbol)
		}
	}

	// datasource config map
	datasourceConfigMapJson := os.Getenv("DATASOURCE_CONFIG_MAP")
	datasourceConfigMap := map[string]json.RawMessage{}

	if datasourceConfigMapJson != "" {
		err = json.Unmarshal([]byte(datasourceConfigMapJson), &datasourceConfigMap)
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
