package types

import "github.com/NibiruChain/nibiru/x/common"

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
	GetPrice(pair common.AssetPair) Price
	// Close shuts down the PriceProvider.
	Close()
}
