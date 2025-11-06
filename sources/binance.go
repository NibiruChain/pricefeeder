package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	SourceNameBinance = "binance"
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
func BinancePriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	url := "https://api.binance.us/api/v3/ticker/price?symbols=%5B" + BinanceSymbolCsv(symbols) + "%5D"
	resp, err := http.Get(url)
	if err != nil {
		logger.Err(err).Msg("failed to fetch prices from Binance")
		metrics.PriceSourceCounter.WithLabelValues(SourceNameBinance, "false").Inc()
		return nil, err
	}
	defer func() {
		errClose := resp.Body.Close()
		if errClose != nil {
			errClose = fmt.Errorf("error closing response body: %w", errClose)
			logger.Err(errClose).Str("source", SourceNameBinance).Msg(errClose.Error())
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("failed to read response body from Binance")
		metrics.PriceSourceCounter.WithLabelValues(SourceNameBinance, "false").Inc()
		return nil, err
	}

	tickers := make([]BinanceTicker, len(symbols))

	err = json.Unmarshal(b, &tickers)
	if err != nil {
		logger.Err(err).Msg("failed to unmarshal response body from Binance")
		metrics.PriceSourceCounter.WithLabelValues(SourceNameBinance, "false").Inc()
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)
	for _, ticker := range tickers {
		rawPrices[types.Symbol(ticker.Symbol)] = ticker.Price
		logger.Debug().Msgf("fetched price for %s on data source %s: %f", ticker.Symbol, SourceNameBinance, ticker.Price)
	}
	metrics.PriceSourceCounter.WithLabelValues(SourceNameBinance, "true").Inc()

	return rawPrices, nil
}
