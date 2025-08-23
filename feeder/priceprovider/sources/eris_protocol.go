package sources

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc/credentials/insecure"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
)

const (
	SourceErisProtocol = "eris_protocol"
	// Configuration constants
	stateQuery = `{"state": {}}`
)

var grpcReadEndpoint string

var _ types.FetchPricesFunc = ErisProtocolPriceUpdate

// newGRPCConnection creates a new gRPC connection with TLS
func newGRPCConnection() (*grpc.ClientConn, error) {
	grpcReadEndpoint = os.Getenv("GRPC_READ_ENDPOINT")
	if grpcReadEndpoint == "" {
		grpcReadEndpoint = os.Getenv("GRPC_ENDPOINT")
	}
	if grpcReadEndpoint == "" {
		grpcReadEndpoint = "localhost:9090"
	}
	enableTLS := os.Getenv("ENABLE_TLS") == "true" || strings.Contains(grpcReadEndpoint, ":443")
	creds := insecure.NewCredentials()
	if enableTLS {
		creds = credentials.NewTLS(
			&tls.Config{
				InsecureSkipVerify: false,
			},
		)
	}
	transportDialOpt := grpc.WithTransportCredentials(creds)
	return grpc.Dial(grpcReadEndpoint, transportDialOpt)
}

// ErisProtocolPriceUpdate retrieves the exchange rate for stNIBI to NIBI (ustnibi:unibi) from the Eris Protocol smart contract.
// Note: This function ignores the input symbols and always returns the exchange rate for "ustnibi:unibi".
func ErisProtocolPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (rawPrices map[types.Symbol]float64, err error) {
	conn, err := newGRPCConnection()
	if err != nil {
		logger.Err(err).Msgf("failed to connect to gRPC endpoint %s", grpcReadEndpoint)
		metrics.PriceSourceCounter.WithLabelValues(SourceErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to connect to gRPC endpoint %s: %w", grpcReadEndpoint, err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logger.Err(closeErr).Str("source", SourceErisProtocol).Msg("failed to close gRPC connection")
		}
	}()

	contractAddr := os.Getenv("ERIS_PROTOCOL_CONTRACT_ADDRESS")
	if contractAddr == "" {
		contractAddr = "nibi1udqqx30cw8nwjxtl4l28ym9hhrp933zlq8dqxfjzcdhvl8y24zcqpzmh8m" // mainnet
	}

	wasmClient := wasmtypes.NewQueryClient(conn)
	query := wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: []byte(stateQuery),
	}

	resp, err := wasmClient.SmartContractState(context.Background(), &query)
	if err != nil {
		logger.Err(err).Msg("failed to query SmartContractState")
		metrics.PriceSourceCounter.WithLabelValues(SourceErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to query SmartContractState: %w", err)
	}

	if resp == nil || len(resp.Data) == 0 {
		logger.Error().Msg("received nil or empty response from SmartContractState")
		metrics.PriceSourceCounter.WithLabelValues(SourceErisProtocol, "false").Inc()
		return nil, fmt.Errorf("nil response from SmartContractState")
	}

	var responseObj struct {
		ExchangeRate string `json:"exchange_rate"`
	}
	if err := json.Unmarshal(resp.Data, &responseObj); err != nil {
		logger.Err(err).Msg("failed to unmarshal SmartContractState response data")
		metrics.PriceSourceCounter.WithLabelValues(SourceErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to unmarshal SmartContractState response data: %w", err)
	}

	exchangeRate, err := strconv.ParseFloat(responseObj.ExchangeRate, 64)
	if err != nil {
		logger.Err(err).Msg("failed to convert exchange_rate to float")
		metrics.PriceSourceCounter.WithLabelValues(SourceErisProtocol, "false").Inc()
		return nil, fmt.Errorf("failed to convert exchange_rate to float: %w", err)
	}

	// Validate the exchange rate
	if exchangeRate <= 0 {
		errMsg := "received invalid exchange rate: value must be positive"
		logger.Error().Float64("exchange_rate", exchangeRate).Msg(errMsg)
		metrics.PriceSourceCounter.WithLabelValues(SourceErisProtocol, "false").Inc()
		return nil, errors.New(errMsg)
	}

	rawPrices = make(map[types.Symbol]float64)
	rawPrices[types.Symbol("ustnibi:unibi")] = exchangeRate

	logger.Debug().
		Str("source", SourceErisProtocol).
		Float64("exchange_rate", exchangeRate).
		Str("symbols", fmt.Sprint(symbols)).
		Msg("fetched prices")

	metrics.PriceSourceCounter.WithLabelValues(SourceErisProtocol, "true").Inc()
	return rawPrices, nil
}
