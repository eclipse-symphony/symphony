package expectations

import (
	"context"
	"fmt"
	"strings"

	econtext "github.com/eclipse-symphony/symphony/packages/testutils/internal/context"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/google/uuid"
)

type (
	AllExpectation struct {
		expectations []types.Expectation
		successCache map[string]struct{}
		shouldCache  bool
		level        int
		id           string
	}
	AnyExpectation struct {
		level        int
		id           string
		expectations []types.Expectation
	}
)

var (
	_ types.Expectation = &AllExpectation{}
	_ types.Expectation = &AnyExpectation{}
)

// Verify implements types.Expectation.
func (e *AnyExpectation) Verify(c context.Context) error {
	ctx := econtext.From(c)
	e.level = ctx.Level()
	e.log("checking if any expectation is satisfied")
	for i, expectation := range e.expectations {
		e.log("checking expectation %d of %d: [%s]", i+1, len(e.expectations), expectation.Description())
		if err := expectation.Verify(ctx.Nested()); err == nil {
			e.log("expectation %d of %d was satisfied", i+1, len(e.expectations))
			return nil
		}
	}
	return fmt.Errorf("no expectation was satisfied")
}

// Verify implements types.Expectation.
func (e *AllExpectation) Verify(c context.Context) error {
	ctx := econtext.From(c)
	e.level = ctx.Level()
	e.log("checking if all expectations are satisfied")
	for i, expectation := range e.expectations {
		if _, ok := e.successCache[expectation.Id()]; ok && e.shouldCache {
			e.log("expectation %d of %d was satisfied (cached) [%s]: skipping...", i+1, len(e.expectations), expectation.Description())
			continue
		}
		e.log("checking expectation %d of %d: [%s]", i+1, len(e.expectations), expectation.Description())
		if err := expectation.Verify(ctx.Nested()); err != nil {
			e.log("expectation %d of %d failed: %s", i+1, len(e.expectations), err)
			return err
		}
		e.log("expectation %d of %d was satisfied [%s]", i+1, len(e.expectations), expectation.Description())
		if e.shouldCache {
			e.successCache[expectation.Id()] = struct{}{}
		}
	}
	e.log("all expectations were satisfied")
	return nil
}

// Description implements types.Expectation.
func (e *AnyExpectation) Description() string {
	return "any expectation"
}

// Id implements types.Expectation.
func (e *AnyExpectation) Id() string {
	return e.id
}

// Id implements types.Expectation.
func (e *AllExpectation) Id() string {
	return e.id
}

// Description implements types.Expectation.
func (e *AllExpectation) Description() string {
	return "all expectation"
}

// WithCaching returns a new expectation that caches the result of each expectation successfull expectation
// so that it is not verified again in future calls to Verify.
func (a *AllExpectation) WithCaching() *AllExpectation {
	na := All(a.expectations...)
	na.shouldCache = true
	return na
}

// All returns an expectation that is satisfied if all of the given expectations are satisfied.
func All(expectations ...types.Expectation) *AllExpectation {
	return &AllExpectation{
		id:           uuid.NewString(),
		expectations: expectations,
		successCache: make(map[string]struct{}),
	}
}

// Any returns an expectation that is satisfied if any of the given expectations is satisfied.
func Any(expectations ...types.Expectation) *AnyExpectation {
	return &AnyExpectation{
		id:           uuid.NewString(),
		expectations: expectations,
	}
}

func (e *AnyExpectation) log(str string, args ...interface{}) {
	s := fmt.Sprintf(str, args...)
	logger.GetDefaultLogger()("%s[%s]: %s\n", strings.Repeat(" ", e.level), e.Description(), s)
}

func (e *AllExpectation) log(str string, args ...interface{}) {
	s := fmt.Sprintf(str, args...)
	logger.GetDefaultLogger()("%s[%s]: %s\n", strings.Repeat(" ", e.level), e.Description(), s)
}
