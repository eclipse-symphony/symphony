package conditions

import (
	"context"
	"fmt"

	"reflect"
)

// CountComparator returns a condition that checks the count of a countable
// based on the given comparator function.
func CountComparator(fn func(int) bool, failMessage string) basic {
	b := Basic(func(ctx context.Context, i interface{}) error {
		switch reflect.TypeOf(i).Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			if !fn(reflect.ValueOf(i).Len()) {
				return fmt.Errorf(failMessage)
			}
		default:
			return fmt.Errorf("expected countable, got %T", i)
		}
		return nil
	})
	return b
}

// Count returns a condition that checks the count of a countable.
func Count(count int) basic {
	return CountComparator(func(c int) bool {
		return c == count
	}, fmt.Sprintf("expected count %d", count))
}

// GreaterThan returns a condition that checks the count of a countable.
func GreaterThan(count int) basic {
	return CountComparator(func(c int) bool {
		return c > count
	}, fmt.Sprintf("expected count greater than %d", count))
}
