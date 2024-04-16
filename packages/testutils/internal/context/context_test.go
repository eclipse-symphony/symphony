package context

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	c, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	ctx := From(c)
	require.Equal(t, 0, ctx.Level())

	ctx = ctx.Nested()
	require.Equal(t, 2, ctx.Level())

	c, cancel = context.WithTimeout(ctx, time.Millisecond) // embeds regular context
	defer cancel()

	c, cancel = context.WithTimeout(c, time.Millisecond) // embeds expectation context further
	defer cancel()

	ctx = From(c)
	require.Equal(t, 2, ctx.Level())

	ctx = From(nil)
	require.NotNil(t, ctx)
	require.Equal(t, 0, ctx.Level())

	ctx2 := From(ctx)
	require.Equal(t, ctx2, ctx)

}
