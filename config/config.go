package config

import (
	"encoding/json"
	"fmt"
	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/price-feeder/feeder"
	"github.com/NibiruChain/price-feeder/feeder/events"
	"github.com/NibiruChain/price-feeder/feeder/priceprovider"
	"github.com/NibiruChain/price-feeder/feeder/tx"
	"github.com/ghodss/yaml"
	"github.com/rs/zerolog"
	"io"
	"os"
	"path/filepath"
)

const (
	DefaultConfigName   = "config.yaml"
	EnvCustomConfigPath = "PF_CUSTOM_CONFIG_PATH"
)

var DefaultConfigPath = func() string {
	hd, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(hd, DefaultConfigName)
}()

func MustGet() *Config {
	conf, err := Get()
	if err != nil {
		panic(err)
	}
	return conf
}
func Get() (*Config, error) {
	var path = DefaultConfigPath
	if customPath, ok := os.LookupEnv(EnvCustomConfigPath); ok {
		path = customPath
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	yamlBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return nil, err
	}

	conf := new(Config)
	err = json.Unmarshal(jsonBytes, conf)
	if err != nil {
		return nil, err
	}
	return conf, conf.Validate()
}

type Config struct {
	ExchangesToPairToSymbolMap map[string]map[common.AssetPair]string
	GRPCEndpoint               string
	TendermintEndpoint         string
	FeederPrivateKeyHex        string
	ChainID                    string
}

type config struct {
	ExchangesToPairToSymbolMap map[string]map[string]string `json:"exchanges_to_pair_to_symbol_map,omitempty"`
	GRPCEndpoint               string                       `json:"grpc_endpoint,omitempty"`
	TendermintEndpoint         string                       `json:"tendermint_endpoint,omitempty"`
	FeederPrivateKeyHex        string                       `json:"feeder_private_key_hex,omitempty"`
	ChainID                    string                       `json:"chain_id,omitempty"`
}

func (c *Config) UnmarshalJSON(b []byte) error {
	conf := new(config)
	err := json.Unmarshal(b, conf)
	if err != nil {
		return err
	}

	c.ChainID = conf.ChainID
	c.GRPCEndpoint = conf.GRPCEndpoint
	c.TendermintEndpoint = conf.TendermintEndpoint
	c.FeederPrivateKeyHex = conf.FeederPrivateKeyHex

	c.ExchangesToPairToSymbolMap = map[string]map[common.AssetPair]string{}
	for e, p := range conf.ExchangesToPairToSymbolMap {
		c.ExchangesToPairToSymbolMap[e] = map[common.AssetPair]string{}
		for assetPairStr, symbol := range p {
			assetPair, err := common.NewAssetPair(assetPairStr)
			if err != nil {
				return fmt.Errorf("%w: %s", err, assetPairStr)
			}
			c.ExchangesToPairToSymbolMap[e][assetPair] = symbol
		}
	}
	return nil
}

func (c *Config) MarshalJSON() ([]byte, error) {
	m := map[string]map[string]string{}

	for e, p := range c.ExchangesToPairToSymbolMap {
		psm := map[string]string{}
		for ap, s := range p {
			psm[ap.String()] = s
		}
		m[e] = psm
	}

	return json.Marshal(config{
		ExchangesToPairToSymbolMap: m,
		GRPCEndpoint:               c.GRPCEndpoint,
		TendermintEndpoint:         c.TendermintEndpoint,
		FeederPrivateKeyHex:        c.FeederPrivateKeyHex,
		ChainID:                    c.ChainID,
	})
}

func (c *Config) Validate() error {
	if c.ChainID == "" {
		return fmt.Errorf("no chain ID")
	}
	if c.FeederPrivateKeyHex == "" {
		return fmt.Errorf("no private key")
	}
	if c.TendermintEndpoint == "" {
		return fmt.Errorf("no tendermint endpoint")
	}
	if c.GRPCEndpoint == "" {
		return fmt.Errorf("no grpc endpoint")
	}
	return nil
}

func (c *Config) Feeder() *feeder.Feeder {
	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	eventsStream := events.Dial(c.TendermintEndpoint, c.GRPCEndpoint, log)
	priceProvider := priceprovider.NewAggregatePriceProvider(c.ExchangesToPairToSymbolMap, log)
	kb, valAddr, feederAddr := getAuth(c.FeederPrivateKeyHex)
	pricePoster := tx.Dial(c.GRPCEndpoint, c.ChainID, kb, valAddr, feederAddr, log)
	return feeder.Run(eventsStream, pricePoster, priceProvider, log)
}
