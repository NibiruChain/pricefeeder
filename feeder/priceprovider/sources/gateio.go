package sources

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	GateIo = "gateio"
)

var _ types.FetchPricesFunc = GateIoPriceUpdate

// BinancePriceUpdate returns the prices given the symbols or an error.
// Uses the Binance API at https://docs.binance.us/#price-data.
func GateIoPriceUpdate(symbols set.Set[types.Symbol]) (rawPrices map[types.Symbol]float64, err error) {
	url := "https://api.gateio.ws/api/v4/spot/tickers"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tickers []map[string]interface{}
	err = json.Unmarshal(b, &tickers)
	if err != nil {
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)
	for _, ticker := range tickers {
		pair := ticker["currency_pair"].(string)
		symbol := types.Symbol(strings.Replace(pair, "_", "", -1))

		lastPrice, err := strconv.ParseFloat(ticker["last"].(string), 64)
		if err != nil {
			return rawPrices, err
		}
		if _, ok := symbols[symbol]; ok {
			rawPrices[symbol] = lastPrice
		}

	}

	return rawPrices, nil
}
