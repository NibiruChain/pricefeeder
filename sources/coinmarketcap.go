// Credit: @oleksandrmarkelov https://github.com/NibiruChain/pricefeeder/pull/27
package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	SourceNameCoinMarketCap = "coinmarketcap"
	link                    = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest"
	apiKeyHeaderParam       = "X-CMC_PRO_API_KEY"
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

// CoinmarketcapPriceUpdate returns a fetch function that obtains USD prices for the given symbols from CoinMarketCap.
// 
// The coinmarketcapConfig parameter is an optional JSON configuration that may contain the API key (field `api_key`).
// The returned function accepts a set of symbols and a logger, and returns a map from symbol to USD price or an error.
// It sets the CoinMarketCap API key header when provided, logs failures, and updates the price source metrics labeled with
// SourceNameCoinMarketCap to indicate success or failure.
func CoinmarketcapPriceUpdate(coinmarketcapConfig json.RawMessage) types.FetchPricesFunc {
	return func(symbols set.Set[types.Symbol], logger zerolog.Logger) (map[types.Symbol]float64, error) {
		config, err := getConfig(coinmarketcapConfig)
		if err != nil {
			logger.Err(err).Msg("failed to extract coinmarketcap config")
			metrics.PriceSourceCounter.WithLabelValues(SourceNameCoinMarketCap, "false").Inc()
			return nil, err
		}

		req, err := buildReq(symbols, config)
		if err != nil {
			logger.Err(err).Msg("failed to build request for Coinmarketcap")
			metrics.PriceSourceCounter.WithLabelValues(SourceNameCoinMarketCap, "false").Inc()
			return nil, err
		}

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			logger.Err(err).Msg("failed to fetch prices from Coinmarketcap")
			metrics.PriceSourceCounter.WithLabelValues(SourceNameCoinMarketCap, "false").Inc()
			return nil, err
		}

		defer func() {
			errClose := res.Body.Close()
			if errClose != nil {
				errClose = fmt.Errorf("error closing response body: %w", errClose)
				logger.Err(errClose).Str("source", SourceNameCoinMarketCap).Msg(errClose.Error())
			}
		}()

		response, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Err(err).Msg("failed to read response body from Coinmarketcap")
			metrics.PriceSourceCounter.WithLabelValues(SourceNameCoinMarketCap, "false").Inc()
			return nil, err
		}

		rawPrices, err := getPricesFromResponse(symbols, response, logger)
		if err != nil {
			logger.Err(err).Msg("failed to extract prices from Coinmarketcap response")
			metrics.PriceSourceCounter.WithLabelValues(SourceNameCoinMarketCap, "false").Inc()
			return nil, err
		}

		metrics.PriceSourceCounter.WithLabelValues(SourceNameCoinMarketCap, "true").Inc()
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

// getPricesFromResponse parses a CoinMarketCap JSON response and extracts USD prices for the requested symbols.
// 
// It returns a map from the requested symbols to their USD price for every symbol found in the response.
// If the response cannot be unmarshaled, an error is returned and the map is nil. Symbols not present in the
// response are omitted from the returned map.
func getPricesFromResponse(symbols set.Set[types.Symbol], response []byte, logger zerolog.Logger) (map[types.Symbol]float64, error) {
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
			logger.Debug().Msg(fmt.Sprintf("fetched price for %s on data source %s: %f", symbol, SourceNameCoinMarketCap, price))
		} else {
			logger.Err(err).Msg(fmt.Sprintf("failed to parse price for %s on data source %s", symbol, SourceNameCoinMarketCap))
			continue
		}
	}

	return rawPrices, err
}

func buildReq(symbols set.Set[types.Symbol], c *CoinmarketcapConfig) (*http.Request, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create a request with link: %s", link)
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