package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var PriceSourceCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "price_source_counter",
	Help: "The total number prices scraped, by source and success status",
}, []string{"source", "success"})
