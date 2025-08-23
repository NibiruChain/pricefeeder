package sources

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"

	"github.com/NibiruChain/pricefeeder/types"
	"github.com/NibiruChain/pricefeeder/types/chainlink"
)

var _ types.FetchPricesFunc = ChainlinkPriceUpdate

const (
	SourceChainLink = "chainlink"
)

// ChainType represents different blockchain networks
type ChainType string

const (
	ChainB2 ChainType = "b2"
)

// ChainlinkConfig represents configuration for a specific Chainlink oracle
type ChainlinkConfig struct {
	Chain           ChainType
	ContractAddress common.Address
	Description     string        // Expected description (for sanity check)
	MaxDataAge      time.Duration // Maximum acceptable data age
}

// chainlinkConfigMap maps trading pair symbols to their Chainlink oracle configurations
var chainlinkConfigMap = map[types.Symbol]ChainlinkConfig{
	"uBTC/BTC": {
		Chain:           ChainB2,
		ContractAddress: common.HexToAddress("0xA2ed2B84073B3BA4F3Bd6528260d85EdDFD72fF2"),
		Description:     "uBTC/BTC Exchange Rate",
		MaxDataAge:      0, // No age limit for this example
	},
}

// chainConnectors maps chain types to their connection functions
var chainConnectors = map[ChainType]func(time.Duration, zerolog.Logger) (*ethclient.Client, error){
	ChainB2: types.ConnectToB2,
}

// ChainlinkPriceUpdate retrieves exchange rates from various Chainlink oracles across different chains
func ChainlinkPriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (map[types.Symbol]float64, error) {
	timeout := 10 * time.Second
	prices := make(map[types.Symbol]float64)

	// Group symbols by chain to optimize connections
	symbolsByChain := groupSymbolsByChain(symbols, logger)

	// Process each chain separately
	for chain, chainSymbols := range symbolsByChain {
		chainPrices, err := fetchPricesFromChain(chain, chainSymbols, timeout, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch prices from %s: %w", chain, err)
		}

		// Merge results
		for symbol, price := range chainPrices {
			prices[symbol] = price
		}
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no prices were successfully fetched")
	}

	return prices, nil
}

// groupSymbolsByChain organizes symbols by their target blockchain
func groupSymbolsByChain(symbols set.Set[types.Symbol], logger zerolog.Logger) map[ChainType][]types.Symbol {
	symbolsByChain := make(map[ChainType][]types.Symbol)

	for _, symbol := range symbols.ToSlice() {
		config, ok := chainlinkConfigMap[symbol]
		if !ok {
			logger.Warn().
				Str("symbol", string(symbol)).
				Msg("unsupported symbol for Chainlink oracle")
			continue
		}

		symbolsByChain[config.Chain] = append(symbolsByChain[config.Chain], symbol)
	}

	return symbolsByChain
}

// fetchPricesFromChain fetches prices for all symbols on a specific chain
func fetchPricesFromChain(
	chain ChainType,
	symbols []types.Symbol,
	timeout time.Duration,
	logger zerolog.Logger,
) (map[types.Symbol]float64, error) {
	// Get the appropriate connector function
	connectFunc, ok := chainConnectors[chain]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}

	// Connect to the chain
	client, err := connectFunc(timeout, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", chain, err)
	}
	defer client.Close()

	ctx := context.Background()
	prices := make(map[types.Symbol]float64)

	// Fetch price for each symbol on this chain
	for _, symbol := range symbols {
		config := chainlinkConfigMap[symbol] // We know it exists from grouping
		price, err := fetchPriceFromOracle(ctx, client, symbol, config, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch price for %s on %s: %w", symbol, chain, err)
		}

		prices[symbol] = price
	}
	return prices, nil
}

// fetchPriceFromOracle fetches price from a single Chainlink oracle
func fetchPriceFromOracle(
	ctx context.Context,
	client *ethclient.Client,
	symbol types.Symbol,
	config ChainlinkConfig,
	logger zerolog.Logger,
) (float64, error) {
	// Create oracle contract instance
	oracle, err := chainlink.NewChainlinkAggregator(config.ContractAddress, client)
	if err != nil {
		return 0, fmt.Errorf("failed to create oracle contract: %w", err)
	}

	// Get oracle metadata
	description, err := oracle.Description(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, fmt.Errorf("failed to get description: %w", err)
	}

	decimals, err := oracle.Decimals(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, fmt.Errorf("failed to get decimals: %w", err)
	}

	// Validate oracle
	if config.Description != "" && description != config.Description {
		logger.Warn().
			Str("symbol", string(symbol)).
			Str("expected", config.Description).
			Str("actual", description).
			Msg("oracle description mismatch")
	}

	// Get latest round data
	roundData, err := oracle.LatestRoundData(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, fmt.Errorf("failed to get latest round data: %w", err)
	}

	// Check data freshness
	now := time.Now()
	updatedAt := time.Unix(int64(roundData.UpdatedAt.Uint64()), 0)
	dataAge := now.Sub(updatedAt)

	if config.MaxDataAge > 0 && dataAge > config.MaxDataAge {
		logger.Warn().
			Str("symbol", string(symbol)).
			Dur("age", dataAge).
			Dur("max_age", config.MaxDataAge).
			Msg("oracle data is stale")
	}

	// Convert price
	price, err := convertChainlinkPrice(roundData.Answer, decimals)
	if err != nil {
		return 0, fmt.Errorf("failed to convert price: %w", err)
	}
	return price, nil
}

// convertChainlinkPrice converts Chainlink's raw price answer to a float64
func convertChainlinkPrice(answer *big.Int, decimals uint8) (float64, error) {
	if answer == nil {
		return 0, fmt.Errorf("answer is nil")
	}

	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	answerFloat := new(big.Float).SetInt(answer)
	result := new(big.Float).Quo(answerFloat, divisor)

	price, _ := result.Float64()
	return price, nil
}
