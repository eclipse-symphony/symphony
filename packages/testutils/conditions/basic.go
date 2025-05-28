package conditions

import (
	"context"
	"fmt"

	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/google/uuid"
)

type (
	basic struct {
		id      string
		desc    string
		fn      func(context.Context, interface{}) error
		failMsg func(interface{}, error) string
	}
	basicopts func(*basic)
)

var (
	_ types.Condition = basic{}
)

// IsSatisfiedBy implements types.Condition.
func (b basic) IsSatisfiedBy(c context.Context, resource interface{}) error {
	if err := b.fn(c, resource); err != nil {
		return fmt.Errorf("%s", b.failMsg(resource, err))
	}
	return nil
}

// Id implements types.Condition.
func (ec basic) Id() string {
	return ec.id
}

// Description implements types.Condition.
func (ec basic) Description() string {
	if ec.desc == "" {
		return "basic condition"
	}
	return ec.desc
}

// WithBasicDescription sets the description of the condition.
func WithBasicDescription(desc string) basicopts {
	return func(b *basic) {
		b.desc = desc
	}
}

// WithBasicFailureMessage sets the failure message of the condition.
func WithBasicFailureMessage(msg func(interface{}, error) string) basicopts {
	return func(b *basic) {
		b.failMsg = msg
	}
}

// Basic returns a basic condition.
func Basic(fn func(context.Context, interface{}) error, opts ...basicopts) basic {
	b := basic{
		id:      uuid.New().String(),
		desc:    "",
		fn:      fn,
		failMsg: func(i interface{}, e error) string { return fmt.Sprintf("condition failed: %s", e.Error()) },
	}

	for _, opt := range opts {
		opt(&b)
	}

	return b
}
