package conditions

import (
	"context"
	"fmt"
	"strings"

	ectx "github.com/eclipse-symphony/symphony/packages/testutils/internal/context"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"

	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/google/uuid"
)

type (
	AnyCondition struct {
		id         string
		conditions []types.Condition
		level      int
	}
)

var (
	_ types.Condition = &AnyCondition{}
)

// Description implements types.Condition.
func (a *AnyCondition) Description() string {
	return "any group"
}

// Id implements types.Condition.
func (a *AnyCondition) Id() string {
	return a.id
}

// IsSatisfiedBy implements types.Condition.
func (c *AnyCondition) IsSatisfiedBy(oc context.Context, resource interface{}) error {
	ctx := ectx.From(oc)
	c.level = ctx.Level()
	c.log("checking if any condition is satisfied")
	for i, condition := range c.conditions {
		c.log("checking condition %d of %d: [%s]", i+1, len(c.conditions), condition.Description())
		if err := condition.IsSatisfiedBy(ctx.Nested(), resource); err == nil {
			c.log("condition %d of %d was satisfied: [%s]", i+1, len(c.conditions), condition.Description())
			return nil
		}
	}
	return fmt.Errorf("none of the conditions were satisfied")
}

func (c *AnyCondition) log(str string, args ...interface{}) {
	s := fmt.Sprintf(str, args...)
	logger.GetDefaultLogger()("%s[%s]: %s\n", strings.Repeat(" ", c.level), c.Description(), s)
}

// Any returns a condition that is satisfied if any of the given conditions are satisfied.
func Any(conditions ...types.Condition) *AnyCondition {
	return &AnyCondition{
		conditions: conditions,
		id:         uuid.NewString(),
	}
}
