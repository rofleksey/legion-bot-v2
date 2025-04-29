package unknown

import (
	"context"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPositive(t *testing.T) {
	detector := New()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := detector.Detect(ctx, "rofleksey")
	require.NoError(t, err)

	assert.Equal(t, true, res)
}

func TestNegative(t *testing.T) {
	detector := New()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := detector.Detect(ctx, "sbinka")
	require.NoError(t, err)

	assert.Equal(t, false, res)
}
