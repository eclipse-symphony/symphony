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
	AllCondition struct {
		id           string
		conditions   []types.Condition
		successCache map[string]struct{}
		shouldCache  bool
		level        int
	}
)

var (
	_ types.Condition = &AllCondition{}
)

// Description implements types.Condition.
func (a *AllCondition) Description() string {
	return "all group"
}

// Id implements types.Condition.
func (a *AllCondition) Id() string {
	return a.id
}

// IsSatisfiedBy implements types.Condition.
func (c *AllCondition) IsSatisfiedBy(oc context.Context, resource interface{}) error {
	ctx := ectx.From(oc)
	c.level = ctx.Level()
	c.log("checking if all conditions are satisfied")
	for i, condition := range c.conditions {
		if _, ok := c.successCache[condition.Id()]; c.shouldCache && ok {
			c.log("condition %d of %d was satisfied (cached) [%s]: skipping...", i+1, len(c.conditions), condition.Description())
			continue
		}
		c.log("checking condition %d of %d: [%s]", i+1, len(c.conditions), condition.Description())
		if err := condition.IsSatisfiedBy(ctx.Nested(), resource); err != nil {
			c.log("condition %d of %d failed: %s", i+1, len(c.conditions), err)
			return err
		}
		c.log("condition %d of %d was satisfied: [%s]", i+1, len(c.conditions), condition.Description())
		if c.shouldCache {
			c.successCache[condition.Id()] = struct{}{}
		}
	}
	c.log("all conditions were satisfied")
	return nil
}

func (c *AllCondition) log(str string, args ...interface{}) {
	s := fmt.Sprintf(str, args...)
	logger.GetDefaultLogger()("%s[%s]: %s\n", strings.Repeat(" ", c.level), c.Description(), s)
}

// All returns a new AllCondition.
func All(conditions ...types.Condition) *AllCondition {
	return &AllCondition{
		conditions:   conditions,
		successCache: make(map[string]struct{}),
		id:           uuid.NewString(),
	}
}

// WithCaching returns a new AllCondition with caching enabled. This means that
// if a condition is satisfied, it will not be checked again.
func (a *AllCondition) WithCaching() *AllCondition {
	na := All(a.conditions...)
	na.shouldCache = true
	return na
}

// For internal use only.
func (a *AllCondition) And(conditions ...types.Condition) types.Condition {
	na := All(a.conditions...)
	na.conditions = append(na.conditions, conditions...)
	na.shouldCache = a.shouldCache
	return na
}
