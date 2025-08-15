package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
)

const (
	Okex = "okex"
)

var _ types.FetchPricesFunc = OkexPriceUpdate

type OkexTicker struct {
	Symbol string `json:"instId"`
	Price  string `json:"last"`
}

type OkexResponse struct {
	Data []OkexTicker `json:"data"`
}

// OkexPriceUpdate returns the prices for given symbols or an error.
// Uses OKEX API at https://www.okx.com/docs-v5/en/#rest-api-market-data.
func OkexPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	url := "https://www.okx.com/api/v5/market/tickers?instType=SPOT"

	resp, err := http.Get(url)
	if err != nil {
		logger.Err(err).Msg("failed to fetch prices from Okex")
		metrics.PriceSourceCounter.WithLabelValues(Okex, "false").Inc()
		return nil, err
	}

	defer func() {
		errClose := resp.Body.Close()
		if errClose != nil {
			errClose = fmt.Errorf("error closing response body: %w", errClose)
			logger.Err(errClose).Str("source", Okex).Msg(errClose.Error())
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("failed to read response body from Okex")
		metrics.PriceSourceCounter.WithLabelValues(Okex, "false").Inc()
		return nil, err
	}

	var response OkexResponse
	err = json.Unmarshal(b, &response)
	if err != nil {
		logger.Err(err).Msg("failed to unmarshal response body from Okex")
		metrics.PriceSourceCounter.WithLabelValues(Okex, "false").Inc()
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)
	for _, ticker := range response.Data {

		symbol := types.Symbol(ticker.Symbol)
		if !symbols.Has(symbol) {
			continue
		}

		price, err := strconv.ParseFloat(ticker.Price, 64)
		if err != nil {
			logger.Err(err).Msg(fmt.Sprintf("failed to parse price for %s on data source %s", symbol, Okex))
			continue
		}

		rawPrices[symbol] = price
		logger.Debug().Msg(fmt.Sprintf("fetched price for %s on data source %s: %f", symbol, Okex, price))
	}

	metrics.PriceSourceCounter.WithLabelValues(Okex, "true").Inc()
	return rawPrices, nil
}
