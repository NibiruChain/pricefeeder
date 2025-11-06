# NibiruChain/pricefeeder for the Oracle Module

<img src="./repo-banner.png">

The `pricefeeder` is a tool developed for Nibiru's [Oracle Module consensus](https://nibiru.fi/docs/ecosystem/oracle/) that runs a process to pull data from various external sources and then broadcasts transactions to vote on exchange rates.

- [Quick Start - Local Development](#quick-start---local-development)
  - [Configuration for the `.env`](#configuration-for-the-env)
  - [Run](#run)
    - [Or, to run the tool as a daemon](#or-to-run-the-tool-as-a-daemon)
- [Hacking](#hacking)
  - [Build](#build)
  - [Delegating "feeder" consent](#delegating-feeder-consent)
  - [Enabling TLS](#enabling-tls)
- [Configuring Price Sources](#configuring-price-sources)
  - [CoinGecko](#coingecko)
- [Uniswap V3 on Ethereum](#uniswap-v3-on-ethereum)
- [Chainlink on B^2 Network](#chainlink-on-b2-network)
- [Eris Protocol for stNIBI price](#eris-protocol-for-stnibi-price)
  - [Eris Protocol for stNIBI price](#eris-protocol-for-stnibi-price-1)
  - [Avalon Finance for sUSDa, USDa](#avalon-finance-for-susda-usda)
- [Glossary](#glossary)

## Quick Start - Local Development

### Configuration for the `.env`

Running a `feeder` requires the setting environment variables in your `.env` in like the following:

```ini
CHAIN_ID="nibiru-localnet-0"
GRPC_ENDPOINT="localhost:9090"
WEBSOCKET_ENDPOINT="ws://localhost:26657/websocket"
FEEDER_MNEMONIC="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
EXCHANGE_SYMBOLS_MAP='{"bitfinex": {"ubtc:unusd": "tBTCUSD", "ueth:unusd": "tETHUSD", "uusd:unusd": "tUSTUSD"}}'
```

This would allow you to run `pricefeeder` using a local instance of the network. To set up a local network, you can run:

```bash
git clone git@github.com:NibiruChain/nibiru.git
cd nibiru
git checkout v1.0.3
make localnet
```

### Run

With your environment set to a live network, you can now run the price feeder:

```sh
make run
```

#### Or, to run the tool as a daemon

1. Build a docker image for use with docker compose.

    ```bash
    make build-docker
    ```

2. Run the 'price_feeder' service defined in the `docker-compose.yaml`.

    ```bash
    make docker-compose up -d price_feeder
    ```

## Hacking

Connecters for data sources like Binance and Bitfinex are defined in the `feeder/sources` directory. Each of these sources must implement a `FetchPricesFunc` function for querying external data.

### Build

Builds the binary for the package:

```sh
make build
```

### Delegating "feeder" consent

Votes for exhange rates in the [Oracle Module](https://nibiru.fi/docs/ecosystem/oracle/) are posted by validator nodes, however a validator can give consent a `feeder` account to post prices on its behalf. This way, the validator won't have to use their validator's mnemonic to send transactions.  

In order to be able to delegate consent to post prices, you need to set the
`VALIDATOR_ADDRESS` env variable to the "valoper" address the `feeder` will represent.

```ini
VALIDATOR_ADDRESS="nibivaloper1..."
```

To delegate consent from a validator node to some `feeder` address, you must execute a `MsgDelegateFeedConsent` message:

```go
type MsgDelegateFeedConsent struct {
 Operator string 
 Delegate string
}
```

This is possible using the `set-feeder` subcommand of the `nibid` CLI:

```bash
nibid tx oracle set-feeder [feeder-address] --from validator
```

### Enabling TLS

To enable TLS, you need to set the following env vars:

```ini
TLS_ENABLED="true"
```

## Configuring Price Sources

### CoinGecko

Coingecko source allows to use paid api key to get more requests per minute. In order to configure it,
you need to set env var:

```ini
DATASOURCE_CONFIG_MAP='{"coingecko": {"api_key": "0123456789"}}'
```

## Uniswap V3 on Ethereum

Some token prices are retrieved from Uniswap V3 on Ethereum. 
To configure the Uniswap V3 data source, you could set the following environment variables.
If not set, the price feeder will use default public RPC endpoints.

```ini
# optional, used for Uniswap V3 prices
ETHEREUM_RPC_ENDPOINT="https://mainnet.infura.io/v3/<INFURA_API_KEY>"

# optional, used if custom exclusive ETHEREUM_RPC_ENDPOINT not set 
ETHEREUM_RPC_PUBLIC_ENDPOINTS="https://eth.llamarpc.com,https://cloudflare-eth.com/,https://rpc.flashbots.net/"
```

### Eris Protocol for stNIBI price

The price of stNIBI is fetched from the Eris Protocol (CosmWasm) by GRPC.

```bash
ERIS=nibi1udqqx30cw8nwjxtl4l28ym9hhrp933zlq8dqxfjzcdhvl8y24zcqpzmh8m
nibid q wasm contract-state smart $ERIS '{"state": {}}' | jq
{
  "data": {
    "total_ustake": "51448235827733",
    "total_utoken": "67090804824813",
    "exchange_rate": "1.304044808250702453",
    "unlocked_coins": [
      {
        "denom": "unibi",
        "amount": "509884"
      }
    ],
    "unbonding": "749733010863",
    "available": "552036314217",
    "tvl_utoken": "68392574149893"
  }
}
```

By default, the pricefeeder will use the local GRPC endpoint, the same as pricefeeder is using to submit the prices.
To allow fetching the mainnet price and submit it to testnet/localnet, you can set the following environment variable:

```ini
# Mainnet
GRPC_READ_ENDPOINT="grpc.nibiru.fi:443"

# Testnet-2
GRPC_READ_ENDPOINT="grpc-testnet-2.nibiru.fi:443"
```

Eris protocol contract address is defaulted to mainnet address.
To customize it, you can set the following environment variable:

```ini
# Mainnet
ERIS_PROTOCOL_CONTRACT_ADDRESS="nibi1udqqx30cw8nwjxtl4l28ym9hhrp933zlq8dqxfjzcdhvl8y24zcqpzmh8m"

# Testnet-2
ERIS_PROTOCOL_CONTRACT_ADDRESS="nibi1keqw4dllsczlldd7pmzp25wyl04fw5anh3wxljhg4fjuqj9xnxuqa82rpg"
```

### Avalon Finance for sUSDa, USDa


[Avalon Finance](https://avalonfinance.xyz) is a CeDeFi lending project that
powers both USDa and sUSDa. The redeem rate between USDa and its yield-bearing variant, sUSDa, is retrieved from the API provided by Avalon Labs. 

This data source adds queries for the "susda:usda" and "susda:usd" asset pairs.

### Chainlink on B^2 Network

Some token prices are retrieved from other chains Chainlink oracles.
To configure the Chainlink data source, you could set the following environment variables.
If not set, the price feeder will use default public RPC endpoints.

```ini
# Optional, used for Chainlink prices, defaults to pulic B^2 Network RPC endpoints
B2_RPC_ENDPOINT="https://mainnet.b2-rpc.com"

# Optional, used if custom exclusive B2_RPC_ENDPOINT not set, defaults to public endpoints (see in the code)
B2_RPC_PUBLIC_ENDPOINTS="https://rpc.bsquared.network,https://mainnet.b2-rpc.com"
```

## Glossary

- **Data source**: A data source is an external service that provides data. For example, Binance is a data source that provides the price of various assets.
- **Symbol**: A symbol is a string that represents a pair of assets on an external data source. For example, `tBTCUSD` is a symbol on Bitfinex that represents the price of Bitcoin in US Dollars.
- **Ticker**: Synonymous with **Symbol**. Exchanges generally use the term "ticker" to refer to a symbol.
- **Pair**: A pair is a combination of two assets recognized by Nibiru Chain. For example, `ubtc:uusd` is a pair that represents Bitcoin and USD.
