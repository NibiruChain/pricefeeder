package sources

import (
	json2 "encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/NibiruChain/price-feeder/types"
	"github.com/tendermint/tendermint/libs/json"
)

const (
	Coingecko = "coingecko"
)

type CoingeckoTicker struct {
	Price float64 `json:"usd,string"`
}

func CoingeckoPriceUpdate(config json2.RawMessage) types.FetchPricesFunc {
	return func(symbols []types.Symbol) (map[types.Symbol]float64, error) {
		baseURL := buildURL(symbols)

		res, err := http.Get(baseURL)
		if err != nil {
			fmt.Println(err)
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

func extractPricesFromResponse(symbols []types.Symbol, response []byte) (map[types.Symbol]float64, error) {
	var result map[string]CoingeckoTicker
	err := json.Unmarshal(response, &result)
	if err != nil {
		return nil, err
	}

	rawPrices := make(map[types.Symbol]float64)
	for _, symbol := range symbols {
		if price, ok := result[string(symbol)]; ok {
			rawPrices[symbol] = price.Price
		} else {
			return nil, fmt.Errorf("symbol %s not found in response: %s\n", symbol, response)
		}
	}

	return rawPrices, err
}

func buildURL(symbols []types.Symbol) string {
	baseURL := "https://api.coingecko.com/api/v3/simple/price?"

	params := url.Values{}
	params.Add("ids", coingeckoSymbolCsv(symbols))
	params.Add("vs_currencies", "usd")

	baseURL = baseURL + params.Encode()
	return baseURL
}

// coingeckoSymbolCsv returns the symbols as a comma separated string.
func coingeckoSymbolCsv(symbols []types.Symbol) string {
	s := ""
	for _, symbol := range symbols {
		s += string(symbol) + ","
	}

	return s[:len(s)-1]
}
