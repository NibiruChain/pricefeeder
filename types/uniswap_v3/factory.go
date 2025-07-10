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

// UniswapV3FactoryMetaData contains all meta data concerning the UniswapV3Factory contract.
var UniswapV3FactoryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenA\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"tokenB\",\"type\":\"address\"},{\"internalType\":\"uint24\",\"name\":\"fee\",\"type\":\"uint24\"}],\"name\":\"getPool\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"pool\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// UniswapV3FactoryABI is the input ABI used to generate the binding from.
// Deprecated: Use UniswapV3FactoryMetaData.ABI instead.
var UniswapV3FactoryABI = UniswapV3FactoryMetaData.ABI

// UniswapV3Factory is an auto generated Go binding around an Ethereum contract.
type UniswapV3Factory struct {
	UniswapV3FactoryCaller     // Read-only binding to the contract
	UniswapV3FactoryTransactor // Write-only binding to the contract
	UniswapV3FactoryFilterer   // Log filterer for contract events
}

// UniswapV3FactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type UniswapV3FactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapV3FactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UniswapV3FactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapV3FactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UniswapV3FactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapV3FactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UniswapV3FactorySession struct {
	Contract     *UniswapV3Factory // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UniswapV3FactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UniswapV3FactoryCallerSession struct {
	Contract *UniswapV3FactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// UniswapV3FactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UniswapV3FactoryTransactorSession struct {
	Contract     *UniswapV3FactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// UniswapV3FactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type UniswapV3FactoryRaw struct {
	Contract *UniswapV3Factory // Generic contract binding to access the raw methods on
}

// UniswapV3FactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UniswapV3FactoryCallerRaw struct {
	Contract *UniswapV3FactoryCaller // Generic read-only contract binding to access the raw methods on
}

// UniswapV3FactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UniswapV3FactoryTransactorRaw struct {
	Contract *UniswapV3FactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUniswapV3Factory creates a new instance of UniswapV3Factory, bound to a specific deployed contract.
func NewUniswapV3Factory(address common.Address, backend bind.ContractBackend) (*UniswapV3Factory, error) {
	contract, err := bindUniswapV3Factory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UniswapV3Factory{UniswapV3FactoryCaller: UniswapV3FactoryCaller{contract: contract}, UniswapV3FactoryTransactor: UniswapV3FactoryTransactor{contract: contract}, UniswapV3FactoryFilterer: UniswapV3FactoryFilterer{contract: contract}}, nil
}

// NewUniswapV3FactoryCaller creates a new read-only instance of UniswapV3Factory, bound to a specific deployed contract.
func NewUniswapV3FactoryCaller(address common.Address, caller bind.ContractCaller) (*UniswapV3FactoryCaller, error) {
	contract, err := bindUniswapV3Factory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UniswapV3FactoryCaller{contract: contract}, nil
}

// NewUniswapV3FactoryTransactor creates a new write-only instance of UniswapV3Factory, bound to a specific deployed contract.
func NewUniswapV3FactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*UniswapV3FactoryTransactor, error) {
	contract, err := bindUniswapV3Factory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UniswapV3FactoryTransactor{contract: contract}, nil
}

// NewUniswapV3FactoryFilterer creates a new log filterer instance of UniswapV3Factory, bound to a specific deployed contract.
func NewUniswapV3FactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*UniswapV3FactoryFilterer, error) {
	contract, err := bindUniswapV3Factory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UniswapV3FactoryFilterer{contract: contract}, nil
}

// bindUniswapV3Factory binds a generic wrapper to an already deployed contract.
func bindUniswapV3Factory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := UniswapV3FactoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UniswapV3Factory *UniswapV3FactoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UniswapV3Factory.Contract.UniswapV3FactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UniswapV3Factory *UniswapV3FactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UniswapV3Factory.Contract.UniswapV3FactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UniswapV3Factory *UniswapV3FactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UniswapV3Factory.Contract.UniswapV3FactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UniswapV3Factory *UniswapV3FactoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UniswapV3Factory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UniswapV3Factory *UniswapV3FactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UniswapV3Factory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UniswapV3Factory *UniswapV3FactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UniswapV3Factory.Contract.contract.Transact(opts, method, params...)
}

// GetPool is a free data retrieval call binding the contract method 0x1698ee82.
//
// Solidity: function getPool(address tokenA, address tokenB, uint24 fee) view returns(address pool)
func (_UniswapV3Factory *UniswapV3FactoryCaller) GetPool(opts *bind.CallOpts, tokenA common.Address, tokenB common.Address, fee *big.Int) (common.Address, error) {
	var out []interface{}
	err := _UniswapV3Factory.contract.Call(opts, &out, "getPool", tokenA, tokenB, fee)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPool is a free data retrieval call binding the contract method 0x1698ee82.
//
// Solidity: function getPool(address tokenA, address tokenB, uint24 fee) view returns(address pool)
func (_UniswapV3Factory *UniswapV3FactorySession) GetPool(tokenA common.Address, tokenB common.Address, fee *big.Int) (common.Address, error) {
	return _UniswapV3Factory.Contract.GetPool(&_UniswapV3Factory.CallOpts, tokenA, tokenB, fee)
}

// GetPool is a free data retrieval call binding the contract method 0x1698ee82.
//
// Solidity: function getPool(address tokenA, address tokenB, uint24 fee) view returns(address pool)
func (_UniswapV3Factory *UniswapV3FactoryCallerSession) GetPool(tokenA common.Address, tokenB common.Address, fee *big.Int) (common.Address, error) {
	return _UniswapV3Factory.Contract.GetPool(&_UniswapV3Factory.CallOpts, tokenA, tokenB, fee)
}
