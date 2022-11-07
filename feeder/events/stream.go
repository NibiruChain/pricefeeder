package events

import (
	"context"
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	"github.com/NibiruChain/price-feeder/feeder/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"sync"
	"sync/atomic"
	"time"
)

// wsI exists for testing purposes.
type wsI interface {
	message() <-chan []byte
	close()
}

type Stream struct {
	stop chan struct{}
	wg   *sync.WaitGroup

	signalVotingPeriod chan types.VotingPeriod
	signalParams       chan types.Params

	params *atomic.Pointer[types.Params]
}

func (s *Stream) votePeriodLoop(ws wsI, log zerolog.Logger) {
	defer func() {
		log.Info().Msg("exited loop")
	}()
	defer s.wg.Done()
	defer ws.close()

	for {
		select {
		case <-s.stop:
			return
		case msg := <-ws.message():
			log.Debug().Bytes("message", msg).Msg("recv")
			blockHeight, err := getBlockHeight(msg)
			if err != nil {
				log.Err(err).Msg("whilst getting block height")
				break
			}
			if blockHeight == 0 {
				break
			}
			p := s.params.Load()
			if blockHeight+1%p.VotePeriodBlocks != 0 {
				break
			}

			select {
			case <-s.stop:
				log.Warn().Uint64("height", blockHeight).Msg("dropped voting period signal")
			case s.signalVotingPeriod <- types.VotingPeriod{Height: blockHeight}:
			}
		}
	}
}

func (s *Stream) paramsLoop(c oracletypes.QueryClient, log zerolog.Logger) {
	defer func() {
		log.Info().Msg("exited loop")
	}()
	defer s.wg.Done()

	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()

	updateParams := func() (types.Params, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		paramsResp, err := c.Params(ctx, &oracletypes.QueryParamsRequest{})
		if err != nil {
			return types.Params{}, err
		}

		panic("todo")
		return types.Params{
			Pairs:            nil,
			VotePeriodBlocks: paramsResp.Params.VotePeriod,
		}, nil
	}

	for {
		select {
		case <-tick.C:
			newParams, err := updateParams()
			if err != nil {
				log.Err(err).Msg("param update")
				break
			}

			oldParams := s.params.Swap(&newParams)
			if oldParams.Equal(newParams) {
				break
			}
			select {
			case <-s.stop:
				log.Warn().Msg("dropped params update due to shutdown")
			case s.signalParams <- newParams:

			}
		case <-s.stop:
			return
		}
	}
}

func (s *Stream) Close() {
	close(s.stop)
	s.wg.Wait()
}

func NewStream(tendermintRPCEndpoint string, grpcEndpoint string, log zerolog.Logger) *Stream {
	grpcConn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	oracle := oracletypes.NewQueryClient(grpcConn)

	const newBlockSubscribe = `{"jsonrpc":"2.0","method":"subscribe","id":0,"params":{"query":"tm.event='NewBlock'"}}`
	ws := dial(tendermintRPCEndpoint, []byte(newBlockSubscribe), log)
	return newStream(ws, oracle, log)
}

func newStream(ws wsI, oracle oracletypes.QueryClient, log zerolog.Logger) *Stream {
	stream := &Stream{
		stop:   make(chan struct{}),
		wg:     new(sync.WaitGroup),
		params: new(atomic.Pointer[types.Params]),
	}

	log = log.With().Str("component", "events.Stream").Logger()
	stream.wg.Add(2)
	go stream.votePeriodLoop(ws, log.With().Str("loop", "vote-period").Logger())
	go stream.paramsLoop(oracle, log.With().Str("loop", "params").Logger())
	return stream
}
