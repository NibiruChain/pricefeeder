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
	// Configuration constants
	grpcEndpoint = "grpc.nibiru.fi:443"
	contractAddr = "nibi1udqqx30cw8nwjxtl4l28ym9hhrp933zlq8dqxfjzcdhvl8y24zcqpzmh8m"
	stateQuery   = `{"state": {}}`
)

var _ types.FetchPricesFunc = ErisProtocolPriceUpdate

// newGRPCConnection creates a new gRPC connection with TLS
func newGRPCConnection() (*grpc.ClientConn, error) {
	transportDialOpt := grpc.WithTransportCredentials(
		credentials.NewTLS(
			&tls.Config{
				InsecureSkipVerify: false,
			},
		),
	)
	return grpc.Dial(grpcEndpoint, transportDialOpt)
}

// ErisProtocolPriceUpdate retrieves the exchange rate for stNIBI to NIBI (ustnibi:unibi) from the Eris Protocol smart contract.
// Note: This function ignores the input symbols and always returns the exchange rate for "ustnibi:unibi".
func ErisProtocolPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	conn, err := newGRPCConnection()
	if err != nil {
		logger.Err(err).Msgf("failed to connect to gRPC endpoint %s", grpcEndpoint)
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to connect to gRPC endpoint %s: %w", grpcEndpoint, err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logger.Err(closeErr).Msg("failed to close gRPC connection")
		}
	}()

	wasmClient := wasmtypes.NewQueryClient(conn)
	query := wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: []byte(stateQuery),
	}

	resp, err := wasmClient.SmartContractState(context.Background(), &query)
	if err != nil {
		logger.Err(err).Msg("failed to query SmartContractState")
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to query SmartContractState: %w", err)
	}

	if resp == nil || len(resp.Data) == 0 {
		logger.Error().Msg("received nil or empty response from SmartContractState")
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("nil response from SmartContractState")
	}

	var responseObj struct {
		ExchangeRate string `json:"exchange_rate"`
	}
	if err := json.Unmarshal(resp.Data, &responseObj); err != nil {
		logger.Err(err).Msg("failed to unmarshal SmartContractState response data")
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to unmarshal SmartContractState response data: %w", err)
	}

	exchangeRate, err := strconv.ParseFloat(responseObj.ExchangeRate, 64)
	if err != nil {
		logger.Err(err).Msg("failed to convert exchange_rate to float")
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to convert exchange_rate to float: %w", err)
	}

	// Validate the exchange rate
	if exchangeRate <= 0 {
		errMsg := "received invalid exchange rate: value must be positive"
		logger.Error().Float64("exchange_rate", exchangeRate).Msg(errMsg)
		metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "false").Inc()
		return nil, fmt.Errorf(errMsg)
	}

	rawPrices = make(map[types.Symbol]float64)
	rawPrices[types.Symbol("ustnibi:unibi")] = exchangeRate

	logger.Info().
		Str("symbols", fmt.Sprint(symbols)).
		Str("source", ErisProtocol).
		Float64("exchange_rate", exchangeRate).
		Msg("fetched prices")

	metrics.PriceSourceCounter.WithLabelValues(ErisProtocol, "true").Inc()
	return rawPrices, nil
}
