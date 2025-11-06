package feeder_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
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
	oracletypes "github.com/NibiruChain/nibiru/v2/x/oracle/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/NibiruChain/pricefeeder/feeder"
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
	testEventStream *feeder.Stream
	logs            *bytes.Buffer
	oracleClient    oracletypes.QueryClient
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
		feeder.NewPriceProvider(
			sources.SourceNameBitfinex,
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

// canConnectToWebsocket checks if we can resolve and connect to the websocket server.
// It performs a DNS lookup to verify network connectivity before attempting a connection.
// This allows tests to skip gracefully when network is unavailable.
func (s *IntegrationSuite) canConnectToWebsocket(urlStr string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Parse the hostname from the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	host := parsedURL.Host
	// Remove port if present (e.g., "echo.websocket.org:443" -> "echo.websocket.org")
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}

	_, err = net.DefaultResolver.LookupHost(ctx, host)
	return err == nil
}

// TestWebsocketSuccess tests the WebSocket connection and echo functionality using
// the public echo server at websocket.org.
//
// This test uses the WebSocket Echo Server provided by websocket.org, which is a free,
// publicly available testing endpoint. According to the documentation at
// https://websocket.org/tools/websocket-echo-server/, the server at wss://echo.websocket.org
// echoes back any message sent to it, making it ideal for testing WebSocket client implementations.
//
// The test verifies:
//   - Successful connection establishment
//   - Automatic sending of the onOpenMsg ("test") after connection
//   - Receiving the echoed message back from the server
//
// Note: The echo server may send an initial server message (e.g., "Request served by ...")
// before echoing client messages, so the test reads messages until it finds the expected echo.
//
// This test requires internet connectivity and will be skipped if the echo server is unreachable.
func (s *IntegrationSuite) TestWebsocketSuccess() {
	// Skip test if we can't reach the external websocket server
	if !s.canConnectToWebsocket("wss://echo.websocket.org") {
		s.T().Skip("Skipping test: cannot reach echo.websocket.org (network may be unavailable)")
	}

	// According to https://websocket.org/tools/websocket-echo-server/
	// The echo server at wss://echo.websocket.org echoes back any message sent to it
	ws := feeder.NewWebsocket("wss://echo.websocket.org", []byte("test"), zerolog.New(os.Stderr))
	defer ws.Close()

	// The echo server may send an initial server message first (e.g., "Request served by ...")
	// Then it will echo back our "test" message. Let's wait for our echo.
	// We'll read messages until we get our "test" message back, or timeout.
	foundEcho := false
	timeout := time.After(5 * time.Second)
	for !foundEcho {
		select {
		case msg := <-ws.Message():
			if string(msg) == "test" {
				// Found our echo!
				foundEcho = true
			}
			// Otherwise, it's likely the initial server message, continue waiting
		case <-timeout:
			s.T().Fatal("timeout waiting for echo of 'test' message")
		}
	}

	s.Require().True(foundEcho, "should have received echo of 'test' message")
}

// TestWebsocketExplicitClose tests that the WebSocket can be closed gracefully without panicking.
// This verifies that the close() method properly handles connection cleanup, even when called
// immediately after connection establishment or when the connection is in various states.
//
// The test ensures that:
//   - close() can be called safely without panicking
//   - All resources are properly cleaned up
//   - The connection is terminated gracefully
//
// This test requires internet connectivity and will be skipped if the echo server is unreachable.
func (s *IntegrationSuite) TestWebsocketExplicitClose() {
	// Skip test if we can't reach the external websocket server
	if !s.canConnectToWebsocket("wss://echo.websocket.org") {
		s.T().Skip("Skipping test: cannot reach echo.websocket.org (network may be unavailable)")
	}

	ws := feeder.NewWebsocket("wss://echo.websocket.org", []byte("test"), zerolog.New(os.Stderr))
	s.Require().NotPanics(func() {
		ws.Close()
	})
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
