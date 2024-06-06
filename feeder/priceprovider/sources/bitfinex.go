package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
)

const (
	Bitfinex = "bitfinex"
)

var _ types.FetchPricesFunc = BitfinexPriceUpdate

func BitfinexSymbolCsv(symbols set.Set[types.Symbol]) string {
	s := ""
	for symbol := range symbols {
		s += string(symbol) + ","
	}
	return s[:len(s)-1]
}

// BitfinexPriceUpdate returns the prices given the symbols or an error.
func BitfinexPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	type ticker []interface{}
	const size = 11
	const lastPriceIndex = 7
	const symbolNameIndex = 0

	var url string = "https://api-pub.bitfinex.com/v2/tickers?symbols=" + BitfinexSymbolCsv(symbols)
	resp, err := http.Get(url)
	if err != nil {
		logger.Err(err).Msg("failed to fetch prices from Bitfinex")
		metrics.PriceSourceCounter.WithLabelValues(Bitfinex, "false").Inc()
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("failed to read response body from Bitfinex")
		metrics.PriceSourceCounter.WithLabelValues(Bitfinex, "false").Inc()
		return nil, err
	}
	var tickers []ticker

	err = json.Unmarshal(b, &tickers)
	if err != nil {
		logger.Err(err).Msg("failed to unmarshal response body from Bitfinex")
		metrics.PriceSourceCounter.WithLabelValues(Bitfinex, "false").Inc()
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)
	for _, ticker := range tickers {
		if len(ticker) != size {
			return nil, fmt.Errorf("impossible to parse ticker size %d, %#v", len(ticker), ticker) // TODO(mercilex): return or log and continue?
		}
		symbol := types.Symbol(ticker[symbolNameIndex].(string))
		lastPrice := ticker[lastPriceIndex].(float64)

		rawPrices[symbol] = lastPrice
		logger.Debug().Msg(fmt.Sprintf("fetched price for %s on data source %s: %f", symbol, Bitfinex, lastPrice))
	}

	metrics.PriceSourceCounter.WithLabelValues(Bitfinex, "true").Inc()

	return rawPrices, nil
}
