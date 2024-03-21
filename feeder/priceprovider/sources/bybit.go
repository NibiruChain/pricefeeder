package sources

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/NibiruChain/nibiru/x/common/set"
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

// BybitPriceUpdate returns the prices for given symbols or an error.
// Uses BYBIT API at https://bybit-exchange.github.io/docs/v5/market/tickers.
func BybitPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	url := "https://api.bybit.com/v5/market/tickers?category=spot"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response BybitResponse
	err = json.Unmarshal(b, &response)
	if err != nil {
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
	return rawPrices, nil
}
