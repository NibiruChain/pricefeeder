package priceposter

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/nibiru/simapp"
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	testutilcli "github.com/NibiruChain/nibiru/x/testutil/cli"
	"github.com/NibiruChain/price-feeder/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutilcli.Config
	network *testutilcli.Network

	client *Client
	logs   *bytes.Buffer
}

func (s *IntegrationTestSuite) SetupSuite() {
	app.SetPrefixes(app.AccountAddressPrefix)
	s.cfg = testutilcli.BuildNetworkConfig(simapp.NewTestGenesisStateFromDefault())
	s.network = testutilcli.NewNetwork(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	require.NoError(s.T(), err)

	val := s.network.Validators[0]
	grpcEndpoint, tmEndpoint := val.AppConfig.GRPC.Address, val.RPCAddress
	url, err := url.Parse(tmEndpoint)
	require.NoError(s.T(), err)

	url.Scheme = "ws"
	url.Path = "/websocket"

	s.logs = new(bytes.Buffer)
	s.client = Dial(grpcEndpoint, s.cfg.ChainID, val.ClientCtx.Keyring, val.ValAddress, val.Address, zerolog.New(io.MultiWriter(os.Stderr, s.logs)))
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.network.Cleanup()
	s.client.Close()
}

func (s *IntegrationTestSuite) TestClientWorks() {
	s.client.SendPrices(types.VotingPeriod{}, s.randomPrices())

	// assert vote was skipped because no previous prevote
	require.Contains(s.T(), s.logs.String(), "skipping vote preparation as there is no old prevote")
	require.NotContains(s.T(), s.logs.String(), "prepared vote message")

	// wait for next vote period
	s.waitNextVotePeriod()
	s.client.SendPrices(types.VotingPeriod{}, s.randomPrices())
	require.Contains(s.T(), s.logs.String(), "prepared vote message")
}

func (s *IntegrationTestSuite) randomPrices() []types.Price {
	vt, err := s.client.deps.oracleClient.(oracletypes.QueryClient).VoteTargets(context.Background(), &oracletypes.QueryVoteTargetsRequest{})
	require.NoError(s.T(), err)
	prices := make([]types.Price, len(vt.VoteTargets))
	for i, assetPair := range vt.VoteTargets {
		prices[i] = types.Price{
			Pair:       assetPair,
			Price:      float64(i),
			SourceName: "test",
			Valid:      true,
		}
	}
	return prices
}

func (s *IntegrationTestSuite) waitNextVotePeriod() {
	params, err := s.client.deps.oracleClient.(oracletypes.QueryClient).Params(context.Background(), &oracletypes.QueryParamsRequest{})
	require.NoError(s.T(), err)
	height, err := s.network.LatestHeight()
	require.NoError(s.T(), err)
	targetHeight := height + int64(uint64(height)%params.Params.VotePeriod)
	_, err = s.network.WaitForHeight(targetHeight)
	require.NoError(s.T(), err)
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
