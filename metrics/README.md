# Metrics in metrics.go

This file defines two metrics that are used to monitor the performance and behavior of the pricefeeder application.

## pricefeeder_voting_periods_total

This is a counter metric that keeps track of the total number of voting periods the pricefeeder has participated in. A voting period is a specific duration during which votes are collected. This metric is incremented each time the pricefeeder participates in a voting period.

```go
var NumVotingPeriods = promauto.NewCounter(prometheus.CounterOpts{
    Name: "pricefeeder_voting_periods_total",
    Help: "The total number of voting periods this pricefeeder has participated in",
})
```

## pricefeeder_price_source_latency_seconds

This is a histogram metric that measures the latency of querying a price source in seconds. The latency is the amount of time it takes for the pricefeeder to receive a response after it has sent a request to the price source. This metric is useful for monitoring the performance of the price sources and identifying any potential issues or delays.

```go
var PriceSourceLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
    Name:    "pricefeeder_price_source_latency_seconds",
    Help:    "The latency of querying a price source in seconds",
    Buckets: prometheus.DefBuckets,
}, []string{"source"})
```

The source label is used to differentiate the latencies of different price sources.
