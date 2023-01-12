package types

import (
	"time"

	"github.com/NibiruChain/nibiru/x/common"
)

const (
	PriceTimeout = 15 * time.Second
)

// Price defines the price of a symbol.
type Price struct {
	// Pair defines the symbol we're posting prices for.
	Pair common.AssetPair
	// Price defines the symbol's price.
	Price float64
	// SourceName defines the source which is providing the prices.
	SourceName string
	// When the price was fetched
	UpdateTime time.Time

	// Valid reports whether the price is valid or not.
	// If not valid then an abstain vote will be posted.
	// Computed from the update time.
	Valid bool
}
