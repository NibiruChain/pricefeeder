package feeder_test

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/nibiru/simapp"
	"github.com/NibiruChain/nibiru/x/common"
	testutilcli "github.com/NibiruChain/nibiru/x/testutil/cli"
	"github.com/NibiruChain/price-feeder/feeder"
	"github.com/NibiruChain/price-feeder/feeder/eventstream"
	"github.com/NibiruChain/price-feeder/feeder/priceposter"
	"github.com/NibiruChain/price-feeder/feeder/priceprovider"
	"github.com/NibiruChain/price-feeder/feeder/priceprovider/sources"
	"github.com/NibiruChain/price-feeder/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutilcli.Config
	network *testutilcli.Network

	feeder *feeder.Feeder
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
	u, err := url.Parse(tmEndpoint)
	require.NoError(s.T(), err)
	u.Scheme = "ws"
	u.Path = "/websocket"

	s.logs = new(bytes.Buffer)
	log := zerolog.New(io.MultiWriter(os.Stderr, s.logs)).Level(zerolog.InfoLevel)

	eventStream := eventstream.Dial(u.String(), grpcEndpoint, log)
	priceProvider := priceprovider.NewPriceProvider(sources.Bitfinex, map[common.AssetPair]types.Symbol{
		common.Pair_BTC_NUSD: "tBTCUSD",
		common.Pair_ETH_NUSD: "tETHUSD",
	}, log)
	pricePoster := priceposter.Dial(grpcEndpoint, s.cfg.ChainID, val.ClientCtx.Keyring, val.ValAddress, val.Address, log)
	s.feeder = feeder.NewFeeder(eventStream, priceProvider, pricePoster, log)
	s.feeder.Run()
}

func (s *IntegrationTestSuite) TestOk() {
	<-time.After(30 * time.Second) // TODO
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.network.Cleanup()
	s.feeder.Close()
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
