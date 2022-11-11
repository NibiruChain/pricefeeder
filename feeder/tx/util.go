package tx

import (
	"context"
	"time"
)

func tryUntilDone(ctx context.Context, wait time.Duration, f func() error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		err := f()
		if err == nil {
			return nil
		}
		time.Sleep(wait)
	}
}
