package feeder_test

import (
	"time"
)

func (s *IntegrationSuite) TestStreamWorks() {
	// Use the separate testEventStream, not the one consumed by the feeder
	select {
	case <-s.testEventStream.ParamsUpdate():
	case <-time.After(15 * time.Second):
		s.T().Fatal("params timeout")
	}
	select {
	case <-s.testEventStream.VotingPeriodStarted():
	case <-time.After(15 * time.Second):
		s.T().Fatal("vote period timeout")
	}
	<-time.After(10 * time.Second)
	// assert if params don't change, then no updates are sent
	s.Require().Contains(s.logs.String(), "skipping params update as they're not different from the old ones")
	// assert new voting period was signaled
	s.Require().Contains(s.logs.String(), "signaled new voting period")
}
