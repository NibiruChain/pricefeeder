package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
)

const (
	Bybit = "bybit"
)

var _ types.FetchPricesFunc = BybitPriceUpdate

type BybitResponse struct {
	Data struct {
		List []struct {
			Symbol string `json:"symbol"`
			Price  string `json:"lastPrice"`
		} `json:"list"`
	} `json:"result"`
}

const ErrBybitBlockAccess = "configured to block access from your country"

// BybitPriceUpdate returns the prices for given symbols or an error.
// Uses BYBIT API at https://bybit-exchange.github.io/docs/v5/market/tickers.
func BybitPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	url := "https://api.bybit.com/v5/market/tickers?category=spot"

	resp, err := http.Get(url)
	if err != nil {
		logger.Err(err).Msg("failed to fetch prices from Bybit")
		metrics.PriceSourceCounter.WithLabelValues(Bybit, "false").Inc()
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("failed to read response body from Bybit")
		metrics.PriceSourceCounter.WithLabelValues(Bybit, "false").Inc()
		return nil, err
	}

	var response BybitResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		if strings.Contains(string(respBody), ErrBybitBlockAccess) {
			err = fmt.Errorf("%s: %w", ErrBybitBlockAccess, err)
		}
		logger.Err(err).Msg("failed to unmarshal response body from Bybit")
		metrics.PriceSourceCounter.WithLabelValues(Bybit, "false").Inc()
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)

	for _, ticker := range response.Data.List {
		symbol := types.Symbol(ticker.Symbol)
		price, err := strconv.ParseFloat(ticker.Price, 64)
		if err != nil {
			logger.Err(err).Msgf("failed to parse price for %s on data source %s", symbol, Bybit)
			continue
		}

		if _, ok := symbols[symbol]; ok {
			rawPrices[symbol] = price
		}
	}
	logger.Debug().Msgf("fetched prices for %s on data source %s: %v", symbols, Bybit, rawPrices)
	metrics.PriceSourceCounter.WithLabelValues(Bybit, "true").Inc()
	return rawPrices, nil
}
