package pipelines

import "context"

func ExampleNewOperation() {
	values := NewOperation[int, int](
		context.Background(),
		OpFunc[int, int](func(ctx context.Context, value int) int {
			return value * 2
		}),
		NewGenerator(context.Background(), 1, 2, 3))

	Print(values)

	// Output:
	// 2
	// 4
	// 6
}
