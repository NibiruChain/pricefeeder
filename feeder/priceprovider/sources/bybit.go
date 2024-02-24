package sources

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	Bybit = "bybit"
)

var _ types.FetchPricesFunc = BybitPriceUpdate

type BybitTicker struct {
	Symbol string `json:"symbol"`
	Price  string `json:"lastPrice"`
}

type Response struct {
	Data struct {
		List []BybitTicker `json:"list"`
	} `json:"result"`
}

// BybitPriceUpdate returns the prices for given symbols or an error.
// Uses BYBIT API at https://bybit-exchange.github.io/docs/v5/market/tickers.
func BybitPriceUpdate(symbols set.Set[types.Symbol]) (rawPrices map[types.Symbol]float64, err error) {
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

	var response Response
	err = json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)

	for _, ticker := range response.Data.List {
		symbol := types.Symbol(ticker.Symbol)
		if !symbols.Has(symbol) {
			continue
		}

		price, err := strconv.ParseFloat(ticker.Price, 64)
		if err != nil {
			price = -1
		}
		rawPrices[symbol] = price
	}
	return rawPrices, nil
}
