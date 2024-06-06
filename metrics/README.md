# Pricefeeder Metrics

## Available metrics

### `fetched_prices_total`

The total number of prices fetched from the data sources. This metric is incremented every time the price feeder fetches prices from the data sources.

**labels**:

- `source`: The data source from which the price was fetched, e.g. `Bybit`.
- `success`: The result of the fetch operation. Possible values are 'true' and 'false'.

### `aggregate_prices_total`

The total number of times the `AggregatePriceProvider` is called to return a price. It randomly selects a source for each pair from its map of price providers. This metric is incremented every time the `AggregatePriceProvider` is called.

**labels**:

- `pair`: The pair for which the price was aggregated.
- `source`: The data source from which the price was fetched, e.g. `Bybit`.
- `success`: The result of the fetch operation. Possible values are 'true' and 'false'.

### `prices_posted_total`

The total number of txs sent to the on-chain oracle module. This metric is incremented every time the price feeder posts a price to the on-chain oracle module.

**labels**:

- `success`: The result of the post operation. Possible values are 'true' and 'false'.
