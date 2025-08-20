# Go Types Generator for Chainlink Aggregator

## Install abigen
```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```

## Generate Go types
```bash
abigen --abi chainlink.abi --pkg chainlink --type ChainlinkAggregator --out chainlink.go
```