package types

import "github.com/NibiruChain/nibiru/v2/x/common/asset"

// PriceProvider defines an exchange API
// which provides prices for the given assets.
// PriceProvider must handle failures by itself.
//
//go:generate mockgen --destination mocks/price_provider.go . PriceProvider
type PriceProvider interface {
	// GetPrice returns the Price for the given symbol.
	// Price.Pair, Price.Source must always be non-empty.
	// If there are errors whilst fetching prices, then
	// Price.Valid must be set to false.
	GetPrice(pair asset.Pair) Price
	// Close shuts down the PriceProvider.
	Close()
}

var _ PriceProvider = (*NullPriceProvider)(nil)

// NullPriceProvider is a no-op implementation of [PriceProvider] that always
// returns invalid prices with [PriceAbstain] values.
//
// Use [NullPriceProvider] as a fallback when source initialization fails,
// allowing the application to continue operating with other available providers
// rather than crashing.
//
// [NullPriceProvider] implements the null object pattern for graceful error handling.
type NullPriceProvider struct{}

func (pp NullPriceProvider) GetPrice(pair asset.Pair) Price {
	return Price{
		Pair:       pair,
		Price:      PriceAbstain,
		SourceName: "null",
		Valid:      false,
	}
}

func (pp NullPriceProvider) Close() {}
