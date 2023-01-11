package priceposter

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	"github.com/NibiruChain/price-feeder/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
)

var (
	// MaxSaltNumber is the maximum salt number we can use for randomness.
	// NOTE: max length of the salt is 4.
	// TODO(mercilex): if we used digits + alphanumerics it's more randomized
	MaxSaltNumber = big.NewInt(9999) // NOTE(mercilex): max salt length is 4
)

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
	var msgs = []sdk.Msg{newPrevote.msg}
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
		// TODO(mercilex): a better way?
		if strings.Contains(err.Error(), oracletypes.ErrNoAggregatePrevote.Error()) {
			log.Warn().Msg("no aggregate prevote found for this voting period")
			return nil, nil
		} else {
			return nil, err
		}
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

type prevote struct {
	msg  *oracletypes.MsgAggregateExchangeRatePrevote
	salt string
	vote string
}

func newPrevote(prices []types.Price, validator sdk.ValAddress, feeder sdk.AccAddress) *prevote {
	tuple := make(oracletypes.ExchangeRateTuples, len(prices))
	for i, price := range prices {
		tuple[i] = oracletypes.ExchangeRateTuple{
			Pair:         price.Pair.String(),
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
	// TODO(mercilex): precision for numbers with a lot of decimal digits
	return sdk.MustNewDecFromStr(fmt.Sprintf("%f", price))
}
