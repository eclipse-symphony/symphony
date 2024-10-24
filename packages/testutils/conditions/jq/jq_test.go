package jq

import (
	"context"
	"errors"
	"testing"

	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/stretchr/testify/require"
)

type (
	jqTestCase struct {
		name     string
		j        *JqCondition
		resource any
		wantErr  bool
	}
	mmap = map[string]interface{}
)

func TestJQCondition_IsSatisfiedBy(t *testing.T) {
	var (
		foobar = mmap{
			"foo": mmap{
				"bar":       "baz",
				"nullValue": nil,
			},
		}
		basicCases = []jqTestCase{
			{
				name:     "should return no error if jq query matches value",
				j:        MustNew(`.foo.bar`, WithValue("baz")),
				resource: foobar,
			},
			{
				name:     "should return error if jq query dosesn't match value",
				j:        MustNew(`.foo.bar`, WithValue("wrong")),
				resource: foobar,
				wantErr:  true,
			},
			{
				name:     "should return no error if jq query matches for custom matcher",
				j:        MustNew(`.foo.bar == "baz"`, WithCustomMatcher(customTruthyTestMatcher)),
				resource: foobar,
			},
			{
				name:     "should return error if jq query does not match for custom matcher",
				j:        MustNew(`.foo.bar != "baz"`, WithCustomMatcher(customTruthyTestMatcher)),
				resource: foobar,
				wantErr:  true,
			},
			{
				name:     "should return error when path resolves to non-existent and no matcher is specified",
				j:        MustNew(".foo.bar.nonexistent"),
				resource: foobar,
				wantErr:  true,
			},
			{
				name:     "should return error when path resolves to nil value and no matcher is specified",
				j:        MustNew(".foo.nullValue"),
				resource: foobar,
				wantErr:  true,
			},
			{
				name:     "should handle nil resource",
				j:        Equality(".foo.nullValue", nil),
				resource: foobar,
			},
			{
				name:     "should handle nil resource and nil wrong value",
				j:        Equality(".foo.nullValue", "wrong"),
				resource: foobar,
				wantErr:  true,
			},
			{
				name:     "should handle incompativle types",
				j:        Equality(".foo.bar", 1),
				resource: foobar,
				wantErr:  true,
			},
			{
				name:     "should return no error when path resolves to non nil and no matcher is specified",
				j:        MustNew(".foo.bar"),
				resource: foobar,
			},
			{
				name:     "should accept a custom logger",
				j:        MustNew(".foo.bar", WithLogger(customTestLogger)),
				resource: foobar,
			},
			{
				name:     "should return error for unsupported type",
				j:        MustNew(".foo.bar", WithValue("baz")),
				resource: struct{ name string }{name: "test"},
				wantErr:  true,
			},
			{
				name:     "should pass ussing helper function",
				j:        Equality(".foo.bar", "baz"),
				resource: foobar,
			},
			{
				name:     "should fail ussing helper function",
				j:        Equality(".foo.bar", "wrong"),
				resource: foobar,
				wantErr:  true,
			},
		}
	)
	for _, tt := range basicCases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.j.IsSatisfiedBy(context.Background(), tt.resource)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHasId(t *testing.T) {
	jq := MustNew(`.foo.bar`)
	require.NotEmpty(t, jq.Id())
}

func TestJQCondition_Description(t *testing.T) {
	tests := []struct {
		name        string
		description string
		path        string
		want        string
	}{
		{
			name:        "should return description if set",
			description: "test description",
			path:        `.foo == "bar"`,
			want:        "test description",
		},
		{
			name: "should return path if description not set",
			path: `.foo == "bar"`,
			want: `.foo == "bar"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := must(New(tt.path, WithDescription(tt.description)))

			require.Equal(t, tt.want, j.Description())
		})
	}
}

func TestShouldPanicForInvalidJQ(t *testing.T) {
	require.Panics(t, func() {
		MustNew(`some invalid jq query`)
	})
}

func customTruthyTestMatcher(ctx context.Context, value, root interface{}, log logger.Logger) error {
	if value.(bool) == true {
		return nil
	}
	return errors.New("jq query did not match")
}

func customTestLogger(format string, args ...interface{}) {}
