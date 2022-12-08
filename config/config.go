package config

import (
	"encoding/json"
	"fmt"
	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/price-feeder/feeder"
	"github.com/NibiruChain/price-feeder/feeder/events"
	"github.com/NibiruChain/price-feeder/feeder/priceprovider"
	"github.com/NibiruChain/price-feeder/feeder/tx"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"os"
)

func MustGet() *Config {
	conf, err := Get()
	if err != nil {
		panic(fmt.Sprintf("config error! check the environment: %v", err))
	}
	return conf
}
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

	conf.ExchangesToPairToSymbolMap = map[string]map[common.AssetPair]string{}
	for e, p := range exchangeSymbolsMap {
		conf.ExchangesToPairToSymbolMap[e] = map[common.AssetPair]string{}
		for assetPairStr, symbol := range p {
			assetPair, err := common.NewAssetPair(assetPairStr)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", err, assetPairStr)
			}
			conf.ExchangesToPairToSymbolMap[e][assetPair] = symbol
		}
	}
	return conf, conf.Validate()
}

type Config struct {
	ExchangesToPairToSymbolMap map[string]map[common.AssetPair]string
	GRPCEndpoint               string
	WebsocketEndpoint          string
	FeederMnemonic             string
	ChainID                    string
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

func (c *Config) Feeder() *feeder.Feeder {
	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	eventsStream := events.Dial(c.WebsocketEndpoint, c.GRPCEndpoint, log)
	priceProvider := priceprovider.NewAggregatePriceProvider(c.ExchangesToPairToSymbolMap, log)
	kb, valAddr, feederAddr := getAuth(c.FeederMnemonic)
	pricePoster := tx.Dial(c.GRPCEndpoint, c.ChainID, kb, valAddr, feederAddr, log)
	return feeder.Run(eventsStream, pricePoster, priceProvider, log)
}
