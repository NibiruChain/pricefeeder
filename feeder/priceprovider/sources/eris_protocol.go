package sources

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/NibiruChain/nibiru/x/common/set"
	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	ErisProtocol = "eris_protocol"
)

var _ types.FetchPricesFunc = ErisProtocolPriceUpdate

type ErisResponse struct {
	Data struct {
		List []struct {
			Symbol string `json:"symbol"`
			Price  string `json:"lastPrice"`
		} `json:"list"`
	} `json:"result"`
}

// ErisProtocolPriceUpdate only returns the exchange rate for stNIBI to NIBI (ustnibi:unibi) from the Eris Protocol smart contract.
func ErisProtocolPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	grpcEndpoint := "grpc.nibiru.fi:443"
	transportDialOpt := grpc.WithTransportCredentials(
		credentials.NewTLS(
			&tls.Config{
				InsecureSkipVerify: false,
			},
		),
	)

	conn, err := grpc.Dial(grpcEndpoint, transportDialOpt)
	if err != nil {
		// Handle gRPC connection error
		logger.Err(err).Msgf("failed to connect to gRPC endpoint %s", grpcEndpoint)
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to connect to gRPC endpoint %s: %w", grpcEndpoint, err)
	}
	wasmClient := wasmtypes.NewQueryClient(conn)

	query := wasmtypes.QuerySmartContractStateRequest{
		Address:   "nibi1udqqx30cw8nwjxtl4l28ym9hhrp933zlq8dqxfjzcdhvl8y24zcqpzmh8m",
		QueryData: []byte(`{"state": {}}`),
	}
	resp, err := wasmClient.SmartContractState(context.Background(), &query)
	if err != nil {
		// Handle SmartContractState query error
		logger.Err(err).Msg("failed to query SmartContractState")
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to query SmartContractState: %w", err)
	}

	// Print the response for demonstration purposes
	if resp == nil || len(resp.Data) == 0 {
		// Handle nil or empty response from SmartContractState
		logger.Err(fmt.Errorf("nil response from SmartContractState")).Msg("received nil or empty response")
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("nil response from SmartContractState")
	}

	responseObj := make(map[string]any)
	if err := json.Unmarshal(resp.Data, &responseObj); err != nil {
		// Handle JSON unmarshal error
		logger.Err(err).Msg("failed to unmarshal SmartContractState response data")
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to unmarshal SmartContractState response data: %w", err)
	}

	exchange_rate, err := strconv.ParseFloat(responseObj["exchange_rate"].(string), 64)
	if err != nil {
		// Handle conversion error
		logger.Err(err).Msg("failed to convert exchange_rate to float")
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to convert exchange_rate to float: %w", err)
	}

	// Close the gRPC connection when done
	if err := conn.Close(); err != nil {
		// Handle connection close error
		logger.Err(err).Msg("failed to close gRPC connection")
	}

	logger.Debug().Msgf("fetched prices for %s on data source %s: %v", symbols, ErisProtocol, rawPrices)
	metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "true").Inc()
	rawPrices = make(map[types.Symbol]float64)
	rawPrices[types.Symbol("ustnibi:unibi")] = exchange_rate

	return rawPrices, nil
}
