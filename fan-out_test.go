package pipelines

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFanOut(t *testing.T) {
	ctx := context.Background()

	fanOut := NewFanOut(
		ctx,
		NewGenerator(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9),
		3,
	)

	values := NewFanIn[int](ctx, fanOut...)

	var count int
	for _ = range values {
		count++
	}

	require.Equal(t, 9, count)
}
