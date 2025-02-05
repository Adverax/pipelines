package pipelines

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func ExampleNewFanIn() {
	ctx := context.Background()
	values := NewFanIn(
		context.Background(),
		NewGenerator(ctx, 1, 2, 3),
		NewGenerator(ctx, 4, 5, 6),
		NewGenerator(ctx, 7, 8, 9),
	)

	var count int
	for _ = range values {
		count++
	}
	fmt.Println(count)

	// Output:
	// 9
}

func TestFanIn(t *testing.T) {
	ctx := context.Background()

	values := NewFanIn(
		ctx,
		NewGenerator(ctx, 1, 2, 3),
		NewGenerator(ctx, 4, 5, 6),
		NewGenerator(ctx, 7, 8, 9),
	)

	var count int
	for _ = range values {
		count++
	}

	require.Equal(t, 9, count)
}
