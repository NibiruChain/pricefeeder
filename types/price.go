package types

import "github.com/NibiruChain/nibiru/x/common"

// Price defines the price of a symbol.
type Price struct {
	// Pair defines the symbol we're posting prices for.
	Pair common.AssetPair
	// Price defines the symbol's price.
	Price float64
	// ExchangeName defines the source which is providing the prices.
	ExchangeName string
	// Valid reports whether the price is valid or not.
	// If not valid then an abstain vote will be posted.
	Valid bool
}
