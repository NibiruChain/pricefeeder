package feeder_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/v2/app"
	"github.com/NibiruChain/nibiru/v2/gosdk"
	"github.com/NibiruChain/nibiru/v2/x/common/asset"
	"github.com/NibiruChain/nibiru/v2/x/common/denoms"
	"github.com/NibiruChain/nibiru/v2/x/common/testutil/genesis"
	"github.com/NibiruChain/nibiru/v2/x/common/testutil/testnetwork"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/NibiruChain/pricefeeder/feeder"
	"github.com/NibiruChain/pricefeeder/feeder/eventstream"
	"github.com/NibiruChain/pricefeeder/feeder/priceprovider"
	"github.com/NibiruChain/pricefeeder/sources"
	"github.com/NibiruChain/pricefeeder/types"
)

type IntegrationSuite struct {
	suite.Suite

	cfg               testnetwork.Config
	network           *testnetwork.Network
	feeder            *feeder.Feeder
	pricePosterClient *feeder.ClientPricePoster
	logs              *bytes.Buffer
}

func (s *IntegrationSuite) SetupSuite() {
	gosdk.EnsureNibiruPrefix()
	s.cfg = testnetwork.BuildNetworkConfig(genesis.NewTestGenesisState(app.MakeEncodingConfig().Codec))
	network, err := testnetwork.New(
		s.T(),
		s.T().TempDir(),
		s.cfg,
	)
	s.Require().NoError(err)
	s.network = network

	_, err = s.network.WaitForHeight(1)
	require.NoError(s.T(), err)

	val := s.network.Validators[0]
	grpcEndpoint, tmEndpoint := val.AppConfig.GRPC.Address, val.RPCAddress
	u, err := url.Parse(tmEndpoint)
	require.NoError(s.T(), err)
	u.Scheme = "ws"
	u.Path = "/websocket"

	s.logs = new(bytes.Buffer)
	log := zerolog.New(io.MultiWriter(os.Stderr, s.logs)).Level(zerolog.InfoLevel)

	enableTLS := false
	eventStream := eventstream.DialEventStream(u.String(), grpcEndpoint, enableTLS, log)
	priceProvider := priceprovider.NewPriceProvider(sources.SourceBitfinex, map[asset.Pair]types.Symbol{
		asset.Registry.Pair(denoms.BTC, denoms.NUSD): "tBTCUSD",
		asset.Registry.Pair(denoms.ETH, denoms.NUSD): "tETHUSD",
	}, json.RawMessage{}, log)
	pricePoster := feeder.DialPricePoster(
		grpcEndpoint,
		s.cfg.ChainID,
		enableTLS,
		val.ClientCtx.Keyring, val.ValAddress, val.Address, log)
	s.feeder = feeder.NewFeeder(eventStream, priceProvider, pricePoster, log)
	s.feeder.Run()

	// Set up a separate pricePoster client for direct testing (shares the same logs)
	s.pricePosterClient = feeder.DialPricePoster(
		grpcEndpoint,
		s.cfg.ChainID,
		enableTLS,
		val.ClientCtx.Keyring, val.ValAddress, val.Address, log)
}

func (s *IntegrationSuite) TestOk() {
	<-time.After(30 * time.Second) // TODO
}

func (s *IntegrationSuite) TearDownSuite() {
	if s.feeder != nil {
		s.feeder.Close()
	}
	if s.pricePosterClient != nil {
		s.pricePosterClient.Close()
	}
	s.network.Cleanup()
}

func TestFeederIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}
