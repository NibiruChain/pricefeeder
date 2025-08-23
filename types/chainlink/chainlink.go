// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package chainlink

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

// ChainlinkAggregatorMetaData contains all meta data concerning the ChainlinkAggregator contract.
var ChainlinkAggregatorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"description\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"latestRoundData\",\"outputs\":[{\"internalType\":\"uint80\",\"name\":\"roundId\",\"type\":\"uint80\"},{\"internalType\":\"int256\",\"name\":\"answer\",\"type\":\"int256\"},{\"internalType\":\"uint256\",\"name\":\"startedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint80\",\"name\":\"answeredInRound\",\"type\":\"uint80\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ChainlinkAggregatorABI is the input ABI used to generate the binding from.
// Deprecated: Use ChainlinkAggregatorMetaData.ABI instead.
var ChainlinkAggregatorABI = ChainlinkAggregatorMetaData.ABI

// ChainlinkAggregator is an auto generated Go binding around an Ethereum contract.
type ChainlinkAggregator struct {
	ChainlinkAggregatorCaller     // Read-only binding to the contract
	ChainlinkAggregatorTransactor // Write-only binding to the contract
	ChainlinkAggregatorFilterer   // Log filterer for contract events
}

// ChainlinkAggregatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type ChainlinkAggregatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChainlinkAggregatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ChainlinkAggregatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChainlinkAggregatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ChainlinkAggregatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChainlinkAggregatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ChainlinkAggregatorSession struct {
	Contract     *ChainlinkAggregator // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ChainlinkAggregatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ChainlinkAggregatorCallerSession struct {
	Contract *ChainlinkAggregatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// ChainlinkAggregatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ChainlinkAggregatorTransactorSession struct {
	Contract     *ChainlinkAggregatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// ChainlinkAggregatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type ChainlinkAggregatorRaw struct {
	Contract *ChainlinkAggregator // Generic contract binding to access the raw methods on
}

// ChainlinkAggregatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ChainlinkAggregatorCallerRaw struct {
	Contract *ChainlinkAggregatorCaller // Generic read-only contract binding to access the raw methods on
}

// ChainlinkAggregatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ChainlinkAggregatorTransactorRaw struct {
	Contract *ChainlinkAggregatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewChainlinkAggregator creates a new instance of ChainlinkAggregator, bound to a specific deployed contract.
func NewChainlinkAggregator(address common.Address, backend bind.ContractBackend) (*ChainlinkAggregator, error) {
	contract, err := bindChainlinkAggregator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ChainlinkAggregator{ChainlinkAggregatorCaller: ChainlinkAggregatorCaller{contract: contract}, ChainlinkAggregatorTransactor: ChainlinkAggregatorTransactor{contract: contract}, ChainlinkAggregatorFilterer: ChainlinkAggregatorFilterer{contract: contract}}, nil
}

// NewChainlinkAggregatorCaller creates a new read-only instance of ChainlinkAggregator, bound to a specific deployed contract.
func NewChainlinkAggregatorCaller(address common.Address, caller bind.ContractCaller) (*ChainlinkAggregatorCaller, error) {
	contract, err := bindChainlinkAggregator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ChainlinkAggregatorCaller{contract: contract}, nil
}

// NewChainlinkAggregatorTransactor creates a new write-only instance of ChainlinkAggregator, bound to a specific deployed contract.
func NewChainlinkAggregatorTransactor(address common.Address, transactor bind.ContractTransactor) (*ChainlinkAggregatorTransactor, error) {
	contract, err := bindChainlinkAggregator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ChainlinkAggregatorTransactor{contract: contract}, nil
}

// NewChainlinkAggregatorFilterer creates a new log filterer instance of ChainlinkAggregator, bound to a specific deployed contract.
func NewChainlinkAggregatorFilterer(address common.Address, filterer bind.ContractFilterer) (*ChainlinkAggregatorFilterer, error) {
	contract, err := bindChainlinkAggregator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ChainlinkAggregatorFilterer{contract: contract}, nil
}

