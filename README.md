# Nibiru x/oracle Price Feeder

Submits prices to the nibiru decentralized oracle.


## Configuration using `.env`

Feeder requires the following environment variables to run:

```ini
CHAIN_ID="nibiru-localnet-0"
GRPC_ENDPOINT="localhost:9090"
WEBSOCKET_ENDPOINT="ws://localhost:26657/websocket"
FEEDER_MNEMONIC="..."
EXCHANGE_SYMBOLS_MAP='{"bitfinex": {"ubtc:unusd": "tBTCUSD", "ueth:unusd": "tETHUSD", "uusd:unusd": "tUSTUSD"}}'
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
```
make docker-compose up -d price_feeder
```


