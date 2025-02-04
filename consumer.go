package pipelines

import (
	"context"
)

type Aggregator[T any] interface {
	Aggregate(ctx context.Context, val T) error
}

func Consume[T any](
	ctx context.Context,
	income <-chan T,
	aggregator Aggregator[T],
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case val, ok := <-income:
			if !ok {
				return nil
			}

			err := aggregator.Aggregate(ctx, val)
			if err != nil {
				return err
			}
		}
	}
}
