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
		success:         false,
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

	success bool
}

func (e *exContext) do() {
	defer close(e.done)
	// after this we can loop attempting to send the new vote && prevote tx.
	for {
		select {
		case <-e.ctx.Done():
			// vote miss, exit, no signal success
			return
		default:
		}

		resp, err := vote(e.ctx, e.currentPrevote, e.previousPrevote, e.validator, e.feeder, e.deps, e.log)
		if err == nil {
			e.log.Info().Str("tx-hash", resp.TxHash).Msg("transaction successfully sent")
			break
		}
		e.handleFailure(resp, err)
	}
	e.handleSuccess()
}

func (e *exContext) terminate() {
	e.cancel()
	<-e.done
}

func (e *exContext) isSuccess() bool {
	return e.success
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
	e.success = true
}
