package pipelines

import (
	"context"
)

// NewIterator is constructor for object, that generates data into a single output by traverse items from iterator
func NewIterator[T any](
	ctx context.Context,
	iterator func(yield func(request T) bool),
) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		iterator(
			func(v T) bool {
				select {
				case <-ctx.Done():
					return false
				case out <- v:
					return true
				}
			},
		)
	}()

	return out
}
