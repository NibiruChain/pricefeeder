package events

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/nibiru/simapp"
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	testutilcli "github.com/NibiruChain/nibiru/x/testutil/cli"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutilcli.Config
	network *testutilcli.Network

	eventsStream *Stream
	logs         *bytes.Buffer
	oracleClient oracletypes.QueryClient
}

func (s *IntegrationTestSuite) SetupSuite() {
	app.SetPrefixes(app.AccountAddressPrefix)
	s.cfg = testutilcli.BuildNetworkConfig(simapp.NewTestGenesisStateFromDefault())
	s.network = testutilcli.NewNetwork(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	require.NoError(s.T(), err)

	val := s.network.Validators[0]
	grpcEndpoint, tmEndpoint := val.AppConfig.GRPC.Address, val.RPCAddress
	u, err := url.Parse(tmEndpoint)
	require.NoError(s.T(), err)
	u.Scheme = "ws"
	u.Path = "/websocket"

	s.logs = new(bytes.Buffer)
	s.eventsStream = Dial(u.String(), grpcEndpoint, zerolog.New(s.logs))

	conn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	require.NoError(s.T(), err)
	s.oracleClient = oracletypes.NewQueryClient(conn)
}

func (s *IntegrationTestSuite) TestStreamWorks() {
	select {
	case <-s.eventsStream.ParamsUpdate():
	case <-time.After(15 * time.Second):
		s.T().Fatal("params timeout")
	}
	select {
	case <-s.eventsStream.VotingPeriodStarted():
	case <-time.After(15 * time.Second):
		s.T().Fatal("vote period timeout")
	}
	<-time.After(10 * time.Second)
	// assert params update
	require.Contains(s.T(), s.logs.String(), `"params":{"Pairs":[{"token0":"ubtc","token1":"unusd"},{"token0":"uusdc","token1":"unusd"},{"token0":"ueth","token1":"unusd"},{"token0":"unibi","token1":"unusd"}]`)
	// assert if params don't change, then no updates are sent
	require.Contains(s.T(), s.logs.String(), "skipping params update as they're not different from the old ones")
	// assert new voting period was signaled
	require.Contains(s.T(), s.logs.String(), "signaled new voting period")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.network.Cleanup()
	s.eventsStream.Close()
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
