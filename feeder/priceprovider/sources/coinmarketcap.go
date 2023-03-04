package sources

import (
	json2 "encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	Coinmarketcap   = "coinmarketcap"
	Link    = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest"
	CMCApiKeyParam = "X-CMC_PRO_API_KEY"
)

type Cmcquoteprice struct {
	Price float64
}

type Cmcquote struct {
	USD Cmcquoteprice
}

type CmcTicker struct {
	Slug string
	Quote Cmcquote
}

type CmcResponse struct {
	Data map[string]CmcTicker
}

type CoinmarketcapConfig struct {
	ApiKey string `json:"api_key"`
}

func CoinmarketcapPriceUpdate(jsonConfig json2.RawMessage) types.FetchPricesFunc {
	return func(symbols set.Set[types.Symbol]) (map[types.Symbol]float64, error) {
		client := &http.Client{}

		c, err := getConfig(jsonConfig)
		if err != nil {
			return nil, err
		}

		req, err := buildReq(symbols, c)
		if err != nil {
			return nil, err
		}

		res, err := client.Do(req);
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
func getConfig(jsonConfig json2.RawMessage) (*CoinmarketcapConfig, error) {
	c := &CoinmarketcapConfig{}
	if len(jsonConfig) > 0 {
		err := json2.Unmarshal(jsonConfig, c)
		if err != nil {
			return nil, fmt.Errorf("invalid coinmarketcap config: %w", err)
		}
	}
	return c, nil
}

func getPricesFromResponse(symbols set.Set[types.Symbol], response []byte) (map[types.Symbol]float64, error) {
	var respCmc CmcResponse
	err := json2.Unmarshal(response, &respCmc)
	if err != nil {
		return nil, err
	}
	
	cmcPrice := make(map[string]float64)
	for _,value  := range respCmc.Data {
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

	req, err := http.NewRequest("GET", Link, nil)
	if err != nil {
		return nil, fmt.Errorf("Can not create a requet with link: %s\n", Link)
	}

	params := url.Values{}
	params.Add("slug", coinmarketcapSymbolCsv(symbols))

	req.Header.Set("Accepts", "application/json")
	req.Header.Add(CMCApiKeyParam, c.ApiKey)
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
