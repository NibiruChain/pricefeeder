package types

// EventStream defines the asynchronous stream
// of events required by the feeder's Loop function.
// EventStream must handle failures by itself.
//
//go:generate mockgen --destination ../mocks/feeder/types/events_stream.go . EventStream
type EventStream interface {
	// ParamsUpdate signals a new Params update.
	// EventStream must provide, on startup, the
	// initial Params found on the chain.
	ParamsUpdate() <-chan Params
	// VotingPeriodStarted signals a new x/oracle
	// voting period has just started.
	VotingPeriodStarted() <-chan VotingPeriod
	// Close shuts down the EventStream.
	Close()
}
