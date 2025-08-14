package sources

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/NibiruChain/pricefeeder/types/uniswap_v3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

var _ types.FetchPricesFunc = UniswapV3PriceUpdate

const (
	UniswapV3               = "uniswap_v3"
	UniswapV3factoryAddress = "0x1F98431c8aD98523631AE4a59f267346ea31F984"
)

type TokenInfo struct {
	Address  string
	Decimals int
}

var tokenInfoMap = map[types.Symbol]TokenInfo{
	"USDa": {
		Address:  "0x8a60e489004ca22d775c5f2c657598278d17d9c2",
		Decimals: 18,
	},
	"USDT": {
		Address:  "0xdac17f958d2ee523a2206206994597c13d831ec7",
		Decimals: 6,
	},
}

// Pool represents a Uniswap V3 pool with its metadata
type Pool struct {
	Address   common.Address
	Fee       uint32
	Liquidity *big.Int
}

// TokenPair represents a sorted token pair for Uniswap V3
type TokenPair struct {
	Token0         common.Address
	Token1         common.Address
	Token0Decimals int
	Token1Decimals int
	IsReversed     bool // true if the original order was reversed for sorting
}

// NewTokenPair creates a properly sorted token pair for Uniswap V3
func NewTokenPair(tokenA, tokenB common.Address, decimalsA, decimalsB int) TokenPair {
	// Sort tokens by address (token0 < token1)
	if tokenA.Hex() < tokenB.Hex() {
		return TokenPair{
			Token0:         tokenA,
			Token1:         tokenB,
			Token0Decimals: decimalsA,
			Token1Decimals: decimalsB,
			IsReversed:     false,
		}
	} else {
		return TokenPair{
			Token0:         tokenB,
			Token1:         tokenA,
			Token0Decimals: decimalsB,
			Token1Decimals: decimalsA,
			IsReversed:     true,
		}
	}
}

// UniswapV3PriceUpdate retrieves the exchange rates for the given symbols from the Uniswap V3 protocol.
func UniswapV3PriceUpdate(symbols set.Set[types.Symbol], logger zerolog.Logger) (map[types.Symbol]float64, error) {
	endpoints := types.GetEthereumRPCEndpoints()

	var client *ethclient.Client
	var connErr error

	// If ETHEREUM_RPC_ENDPOINT is set, use it exclusively
	if os.Getenv("ETHEREUM_RPC_ENDPOINT") != "" {
		client, connErr = ethclient.Dial(endpoints[0])
		if connErr != nil {
			return nil, fmt.Errorf("failed to connect to Ethereum client: %w", connErr)
		}
	} else {
		// Try each endpoint with 10 second timeout
		timeout := 10 * time.Second

		for _, endpoint := range endpoints {
			client, connErr = types.TryEthereumRPCEndpoint(endpoint, timeout, logger)
			if connErr == nil {
				break
			}
			logger.Warn().
				Str("endpoint", endpoint).
				Err(connErr).
				Msg("failed to connect to RPC endpoint, trying next")
		}
		if client == nil {
			return nil, fmt.Errorf("failed to connect to any Ethereum RPC endpoint. Last error: %w", connErr)
		}
	}
	defer client.Close()

	factory, err := uniswap_v3.NewUniswapV3Factory(common.HexToAddress(UniswapV3factoryAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create factory contract: %w", err)
	}

	ctx := context.Background()
	prices := make(map[types.Symbol]float64)

	for _, symbol := range symbols.ToSlice() {
		// symbol is actually a pair like "USDa:USDT", so split it
		parts := strings.Split(string(symbol), ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid symbol: %s", symbol)
		}

		baseTokenSymbol, quoteTokenSymbol := parts[0], parts[1]
		baseTokenInfo, ok := tokenInfoMap[types.Symbol(baseTokenSymbol)]
		if !ok {
			return nil, fmt.Errorf("unsupported token %s in pair %s", baseTokenSymbol, symbol)
		}
		quoteTokenInfo, ok := tokenInfoMap[types.Symbol(quoteTokenSymbol)]
		if !ok {
			return nil, fmt.Errorf("unsupported token %s in pair %s", quoteTokenSymbol, symbol)
		}

		// Create properly sorted token pair
		tokenPair := NewTokenPair(
			common.HexToAddress(baseTokenInfo.Address),
			common.HexToAddress(quoteTokenInfo.Address),
			baseTokenInfo.Decimals,
			quoteTokenInfo.Decimals,
		)

		logger.Debug().
			Str("symbol", string(symbol)).
			Str("base_token", baseTokenSymbol).
			Str("quote_token", quoteTokenSymbol).
			Str("token0", tokenPair.Token0.Hex()).
			Str("token1", tokenPair.Token1.Hex()).
			Bool("is_reversed", tokenPair.IsReversed).
			Msg("processing token pair")

		bestPoolAddress, err := findPoolWithHighestLiquidity(
			ctx,
			factory,
			client,
			tokenPair,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to find pool for %s/%s: %w", baseTokenSymbol, quoteTokenSymbol, err)
		}
		if bestPoolAddress == nil {
			return nil, fmt.Errorf("no pool found for %s/%s", baseTokenSymbol, quoteTokenSymbol)
		}

		price, err := getPriceFromPool(ctx, client, *bestPoolAddress, tokenPair)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to get price from pool %s for %s/%s: %w",
				bestPoolAddress.Hex(),
				baseTokenSymbol, quoteTokenSymbol,
				err,
			)
		}
		// NOTE: Limit USDa:USDT price to be >= 1.01 to prevent oracle price
		// manipulations on Uniswap pools. The token is backed by USDT.
		if symbol == "USDa:USDT" && price < 1.01 {
			price = 1.01
		}
		prices[symbol] = price
	}
	return prices, nil
}

