package tx

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_tryUntilDone(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		require.NoError(t, tryUntilDone(context.Background(), func() error {
			return nil
		}))
	})

	t.Run("retries", func(t *testing.T) {
		i := 0
		err := tryUntilDone(context.Background(), func() error {
			if i == 0 {
				i++
				return fmt.Errorf("some error")
			}

			return nil
		})

		require.NoError(t, err)
		require.Equal(t, 1, i)
	})

	t.Run("retries until ctx cancelled", func(t *testing.T) {
		i := 0
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := tryUntilDone(ctx, func() error {
			i++
			if i == 5 {
				cancel()
				return fmt.Errorf("ctx cancel")
			}
			return fmt.Errorf("an error")
		})

		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, 5, i)
	})
}
