// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package uniswap_v3

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// UniswapV3PoolMetaData contains all meta data concerning the UniswapV3Pool contract.
var UniswapV3PoolMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"slot0\",\"outputs\":[{\"internalType\":\"uint160\",\"name\":\"sqrtPriceX96\",\"type\":\"uint160\"},{\"internalType\":\"int24\",\"name\":\"tick\",\"type\":\"int24\"},{\"internalType\":\"uint16\",\"name\":\"observationIndex\",\"type\":\"uint16\"},{\"internalType\":\"uint16\",\"name\":\"observationCardinality\",\"type\":\"uint16\"},{\"internalType\":\"uint16\",\"name\":\"observationCardinalityNext\",\"type\":\"uint16\"},{\"internalType\":\"uint8\",\"name\":\"feeProtocol\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"unlocked\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"liquidity\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// UniswapV3PoolABI is the input ABI used to generate the binding from.
// Deprecated: Use UniswapV3PoolMetaData.ABI instead.
var UniswapV3PoolABI = UniswapV3PoolMetaData.ABI

// UniswapV3Pool is an auto generated Go binding around an Ethereum contract.
type UniswapV3Pool struct {
	UniswapV3PoolCaller     // Read-only binding to the contract
	UniswapV3PoolTransactor // Write-only binding to the contract
	UniswapV3PoolFilterer   // Log filterer for contract events
}

