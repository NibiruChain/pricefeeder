package eventstream

import (
	"context"
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	"github.com/NibiruChain/price-feeder/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

// DialParamsStream opens a connection to the given endpoint for the oracle grpc.
func DialParamsStream(grpcEndpoint string, logger zerolog.Logger) ParamsStream {
	grpcConn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	oracleClient := oracletypes.NewQueryClient(grpcConn)

	return ParamsStream{
		logger:       logger,
		oracleClient: oracleClient,
		params:       new(atomic.Pointer[types.Params]),
	}
}

type ParamsStream struct {
	logger       zerolog.Logger
	oracleClient oracletypes.QueryClient

	paramsChannel chan types.Params
	stopSignal    chan struct{}

	params *atomic.Pointer[types.Params]
}

// paramsLoop calls every 10 seconds the oracle grpc to obtain the current params as a way to keep the params up to date.
func (p ParamsStream) paramsLoop(oracleClient oracletypes.QueryClient, logger zerolog.Logger) {
	tick := time.NewTicker(10 * time.Second)
	defer func() {
		logger.Info().Msg("exited loop")
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
				logger.Err(err).Msg("param update")
				break
			}

			oldParams := p.params.Swap(&newParams)
			if oldParams != nil && oldParams.Equal(newParams) {
				logger.Debug().Msg("skipping params update as they're not different from the old ones")
				break
			}

			select {
			case <-p.stopSignal:
				logger.Warn().Msg("dropped params update due to shutdown")
			case p.paramsChannel <- newParams:
				logger.Info().Interface("params", newParams).Msg("signaling new params update")
			}

		case <-p.stopSignal:
			return
		}
	}
}

func (p ParamsStream) Close() {
	close(p.stopSignal)
}

func (p ParamsStream) ParamsUpdate() <-chan types.Params {
	return p.paramsChannel
}
