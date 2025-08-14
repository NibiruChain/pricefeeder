package sources

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/NibiruChain/pricefeeder/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JSON-RPC request/response structures
type JSONRPCRequest struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type JSONRPCResponse struct {
	ID     int           `json:"id"`
	Result interface{}   `json:"result"`
	Error  *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Test helper functions
func setTestRPC(t *testing.T, url string) func() {
	original := os.Getenv("ETHEREUM_RPC_ENDPOINT")
	require.NoError(t, os.Setenv("ETHEREUM_RPC_ENDPOINT", url))

	return func() {
		if original != "" {
			require.NoError(t, os.Setenv("ETHEREUM_RPC_ENDPOINT", original))
		} else {
			_ = os.Unsetenv("ETHEREUM_RPC_ENDPOINT")
		}
	}
}

func withMockServer(t *testing.T) (*httptest.Server, func()) {
	server := createMockEthereumServer()
	cleanup := setTestRPC(t, server.URL)

	return server, func() {
		server.Close()
		cleanup()
	}
}

// Mock server that responds to Ethereum JSON-RPC calls
func createMockEthereumServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		var response JSONRPCResponse
		response.ID = req.ID

		switch req.Method {
		case "eth_call":
			// Extract the 'to' address and 'data' from params
			if len(req.Params) > 0 {
				callParams := req.Params[0].(map[string]interface{})
				to := callParams["to"].(string)
				data := callParams["data"].(string)

				// Mock responses based on contract address and method signature
				if to == strings.ToLower(UniswapV3factoryAddress) { // Uniswap V3 Factory (lowercase)
					response.Result = handleFactoryCall(data)
				} else {
					// Assume it's a pool contract call
					response.Result = handlePoolCall(data)
				}
			}

		default:
			response.Error = &JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func handleFactoryCall(data string) string {
	// getPool(address,address,uint24) method signature: 0x1698ee82
	if len(data) >= 10 && data[:10] == "0x1698ee82" {
		// Return a mock pool address
		// 32-byte address padded with zeros: 0x1111111111111111111111111111111111111111
		return "0x0000000000000000000000001111111111111111111111111111111111111111"
	}
	return "0x0000000000000000000000000000000000000000000000000000000000000000"
}

func handlePoolCall(data string) string {
	// Check method signatures
	if len(data) >= 10 {
		methodSig := data[:10]

		switch methodSig {
		case "0x1a686502": // liquidity() method
			// Return a mock liquidity value: 1000000 (as hex, 32-byte padded)
			return "0x00000000000000000000000000000000000000000000000000000000000f4240"

		case "0x3850c7bd": // slot0() method
			// Return mock slot0 data
			// This is a struct with multiple fields, but we only need sqrtPriceX96 (first field)
			// Using sqrt(1) * 2^96 = 79228162514264337593543950336 in hex
			sqrtPriceX96 := "0000000000000000000000000000000000000000000001000000000000000000" // sqrt(1) * 2^96
			tick := "0000000000000000000000000000000000000000000000000000000000000000"         // tick = 0
			observationIndex := "0000000000000000000000000000000000000000000000000000000000000000"
			observationCardinality := "0000000000000000000000000000000000000000000000000000000000000000"
			observationCardinalityNext := "0000000000000000000000000000000000000000000000000000000000000000"
			feeProtocol := "0000000000000000000000000000000000000000000000000000000000000000"
			unlocked := "0000000000000000000000000000000000000000000000000000000000000001"

			return "0x" + sqrtPriceX96 + tick + observationIndex + observationCardinality + observationCardinalityNext + feeProtocol + unlocked
		}
	}

	return "0x0000000000000000000000000000000000000000000000000000000000000000"
}

func TestUniswapV3PriceUpdate_WithHTTPMock_Success(t *testing.T) {
	_, cleanup := withMockServer(t)
	defer cleanup()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]()
	symbols.Add("USDa:USDT")

	prices, err := UniswapV3PriceUpdate(symbols, logger)

	require.NoError(t, err)
	require.NotNil(t, prices)
	require.Contains(t, prices, types.Symbol("USDa:USDT"))

	price := prices[types.Symbol("USDa:USDT")]
	assert.Greater(t, price, 0.0, "Price should be positive")
	t.Logf("Mocked USDa:USDT price: %f", price)
}

func TestUniswapV3PriceUpdate_WithHTTPMock_NoPoolFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		response := JSONRPCResponse{ID: req.ID}
		if req.Method == "eth_call" {
			response.Result = "0x0000000000000000000000000000000000000000000000000000000000000000"
		} else {
			response.Error = &JSONRPCError{Code: -32601, Message: "Method not found"}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer setTestRPC(t, server.URL)()
	defer server.Close()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]()
	symbols.Add("USDa:USDT")

	_, err := UniswapV3PriceUpdate(symbols, logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no pools found")
}

func TestUniswapV3PriceUpdate_WithHTTPMock_RPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		response := JSONRPCResponse{
			ID:    req.ID,
			Error: &JSONRPCError{Code: -32000, Message: "execution reverted"},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer setTestRPC(t, server.URL)()
	defer server.Close()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]()
	symbols.Add("USDa:USDT")

	_, err := UniswapV3PriceUpdate(symbols, logger)
	require.Error(t, err)
	t.Logf("RPC error: %v", err)
}

