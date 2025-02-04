package pipelines

import (
	"context"
)

// NewGenerator is constructor for object, that generates data into a single output
func NewGenerator[T any](ctx context.Context, vals ...T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for _, v := range vals {
			select {
			case <-ctx.Done():
				return
			case out <- v:
			}
		}
	}()

	return out
}