// findPoolWithHighestLiquidity finds the Uniswap V3 pool with the highest liquidity for the given token pair
func findPoolWithHighestLiquidity(
	ctx context.Context,
	factory *uniswap_v3.UniswapV3Factory,
	client *ethclient.Client,
	tokenPair TokenPair,
) (*common.Address, error) {
	feeTiers := []uint32{500, 3000, 10000} // 0.05%, 0.3%, 1%
	var pools []Pool

	// Find all available pools for the token pair (using sorted addresses)
	for _, fee := range feeTiers {
		poolAddress, err := factory.GetPool(
			&bind.CallOpts{Context: ctx},
			tokenPair.Token0,
			tokenPair.Token1,
			big.NewInt(int64(fee)),
		)
		if err != nil {
			continue
		}

		if poolAddress != (common.Address{}) {
			pools = append(pools, Pool{
				Address: poolAddress,
				Fee:     fee,
			})
		}
	}

	if len(pools) == 0 {
		return nil, fmt.Errorf("no pools found for token pair %s/%s", tokenPair.Token0.Hex(), tokenPair.Token1.Hex())
	}

	// Get liquidity for each pool
	var poolsWithLiquidity []Pool
	for _, poolInfo := range pools {
		pool, err := uniswap_v3.NewUniswapV3Pool(poolInfo.Address, client)
		if err != nil {
			return nil, fmt.Errorf("failed to create pool contract: %w", err)
		}
		liquidity, err := pool.Liquidity(&bind.CallOpts{Context: ctx})
		if err != nil {
			return nil, fmt.Errorf("failed to get pool liquidity: %w", err)
		}
		poolInfo.Liquidity = liquidity
		poolsWithLiquidity = append(poolsWithLiquidity, poolInfo)
	}

	if len(poolsWithLiquidity) == 0 {
		return nil, fmt.Errorf("could not get liquidity data for any pools")
	}

	// Sort by liquidity (highest first)
	sort.Slice(poolsWithLiquidity, func(i, j int) bool {
		return poolsWithLiquidity[i].Liquidity.Cmp(poolsWithLiquidity[j].Liquidity) > 0
	})

	bestPool := poolsWithLiquidity[0]
	return &bestPool.Address, nil
}

// getPriceFromPool retrieves the current price from a Uniswap V3 pool
func getPriceFromPool(
	ctx context.Context,
	client *ethclient.Client,
	poolAddress common.Address,
	tokenPair TokenPair,
) (float64, error) {
	pool, err := uniswap_v3.NewUniswapV3Pool(poolAddress, client)
	if err != nil {
		return 0, fmt.Errorf("failed to create pool contract: %w", err)
	}

	slot0, err := pool.Slot0(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, fmt.Errorf("failed to get slot0: %w", err)
	}

	// Calculate the raw price (token0 in terms of token1)
	rawPrice := sqrtPriceX96ToPrice(slot0.SqrtPriceX96, tokenPair.Token0Decimals, tokenPair.Token1Decimals)

	// If the tokens were reversed during sorting, we need to invert the price
	// to get the price in the original requested order (base:quote)
	var finalPrice float64
	if tokenPair.IsReversed {
		// If reversed, rawPrice is quote/base, so we need base/quote
		finalPrice = 1.0 / rawPrice
	} else {
		// If not reversed, rawPrice is already base/quote
		finalPrice = rawPrice
	}

	return finalPrice, nil
}

// sqrtPriceX96ToPrice converts Uniswap V3's sqrtPriceX96 format to a regular price.
//
// Inputs:
// - sqrtPriceX96: The square root of the price in Uniswap V3's fixed-point Q96 format.
// - decimals0: The number of decimal places for token0.
// - decimals1: The number of decimal places for token1.
//
// Algorithm:
// 1. Uniswap V3 represents prices using sqrtPriceX96, which is the square root of the price scaled by 2^96.
// 2. To calculate the actual price, we square sqrtPriceX96 and divide by 2^192 (since (2^96)^2 = 2^192).
// 3. Adjust the price based on the difference in token decimals to ensure consistent scaling.
// 4. Scale the result to 18 decimal places for precision and convert it to a float64.
//
// Constants:
// - Q96 = 2^96: Used for scaling sqrtPriceX96.
// - Q192 = 2^192: Used for scaling the squared price.
func sqrtPriceX96ToPrice(sqrtPriceX96 *big.Int, decimals0, decimals1 int) float64 {
	// Convert sqrtPriceX96 to price
	sqrtPrice := new(big.Int).Set(sqrtPriceX96)

	// Q96 = 2^96
	Q96 := new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil)

	// Calculate price = (sqrtPriceX96)^2 / 2^192
	priceX192 := new(big.Int).Mul(sqrtPrice, sqrtPrice)

	// Adjust for token decimals
	decimalDiff := decimals0 - decimals1
	decimalAdjustment := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs(decimalDiff))), nil)

	var adjustedPriceX192 *big.Int
	if decimalDiff >= 0 {
		adjustedPriceX192 = new(big.Int).Mul(priceX192, decimalAdjustment)
	} else {
		adjustedPriceX192 = new(big.Int).Div(priceX192, decimalAdjustment)
	}

	// Q192 = 2^192
	Q192 := new(big.Int).Mul(Q96, Q96)

	// Use 18 decimal places for precision
	scaleFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

	scaledPrice := new(big.Int).Mul(adjustedPriceX192, scaleFactor)
	scaledPrice = new(big.Int).Div(scaledPrice, Q192)

	// Convert to float64
	priceFloat, _ := strconv.ParseFloat(scaledPrice.String(), 64)
	return priceFloat / 1e18
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
