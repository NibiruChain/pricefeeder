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

// NetworkConfig holds configuration for an EVM network
type NetworkConfig struct {
	Name               string
	DefaultEndpoints   []string
	EnvEndpoint        string // Environment variable for single endpoint
	EnvPublicEndpoints string // Environment variable for multiple endpoints
}

// DefaultNetworkConfigs provides configurations for supported networks
var DefaultNetworkConfigs = map[string]NetworkConfig{
	"ethereum": {
		Name: "ethereum",
		DefaultEndpoints: []string{
			"https://eth.llamarpc.com",
			"https://eth-mainnet.public.blastapi.io",
			"https://rpc.flashbots.net/",
			"https://cloudflare-eth.com/",
			"https://ethereum.publicnode.com",
		},
		EnvEndpoint:        "ETHEREUM_RPC_ENDPOINT",
		EnvPublicEndpoints: "ETHEREUM_RPC_PUBLIC_ENDPOINTS",
	},
	"b2": {
		Name: "b2",
		DefaultEndpoints: []string{
			"https://rpc.bsquared.network",
			"https://mainnet.b2-rpc.com",
			"https://rpc.ankr.com/b2",
			"https://b2-mainnet.alt.technology",
		},
		EnvEndpoint:        "B2_RPC_ENDPOINT",
		EnvPublicEndpoints: "B2_RPC_PUBLIC_ENDPOINTS",
	},
}

// Global variables to track the last working RPC endpoint per network
var (
	lastWorkingRPCs = make(map[string]string)
	rpcMutex        sync.RWMutex
)

// GetRPCEndpoints returns the list of RPC endpoints to try for a given network
func GetRPCEndpoints(networkName string) ([]string, error) {
	config, exists := DefaultNetworkConfigs[networkName]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", networkName)
	}

	// Check if priority endpoint is set
	if endpoint := os.Getenv(config.EnvEndpoint); endpoint != "" {
		return []string{endpoint}, nil
	}

	// Get public endpoints from environment or use defaults
	publicEndpoints := os.Getenv(config.EnvPublicEndpoints)
	var endpoints []string
	if publicEndpoints != "" {
		endpoints = strings.Split(publicEndpoints, ",")
	} else {
		endpoints = config.DefaultEndpoints
	}

	// Trim whitespace from each endpoint
	for i, endpoint := range endpoints {
		endpoints[i] = strings.TrimSpace(endpoint)
	}

	// If we have a last working RPC for this network, move it to the front
	rpcMutex.RLock()
	lastRPC := lastWorkingRPCs[networkName]
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
	return endpoints, nil
}

// TryRPCEndpoint attempts to connect to a single RPC endpoint with timeout
func TryRPCEndpoint(networkName, endpoint string, timeout time.Duration, logger zerolog.Logger) (*ethclient.Client, error) {
	logger.Debug().
		Str("network", networkName).
		Str("endpoint", endpoint).
		Msg("trying RPC endpoint")

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
			logger.Debug().
				Str("network", networkName).
				Str("endpoint", endpoint).
				Msg("successfully connected to RPC endpoint")

			// Update last working RPC for this network
			rpcMutex.Lock()
			lastWorkingRPCs[networkName] = endpoint
			rpcMutex.Unlock()
		}
		return res.client, res.err
	case <-ctx.Done():
		return nil, fmt.Errorf("connection timeout after %v", timeout)
	}
}

// ConnectToNetwork creates a connection to the specified EVM network
func ConnectToNetwork(networkName string, timeout time.Duration, logger zerolog.Logger) (*ethclient.Client, error) {
	config, exists := DefaultNetworkConfigs[networkName]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", networkName)
	}

	endpoints, err := GetRPCEndpoints(networkName)
	if err != nil {
		return nil, err
	}

	var client *ethclient.Client
	var connErr error

	// If priority endpoint is set, use it exclusively
	if os.Getenv(config.EnvEndpoint) != "" {
		client, connErr = ethclient.Dial(endpoints[0])
		if connErr != nil {
			return nil, fmt.Errorf("failed to connect to %s client: %w", networkName, connErr)
		}
		return client, nil
	}

	// Try each endpoint with the specified timeout
	for _, endpoint := range endpoints {
		client, connErr = TryRPCEndpoint(networkName, endpoint, timeout, logger)
		if connErr == nil {
			return client, nil
		}
		logger.Warn().
			Str("network", networkName).
			Str("endpoint", endpoint).
			Err(connErr).
			Msg("failed to connect to RPC endpoint, trying next")
	}

	return nil, fmt.Errorf("failed to connect to any %s RPC endpoint. Last error: %w", networkName, connErr)
}

func TryEthereumRPCEndpoint(endpoint string, timeout time.Duration, logger zerolog.Logger) (*ethclient.Client, error) {
	return TryRPCEndpoint("ethereum", endpoint, timeout, logger)
}

func ConnectToEthereum(timeout time.Duration, logger zerolog.Logger) (*ethclient.Client, error) {
	return ConnectToNetwork("ethereum", timeout, logger)
}

func ConnectToB2(timeout time.Duration, logger zerolog.Logger) (*ethclient.Client, error) {
	return ConnectToNetwork("b2", timeout, logger)
}
