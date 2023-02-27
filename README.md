# Nibiru x/oracle Price Feeder

Submits prices to the nibiru decentralized oracle.

## Configuration using `.env`

Feeder requires the following environment variables to run:

```ini
CHAIN_ID="nibiru-localnet-0"
GRPC_ENDPOINT="localhost:9090"
WEBSOCKET_ENDPOINT="ws://localhost:26657/websocket"
FEEDER_MNEMONIC="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
EXCHANGE_SYMBOLS_MAP='{ "bitfinex": { "ubtc:uusd": "tBTCUSD", "ueth:uusd": "tETHUSD", "uusdt:uusd": "tUSTUSD" }, "binance": { "ubtc:uusd": "BTCUSD", "ueth:uusd": "ETHUSD", "uusdt:uusd": "USDTUSD", "uusdc:uusd": "USDCUSD", "uatom:uusd": "ATOMUSD", "ubnb:uusd": "BNBUSD", "uavax:uusd": "AVAXUSD", "usol:uusd": "SOLUSD", "uada:uusd": "ADAUSD", "ubtc:unusd": "BTCUSD", "ueth:unusd": "ETHUSD", "uusdt:unusd": "USDTUSD", "uusdc:unusd": "USDCUSD", "uatom:unusd": "ATOMUSD", "ubnb:unusd": "BNBUSD", "uavax:unusd": "AVAXUSD", "usol:unusd": "SOLUSD", "uada:unusd": "ADAUSD" } }'
```

### Delegating post pricing

In order to be able to delegate the post pricing you need to set the
env variable for the validator that delegated you the post pricing:

```ini
VALIDATOR_ADDRESS="nibiruvaloper1..."
```

And from your validator node, you need to delegate responsibilites to the feeder address

```sh
nibid tx oracle set-feeder <feeder address> --from validator
```

### Configuring specific exchanges

#### CoinGecko

Coingecko source allows to use paid api key to get more requests per minute. In order to configure it,
you need to set env var:

```ini
DATASOURCE_CONFIG_MAP='{"coingecko": {"api_key": "0123456789"}}'
```

## Build

```sh
make build-feeder
```

## Run

```sh
make run
```

or to run as a daemon:

```sh
make docker-compose up -d price_feeder
```
