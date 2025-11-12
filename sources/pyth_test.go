package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/v2/x/common/set"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/pricefeeder/types"
)

func TestPythPriceUpdate_CustomEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/updates/price/latest", r.URL.Path)
		require.ElementsMatch(t, []string{"feed-1", "feed-2"}, r.URL.Query()[pythQueryParamIDs])
		require.Equal(t, pythQueryValueParsed, r.URL.Query().Get(pythQueryParamParsed))

		publishTime := time.Now().Add(-30 * time.Second).Unix()
		fmt.Fprintf(w, `{"parsed":[
			{"id":"feed-1","price":{"price":"123456","expo":-2,"publish_time":%d}},
			{"id":"feed-2","price":{"price":"987654321","expo":-4,"publish_time":%d}}
		]}`, publishTime, publishTime)
	}))
	defer server.Close()

	cfgBytes, err := json.Marshal(pythConfig{
		Endpoint:           server.URL,
		TimeoutSeconds:     1,
		MaxPriceAgeSeconds: 60,
	})
	require.NoError(t, err)

	fetchPrices := PythPriceUpdate(cfgBytes)
	symbols := set.New[types.Symbol]("feed-1", "feed-2")
	logger := zerolog.New(io.Discard)

	prices, err := fetchPrices(symbols, logger)
	require.NoError(t, err)
	require.Len(t, prices, 2)
	assert.InDelta(t, 1234.56, prices["feed-1"], 1e-9)
	assert.InDelta(t, 98765.4321, prices["feed-2"], 1e-9)
}

func TestPythPriceUpdate_EmptyParsedReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, writeErr := w.Write([]byte(`{"parsed":[]}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	cfgBytes, err := json.Marshal(pythConfig{Endpoint: server.URL})
	require.NoError(t, err)

	fetchPrices := PythPriceUpdate(cfgBytes)
	logger := zerolog.New(io.Discard)
	_, err = fetchPrices(set.New[types.Symbol]("feed-1"), logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing parsed prices")
}

func TestConvertPythPrice(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		price, err := convertPythPrice("123456", -2)
		require.NoError(t, err)
		assert.InDelta(t, 1234.56, price, 1e-9)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := convertPythPrice("not-a-number", 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid price string")
	})
}
