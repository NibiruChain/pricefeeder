package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var NumVotingPeriods = promauto.NewCounter(prometheus.CounterOpts{
	Name: "pricefeeder_voting_periods_total",
	Help: "The total number of voting periods this pricefeeder has participated in",
})

var PriceSourceLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "pricefeeder_price_source_latency_seconds",
	Help:    "The latency of querying a price source in seconds",
	Buckets: prometheus.DefBuckets,
}, []string{"source"})
