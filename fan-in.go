package pipelines

import (
	"context"
	"sync"
)

// NewFanIn is used to combine data from many sources into one.
func NewFanIn[T any](ctx context.Context, incomes ...<-chan T) <-chan T {
	outcome := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(incomes))

	for _, income := range incomes {
		go func(income <-chan T) {
			defer wg.Done()

			for v := range income {
				select {
				case <-ctx.Done():
					return
				case outcome <- v:
				}
			}
		}(income)
	}

	go func() {
		wg.Wait()
		close(outcome)
	}()

	return outcome
}
