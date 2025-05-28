package context

import (
	"context"
)

type (
	key int

	nestContext struct {
		context.Context
		level int
	}
)

const (
	nestKey key = iota
)

func (c *nestContext) Level() int {
	return c.level
}

func (c *nestContext) Nested() *nestContext {
	return &nestContext{
		Context: c,
		level:   c.level + 2,
	}
}

func (c *nestContext) Value(key any) any {
	if key == nestKey {
		return c
	}
	return c.Context.Value(key)
}

// From returns a context.Context that wraps the given context. This context
// allows to track the nesting level of the context. Useful for logging.
func From(ctx context.Context) *nestContext {
	if ctx == nil {
		return &nestContext{
			Context: context.Background(),
		}
	}
	if ec, ok := ctx.(*nestContext); ok {
		return ec
	}
	ec, ok := ctx.Value(nestKey).(*nestContext)
	if !ok {
		return &nestContext{
			Context: ctx,
		}
	}
	return ec
}
