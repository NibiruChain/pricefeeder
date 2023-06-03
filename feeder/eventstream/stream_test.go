package eventstream

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/app"
	testutilcli "github.com/NibiruChain/nibiru/x/common/testutil/cli"
	"github.com/NibiruChain/nibiru/x/common/testutil/genesis"
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutilcli.Config
	network *testutilcli.Network

	eventStream  *Stream
	logs         *bytes.Buffer
	oracleClient oracletypes.QueryClient
}

func (s *IntegrationTestSuite) SetupSuite() {
	app.SetPrefixes(app.AccountAddressPrefix)
	s.cfg = testutilcli.BuildNetworkConfig(genesis.NewTestGenesisState())
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
	enableTLS := false
	s.eventStream = Dial(
		u.String(),
		grpcEndpoint,
		enableTLS,
		zerolog.New(s.logs))

	conn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	require.NoError(s.T(), err)
	s.oracleClient = oracletypes.NewQueryClient(conn)
}

func (s *IntegrationTestSuite) TestStreamWorks() {
	select {
	case <-s.eventStream.ParamsUpdate():
	case <-time.After(15 * time.Second):
		s.T().Fatal("params timeout")
	}
	select {
	case <-s.eventStream.VotingPeriodStarted():
	case <-time.After(15 * time.Second):
		s.T().Fatal("vote period timeout")
	}
	<-time.After(10 * time.Second)
	// assert if params don't change, then no updates are sent
	require.Contains(s.T(), s.logs.String(), "skipping params update as they're not different from the old ones")
	// assert new voting period was signaled
	require.Contains(s.T(), s.logs.String(), "signaled new voting period")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.eventStream.Close()
	s.network.Cleanup()
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
