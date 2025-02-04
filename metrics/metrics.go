package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const PrometheusNamespace = "pricefeeder"

var PriceSourceCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: PrometheusNamespace,
	Name:      "fetched_prices_total",
	Help:      "The total number prices fetched, by source and success status",
}, []string{"source", "success"})
