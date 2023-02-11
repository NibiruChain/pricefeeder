package types

// Source defines a source for price provision.
// This source has no knowledge of nibiru internals
// and mappings across asset.Pair and the Source
// symbols.
type Source interface {
	// PriceUpdates is a readonly channel which provides
	// the latest prices update. Updates can be provided
	// for one asset only or in batches, hence the map.
	PriceUpdates() <-chan map[Symbol]RawPrice
	// Close closes the Source.
	Close()
}
