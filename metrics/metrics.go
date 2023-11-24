package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var NumVotingPeriods = promauto.NewCounter(prometheus.CounterOpts{
	Name: "pricefeeder_voting_periods_total",
	Help: "The total number of voting periods this pricefeeder has participated in",
})
