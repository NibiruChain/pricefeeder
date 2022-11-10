package tx

import (
	"context"
	"github.com/NibiruChain/price-feeder/feeder/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
)

func newContext(
	ctx context.Context,
	newPrices []types.Price,
	previousPrevote *prevote,
	deps deps,
	validator sdk.ValAddress,
	feeder sdk.AccAddress,
	log zerolog.Logger) *exContext {
	ctx, cancel := context.WithCancel(ctx)
	e := &exContext{
		log:             log.With().Str("sub-component", "tx-execution-context").Logger(),
		cancel:          cancel,
		ctx:             ctx,
		done:            make(chan struct{}),
		signalSuccess:   make(chan struct{}),
		currentPrevote:  newPrevote(newPrices, validator, feeder),
		previousPrevote: previousPrevote,
		deps:            deps,
		validator:       validator,
		feeder:          feeder,
	}
	go e.do()
	return e
}

type exContext struct {
	log           zerolog.Logger
	cancel        context.CancelFunc
	ctx           context.Context
	done          chan struct{}
	signalSuccess chan struct{}

	currentPrevote  *prevote
	previousPrevote *prevote
	deps            deps

	validator sdk.ValAddress
	feeder    sdk.AccAddress
}

func (e *exContext) do() {
	defer close(e.done)
	var resp *sdk.TxResponse
	err := tryUntilDone(e.ctx, func() error {
		voteResp, err := vote(e.ctx, e.currentPrevote, e.previousPrevote, e.validator, e.feeder, e.deps, e.log)
		if err != nil {
			e.log.Err(err).Msg("failed to vote")
			return err
		}
		resp = voteResp
		return nil
	})
	if err != nil {
		e.handleFailure(resp, err)
	} else {
		e.handleSuccess()
	}
}

func (e *exContext) terminate() {
	e.cancel()
	<-e.done
}

func (e *exContext) isSuccess() bool {
	select {
	case <-e.signalSuccess:
		return true
	default:
		return false
	}
}

func (e *exContext) handleFailure(resp *sdk.TxResponse, err error) {
	record := e.log.Err(err)
	if resp != nil {
		record.Interface("tx-response", resp)
	}
	record.Msg("failed to send tx")
}

func (e *exContext) handleSuccess() {
	close(e.signalSuccess)
}
