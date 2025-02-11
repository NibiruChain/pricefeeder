package priceposter

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/NibiruChain/nibiru/app"
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txservice "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var _ types.PricePoster = (*Client)(nil)

type Oracle interface {
	AggregatePrevote(context.Context, *oracletypes.QueryAggregatePrevoteRequest, ...grpc.CallOption) (*oracletypes.QueryAggregatePrevoteResponse, error)
}

type Auth interface {
	Account(context.Context, *authtypes.QueryAccountRequest, ...grpc.CallOption) (*authtypes.QueryAccountResponse, error)
}

type TxService interface {
	BroadcastTx(context.Context, *txservice.BroadcastTxRequest, ...grpc.CallOption) (*txservice.BroadcastTxResponse, error)
}

type deps struct {
	oracleClient Oracle
	authClient   Auth
	txClient     TxService
	keyBase      keyring.Keyring
	txConfig     client.TxConfig
	ir           codectypes.InterfaceRegistry
	chainID      string
}

func Dial(
	grpcEndpoint string,
	chainID string,
	enableTLS bool,
	keyBase keyring.Keyring,
	validator sdk.ValAddress,
	feeder sdk.AccAddress,
	logger zerolog.Logger,
) *Client {
	var transportDialOpt grpc.DialOption

	if enableTLS {
		transportDialOpt = grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					InsecureSkipVerify: false,
				},
			),
		)
	} else {
		transportDialOpt = grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		)
	}

	conn, err := grpc.Dial(grpcEndpoint, transportDialOpt)
	if err != nil {
		panic(err)
	}

	encoding := app.MakeEncodingConfig()
	deps := deps{
		oracleClient: oracletypes.NewQueryClient(conn),
		authClient:   authtypes.NewQueryClient(conn),
		txClient:     txservice.NewServiceClient(conn),
		keyBase:      keyBase,
		txConfig:     encoding.TxConfig,
		ir:           encoding.InterfaceRegistry,
		chainID:      chainID,
	}

	return &Client{
		logger:    logger,
		validator: validator,
		feeder:    feeder,
		deps:      deps,
	}
}

type Client struct {
	logger zerolog.Logger

	validator sdk.ValAddress
	feeder    sdk.AccAddress

	previousPrevote *prevote
	deps            deps
}

func (c *Client) Whoami() sdk.ValAddress {
	return c.validator
}

var pricePosterCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: metrics.PrometheusNamespace,
	Name:      "prices_posted_total",
	Help:      "The total number of price update txs sent to the chain, by success status",
}, []string{"success"})

func (c *Client) SendPrices(vp types.VotingPeriod, prices []types.Price) {
	logger := c.logger.With().Uint64("voting-period-height", vp.Height).Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	newPrevote := newPrevote(prices, c.validator, c.feeder)
	resp, err := vote(ctx, newPrevote, c.previousPrevote, c.validator, c.feeder, c.deps, logger)
	if err != nil {
		logger.Err(err).Msg("prevote failed")
		pricePosterCounter.WithLabelValues("false").Inc()
		return
	}

	c.previousPrevote = newPrevote
	logger.Info().Str("tx-hash", resp.TxHash).Msg("successfully forwarded prices")
	pricePosterCounter.WithLabelValues("true").Inc()
}

func (c *Client) Close() {
}
