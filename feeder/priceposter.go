package feeder

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NibiruChain/nibiru/v2/app"
	oracletypes "github.com/NibiruChain/nibiru/v2/x/oracle/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
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

	"github.com/NibiruChain/pricefeeder/metrics"
	"github.com/NibiruChain/pricefeeder/types"
)

var _ types.PricePoster = (*ClientPricePoster)(nil)

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

func DialPricePoster(
	grpcEndpoint string,
	chainID string,
	enableTLS bool,
	keyBase keyring.Keyring,
	validator sdk.ValAddress,
	feeder sdk.AccAddress,
	logger zerolog.Logger,
) *ClientPricePoster {
	creds := insecure.NewCredentials()
	if enableTLS {
		creds = credentials.NewTLS(
			&tls.Config{
				InsecureSkipVerify: false,
			},
		)
	}
	transportDialOpt := grpc.WithTransportCredentials(creds)

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

	return &ClientPricePoster{
		logger:    logger,
		validator: validator,
		feeder:    feeder,
		deps:      deps,
	}
}

type ClientPricePoster struct {
	logger zerolog.Logger

	validator sdk.ValAddress
	feeder    sdk.AccAddress

	previousPrevote *prevote
	deps            deps
}

func (c *ClientPricePoster) Whoami() sdk.ValAddress {
	return c.validator
}

var pricePosterCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: metrics.PrometheusNamespace,
	Name:      "prices_posted_total",
	Help:      "The total number of price update txs sent to the chain, by success status",
}, []string{"success"})

func (c *ClientPricePoster) SendPrices(vp types.VotingPeriod, prices []types.Price) {
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

func (c *ClientPricePoster) Close() {
}

// GetOracleClient returns the oracle client for testing purposes.
func (c *ClientPricePoster) GetOracleClient() Oracle {
	return c.deps.oracleClient
}

// TryUntilDone will try to execute the given function until it succeeds or the
// context is cancelled.
func TryUntilDone(
	ctx context.Context,
	wait time.Duration,
	f func() error,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := f(); err == nil {
				return nil
			}
			time.Sleep(wait)
		}
	}
}

/// -------------------------------------------------
/// SIGNING
/// -------------------------------------------------

// sendTx constructs, signs, and broadcasts a transaction containing the provided
// messages. It retrieves account information (account number and sequence) from
// the chain and sets transaction fees and gas limits. Returns the transaction
// response or an error if the broadcast fails or the transaction is rejected.
//
// Panics on configuration errors (missing key, encoding failures) as these
// indicate programmer errors rather than runtime failures.
func sendTx(
	ctx context.Context,
	keyBase keyring.Keyring,
	authClient Auth,
	txClient TxService,
	feeder sdk.AccAddress,
	txConfig client.TxConfig,
	ir codectypes.InterfaceRegistry,
	chainID string,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {
	// get key from keybase, can't fail
	keyInfo, err := keyBase.KeyByAddress(feeder)
	if err != nil {
		panic(err)
	}

	// set msgs, can't fail
	txBuilder := txConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		panic(err)
	}

	txFeeStr := os.Getenv("FEE_AMOUNT_UNIBI")
	feeAmount := int64(125)
	if txFeeStr != "" {
		if v, err := strconv.ParseInt(txFeeStr, 10, 64); err == nil {
			feeAmount = v
		}
	}
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("unibi", feeAmount)))

	gasStr := os.Getenv("GAS_LIMIT")
	gasLimit := uint64(7000)
	if gasStr != "" {
		if v, err := strconv.ParseUint(gasStr, 10, 64); err == nil {
			gasLimit = v
		}
	}
	txBuilder.SetGasLimit(gasLimit)

	// get acc info, can fail
	accNum, sequence, err := getAccount(ctx, authClient, ir, feeder)
	if err != nil {
		return nil, err
	}

	txFactory := tx.Factory{}.
		WithChainID(chainID).
		WithKeybase(keyBase).
		WithTxConfig(txConfig).
		WithAccountNumber(accNum).
		WithSequence(sequence)

	// sign tx, can't fail
	err = tx.Sign(txFactory, keyInfo.Name, txBuilder, true)
	if err != nil {
		panic(err)
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	resp, err := txClient.BroadcastTx(ctx, &txservice.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txservice.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	if err != nil {
		return nil, err
	}
	if resp.TxResponse.Code != abcitypes.CodeTypeOK {
		return resp.TxResponse, fmt.Errorf("tx failed: %s", resp.TxResponse.RawLog)
	}
	return resp.TxResponse, nil
}

// The [getAccount] fn retrieves the account number and sequence for the given feeder
// address from the chain. These values are required for constructing valid
// transactions. Returns an error if the account cannot be found or unpacked.
func getAccount(
	ctx context.Context,
	authClient Auth,
	ir codectypes.InterfaceRegistry,
	feeder sdk.AccAddress,
) (uint64, uint64, error) {
	accRaw, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: feeder.String()})
	if err != nil {
		return 0, 0, err // if account not found it's pointless to continue
	}

	var acc authtypes.AccountI
	err = ir.UnpackAny(accRaw.Account, &acc)
	if err != nil {
		panic(err)
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}

/// -------------------------------------------------
/// Oracle:  vote, MaxSaltNumber, prepareVote, prevote, newPrevote
/// -------------------------------------------------

// MaxSaltNumber is the maximum salt number we can use for randomness.
// NOTE: max length of the salt is 4.
// TODO(mercilex): if we used digits + alphanumerics it's more randomized
var MaxSaltNumber = big.NewInt(9999) // NOTE(mercilex): max salt length is 4

