package tx

import "context"

func tryUntilDone(ctx context.Context, f func() error) error {
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
	}
}
