package events

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	"github.com/NibiruChain/price-feeder/feeder/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var _ types.EventsStream = (*Stream)(nil)

// wsI exists for testing purposes.
type wsI interface {
	message() <-chan []byte
	close()
}

func Dial(tendermintRPCEndpoint string, grpcEndpoint string, logger zerolog.Logger) *Stream {
	grpcConn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	oracleClient := oracletypes.NewQueryClient(grpcConn)

	const newBlockSubscribe = `{"jsonrpc":"2.0","method":"subscribe","id":0,"params":{"query":"tm.event='NewBlock'"}}`
	ws := NewWebsocket(tendermintRPCEndpoint, []byte(newBlockSubscribe), logger)
	return newStream(ws, oracleClient, logger)
}

func newStream(ws wsI, oracle oracletypes.QueryClient, logger zerolog.Logger) *Stream {
	stream := &Stream{
		stopSignal:          make(chan struct{}),
		waitGroup:           new(sync.WaitGroup),
		votingPeriodChannel: make(chan types.VotingPeriod),
		paramsChannel:       make(chan types.Params, 1),
		params:              new(atomic.Pointer[types.Params]),
	}

	stream.waitGroup.Add(2)
	go stream.votePeriodLoop(ws, logger.With().Str("component", "vote-period-loop").Logger())
	go stream.paramsLoop(oracle, logger.With().Str("component", "params-loop").Logger())
	return stream
}

type Stream struct {
	stopSignal          chan struct{} // external signal to stop the stream
	waitGroup           *sync.WaitGroup
	votingPeriodChannel chan types.VotingPeriod
	paramsChannel       chan types.Params
	params              *atomic.Pointer[types.Params]
}

func (s *Stream) votePeriodLoop(ws wsI, logger zerolog.Logger) {
	defer func() {
		logger.Info().Msg("exited loop")
	}()
	defer s.waitGroup.Done()
	defer ws.close()

	for {
		select {
		case <-s.stopSignal:
			return
		case msg := <-ws.message():
			logger.Debug().Bytes("message", msg).Msg("received message from websocket")
			blockHeight, err := getBlockHeight(msg)
			if err != nil {
				logger.Err(err).Msg("whilst getting block height")
				break
			}
			if blockHeight == 0 {
				break
			}
			p := s.params.Load()
			if p == nil {
				break
			}
			if (blockHeight+1)%p.VotePeriodBlocks != 0 {
				break
			}

			logger.Debug().Msg("signaling new voting period")
			select {
			case <-s.stopSignal:
				logger.Warn().Uint64("height", blockHeight+1).Msg("dropped voting period signal")
			case s.votingPeriodChannel <- types.VotingPeriod{Height: blockHeight + 1}:
				logger.Debug().Msg("signaled new voting period")
			}
		}
	}
}

func (s *Stream) paramsLoop(oracleClient oracletypes.QueryClient, logger zerolog.Logger) {
	defer func() {
		logger.Info().Msg("exited loop")
	}()
	defer s.waitGroup.Done()

	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()

	updateParams := func() (types.Params, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		paramsResp, err := oracleClient.Params(ctx, &oracletypes.QueryParamsRequest{})
		if err != nil {
			return types.Params{}, err
		}

		return types.ParamsFromOracleParams(paramsResp.Params), nil
	}

	for {
		select {
		case <-tick.C:
			newParams, err := updateParams()
			if err != nil {
				logger.Err(err).Msg("param update")
				break
			}

			oldParams := s.params.Swap(&newParams)
			if oldParams != nil && oldParams.Equal(newParams) {
				logger.Debug().Msg("skipping params update as they're not different from the old ones")
				break
			}

			select {
			case <-s.stopSignal:
				logger.Warn().Msg("dropped params update due to shutdown")
			case s.paramsChannel <- newParams:
				logger.Info().Interface("params", newParams).Msg("signaling new params update")
			}

		case <-s.stopSignal:
			return
		}
	}
}

func (s *Stream) Close() {
	close(s.stopSignal)
	s.waitGroup.Wait()
}

func (s *Stream) ParamsUpdate() <-chan types.Params {
	return s.paramsChannel
}

func (s *Stream) VotingPeriodStarted() <-chan types.VotingPeriod {
	return s.votingPeriodChannel
}
