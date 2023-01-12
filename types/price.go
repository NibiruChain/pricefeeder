package types

import (
	"time"

	"github.com/NibiruChain/nibiru/x/common"
)

const (
	PriceTimeout = 15 * time.Second
)

type RawPrice struct {
	Price      float64
	UpdateTime time.Time
}

// Price defines the price of a symbol.
type Price struct {
	// Pair defines the symbol we're posting prices for.
	Pair common.AssetPair
	// Price defines the symbol's price.
	Price float64
	// SourceName defines the source which is providing the prices.
	SourceName string
	// Valid reports whether the price is valid or not.
	// If not valid then an abstain vote will be posted.
	// Computed from the update time.
	Valid bool
}

// FetchPricesFunc is the function used to fetch updated prices.
// The symbols passed are the symbols we require prices for.
// The returned map must map symbol to its float64 price, or an error.
// If there's a failure in updating only one price then the map can be returned
// without the provided symbol.
type FetchPricesFunc func(symbols Symbols) (map[Symbol]float64, error)
