package jq

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	ectx "github.com/eclipse-symphony/symphony/packages/testutils/internal/context"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/google/uuid"
	"github.com/itchyny/gojq"
)

type (
	JqCondition struct {
		id          string
		matcher     func(ctx context.Context, values, root interface{}, log logger.Logger) error
		jq          *gojq.Query
		description string
		l           func(format string, args ...interface{})
		path        string
		level       int
	}

	Option func(*JqCondition)
)

var (
	_ types.Condition = &JqCondition{}
)

// WithCustomMatcher specifies the matcher to be used to match the jq result.
func WithCustomMatcher(matcher func(ctx context.Context, value, root interface{}, log logger.Logger) error) Option {
	return func(j *JqCondition) {
		j.matcher = matcher
	}
}

// WithLogger specifies the logger to be used to log the jq operations.
func WithLogger(log func(format string, args ...interface{})) Option {
	return func(j *JqCondition) {
		j.l = log
	}
}

// WithValue does an equality check on the jq result.
func WithValue(value interface{}) Option {
	return func(j *JqCondition) {
		t1 := reflect.TypeOf(value)
		j.matcher = func(ctx context.Context, resolved, root interface{}, log logger.Logger) error {
			j.log("Comparing %v with %v", value, resolved)
			if resolved == nil {
				if value == nil {
					return nil
				}
				return fmt.Errorf("expected %v, got nil", value)
			}

			if !reflect.TypeOf(resolved).ConvertibleTo(t1) {
				return fmt.Errorf("expected %v, got %v", value, resolved)
			}

			if reflect.DeepEqual(value, reflect.ValueOf(resolved).Convert(t1).Interface()) {
				return nil
			}
			return fmt.Errorf("expected %v, got %v", value, resolved)
		}
	}
}

// WithDescription specifies the description of the condition.
func WithDescription(description string) Option {
	return func(j *JqCondition) {
		j.description = description
	}
}

// New returns a new JqCondition.
func New(path string, opts ...Option) (*JqCondition, error) {
	jqc := JqCondition{
		matcher: defaultMatcher,
		path:    path,
		id:      uuid.NewString(),
	}

	jq, err := gojq.Parse(path)
	if err != nil {
		return nil, err
	}
	jqc.jq = jq

	for _, opt := range opts {
		opt(&jqc)
	}

	return &jqc, nil
}

func must(jqc *JqCondition, err error) *JqCondition {
	if err != nil {
		panic(err)
	}
	return jqc
}

// MustNew returns a new JqCondition. It panics if the condition cannot be created.
func MustNew(path string, opts ...Option) *JqCondition {
	return must(New(path, opts...))
}

// IsSatisfiedBy implements condition.Condition.
func (j *JqCondition) IsSatisfiedBy(c context.Context, resource interface{}) error {
	ctx := ectx.From(c)
	j.level = ctx.Level()
	j.log("Evaluating jq condition on resource")

	iter := j.jq.RunWithContext(ctx, resource)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}
		err := j.matcher(ctx, v, resource, j.log)
		if err != nil {
			return err
		}
	}

	return nil
}

// Id implements condition.Condition.
func (j *JqCondition) Id() string {
	return j.id
}

// Description implements condition.Condition.
func (j *JqCondition) Description() string {
	if j.description != "" {
		return j.description
	}
	return j.path
}

func (j *JqCondition) log(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	format = "%s[%s]: %s\n"
	args = []interface{}{strings.Repeat(" ", j.level), j.Description(), s}

	if j.l != nil {
		j.l(format, args...)
	} else {
		logger.GetDefaultLogger()(format, args...)
	}
}

func defaultMatcher(ctx context.Context, value, root interface{}, log logger.Logger) error {
	if value != nil {
		return nil
	}
	return fmt.Errorf("expected non-empty result, got empty result")
}
