package feeder

import (
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/pricefeeder/types"
	mocks "github.com/NibiruChain/pricefeeder/types/mocks"
)

func TestRunPanics(t *testing.T) {
	ctrl := gomock.NewController(t)
	pricePoster := mocks.NewMockPricePoster(ctrl)
	priceProvider := mocks.NewMockPriceProvider(ctrl)
	eventStream := mocks.NewMockEventStream(ctrl)
	eventStream.EXPECT().ParamsUpdate().Return(make(chan types.Params))

	f := NewFeeder(eventStream, priceProvider, pricePoster, zerolog.New(io.Discard))

	require.Panics(t, func() {
		f.Run()
	})
}

func TestParamsUpdate(t *testing.T) {
	tf := initFeeder(t)
	defer tf.feeder.Close()
	p := types.Params{
		Pairs:            []asset.Pair{asset.Registry.Pair(denoms.NIBI, denoms.NUSD)},
		VotePeriodBlocks: 50,
	}

	tf.paramsChannel <- p
	time.Sleep(10 * time.Millisecond)
	require.Equal(t, tf.feeder.params, p)
}

func TestVotingPeriod(t *testing.T) {
	tf := initFeeder(t)
	defer tf.feeder.Close()

	validPrice := types.Price{
		Pair:       asset.Registry.Pair(denoms.BTC, denoms.NUSD),
		Price:      100_000.8,
		SourceName: "mock-source",
		Valid:      true,
	}

	invalidPrice := types.Price{
		Pair:       asset.Registry.Pair(denoms.ETH, denoms.NUSD),
		Price:      7000.11,
		SourceName: "mock-source",
		Valid:      false,
	}

	abstainPrice := invalidPrice
	abstainPrice.Price = 0.0

	tf.mockPriceProvider.EXPECT().GetPrice(asset.Registry.Pair(denoms.BTC, denoms.NUSD)).Return(validPrice)
	tf.mockPriceProvider.EXPECT().GetPrice(asset.Registry.Pair(denoms.ETH, denoms.NUSD)).Return(invalidPrice)
	tf.mockPricePoster.EXPECT().SendPrices(gomock.Any(), []types.Price{validPrice, abstainPrice})
	// trigger voting period.
	tf.newVotingPeriod <- types.VotingPeriod{Height: 100}
	time.Sleep(10 * time.Millisecond)
}

type testFeederHarness struct {
	feeder            *Feeder
	mockPriceProvider *mocks.MockPriceProvider
	mockEventStream   *mocks.MockEventStream
	mockPricePoster   *mocks.MockPricePoster
	newVotingPeriod   chan types.VotingPeriod
	paramsChannel     chan types.Params
}

func initFeeder(t *testing.T) testFeederHarness {
	ctrl := gomock.NewController(t)
	pricePoster := mocks.NewMockPricePoster(ctrl)
	priceProvider := mocks.NewMockPriceProvider(ctrl)
	eventStream := mocks.NewMockEventStream(ctrl)

	paramsChannel := make(chan types.Params, 1)
	eventStream.EXPECT().ParamsUpdate().AnyTimes().Return(paramsChannel)
	paramsChannel <- types.Params{Pairs: []asset.Pair{asset.Registry.Pair(denoms.BTC, denoms.NUSD), asset.Registry.Pair(denoms.ETH, denoms.NUSD)}}

	votingPeriodChannel := make(chan types.VotingPeriod, 1)
	eventStream.EXPECT().VotingPeriodStarted().AnyTimes().Return(votingPeriodChannel)

	feeder := &Feeder{
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
		eventStream:   eventStream,
		pricePoster:   pricePoster,
		priceProvider: priceProvider,
		params:        types.Params{},
		logger:        zerolog.New(io.Discard),
	}
	feeder.Run()

	eventStream.EXPECT().Close()
	priceProvider.EXPECT().Close()
	pricePoster.EXPECT().Close()

	return testFeederHarness{
		feeder:            feeder,
		mockPriceProvider: priceProvider,
		mockEventStream:   eventStream,
		mockPricePoster:   pricePoster,
		newVotingPeriod:   votingPeriodChannel,
		paramsChannel:     paramsChannel,
	}
}
