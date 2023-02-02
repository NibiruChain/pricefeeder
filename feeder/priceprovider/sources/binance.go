package sources

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/price-feeder/types"
)

const (
	Binance = "binance"
)

var _ types.FetchPricesFunc = BinancePriceUpdate

type BinanceTicker struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
}

func BinanceSymbolCsv(symbols set.Set[types.Symbol]) string {
	s := ""
	for symbol := range symbols {
		s += "%22" + string(symbol) + "%22,"
	}
	// chop off trailing comma
	return s[:len(s)-1]
}

// BinancePriceUpdate returns the prices given the symbols or an error.
// Uses the Binance API at https://docs.binance.us/#price-data.
func BinancePriceUpdate(symbols set.Set[types.Symbol]) (rawPrices map[types.Symbol]float64, err error) {
	url := "https://api.binance.us/api/v3/ticker/price?symbols=%5B" + BinanceSymbolCsv(symbols) + "%5D"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tickers := make([]BinanceTicker, len(symbols))

	err = json.Unmarshal(b, &tickers)
	if err != nil {
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)
	for _, ticker := range tickers {
		rawPrices[types.Symbol(ticker.Symbol)] = ticker.Price
	}

	return rawPrices, nil
}
