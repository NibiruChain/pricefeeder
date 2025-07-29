package types

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestGetEthereumRPCEndpoints(t *testing.T) {
	// Reset global state
	lastWorkingRPC = ""

	t.Run("uses ETHEREUM_RPC_ENDPOINT when set", func(t *testing.T) {
		url := "https://custom-endpoint.com"
		require.NoError(t, os.Setenv("ETHEREUM_RPC_ENDPOINT", url))
		defer func() {
			require.NoError(t, os.Unsetenv("ETHEREUM_RPC_ENDPOINT"))
		}()

		endpoints := GetEthereumRPCEndpoints()
		require.Len(t, endpoints, 1)
		require.Equal(t, url, endpoints[0])
	})

	t.Run("uses ETHEREUM_RPC_PUBLIC_ENDPOINTS when set", func(t *testing.T) {
		endpoints := "https://endpoint1.com,https://endpoint2.com"
		require.NoError(t, os.Setenv("ETHEREUM_RPC_PUBLIC_ENDPOINTS", endpoints))
		defer func() {
			require.NoError(t, os.Unsetenv("ETHEREUM_RPC_PUBLIC_ENDPOINTS"))
		}()

		result := GetEthereumRPCEndpoints()
		expected := []string{"https://endpoint1.com", "https://endpoint2.com"}
		require.Equal(t, expected, result)
	})

	t.Run("uses default endpoints when no env vars set", func(t *testing.T) {
		endpoints := GetEthereumRPCEndpoints()
		require.Len(t, endpoints, len(DefaultEthereumEndpoints))
		require.Equal(t, DefaultEthereumEndpoints[0], endpoints[0])
	})

	t.Run("moves last working RPC to front", func(t *testing.T) {
		lastWorkingRPC = "https://ethereum.publicnode.com"

		endpoints := GetEthereumRPCEndpoints()
		require.Equal(t, lastWorkingRPC, endpoints[0])
	})
}

func TestTryEthereumRPCEndpoint(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled) // Disable logging for tests

	t.Run("fails with invalid endpoint", func(t *testing.T) {
		client, err := TryEthereumRPCEndpoint("invalid-url", 1*time.Second, logger)
		require.Error(t, err)
		require.Nil(t, client)
	})

	t.Run("times out quickly", func(t *testing.T) {
		start := time.Now()
		client, err := TryEthereumRPCEndpoint("https://non-existent-endpoint-12345.com", 100*time.Millisecond, logger)
		duration := time.Since(start)

		require.Error(t, err)
		require.Nil(t, client)
		require.Less(t, duration, 2*time.Second, "timeout took too long")
	})

	// This test will only pass if you have internet connection
	t.Run("connects to real endpoint", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping network test in short mode")
		}

		endpoint := "https://cloudflare-eth.com/"
		client, err := TryEthereumRPCEndpoint(endpoint, 5*time.Second, logger)
		if err != nil {
			t.Logf("Network test failed (this is OK if no internet): %v", err)
			return
		}
		require.NotNil(t, client, "expected non-nil client for valid endpoint")
		defer client.Close()

		// Check that lastWorkingRPC was updated
		rpcMutex.RLock()
		currentLastRPC := lastWorkingRPC
		rpcMutex.RUnlock()

		require.Equal(t, endpoint, currentLastRPC, "lastWorkingRPC should be updated")
	})
}
