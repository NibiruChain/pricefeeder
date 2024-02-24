// Credit: @oleksandrmarkelov https://github.com/NibiruChain/pricefeeder/pull/27
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
	CoinMarketCap     = "coinmarketcap"
	link              = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest"
	apiKeyHeaderParam = "X-CMC_PRO_API_KEY"
)

type CmcQuotePrice struct {
	Price float64
}

type CmcQuote struct {
	USD CmcQuotePrice
}

type CmcTicker struct {
	Slug  string
	Quote CmcQuote
}

type CmcResponse struct {
	Data map[string]CmcTicker
}

type CoinmarketcapConfig struct {
	ApiKey string `json:"api_key"`
}

func CoinmarketcapPriceUpdate(rawConfig json.RawMessage) types.FetchPricesFunc {
	return func(symbols set.Set[types.Symbol]) (map[types.Symbol]float64, error) {
		config, err := getConfig(rawConfig)
		if err != nil {
			return nil, err
		}

		req, err := buildReq(symbols, config)
		if err != nil {
			return nil, err
		}

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		response, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		rawPrices, err := getPricesFromResponse(symbols, response)
		if err != nil {
			return nil, err
		}

		return rawPrices, nil
	}
}

// extractConfig tries to get the configuration, if nothing is found, it returns an empty config.
func getConfig(jsonConfig json.RawMessage) (*CoinmarketcapConfig, error) {
	c := &CoinmarketcapConfig{}
	if len(jsonConfig) > 0 {
		err := json.Unmarshal(jsonConfig, c)
		if err != nil {
			return nil, fmt.Errorf("invalid coinmarketcap config: %w", err)
		}
	}
	return c, nil
}

func getPricesFromResponse(symbols set.Set[types.Symbol], response []byte) (map[types.Symbol]float64, error) {
	var respCmc CmcResponse
	err := json.Unmarshal(response, &respCmc)
	if err != nil {
		return nil, err
	}

	cmcPrice := make(map[string]float64)
	for _, value := range respCmc.Data {
		cmcPrice[value.Slug] = value.Quote.USD.Price
	}

	rawPrices := make(map[types.Symbol]float64)
	for symbol := range symbols {
		if price, ok := cmcPrice[string(symbol)]; ok {
			rawPrices[symbol] = price
		} else {
			return nil, fmt.Errorf("symbol %s not found in response: %s\n", symbol, response)
		}
	}

	return rawPrices, err
}

func buildReq(symbols set.Set[types.Symbol], c *CoinmarketcapConfig) (*http.Request, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, fmt.Errorf("Can not create a request with link: %s\n", link)
	}

	params := url.Values{}
	params.Add("slug", coinmarketcapSymbolCsv(symbols))

	req.Header.Set("Accepts", "application/json")
	req.Header.Add(apiKeyHeaderParam, c.ApiKey)
	req.URL.RawQuery = params.Encode()

	return req, nil
}

// coinmarketcapSymbolCsv returns the symbols as a comma separated string.
func coinmarketcapSymbolCsv(symbols set.Set[types.Symbol]) string {
	s := ""
	for symbol := range symbols {
		s += string(symbol) + ","
	}

	return s[:len(s)-1]
}
