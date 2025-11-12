package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"time"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	SourceNamePyth         = "pyth"
	defaultPythEndpoint    = "https://hermes.pyth.network"
	defaultPythTimeout     = 10 * time.Second
	defaultPythMaxPriceAge = 2 * time.Minute
	pythLatestPricePath    = "/v2/updates/price/latest"
	pythQueryParamParsed   = "parsed"
	pythQueryParamIDs      = "ids[]"
	pythQueryValueParsed   = "true"
)

type pythConfig struct {
	Endpoint           string `json:"endpoint"`
	TimeoutSeconds     int    `json:"timeout_seconds"`
	MaxPriceAgeSeconds int    `json:"max_price_age_seconds"`
}

type pythLatestPriceResponse struct {
	Parsed []pythParsedPrice `json:"parsed"`
}

type pythParsedPrice struct {
	ID    string           `json:"id"`
	Price pythPricePayload `json:"price"`
}

type pythPricePayload struct {
	Price       string `json:"price"`
	Expo        int32  `json:"expo"`
	PublishTime int64  `json:"publish_time"`
}

var _ types.FetchPricesFunc = PythPriceUpdate(nil)

// PythPriceUpdate builds a price fetcher backed by the Hermes REST API.
// Optional configuration is provided via the datasource config map and allows
// overriding the endpoint, timeout, and maximum accepted data age.
func PythPriceUpdate(rawCfg json.RawMessage) types.FetchPricesFunc {
	return func(symbols set.Set[types.Symbol], logger zerolog.Logger) (map[types.Symbol]float64, error) {
		cfg := defaultPythConfig()
		if len(rawCfg) > 0 {
			if err := json.Unmarshal(rawCfg, &cfg); err != nil {
				metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "false").Inc()
				return nil, fmt.Errorf("invalid pyth config: %w", err)
			}
		}

		return fetchPythPrices(symbols, cfg, logger)
	}
}

func defaultPythConfig() pythConfig {
	return pythConfig{
		Endpoint:           defaultPythEndpoint,
		TimeoutSeconds:     int(defaultPythTimeout / time.Second),
		MaxPriceAgeSeconds: int(defaultPythMaxPriceAge / time.Second),
	}
}

func fetchPythPrices(symbols set.Set[types.Symbol], cfg pythConfig, logger zerolog.Logger) (map[types.Symbol]float64, error) {
	if symbols.Len() == 0 {
		metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "false").Inc()
		return nil, fmt.Errorf("no symbols configured for pyth")
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = defaultPythEndpoint
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = defaultPythTimeout
	}

	maxAge := time.Duration(cfg.MaxPriceAgeSeconds) * time.Second
	if maxAge <= 0 {
		maxAge = defaultPythMaxPriceAge
	}

	req, err := http.NewRequest(http.MethodGet, endpoint+pythLatestPricePath, nil)
	if err != nil {
		metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "false").Inc()
		return nil, fmt.Errorf("failed to create pyth request: %w", err)
	}

	q := req.URL.Query()
	for symbol := range symbols {
		q.Add(pythQueryParamIDs, string(symbol))
	}
	q.Set(pythQueryParamParsed, pythQueryValueParsed)
	req.URL.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "false").Inc()
		return nil, fmt.Errorf("pyth http request failed: %w", err)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			logger.Err(errClose).Str("source", SourceNamePyth).Msg("failed to close pyth response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "false").Inc()
		return nil, fmt.Errorf("pyth returned non-200 status: %s", resp.Status)
	}

	var parsedResp pythLatestPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "false").Inc()
		return nil, fmt.Errorf("failed to decode pyth response: %w", err)
	}

	if len(parsedResp.Parsed) == 0 {
		metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "false").Inc()
		return nil, fmt.Errorf("pyth response missing parsed prices")
	}

	prices := make(map[types.Symbol]float64, len(parsedResp.Parsed))
	for _, priceEntry := range parsedResp.Parsed {
		price, err := convertPythPrice(priceEntry.Price.Price, priceEntry.Price.Expo)
		if err != nil {
			logger.Err(err).
				Str("id", priceEntry.ID).
				Msg("pyth conversion error")
			continue
		}

		publishTime := time.Unix(priceEntry.Price.PublishTime, 0)
		if age := time.Since(publishTime); age > maxAge {
			logger.Warn().
				Str("id", priceEntry.ID).
				Dur("age", age).
				Dur("max_age", maxAge).
				Msg("pyth price stale")
		}

		prices[types.Symbol(priceEntry.ID)] = price
		logger.Debug().Str("source", SourceNamePyth).
			Str("id", priceEntry.ID).
			Float64("price", price).
			Msg("fetched pyth price")
	}

	if len(prices) == 0 {
		metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "false").Inc()
		return nil, fmt.Errorf("pyth returned zero valid prices")
	}

	metrics.PriceSourceCounter.WithLabelValues(SourceNamePyth, "true").Inc()
	return prices, nil
}

func convertPythPrice(priceStr string, expo int32) (float64, error) {
	intVal, ok := new(big.Int).SetString(priceStr, 10)
	if !ok {
		return 0, fmt.Errorf("invalid price string: %s", priceStr)
	}

	floatVal, _ := new(big.Float).SetInt(intVal).Float64()
	return floatVal * math.Pow10(int(expo)), nil
}
