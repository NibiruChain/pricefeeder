package feeder_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/v2/app"
	"github.com/NibiruChain/nibiru/v2/gosdk"
	"github.com/NibiruChain/nibiru/v2/x/common/asset"
	"github.com/NibiruChain/nibiru/v2/x/common/denoms"
	"github.com/NibiruChain/nibiru/v2/x/common/testutil/genesis"
	"github.com/NibiruChain/nibiru/v2/x/common/testutil/testnetwork"
	oracletypes "github.com/NibiruChain/nibiru/v2/x/oracle/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/NibiruChain/pricefeeder/feeder"
	"github.com/NibiruChain/pricefeeder/feeder/priceprovider"
	"github.com/NibiruChain/pricefeeder/sources"
	"github.com/NibiruChain/pricefeeder/types"
)

type IntegrationSuite struct {
	suite.Suite

	cfg     testnetwork.Config
	network *testnetwork.Network

	feeder            *feeder.Feeder
	pricePosterClient *feeder.ClientPricePoster // TODO: refactor to use field from `Feeder`.

	// testEventStream is a separate eventStream instance used for testing.
	// It exists separately from feeder.EventStream because the feeder's eventStream
	// uses unbuffered channels that can only be consumed by one receiver. Since the
	// feeder's loop() goroutine is already consuming from feeder.EventStream's channels
	// (ParamsUpdate() and VotingPeriodStarted()), the test cannot read from the same
	// channels. By creating a separate testEventStream, the test can independently
	// receive params updates and voting period signals without interfering with the
	// feeder's operation.
	testEventStream  *feeder.Stream
	logs             *bytes.Buffer
	oracleClient     oracletypes.QueryClient
}

func (s *IntegrationSuite) SetupSuite() {
	gosdk.EnsureNibiruPrefix()

	s.cfg = testnetwork.BuildNetworkConfig(
		genesis.NewTestGenesisState(
			app.MakeEncodingConfig().Codec))
	network, err := testnetwork.New(
		s.T(),
		s.T().TempDir(),
		s.cfg,
	)
	s.Require().NoError(err)
	s.network = network

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]
	grpcEndpoint, tmEndpoint := val.AppConfig.GRPC.Address, val.RPCAddress

	s.T().Logf("Set up websocket { tmEndpoint: %s }", tmEndpoint)
	wsUrl, err := url.Parse(tmEndpoint)
	require.NoError(s.T(), err)
	wsUrl.Scheme = "ws"
	wsUrl.Path = "/websocket"

	s.logs = new(bytes.Buffer)
	logger := zerolog.
		New(io.MultiWriter(zerolog.NewTestWriter(s.T()), s.logs)).
		Level(zerolog.InfoLevel)

	enableTLS := false
	s.feeder = feeder.NewFeeder(
		feeder.DialEventStream(
			wsUrl.String(),
			grpcEndpoint,
			enableTLS,
			logger,
		),
		priceprovider.NewPriceProvider(
			sources.SourceBitfinex,
			map[asset.Pair]types.Symbol{
				asset.Registry.Pair(denoms.BTC, denoms.NUSD): "tBTCUSD",
				asset.Registry.Pair(denoms.ETH, denoms.NUSD): "tETHUSD",
			},
			json.RawMessage{},
			logger,
		),
		feeder.DialPricePoster(
			grpcEndpoint,
			s.cfg.ChainID,
			enableTLS,
			val.ClientCtx.Keyring, val.ValAddress, val.Address, logger),
		logger,
	)
	s.feeder.Run()
	s.pricePosterClient = s.feeder.PricePoster.(*feeder.ClientPricePoster)

	s.T().Log("Set up price posing client for direct testing") // (shares the same logs)

	s.T().Log("Set up x/oracle module query client")
	conn, err := grpc.Dial(grpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(s.T(), err)
	s.oracleClient = oracletypes.NewQueryClient(conn)

	// Set up a separate eventStream for testing (not consumed by feeder)
	// Use Debug level logger so we can see "skipping params update" messages
	s.T().Log("Set up separate eventStream for testing")
	testLogger := zerolog.
		New(io.MultiWriter(zerolog.NewTestWriter(s.T()), s.logs)).
		Level(zerolog.DebugLevel)
	s.testEventStream = feeder.DialEventStream(
		wsUrl.String(),
		grpcEndpoint,
		enableTLS,
		testLogger,
	)
}

func (s *IntegrationSuite) TestOk() {
	<-time.After(30 * time.Second) // TODO
}

func (s *IntegrationSuite) TearDownSuite() {
	if s.testEventStream != nil {
		s.testEventStream.Close()
	}
	if s.feeder != nil {
		s.feeder.Close()
	}
	s.network.Cleanup()
}

func TestFeederIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}
