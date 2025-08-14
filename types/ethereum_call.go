package types

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

// DefaultEthereumEndpoints copied from https://ethereumnodes.com/
var DefaultEthereumEndpoints = []string{
	"https://eth.llamarpc.com",
	"https://eth-mainnet.public.blastapi.io",
	"https://rpc.flashbots.net/",
	"https://cloudflare-eth.com/",
	"https://ethereum.publicnode.com",
}

// Global variables to track the last working RPC endpoint
var (
	lastWorkingRPC string
	rpcMutex       sync.RWMutex
)

// GetEthereumRPCEndpoints returns the list of RPC endpoints to try
func GetEthereumRPCEndpoints() []string {
	// Check if ETHEREUM_RPC_ENDPOINT is set (priority endpoint)
	if endpoint := os.Getenv("ETHEREUM_RPC_ENDPOINT"); endpoint != "" {
		return []string{endpoint}
	}

	// Get public endpoints from environment or use defaults
	publicEndpoints := os.Getenv("ETHEREUM_RPC_PUBLIC_ENDPOINTS")
	var endpoints []string
	if publicEndpoints != "" {
		endpoints = strings.Split(publicEndpoints, ",")
	} else {
		// Default public RPC endpoints
		endpoints = DefaultEthereumEndpoints
	}
	// Trim whitespace from each endpoint
	for i, endpoint := range endpoints {
		endpoints[i] = strings.TrimSpace(endpoint)
	}

	// If we have a last working RPC, move it to the front
	rpcMutex.RLock()
	lastRPC := lastWorkingRPC
	rpcMutex.RUnlock()

	if lastRPC != "" {
		// Find and move the last working RPC to the front
		for i, endpoint := range endpoints {
			if endpoint == lastRPC {
				// Move last working rpc to front
				endpoints = append([]string{endpoint}, append(endpoints[:i], endpoints[i+1:]...)...)
				break
			}
		}
	}
	return endpoints
}

// TryEthereumRPCEndpoint attempts to connect to a single RPC endpoint with timeout
func TryEthereumRPCEndpoint(endpoint string, timeout time.Duration, logger zerolog.Logger) (*ethclient.Client, error) {
	logger.Debug().Str("endpoint", endpoint).Msg("trying RPC endpoint")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a channel to receive the result
	type result struct {
		client *ethclient.Client
		err    error
	}
	resultCh := make(chan result, 1)

	// Try to connect in a goroutine
	go func() {
		client, err := ethclient.Dial(endpoint)
		if err != nil {
			resultCh <- result{nil, err}
			return
		}

		// Test the connection by getting the latest block number
		_, err = client.BlockNumber(context.Background())
		if err != nil {
			client.Close()
			resultCh <- result{nil, err}
			return
		}
		resultCh <- result{client, nil}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultCh:
		if res.err == nil {
			logger.Debug().Str("endpoint", endpoint).Msg("successfully connected to RPC endpoint")

			// Update last working RPC
			rpcMutex.Lock()
			lastWorkingRPC = endpoint
			rpcMutex.Unlock()
		}
		return res.client, res.err
	case <-ctx.Done():
		return nil, fmt.Errorf("connection timeout after %v", timeout)
	}
}
