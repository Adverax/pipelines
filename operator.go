package pipelines

import (
	"context"
)

type Op[S, D any] interface {
	Execute(context.Context, S) D
}

type OpFunc[S, D any] func(context.Context, S) D

func (fn OpFunc[S, D]) Execute(ctx context.Context, s S) D {
	return fn(ctx, s)
}

func NewOperation[S, D any](
	ctx context.Context,
	op Op[S, D],
	income <-chan S,
) <-chan D {
	outcome := make(chan D)

	go func() {
		defer close(outcome)

		for value := range income {
			result := op.Execute(ctx, value)

			select {
			case <-ctx.Done():
				return
			case outcome <- result:
			}
		}
	}()

	return outcome
}

func NewOperations[S, D any](
	ctx context.Context,
	op Op[S, D],
	incomes ...<-chan S,
) []<-chan D {
	outcomes := make([]<-chan D, 0, len(incomes))

	for _, income := range incomes {
		outcomes = append(
			outcomes,
			NewOperation[S, D](ctx, op, income),
		)
	}

	return outcomes
}
