package tx

import (
	"context"
	"github.com/NibiruChain/nibiru/app"
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	"github.com/NibiruChain/price-feeder/feeder/types"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txservice "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"sync"
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

func Dial(grpcEndpoint string, chainID string, keyBase keyring.Keyring, validator sdk.ValAddress, feeder sdk.AccAddress, log zerolog.Logger) *Client {
	conn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	encoding := app.MakeTestEncodingConfig()
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
		log:                     log,
		validator:               validator,
		feeder:                  feeder,
		mu:                      new(sync.Mutex),
		currentExecutionContext: nil,
		deps:                    deps,
	}
}

type Client struct {
	log zerolog.Logger

	validator sdk.ValAddress
	feeder    sdk.AccAddress

	mu                      *sync.Mutex
	currentExecutionContext *exContext

	deps deps
}

func (c *Client) Whoami() sdk.ValAddress {
	return c.validator
}

func (c *Client) SendPrices(ctx context.Context, vp types.VotingPeriod, prices []types.Price) (signalSuccess chan struct{}) {
	defer c.mu.Unlock()
	c.mu.Lock()
	// first time executing
	log := c.log.With().Uint64("voting-period-height", vp.Height).Logger()
	if c.currentExecutionContext == nil {
		c.currentExecutionContext = newContext(ctx, prices, nil, c.deps, c.validator, c.feeder, log)
	} else { // already executed once...
		if !c.currentExecutionContext.isSuccess() {
			c.currentExecutionContext = newContext(ctx, prices, nil, c.deps, c.validator, c.feeder, log)
		} else {
			c.currentExecutionContext = newContext(ctx, prices, c.currentExecutionContext.currentPrevote, c.deps, c.validator, c.feeder, log)
		}
	}
	return c.currentExecutionContext.signalSuccess
}

func (c *Client) Close() {
	defer c.mu.Unlock()
	c.mu.Lock()
	if c.currentExecutionContext != nil {
		c.currentExecutionContext.terminate()
	}
}
