package tx

import (
	"bytes"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient_handleFailure(t *testing.T) {
	t.Run("no tx response", func(t *testing.T) {
		logs := new(bytes.Buffer)
		c := &exContext{
			log: zerolog.New(logs),
		}

		c.handleFailure(nil, fmt.Errorf("some failure"))
		require.Contains(t, logs.String(), "some failure")
		require.NotContains(t, logs.String(), "tx-response")
	})

	t.Run("with tx response", func(t *testing.T) {
		logs := new(bytes.Buffer)
		c := &exContext{
			log: zerolog.New(logs),
		}
		c.handleFailure(&sdk.TxResponse{
			RawLog: "some log that we can check exist in the test",
		}, fmt.Errorf("some failure"))
		require.Contains(t, logs.String(), "tx-response")
		require.Contains(t, logs.String(), "some log that we can check exist in the test")
		require.Contains(t, logs.String(), "some failure")
	})
}
