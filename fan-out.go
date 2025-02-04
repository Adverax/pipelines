package pipelines

import (
	"context"
)

// NewFanOut is used to distribute data from one input to many outputs.
func NewFanOut[T any](ctx context.Context, income <-chan T, workers int) []<-chan T {
	outcomes := make([]<-chan T, workers)
	outputs := make([]chan T, workers)
	for i := 0; i < workers; i++ {
		ch := make(chan T)
		outcomes[i] = ch
		outputs[i] = ch
	}

	for _, outcome := range outputs {
		go func(outcome chan T) {
			defer close(outcome)

			for v := range income {
				select {
				case <-ctx.Done():
					return
				case outcome <- v:
				}
			}
		}(outcome)
	}

	return outcomes
}
