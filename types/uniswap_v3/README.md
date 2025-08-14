# Go Types Generator for Uniswap V3

## Install abigen
```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```

## Generate Go types
```bash
abigen --abi factory.abi --pkg uniswap_v3 --type UniswapV3Factory --out factory.go
abigen --abi pool.abi --pkg uniswap_v3 --type UniswapV3Pool --out pool.go
```