// The [vote] fn submits a prevote message to the chain, and optionally a vote message
// if an old prevote exists. The vote is constructed from the old prevote's
// hash and salt. Messages are ordered such that the vote (if present) is sent
// before the new prevote, as the new prevote will overwrite the old one.
//
// oldPrevote may be nil if no previous prevote exists, in which case only
// the new prevote is submitted.
func vote(
	ctx context.Context,
	newPrevote, oldPrevote *prevote,
	validator sdk.ValAddress,
	feeder sdk.AccAddress,
	deps deps,
	logger zerolog.Logger,
) (txResponse *sdk.TxResponse, err error) {
	// if oldPrevote is not nil then we need to get the vote msg
	var voteMsg *oracletypes.MsgAggregateExchangeRateVote
	if oldPrevote != nil {
		voteMsg, err = prepareVote(ctx, deps.oracleClient, validator, feeder, oldPrevote, logger)
		if err != nil {
			return nil, err
		}
		logger.Info().Interface("vote", voteMsg).Msg("prepared vote message")
	}
	// once we prepared the vote msg we can send the tx
	msgs := []sdk.Msg{newPrevote.msg}
	// if there was a vote then we of course need to vote first and then prevote.
	if voteMsg != nil {
		// note ordering matters because the new prevote will overwrite the old one
		msgs = []sdk.Msg{voteMsg, newPrevote.msg}
	} else {
		logger.Info().Msg("skipping vote preparation as there is no old prevote")
	}

	return sendTx(
		ctx, deps.keyBase, deps.authClient, deps.txClient,
		feeder, deps.txConfig, deps.ir, deps.chainID, msgs...,
	)
}

// The [prepareVote] fn constructs a vote message from an existing prevote on the chain.
// It verifies that the local prevote hash matches the chain's prevote hash
// to ensure the prevote hasn't been tampered with or expired. Returns nil
// if no prevote exists on the chain or if the hashes don't match, indicating
// the prevote has expired or been invalidated.
func prepareVote(
	ctx context.Context,
	oracleClient Oracle,
	validator sdk.ValAddress, feeder sdk.AccAddress,
	prevote *prevote,
	logger zerolog.Logger,
) (*oracletypes.MsgAggregateExchangeRateVote, error) {
	log := logger.With().Str("stage", "prepare-vote").Logger()
	// there might be cases where due to downtimes the prevote
	// has expired. So we check if a prevote exists in the chain, if it does not
	// then we simply return.
	resp, err := oracleClient.AggregatePrevote(ctx, &oracletypes.QueryAggregatePrevoteRequest{
		ValidatorAddr: validator.String(),
	})
	if err != nil {
		log.Err(err).Msg("failed to get aggregate prevote from chain")
		return nil, nil
	}

	// assert equality between feeder's prevote and chain's prevote
	if localHash := oracletypes.GetAggregateVoteHash(prevote.salt, prevote.vote, validator).String(); localHash != resp.AggregatePrevote.Hash {
		log.Warn().Str("chain hash", resp.AggregatePrevote.Hash).Str("local hash", localHash).Msg("chain and local prevote do not match")
		return nil, nil
	}

	return &oracletypes.MsgAggregateExchangeRateVote{
		Salt:          prevote.salt,
		ExchangeRates: prevote.vote,
		Feeder:        feeder.String(),
		Validator:     validator.String(),
	}, nil
}

// The [prevote] struct contains the data needed for a prevote-vote cycle in the
// oracle. The [prevote] message is sent first with a hash of the vote data,
// followed by the actual vote message in the next voting period. This two-phase
// commit pattern prevents front-running and ensures vote integrity.
type prevote struct {
	msg  *oracletypes.MsgAggregateExchangeRatePrevote
	salt string
	vote string
}

// The [newPrevote] fn creates a new prevote message from a list of prices. It generates
// a random salt value, formats the prices as exchange rate tuples, and computes
// the aggregate vote hash. The prevote struct contains both the message to send
// and the data needed to construct the corresponding vote in the next voting period.
func newPrevote(prices []types.Price, validator sdk.ValAddress, feeder sdk.AccAddress) *prevote {
	tuple := make(oracletypes.ExchangeRateTuples, len(prices))
	for i, price := range prices {
		tuple[i] = oracletypes.ExchangeRateTuple{
			Pair:         price.Pair,
			ExchangeRate: float64ToDec(price.Price),
		}
	}

	votesStr, err := tuple.ToString()
	if err != nil {
		panic(err)
	}
	nBig, err := rand.Int(rand.Reader, MaxSaltNumber)
	if err != nil {
		panic(err)
	}
	salt := nBig.String()

	hash := oracletypes.GetAggregateVoteHash(salt, votesStr, validator)

	return &prevote{
		msg:  oracletypes.NewMsgAggregateExchangeRatePrevote(hash, feeder, validator),
		salt: salt,
		vote: votesStr,
	}
}

func float64ToDec(price float64) sdk.Dec {
	formattedPrice := strconv.FormatFloat(price, 'f', -1, 64)

	parts := strings.Split(formattedPrice, ".")
	intPart := parts[0]
	decPart := ""

	if len(parts) > 1 {
		decPart = parts[1]
	}

	if len(decPart) > 18 {
		decPart = decPart[:18]
	}

	if decPart == "" {
		formattedPrice = intPart
	} else {
		formattedPrice = intPart + "." + decPart
	}

	return sdk.MustNewDecFromStr(formattedPrice)
}
