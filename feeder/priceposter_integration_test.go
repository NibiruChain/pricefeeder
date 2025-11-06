package feeder_test

import (
	"context"
	"fmt"
	"testing"

	oracletypes "github.com/NibiruChain/nibiru/v2/x/oracle/types"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/feeder"
	"github.com/NibiruChain/pricefeeder/types"
)

func (s *IntegrationSuite) TestClientWorks() {

	s.pricePosterClient.SendPrices(types.VotingPeriod{}, s.randomPrices())

	// assert vote was skipped because no previous prevote
	require.Contains(s.T(), s.logs.String(), "skipping vote preparation as there is no old prevote")
	require.NotContains(s.T(), s.logs.String(), "prepared vote message")

	// wait for next vote period
	s.waitNextVotePeriod()
	s.pricePosterClient.SendPrices(types.VotingPeriod{}, s.randomPrices())
	require.Contains(s.T(), s.logs.String(), "prepared vote message")
}

func (s *IntegrationSuite) randomPrices() []types.Price {
	vt, err := s.pricePosterClient.GetOracleClient().(oracletypes.QueryClient).VoteTargets(context.Background(), &oracletypes.QueryVoteTargetsRequest{})
	require.NoError(s.T(), err)
	prices := make([]types.Price, len(vt.VoteTargets))
	for i, assetPair := range vt.VoteTargets {
		prices[i] = types.Price{
			Pair:       assetPair,
			Price:      float64(i),
			SourceName: "test",
			Valid:      true,
		}
	}
	return prices
}

func (s *IntegrationSuite) waitNextVotePeriod() {
	params, err := s.pricePosterClient.GetOracleClient().(oracletypes.QueryClient).Params(context.Background(), &oracletypes.QueryParamsRequest{})
	require.NoError(s.T(), err)
	height, err := s.network.LatestHeight()
	require.NoError(s.T(), err)
	targetHeight := height + int64(uint64(height)%params.Params.VotePeriod)
	_, err = s.network.WaitForHeight(targetHeight)
	require.NoError(s.T(), err)
}

func Test_tryUntilDone(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		require.NoError(t, feeder.TryUntilDone(context.Background(), 0, func() error {
			return nil
		}))
	})

	t.Run("retries", func(t *testing.T) {
		i := 0
		err := feeder.TryUntilDone(context.Background(), 0, func() error {
			if i == 0 {
				i++
				return fmt.Errorf("some error")
			}
			return nil
		})

		require.NoError(t, err)
		require.Equal(t, 1, i)
	})

	t.Run("retries until ctx cancelled", func(t *testing.T) {
		i := 0
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := feeder.TryUntilDone(ctx, 0, func() error {
			i++
			if i == 5 {
				cancel()
				return fmt.Errorf("ctx cancel")
			}
			return fmt.Errorf("an error")
		})
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, 5, i)
	})
}