// UniswapV3PoolCaller is an auto generated read-only Go binding around an Ethereum contract.
type UniswapV3PoolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapV3PoolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UniswapV3PoolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapV3PoolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UniswapV3PoolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapV3PoolSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UniswapV3PoolSession struct {
	Contract     *UniswapV3Pool    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UniswapV3PoolCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UniswapV3PoolCallerSession struct {
	Contract *UniswapV3PoolCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// UniswapV3PoolTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UniswapV3PoolTransactorSession struct {
	Contract     *UniswapV3PoolTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// UniswapV3PoolRaw is an auto generated low-level Go binding around an Ethereum contract.
type UniswapV3PoolRaw struct {
	Contract *UniswapV3Pool // Generic contract binding to access the raw methods on
}

// UniswapV3PoolCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UniswapV3PoolCallerRaw struct {
	Contract *UniswapV3PoolCaller // Generic read-only contract binding to access the raw methods on
}

// UniswapV3PoolTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UniswapV3PoolTransactorRaw struct {
	Contract *UniswapV3PoolTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUniswapV3Pool creates a new instance of UniswapV3Pool, bound to a specific deployed contract.
func NewUniswapV3Pool(address common.Address, backend bind.ContractBackend) (*UniswapV3Pool, error) {
	contract, err := bindUniswapV3Pool(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UniswapV3Pool{UniswapV3PoolCaller: UniswapV3PoolCaller{contract: contract}, UniswapV3PoolTransactor: UniswapV3PoolTransactor{contract: contract}, UniswapV3PoolFilterer: UniswapV3PoolFilterer{contract: contract}}, nil
}

// NewUniswapV3PoolCaller creates a new read-only instance of UniswapV3Pool, bound to a specific deployed contract.
func NewUniswapV3PoolCaller(address common.Address, caller bind.ContractCaller) (*UniswapV3PoolCaller, error) {
	contract, err := bindUniswapV3Pool(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UniswapV3PoolCaller{contract: contract}, nil
}

// NewUniswapV3PoolTransactor creates a new write-only instance of UniswapV3Pool, bound to a specific deployed contract.
func NewUniswapV3PoolTransactor(address common.Address, transactor bind.ContractTransactor) (*UniswapV3PoolTransactor, error) {
	contract, err := bindUniswapV3Pool(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UniswapV3PoolTransactor{contract: contract}, nil
}

// NewUniswapV3PoolFilterer creates a new log filterer instance of UniswapV3Pool, bound to a specific deployed contract.
func NewUniswapV3PoolFilterer(address common.Address, filterer bind.ContractFilterer) (*UniswapV3PoolFilterer, error) {
	contract, err := bindUniswapV3Pool(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UniswapV3PoolFilterer{contract: contract}, nil
}

// bindUniswapV3Pool binds a generic wrapper to an already deployed contract.
func bindUniswapV3Pool(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := UniswapV3PoolMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UniswapV3Pool *UniswapV3PoolRaw) Call(opts *bind.CallOpts, result *[]any, method string, params ...any) error {
	return _UniswapV3Pool.Contract.UniswapV3PoolCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UniswapV3Pool *UniswapV3PoolRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UniswapV3Pool.Contract.UniswapV3PoolTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UniswapV3Pool *UniswapV3PoolRaw) Transact(opts *bind.TransactOpts, method string, params ...any) (*types.Transaction, error) {
	return _UniswapV3Pool.Contract.UniswapV3PoolTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UniswapV3Pool *UniswapV3PoolCallerRaw) Call(opts *bind.CallOpts, result *[]any, method string, params ...any) error {
	return _UniswapV3Pool.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UniswapV3Pool *UniswapV3PoolTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UniswapV3Pool.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UniswapV3Pool *UniswapV3PoolTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...any) (*types.Transaction, error) {
	return _UniswapV3Pool.Contract.contract.Transact(opts, method, params...)
}

// Liquidity is a free data retrieval call binding the contract method 0x1a686502.
//
// Solidity: function liquidity() view returns(uint128)
func (_UniswapV3Pool *UniswapV3PoolCaller) Liquidity(opts *bind.CallOpts) (*big.Int, error) {
	var out []any
	err := _UniswapV3Pool.contract.Call(opts, &out, "liquidity")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Liquidity is a free data retrieval call binding the contract method 0x1a686502.
//
// Solidity: function liquidity() view returns(uint128)
func (_UniswapV3Pool *UniswapV3PoolSession) Liquidity() (*big.Int, error) {
	return _UniswapV3Pool.Contract.Liquidity(&_UniswapV3Pool.CallOpts)
}

// Liquidity is a free data retrieval call binding the contract method 0x1a686502.
//
// Solidity: function liquidity() view returns(uint128)
func (_UniswapV3Pool *UniswapV3PoolCallerSession) Liquidity() (*big.Int, error) {
	return _UniswapV3Pool.Contract.Liquidity(&_UniswapV3Pool.CallOpts)
}

// Slot0 is a free data retrieval call binding the contract method 0x3850c7bd.
//
// Solidity: function slot0() view returns(uint160 sqrtPriceX96, int24 tick, uint16 observationIndex, uint16 observationCardinality, uint16 observationCardinalityNext, uint8 feeProtocol, bool unlocked)
func (_UniswapV3Pool *UniswapV3PoolCaller) Slot0(opts *bind.CallOpts) (struct {
	SqrtPriceX96               *big.Int
	Tick                       *big.Int
	ObservationIndex           uint16
	ObservationCardinality     uint16
	ObservationCardinalityNext uint16
	FeeProtocol                uint8
	Unlocked                   bool
}, error) {
	var out []any
	err := _UniswapV3Pool.contract.Call(opts, &out, "slot0")

	outstruct := new(struct {
		SqrtPriceX96               *big.Int
		Tick                       *big.Int
		ObservationIndex           uint16
		ObservationCardinality     uint16
		ObservationCardinalityNext uint16
		FeeProtocol                uint8
		Unlocked                   bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.SqrtPriceX96 = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Tick = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.ObservationIndex = *abi.ConvertType(out[2], new(uint16)).(*uint16)
	outstruct.ObservationCardinality = *abi.ConvertType(out[3], new(uint16)).(*uint16)
	outstruct.ObservationCardinalityNext = *abi.ConvertType(out[4], new(uint16)).(*uint16)
	outstruct.FeeProtocol = *abi.ConvertType(out[5], new(uint8)).(*uint8)
	outstruct.Unlocked = *abi.ConvertType(out[6], new(bool)).(*bool)

	return *outstruct, err

}

// Slot0 is a free data retrieval call binding the contract method 0x3850c7bd.
//
// Solidity: function slot0() view returns(uint160 sqrtPriceX96, int24 tick, uint16 observationIndex, uint16 observationCardinality, uint16 observationCardinalityNext, uint8 feeProtocol, bool unlocked)
func (_UniswapV3Pool *UniswapV3PoolSession) Slot0() (struct {
	SqrtPriceX96               *big.Int
	Tick                       *big.Int
	ObservationIndex           uint16
	ObservationCardinality     uint16
	ObservationCardinalityNext uint16
	FeeProtocol                uint8
	Unlocked                   bool
}, error) {
	return _UniswapV3Pool.Contract.Slot0(&_UniswapV3Pool.CallOpts)
}

// Slot0 is a free data retrieval call binding the contract method 0x3850c7bd.
//
// Solidity: function slot0() view returns(uint160 sqrtPriceX96, int24 tick, uint16 observationIndex, uint16 observationCardinality, uint16 observationCardinalityNext, uint8 feeProtocol, bool unlocked)
func (_UniswapV3Pool *UniswapV3PoolCallerSession) Slot0() (struct {
	SqrtPriceX96               *big.Int
	Tick                       *big.Int
	ObservationIndex           uint16
	ObservationCardinality     uint16
	ObservationCardinalityNext uint16
	FeeProtocol                uint8
	Unlocked                   bool
}, error) {
	return _UniswapV3Pool.Contract.Slot0(&_UniswapV3Pool.CallOpts)
}
