package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	SourceNameGateIo = "gateio"
)

var _ types.FetchPricesFunc = GateIoPriceUpdate

// GateIoPriceUpdate returns the prices given the symbols or an error.
// Uses the GateIo API at https://www.gate.io/docs/developers/apiv4/en/#get-details-of-a-specifc-currency-pair.
func GateIoPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	url := "https://api.gateio.ws/api/v4/spot/tickers"
	resp, err := http.Get(url)
	if err != nil {
		logger.Err(err).Msg("failed to fetch prices from GateIo")
		metrics.PriceSourceCounter.WithLabelValues(SourceNameGateIo, "false").Inc()
		return nil, err
	}

	defer func() {
		errClose := resp.Body.Close()
		if errClose != nil {
			errClose = fmt.Errorf("error closing response body: %w", errClose)
			logger.Err(errClose).Str("source", SourceNameGateIo).Msg(errClose.Error())
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("failed to read response body from GateIo")
		metrics.PriceSourceCounter.WithLabelValues(SourceNameGateIo, "false").Inc()
		return nil, err
	}

	var tickers []map[string]any
	err = json.Unmarshal(b, &tickers)
	if err != nil {
		logger.Err(err).Msg("failed to unmarshal response body from GateIo")
		metrics.PriceSourceCounter.WithLabelValues(SourceNameGateIo, "false").Inc()
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)
	for _, ticker := range tickers {
		symbol := types.Symbol(ticker["currency_pair"].(string))
		if !symbols.Has(symbol) {
			continue
		}

		price, err := strconv.ParseFloat(ticker["last"].(string), 64)
		if err != nil {
			logger.Err(err).Msg(fmt.Sprintf("failed to parse price for %s on data source %s", symbol, SourceNameGateIo))
			continue
		}

		rawPrices[symbol] = price
		logger.Debug().Msg(fmt.Sprintf("fetched price for %s on data source %s: %f", symbol, SourceNameGateIo, price))
	}

	metrics.PriceSourceCounter.WithLabelValues(SourceNameGateIo, "true").Inc()
	return rawPrices, nil
}
