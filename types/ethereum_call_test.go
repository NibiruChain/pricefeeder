package types

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectToB2(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	timeout := 10 * time.Second

	client, err := ConnectToB2(timeout, logger)

	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Test that we can get block number
	blockNumber, err := client.BlockNumber(context.Background())
	require.NoError(t, err)
	assert.Greater(t, blockNumber, uint64(0))
}

func TestConnectToEthereum(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	timeout := 10 * time.Second

	client, err := ConnectToEthereum(timeout, logger)

	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Test that we can get block number
	blockNumber, err := client.BlockNumber(context.Background())
	require.NoError(t, err)
	assert.Greater(t, blockNumber, uint64(0))
}

func TestConnectToBase(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	timeout := 10 * time.Second

	client, err := ConnectToBase(timeout, logger)

	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	blockNumber, err := client.BlockNumber(context.Background())
	require.NoError(t, err)
	assert.Greater(t, blockNumber, uint64(0))
}

func TestGetRPCEndpoints(t *testing.T) {
	endpoints, err := GetRPCEndpoints("b2")

	require.NoError(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Contains(t, endpoints, "https://rpc.bsquared.network")
}

func TestGetRPCEndpoints_UnsupportedNetwork(t *testing.T) {
	endpoints, err := GetRPCEndpoints("unknown")

	require.Error(t, err)
	assert.Nil(t, endpoints)
	assert.Contains(t, err.Error(), "unsupported network")
}
