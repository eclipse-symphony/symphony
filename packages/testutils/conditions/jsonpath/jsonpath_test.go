package jsonpath

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/stretchr/testify/require"
)

type (
	mmap          = map[string]interface{}
	basicTestCase struct {
		name     string
		j        *JpCondition
		resource any
		wantErr  bool
	}
)

func TestShouldPanicForInvalidJsonPath(t *testing.T) {
	require.Panics(t, func() {
		MustNew(`some invalid json path`)
	})
}

func TestIsSatisfiedBy(t *testing.T) {
	type Array[T any] []T

	data := mmap{
		"store": mmap{
			"type": "book store",
			"books": Array[mmap]{
				{
					"title":  "The Catcher in the Rye",
					"author": "J.D. Salinger",
				},
				{
					"title":  "To Kill a Mockingbird",
					"author": "Harper Lee",
				},
			},
		},
	}

	tests := []basicTestCase{
		{
			name:     "should return no error if jsonpath query matches value",
			j:        MustNew(`$.store.books[0].title`, WithValue("The Catcher in the Rye")),
			resource: data,
		},
		{
			name:     "should return error if jsonpath query dosesn't match value",
			j:        MustNew(`$.store.books[0].title`, WithValue("wrong")),
			resource: data,
			wantErr:  true,
		},
		{
			name:     "should return no error if jsonpath query matches for custom matcher",
			j:        MustNew(`$.store.books[?(@.author == 'Harper Lee')].author`, WithCustomMatcher(customAuthorMatcher)),
			resource: data,
		},
		{
			name:     "should return error if jsonpath query does not match for custom matcher",
			j:        MustNew(`$.store.books[?(@.author != 'Harper Lee')].author`, WithCustomMatcher(customAuthorMatcher)),
			resource: data,
			wantErr:  true,
		},
		{
			name:     "should return error when path resolves to nil and no matcher is specified",
			j:        MustNew("$.store.books[0].title.nonexistent"),
			resource: data,
			wantErr:  true,
		},
		{
			name:     "should return no error when path resolves to non nil and no matcher is specified",
			j:        MustNew("$.store.books[0].title"),
			resource: data,
		},
		{
			name:     "should accept a custom logger",
			j:        MustNew("$.store.books[0].title", WithLogger(customTestLogger)),
			resource: data,
		},
	}

	for _, tt := range tests {
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

func TestJsonPathCondition_Description(t *testing.T) {
	tests := []struct {
		name        string
		description string
		path        string
		want        string
	}{
		{
			name:        "should return description if set",
			description: "test description",
			path:        "$.foo.bar",
			want:        "test description",
		},
		{
			name: "should return path if description not set",
			path: "$.foo.bar",
			want: "$.foo.bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := must(New(tt.path, WithDescription(tt.description)))

			require.Contains(t, j.Description(), tt.want)
		})
	}
}

func TestHasId(t *testing.T) {
	jp := MustNew(`$.foo.bar`)
	require.NotEmpty(t, jp.Id())
}

func customAuthorMatcher(ctx context.Context, value, root interface{}, log logger.Logger) error {
	fave := []interface{}{
		"Harper Lee",
	}
	author, ok := value.([]interface{})
	if !ok {
		return errors.New("not an array")
	}
	if reflect.DeepEqual(author, fave) {
		return nil
	}

	return errors.New("jsonpath query did not match")
}

func customTestLogger(format string, args ...interface{}) {}
