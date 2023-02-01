package eventstream

import (
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	"github.com/NibiruChain/price-feeder/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
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
	}
}

type ParamsStream struct {
	logger       zerolog.Logger
	oracleClient oracletypes.QueryClient
}

func (p ParamsStream) ParamsUpdate() <-chan types.Params {
	//TODO implement me
	panic("implement me")
}

func (p ParamsStream) Close() {
	//TODO implement me
	panic("implement me")
}
