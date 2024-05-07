package jsonpath

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	ectx "github.com/eclipse-symphony/symphony/packages/testutils/internal/context"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/google/uuid"
	"github.com/oliveagle/jsonpath"
)

type (
	JpCondition struct {
		id          string
		matcher     func(ctx context.Context, values, root interface{}, log logger.Logger) error
		jp          *jsonpath.Compiled
		description string
		l           func(format string, args ...interface{})
		level       int
	}

	Option func(*JpCondition)
)

var (
	_ types.Condition = &JpCondition{}
)

// WithCustomMatcher specifies the matcher to be used to match the jsonpath result.
func WithCustomMatcher(matcher func(ctx context.Context, value, root interface{}, log logger.Logger) error) Option {
	return func(j *JpCondition) {
		j.matcher = matcher
	}
}

// WithLogger specifies the logger to be used to log the jsonpath operations.
func WithLogger(log func(format string, args ...interface{})) Option {
	return func(j *JpCondition) {
		j.l = log
	}
}

// WithValue does an equality check on the jsonpath result.
func WithValue(value interface{}) Option {
	return func(j *JpCondition) {
		j.matcher = func(ctx context.Context, resolved, root interface{}, log logger.Logger) error {
			j.log("Comparing %v with %s", value, resolved)

			if reflect.DeepEqual(value, resolved) {
				return nil
			}
			return fmt.Errorf("expected %v, got %v", value, resolved)
		}
	}
}

// WithDescription sets the description of the condition.
func WithDescription(description string) Option {
	return func(j *JpCondition) {
		j.description = description
	}
}

// New returns a new JqCondition.
func New(path string, opts ...Option) (*JpCondition, error) {
	jpc := JpCondition{
		matcher: defaultMatcher,
		id:      uuid.NewString(),
	}

	jp, err := jsonpath.Compile(path)
	if err != nil {
		return nil, err
	}
	jpc.jp = jp

	for _, opt := range opts {
		opt(&jpc)
	}

	return &jpc, nil
}

func must(jpc *JpCondition, err error) *JpCondition {
	if err != nil {
		panic(err)
	}
	return jpc
}

// MustNew returns a new JqCondition. It panics if the condition cannot be created.
func MustNew(path string, opts ...Option) *JpCondition {
	return must(New(path, opts...))
}

// IsSatisfiedBy implements condition.Condition.
func (j *JpCondition) IsSatisfiedBy(c context.Context, resource interface{}) error {
	ctx := ectx.From(c)
	j.level = ctx.Level()
	j.log("Evaluating jsonpath condition on resource")
	value, err := j.jp.Lookup(resource)
	if err != nil {
		return err
	}

	return j.matcher(ctx, value, resource, j.log)
}

// Id implements types.Condition.
func (j *JpCondition) Id() string {
	return j.id
}

// Description implements types.Condition.
func (j *JpCondition) Description() string {
	if j.description != "" {
		return j.description
	}
	return j.jp.String()
}

func defaultMatcher(ctx context.Context, value, root interface{}, log logger.Logger) error {
	if value != nil {
		return nil
	}
	return fmt.Errorf("expected non-empty result, got empty result")
}

func (j *JpCondition) log(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	format = "%s[%s]: %s\n"
	args = []interface{}{strings.Repeat(" ", j.level), j.Description(), s}

	if j.l != nil {
		j.l(format, args...)
	} else {
		logger.GetDefaultLogger()(format, args...)
	}
}
