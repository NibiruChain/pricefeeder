package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// PricePoster defines the validator oracle client,
// which sends new prices.
// PricePoster must handle failures by itself.
//
//go:generate mockgen --destination mocks/price_poster.go . PricePoster
type PricePoster interface {
	// Whoami returns the validator address the PricePoster
	// is sending prices for.
	Whoami() sdk.ValAddress
	// SendPrices sends the provided slice of Price.
	SendPrices(vp VotingPeriod, prices []Price)
	// Close shuts down the PricePoster.
	Close()
}
