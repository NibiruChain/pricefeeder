package priceposter

import (
	"context"
	"time"
)

// tryUntilDone will try to execute the given function until it succeeds or the
// context is cancelled.
func tryUntilDone(ctx context.Context, wait time.Duration, f func() error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := f(); err == nil {
				return nil
			}
			time.Sleep(wait)
		}
	}
}
