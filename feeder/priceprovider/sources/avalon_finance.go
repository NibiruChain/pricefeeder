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
	// https://www.avalonfinance.xyz/
	SourceAvalon = "avalon_finance"

	Symbol_sUSDaUSDa types.Symbol = "susda:usda"
	Symbol_sUSDaUSD  types.Symbol = "susda:usd"
)

var _ types.FetchPricesFunc = AvalonPriceUpdate

// AvalonPriceUpdate returns the prices given the symbols or an error.
// Uses the Avalon API at https://www.gate.io/docs/developers/apiv4/en/#get-details-of-a-specifc-currency-pair.
func AvalonPriceUpdate(_ set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	var (
		// API request URL for Avalon Finance sUSDa and USDa redeem ratio.
		url string = "https://avalon-api-world.vercel.app/api/usda/susda-convert-ratio"

		// Response format from the Avalon API.
		// A ratio of "1.08084207433998" means that 1 sUSDa == 1.08 USDa.
		avalonApiResp struct {
			Message string `json:"message"`
			Error   string `json:"error"`
			Data    struct {
				Ratio float64 `json:"ratio"`
			} `json:"data"`
		}
	)

	resp, err := http.Get(url)
	if err != nil {
		logger.Err(err).Msg("failed to fetch prices from Avalon")
		metrics.PriceSourceCounter.WithLabelValues(SourceAvalon, "false").Inc()
		return nil, err
	}

	defer func() {
		errClose := resp.Body.Close()
		if errClose != nil {
			errClose = fmt.Errorf("error closing response body: %w", errClose)
			logger.Err(errClose).Str("source", SourceAvalon).Msg(errClose.Error())
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("failed to read response body from Avalon")
		metrics.PriceSourceCounter.WithLabelValues(SourceAvalon, "false").Inc()
		return nil, err
	}

	err = json.Unmarshal(b, &avalonApiResp)
	if err != nil {
		logger.Err(err).Msg("failed to unmarshal response body from Avalon")
		metrics.PriceSourceCounter.WithLabelValues(SourceAvalon, "false").Inc()
		return nil, err
	}

	rawPrices = make(map[types.Symbol]float64)
	rawPrices[Symbol_sUSDaUSDa] = avalonApiResp.Data.Ratio

	logger.Debug().
		Str("source", SourceAvalon).
		Str("symbol", string(Symbol_sUSDaUSDa)).
		Float64("exchange_rate", avalonApiResp.Data.Ratio).
		Msg("fetched prices")

	metrics.PriceSourceCounter.WithLabelValues(SourceAvalon, "true").Inc()
	return rawPrices, nil
}