func TestUniswapV3PriceUpdate_WithHTTPMock_ServerUnavailable(t *testing.T) {
	defer setTestRPC(t, "http://nonexistent-server-that-should-not-exist.invalid:9999")()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]()
	symbols.Add("USDa:USDT")

	_, err := UniswapV3PriceUpdate(symbols, logger)
	require.Error(t, err)

	errorContainsExpected := strings.Contains(err.Error(), "failed to connect to Ethereum client") ||
		strings.Contains(err.Error(), "no pools found") ||
		strings.Contains(err.Error(), "failed to find pool")
	assert.True(t, errorContainsExpected, "Should get connection or RPC-related error, got: %v", err)
}

func TestUniswapV3PriceUpdate_WithHTTPMock_InvalidURL(t *testing.T) {
	defer setTestRPC(t, "invalid-url-format")()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]()
	symbols.Add("USDa:USDT")

	_, err := UniswapV3PriceUpdate(symbols, logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to Ethereum client")
}

func TestUniswapV3PriceUpdate_WithHTTPMock_MultiplePools(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		var response JSONRPCResponse
		response.ID = req.ID

		if req.Method == "eth_call" && len(req.Params) > 0 {
			callParams := req.Params[0].(map[string]interface{})
			to := callParams["to"].(string)
			data := callParams["data"].(string)

			if to == strings.ToLower(UniswapV3factoryAddress) {
				if data[:10] == "0x1698ee82" { // getPool
					response.Result = "0x0000000000000000000000001111111111111111111111111111111111111111"
				}
			} else {
				// Pool calls
				switch data[:10] {
				case "0x1a686502": // liquidity()
					callCount++
					switch callCount {
					case 1:
						response.Result = "0x00000000000000000000000000000000000000000000000000000000000186a0" // 100000
					case 2:
						response.Result = "0x00000000000000000000000000000000000000000000000000000000000f4240" // 1000000 (highest)
					case 3:
						response.Result = "0x000000000000000000000000000000000000000000000000000000000000c350" // 50000
					default:
						response.Result = "0x00000000000000000000000000000000000000000000000000000000000f4240" // 1000000
					}
				case "0x3850c7bd": // slot0()
					sqrtPriceX96 := "0000000000000000000000000000000000000000000001000000000000000000"
					result := "0x" + sqrtPriceX96 + strings.Repeat("0000000000000000000000000000000000000000000000000000000000000000", 6) + "0000000000000000000000000000000000000000000000000000000000000001"
					response.Result = result
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer setTestRPC(t, server.URL)()
	defer server.Close()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]()
	symbols.Add("USDa:USDT")

	prices, err := UniswapV3PriceUpdate(symbols, logger)

	require.NoError(t, err)
	require.NotNil(t, prices)
	require.Contains(t, prices, types.Symbol("USDa:USDT"))

	price := prices[types.Symbol("USDa:USDT")]
	assert.Greater(t, price, 0.0, "Price should be positive")
	assert.GreaterOrEqual(t, callCount, 3, "Should have called liquidity() for multiple pools")

	t.Logf("Mocked USDa:USDT price: %f, liquidity calls made: %d", price, callCount)
}

// =======================
// Unit Tests for Pure Functions
// =======================

func TestNewTokenPair(t *testing.T) {
	tests := []struct {
		name            string
		tokenA          string
		tokenB          string
		decimalsA       int
		decimalsB       int
		expectedToken0  string
		expectedToken1  string
		expectedReverse bool
	}{
		{
			name:            "tokenA < tokenB",
			tokenA:          "0x1111111111111111111111111111111111111111",
			tokenB:          "0x2222222222222222222222222222222222222222",
			decimalsA:       18,
			decimalsB:       6,
			expectedToken0:  "0x1111111111111111111111111111111111111111",
			expectedToken1:  "0x2222222222222222222222222222222222222222",
			expectedReverse: false,
		},
		{
			name:            "tokenA > tokenB",
			tokenA:          "0x2222222222222222222222222222222222222222",
			tokenB:          "0x1111111111111111111111111111111111111111",
			decimalsA:       18,
			decimalsB:       6,
			expectedToken0:  "0x1111111111111111111111111111111111111111",
			expectedToken1:  "0x2222222222222222222222222222222222222222",
			expectedReverse: true,
		},
		{
			name:            "USDa and USDT from config",
			tokenA:          "0x8a60e489004ca22d775c5f2c657598278d17d9c2", // USDa
			tokenB:          "0xdac17f958d2ee523a2206206994597c13d831ec7", // USDT
			decimalsA:       18,
			decimalsB:       6,
			expectedToken0:  "0x8A60E489004Ca22d775C5F2c657598278d17D9c2", // USDa (smaller when compared as hex strings)
			expectedToken1:  "0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT (larger when compared as hex strings)
			expectedReverse: false,                                        // USDa is tokenA and also token0, so no reversal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenA := common.HexToAddress(tt.tokenA)
			tokenB := common.HexToAddress(tt.tokenB)

			pair := NewTokenPair(tokenA, tokenB, tt.decimalsA, tt.decimalsB)

			assert.Equal(t, tt.expectedToken0, pair.Token0.Hex())
			assert.Equal(t, tt.expectedToken1, pair.Token1.Hex())
			assert.Equal(t, tt.expectedReverse, pair.IsReversed)

			if tt.expectedReverse {
				assert.Equal(t, tt.decimalsB, pair.Token0Decimals)
				assert.Equal(t, tt.decimalsA, pair.Token1Decimals)
			} else {
				assert.Equal(t, tt.decimalsA, pair.Token0Decimals)
				assert.Equal(t, tt.decimalsB, pair.Token1Decimals)
			}
		})
	}
}

func TestSqrtPriceX96ToPrice(t *testing.T) {
	tests := []struct {
		name         string
		sqrtPriceX96 string // hex string for easier readability
		decimals0    int
		decimals1    int
		expected     float64
		tolerance    float64
	}{
		{
			name:         "equal decimals - price 1",
			sqrtPriceX96: "79228162514264337593543950336", // sqrt(1) * 2^96
			decimals0:    18,
			decimals1:    18,
			expected:     1.0,
			tolerance:    0.001,
		},
		{
			name:         "different decimals - 18 vs 6",
			sqrtPriceX96: "79228162514264337593543950336", // sqrt(1) * 2^96
			decimals0:    18,
			decimals1:    6,
			expected:     1e12, // 10^(18-6)
			tolerance:    1e9,
		},
		{
			name:         "different decimals - 6 vs 18",
			sqrtPriceX96: "79228162514264337593543950336", // sqrt(1) * 2^96
			decimals0:    6,
			decimals1:    18,
			expected:     1e-12, // 10^(6-18)
			tolerance:    1e-15,
		},
		{
			name:         "price of 2.0",
			sqrtPriceX96: "112045541949572279837463876454", // sqrt(2) * 2^96
			decimals0:    18,
			decimals1:    18,
			expected:     2.0,
			tolerance:    0.01,
		},
		{
			name:         "price of 0.5",
			sqrtPriceX96: "56022770974786139918731938227", // sqrt(0.5) * 2^96
			decimals0:    18,
			decimals1:    18,
			expected:     0.5,
			tolerance:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqrtPriceX96, _ := new(big.Int).SetString(tt.sqrtPriceX96, 10)
			result := sqrtPriceX96ToPrice(sqrtPriceX96, tt.decimals0, tt.decimals1)
			assert.InDelta(t, tt.expected, result, tt.tolerance)
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{input: 5, expected: 5},
		{input: -5, expected: 5},
		{input: 0, expected: 0},
		{input: -1, expected: 1},
		{input: 1, expected: 1},
		{input: -100, expected: 100},
		{input: 999, expected: 999},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("abs(%d)", tt.input), func(t *testing.T) {
			result := abs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =======================
// Configuration and Constants Tests
// =======================

func TestTokenInfoMap(t *testing.T) {
	// Test that expected tokens are in the map
	expectedTokens := []types.Symbol{"USDa", "USDT"}

	for _, token := range expectedTokens {
		info, exists := tokenInfoMap[token]
		assert.True(t, exists, "token %s should exist in tokenInfoMap", token)
		assert.NotEmpty(t, info.Address, "token %s should have a non-empty address", token)
		assert.Greater(t, info.Decimals, 0, "token %s should have positive decimals", token)

		// Validate address format
		assert.True(t, common.IsHexAddress(info.Address), "token %s should have valid hex address", token)
	}

	// Test specific token values
	usdaInfo := tokenInfoMap["USDa"]
	assert.Equal(t, "0x8a60e489004ca22d775c5f2c657598278d17d9c2", usdaInfo.Address)
	assert.Equal(t, 18, usdaInfo.Decimals)

	usdtInfo := tokenInfoMap["USDT"]
	assert.Equal(t, "0xdac17f958d2ee523a2206206994597c13d831ec7", usdtInfo.Address)
	assert.Equal(t, 6, usdtInfo.Decimals)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "uniswap_v3", UniswapV3)
	assert.Equal(t, "0x1F98431c8aD98523631AE4a59f267346ea31F984", UniswapV3factoryAddress)

	// Validate factory address format
	assert.True(t, common.IsHexAddress(UniswapV3factoryAddress))
}

// =======================
// Input Validation Tests
// =======================

func TestUniswapV3PriceUpdate_InvalidSymbol(t *testing.T) {
	_, cleanup := withMockServer(t)
	defer cleanup()

	logger := zerolog.New(os.Stdout)

	tests := []struct {
		name     string
		symbol   string
		errorMsg string
	}{
		{
			name:     "no colon",
			symbol:   "USDaUSDT",
			errorMsg: "invalid symbol",
		},
		{
			name:     "empty string",
			symbol:   "",
			errorMsg: "invalid symbol",
		},
		{
			name:     "too many parts",
			symbol:   "USDa:USDT:ETH",
			errorMsg: "invalid symbol",
		},
		{
			name:     "empty first part",
			symbol:   ":USDT",
			errorMsg: "unsupported token",
		},
		{
			name:     "empty second part",
			symbol:   "USDa:",
			errorMsg: "unsupported token",
		},
		{
			name:     "only colon",
			symbol:   ":",
			errorMsg: "unsupported token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbols := set.New[types.Symbol]()
			symbols.Add(types.Symbol(tt.symbol))

			_, err := UniswapV3PriceUpdate(symbols, logger)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestUniswapV3PriceUpdate_UnsupportedTokens(t *testing.T) {
	_, cleanup := withMockServer(t)
	defer cleanup()

	logger := zerolog.New(os.Stdout)

	tests := []struct {
		name   string
		symbol string
	}{
		{
			name:   "unknown first token",
			symbol: "UNKNOWN:USDT",
		},
		{
			name:   "unknown second token",
			symbol: "USDa:UNKNOWN",
		},
		{
			name:   "both tokens unknown",
			symbol: "UNKNOWN1:UNKNOWN2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbols := set.New[types.Symbol]()
			symbols.Add(types.Symbol(tt.symbol))

			_, err := UniswapV3PriceUpdate(symbols, logger)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported token")
		})
	}
}

// =======================
// Edge Cases and Error Scenarios
// =======================

func TestUniswapV3PriceUpdate_EmptySymbolSet(t *testing.T) {
	_, cleanup := withMockServer(t)
	defer cleanup()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]() // Empty set

	prices, err := UniswapV3PriceUpdate(symbols, logger)

	require.NoError(t, err)
	assert.NotNil(t, prices)
	assert.Empty(t, prices)
}

func TestUniswapV3PriceUpdate_MultipleSymbols(t *testing.T) {
	_, cleanup := withMockServer(t)
	defer cleanup()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]()
	symbols.Add("USDa:USDT")
	symbols.Add("USDT:USDa") // Reverse pair

	prices, err := UniswapV3PriceUpdate(symbols, logger)

	require.NoError(t, err)
	assert.NotNil(t, prices)
	assert.Len(t, prices, 2)
	assert.Contains(t, prices, types.Symbol("USDa:USDT"))
	assert.Contains(t, prices, types.Symbol("USDT:USDa"))

	// Prices should be reciprocals (roughly)
	price1 := prices[types.Symbol("USDa:USDT")]
	price2 := prices[types.Symbol("USDT:USDa")]
	assert.Greater(t, price1, 0.0)
	assert.Greater(t, price2, 0.0)

	t.Logf("USDa:USDT = %f, USDT:USDa = %f", price1, price2)
}

// =======================
// Benchmark Tests
// =======================

func BenchmarkNewTokenPair(b *testing.B) {
	tokenA := common.HexToAddress("0x1111111111111111111111111111111111111111")
	tokenB := common.HexToAddress("0x2222222222222222222222222222222222222222")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTokenPair(tokenA, tokenB, 18, 6)
	}
}

func BenchmarkSqrtPriceX96ToPrice(b *testing.B) {
	sqrtPriceX96, _ := new(big.Int).SetString("79228162514264337593543950336", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqrtPriceX96ToPrice(sqrtPriceX96, 18, 6)
	}
}

func BenchmarkAbs(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		abs(-12)
	}
}

// =======================
// Pool Selection Logic Tests
// =======================

func TestUniswapV3PriceUpdate_PoolSelection_DifferentLiquidities(t *testing.T) {
	liquidityCallOrder := make([]int, 0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		var response JSONRPCResponse
		response.ID = req.ID

		if req.Method == "eth_call" && len(req.Params) > 0 {
			callParams := req.Params[0].(map[string]interface{})
			to := callParams["to"].(string)
			data := callParams["data"].(string)

			if to == "0x1f98431c8ad98523631ae4a59f267346ea31f984" { // Factory
				response.Result = "0x0000000000000000000000001111111111111111111111111111111111111111"
			} else {
				// Pool calls
				switch data[:10] {
				case "0x1a686502": // liquidity()
					callOrder := len(liquidityCallOrder) + 1
					liquidityCallOrder = append(liquidityCallOrder, callOrder)

					switch callOrder {
					case 1: // Fee 500 - lowest liquidity
						response.Result = "0x0000000000000000000000000000000000000000000000000000000000002710" // 10000
					case 2: // Fee 3000 - highest liquidity
						response.Result = "0x00000000000000000000000000000000000000000000000000000000000f4240" // 1000000
					case 3: // Fee 10000 - medium liquidity
						response.Result = "0x000000000000000000000000000000000000000000000000000000000001869f" // 100000
					default:
						response.Result = "0x00000000000000000000000000000000000000000000000000000000000f4240" // Default to highest
					}
				case "0x3850c7bd": // slot0()
					// Should only be called once for the highest liquidity pool
					sqrtPriceX96 := "0000000000000000000000000000000000000000000001000000000000000000"
					result := "0x" + sqrtPriceX96 + strings.Repeat("0000000000000000000000000000000000000000000000000000000000000000", 6) + "0000000000000000000000000000000000000000000000000000000000000001"
					response.Result = result
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer setTestRPC(t, server.URL)()
	defer server.Close()

	logger := zerolog.New(os.Stdout)
	symbols := set.New[types.Symbol]()
	symbols.Add("USDa:USDT")

	prices, err := UniswapV3PriceUpdate(symbols, logger)

	require.NoError(t, err)
	assert.NotNil(t, prices)
	assert.Contains(t, prices, types.Symbol("USDa:USDT"))

	// Verify that liquidity was checked for all pools
	assert.Equal(t, 3, len(liquidityCallOrder), "Should check liquidity for all 3 fee tiers")

	t.Logf("Liquidity call order: %v", liquidityCallOrder)
	t.Logf("Selected pool price: %f", prices[types.Symbol("USDa:USDT")])
}
