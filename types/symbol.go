package types

// A symbol refers to the ticker name used by the third party data source/exchange.
type Symbol string

type Symbols []Symbol

func (s Symbols) CSV() string {
	var csv string
	for _, symbol := range s {
		csv += string(symbol) + ","
	}
	return csv[:len(csv)-1]
}
