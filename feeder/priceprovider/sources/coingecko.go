package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	Coingecko   = "coingecko"
	FreeLink    = "https://api.coingecko.com/api/v3/"
	PaidLink    = "https://pro-api.coingecko.com/api/v3/"
	ApiKeyParam = "x_cg_pro_api_key"
)

type CoingeckoTicker struct {
	Price float64 `json:"usd"`
}

type CoingeckoConfig struct {
	ApiKey string `json:"api_key"`
}

func CoingeckoPriceUpdate(sourceConfig json.RawMessage) types.FetchPricesFunc {
	return func(symbols set.Set[types.Symbol]) (map[types.Symbol]float64, error) {
		c, err := extractConfig(sourceConfig)
		if err != nil {
			return nil, err
		}

		baseURL := buildURL(symbols, c)

		res, err := http.Get(baseURL)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		response, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		rawPrices, err := extractPricesFromResponse(symbols, response)
		if err != nil {
			return nil, err
		}

		return rawPrices, nil
	}
}

// extractConfig tries to get the configuration, if nothing is found, it returns an empty config.
func extractConfig(jsonConfig json.RawMessage) (*CoingeckoConfig, error) {
	c := &CoingeckoConfig{}
	if len(jsonConfig) > 0 {
		err := json.Unmarshal(jsonConfig, c)
		if err != nil {
			return nil, fmt.Errorf("invalid coingecko config: %w", err)
		}
	}
	return c, nil
}

func extractPricesFromResponse(symbols set.Set[types.Symbol], response []byte) (map[types.Symbol]float64, error) {
	var result map[string]CoingeckoTicker
	err := json.Unmarshal(response, &result)
	if err != nil {
		return nil, err
	}

	rawPrices := make(map[types.Symbol]float64)
	for symbol := range symbols {
		if price, ok := result[string(symbol)]; ok {
			rawPrices[symbol] = price.Price
		} else {
			return nil, fmt.Errorf("symbol %s not found in response: %s\n", symbol, response)
		}
	}

	return rawPrices, err
}

func buildURL(symbols set.Set[types.Symbol], c *CoingeckoConfig) string {
	link := FreeLink
	if c.ApiKey != "" {
		link = PaidLink
	}

	baseURL := fmt.Sprintf("%ssimple/price?", link)

	params := url.Values{}
	params.Add("ids", coingeckoSymbolCsv(symbols))
	params.Add("vs_currencies", "usd")
	if c.ApiKey != "" {
		params.Add(ApiKeyParam, c.ApiKey)
	}

	baseURL = baseURL + params.Encode()
	return baseURL
}

// coingeckoSymbolCsv returns the symbols as a comma separated string.
func coingeckoSymbolCsv(symbols set.Set[types.Symbol]) string {
	s := ""
	for symbol := range symbols {
		s += string(symbol) + ","
	}

	return s[:len(s)-1]
}
