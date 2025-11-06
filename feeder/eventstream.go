package feeder

import (
	"context"
	"crypto/tls"
	"sync"
	"sync/atomic"
	"time"

	oracletypes "github.com/NibiruChain/nibiru/v2/x/oracle/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/NibiruChain/pricefeeder/types"
)

var _ types.EventStream = (*Stream)(nil)

// wsI exists for testing purposes.
type wsI interface {
	Message() <-chan []byte
	Close()
}

// DialEventStream opens two connections to the given endpoint, one for the websocket and one for the oracle grpc.
func DialEventStream(tendermintRPCEndpoint string, grpcEndpoint string, enableTLS bool, logger zerolog.Logger) *Stream {
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

	oracleClient := oracletypes.NewQueryClient(conn)

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

	go stream.votingPeriodStartedLoop(ws, logger.With().Str("component", "voting-period-started-loop").Logger())
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

func (s *Stream) votingPeriodStartedLoop(ws wsI, logger zerolog.Logger) {
	defer func() {
		logger.Info().Msg("exited loop")
		s.waitGroup.Done()
		ws.Close()
	}()

	for {
		select {
		case <-s.stopSignal:
			return
		case msg := <-ws.Message():
			logger.Debug().Bytes("payload", msg).Msg("received message from websocket")
			blockHeight, err := types.GetBlockHeight(msg)
			if err != nil {
				logger.Err(err).Msg("could not obtain block height")
				break
			}
			if blockHeight <= 0 {
				logger.Err(err).Uint64("block-height", blockHeight).Msg("invalid block height")
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

// paramsLoop calls every 10 seconds the oracle grpc to obtain the current params as a way to keep the params up to date.
func (s *Stream) paramsLoop(oracleClient oracletypes.QueryClient, logger zerolog.Logger) {
	tick := time.NewTicker(10 * time.Second)
	defer func() {
		logger.Info().Msg("exited loop")
		s.waitGroup.Done()
		tick.Stop()
	}()

	fetchParams := func() (types.Params, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
			newParams, err := fetchParams()
			if err != nil {
				logger.Err(err).Msg("param update failed")
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
