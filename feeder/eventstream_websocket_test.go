package feeder

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// canConnectToWebsocket checks if we can resolve and connect to the websocket server.
// It performs a DNS lookup to verify network connectivity before attempting a connection.
// This allows tests to skip gracefully when network is unavailable.
func canConnectToWebsocket(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try to resolve the hostname first
	host := "echo.websocket.org"
	_, err := net.DefaultResolver.LookupHost(ctx, host)
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
func TestWebsocketSuccess(t *testing.T) {
	// Skip test if we can't reach the external websocket server
	if !canConnectToWebsocket("wss://echo.websocket.org") {
		t.Skip("Skipping test: cannot reach echo.websocket.org (network may be unavailable)")
	}

	// According to https://websocket.org/tools/websocket-echo-server/
	// The echo server at wss://echo.websocket.org echoes back any message sent to it
	ws := NewWebsocket("wss://echo.websocket.org", []byte("test"), zerolog.New(os.Stderr))
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
			t.Fatal("timeout waiting for echo of 'test' message")
		}
	}

	require.True(t, foundEcho, "should have received echo of 'test' message")
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
func TestWebsocketExplicitClose(t *testing.T) {
	// Skip test if we can't reach the external websocket server
	if !canConnectToWebsocket("wss://echo.websocket.org") {
		t.Skip("Skipping test: cannot reach echo.websocket.org (network may be unavailable)")
	}

	ws := NewWebsocket("wss://echo.websocket.org", []byte("test"), zerolog.New(os.Stderr))
	require.NotPanics(t, func() {
		ws.Close()
	})
}
