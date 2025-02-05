package pipelines

import (
	"context"
)

func ExampleNewIterator() {
	values := NewIterator(
		context.Background(),
		func(yield func(request int) bool) {
			for i := 1; i <= 3; i++ {
				if !yield(i) {
					return
				}
			}
		},
	)
	Print(values)

	// Output:
	// 1
	// 2
	// 3
}
