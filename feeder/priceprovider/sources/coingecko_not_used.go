package sources

// NOTE: we don't use CoinGecko provider as it's free rate limits are too low for production use.

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
	SourceCoingecko = "coingecko"
	FreeLink        = "https://api.coingecko.com/api/v3/"
	PaidLink        = "https://pro-api.coingecko.com/api/v3/"
	ApiKeyParam     = "x_cg_pro_api_key"
)

type CoingeckoTicker struct {
	Price float64 `json:"usd"`
}

type CoingeckoConfig struct {
	ApiKey string `json:"api_key"`
}

func CoingeckoPriceUpdate(sourceConfig json.RawMessage) types.FetchPricesFunc {
	return func(symbols set.Set[types.Symbol], logger zerolog.Logger) (map[types.Symbol]float64, error) {
		c, err := extractConfig(sourceConfig)
		if err != nil {
			logger.Err(err).Msg("failed to extract coingecko config")
			metrics.PriceSourceCounter.WithLabelValues(SourceCoingecko, "false").Inc()
			return nil, err
		}

		res, err := http.Get(buildURL(symbols, c))
		if err != nil {
			logger.Err(err).Msg("failed to fetch prices from Coingecko")
			metrics.PriceSourceCounter.WithLabelValues(SourceCoingecko, "false").Inc()
			return nil, err
		}

		defer func() {
			errClose := res.Body.Close()
			if errClose != nil {
				errClose = fmt.Errorf("error closing response body: %w", errClose)
				logger.Err(errClose).Str("source", SourceCoingecko).Msg(errClose.Error())
			}
		}()

		response, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Err(err).Msg("failed to read response body from Coingecko")
			metrics.PriceSourceCounter.WithLabelValues(SourceCoingecko, "false").Inc()
			return nil, err
		}

		rawPrices, err := extractPricesFromResponse(symbols, response, logger)
		if err != nil {
			logger.Err(err).Msg("failed to extract prices from Coingecko response")
			metrics.PriceSourceCounter.WithLabelValues(SourceCoingecko, "false").Inc()
			return nil, err
		}

		metrics.PriceSourceCounter.WithLabelValues(SourceCoingecko, "true").Inc()
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

func extractPricesFromResponse(symbols set.Set[types.Symbol], response []byte, logger zerolog.Logger) (map[types.Symbol]float64, error) {
	var result map[string]CoingeckoTicker
	err := json.Unmarshal(response, &result)
	if err != nil {
		return nil, err
	}

	rawPrices := make(map[types.Symbol]float64)
	for symbol := range symbols {
		if price, ok := result[string(symbol)]; ok {
			rawPrices[symbol] = price.Price
			logger.Debug().Msg(fmt.Sprintf("fetched price for %s on data source %s: %f", symbol, SourceCoingecko, price.Price))
		} else {
			logger.Err(fmt.Errorf("failed to parse price for %s on data source %s", symbol, SourceCoingecko)).Msg(string(response))
			continue
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

	url := baseURL + params.Encode()
	return url
}

// coingeckoSymbolCsv returns the symbols as a comma separated string.
func coingeckoSymbolCsv(symbols set.Set[types.Symbol]) string {
	s := ""
	for symbol := range symbols {
		s += string(symbol) + ","
	}

	return s[:len(s)-1]
}