// bindChainlinkAggregator binds a generic wrapper to an already deployed contract.
func bindChainlinkAggregator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ChainlinkAggregatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ChainlinkAggregator *ChainlinkAggregatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ChainlinkAggregator.Contract.ChainlinkAggregatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ChainlinkAggregator *ChainlinkAggregatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ChainlinkAggregator.Contract.ChainlinkAggregatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ChainlinkAggregator *ChainlinkAggregatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ChainlinkAggregator.Contract.ChainlinkAggregatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ChainlinkAggregator *ChainlinkAggregatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ChainlinkAggregator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ChainlinkAggregator *ChainlinkAggregatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ChainlinkAggregator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ChainlinkAggregator *ChainlinkAggregatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ChainlinkAggregator.Contract.contract.Transact(opts, method, params...)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ChainlinkAggregator *ChainlinkAggregatorCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ChainlinkAggregator.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ChainlinkAggregator *ChainlinkAggregatorSession) Decimals() (uint8, error) {
	return _ChainlinkAggregator.Contract.Decimals(&_ChainlinkAggregator.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ChainlinkAggregator *ChainlinkAggregatorCallerSession) Decimals() (uint8, error) {
	return _ChainlinkAggregator.Contract.Decimals(&_ChainlinkAggregator.CallOpts)
}

// Description is a free data retrieval call binding the contract method 0x7284e416.
//
// Solidity: function description() view returns(string)
func (_ChainlinkAggregator *ChainlinkAggregatorCaller) Description(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ChainlinkAggregator.contract.Call(opts, &out, "description")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Description is a free data retrieval call binding the contract method 0x7284e416.
//
// Solidity: function description() view returns(string)
func (_ChainlinkAggregator *ChainlinkAggregatorSession) Description() (string, error) {
	return _ChainlinkAggregator.Contract.Description(&_ChainlinkAggregator.CallOpts)
}

// Description is a free data retrieval call binding the contract method 0x7284e416.
//
// Solidity: function description() view returns(string)
func (_ChainlinkAggregator *ChainlinkAggregatorCallerSession) Description() (string, error) {
	return _ChainlinkAggregator.Contract.Description(&_ChainlinkAggregator.CallOpts)
}

// LatestRoundData is a free data retrieval call binding the contract method 0xfeaf968c.
//
// Solidity: function latestRoundData() view returns(uint80 roundId, int256 answer, uint256 startedAt, uint256 updatedAt, uint80 answeredInRound)
func (_ChainlinkAggregator *ChainlinkAggregatorCaller) LatestRoundData(opts *bind.CallOpts) (struct {
	RoundId         *big.Int
	Answer          *big.Int
	StartedAt       *big.Int
	UpdatedAt       *big.Int
	AnsweredInRound *big.Int
}, error) {
	var out []interface{}
	err := _ChainlinkAggregator.contract.Call(opts, &out, "latestRoundData")

	outstruct := new(struct {
		RoundId         *big.Int
		Answer          *big.Int
		StartedAt       *big.Int
		UpdatedAt       *big.Int
		AnsweredInRound *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.RoundId = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Answer = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.StartedAt = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.UpdatedAt = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.AnsweredInRound = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// LatestRoundData is a free data retrieval call binding the contract method 0xfeaf968c.
//
// Solidity: function latestRoundData() view returns(uint80 roundId, int256 answer, uint256 startedAt, uint256 updatedAt, uint80 answeredInRound)
func (_ChainlinkAggregator *ChainlinkAggregatorSession) LatestRoundData() (struct {
	RoundId         *big.Int
	Answer          *big.Int
	StartedAt       *big.Int
	UpdatedAt       *big.Int
	AnsweredInRound *big.Int
}, error) {
	return _ChainlinkAggregator.Contract.LatestRoundData(&_ChainlinkAggregator.CallOpts)
}

// LatestRoundData is a free data retrieval call binding the contract method 0xfeaf968c.
//
// Solidity: function latestRoundData() view returns(uint80 roundId, int256 answer, uint256 startedAt, uint256 updatedAt, uint80 answeredInRound)
func (_ChainlinkAggregator *ChainlinkAggregatorCallerSession) LatestRoundData() (struct {
	RoundId         *big.Int
	Answer          *big.Int
	StartedAt       *big.Int
	UpdatedAt       *big.Int
	AnsweredInRound *big.Int
}, error) {
	return _ChainlinkAggregator.Contract.LatestRoundData(&_ChainlinkAggregator.CallOpts)
}
