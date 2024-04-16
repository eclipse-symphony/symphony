package conditions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/stretchr/testify/require"
)

type (
	basicConditionTest struct {
		name      string
		condition types.Condition
		resource  interface{}
		wantErr   bool
	}
	mmap = map[string]string
)

func TestAll(t *testing.T) {
	tests := []basicConditionTest{
		{
			name:      "should not return error if all conditions are satisfied - 1 right",
			condition: All(Count(2)),
			resource: []mmap{
				{"id": "1"},
				{"id": "2"},
			},
		},
		{
			name:      "should not return error if all conditions are satisfied - 2 right",
			condition: All(Count(2), Count(2)),
			resource: []mmap{
				{"id": "1"},
				{"id": "2"},
			},
		},
		{
			name:      "should return an error if any condition is not satisfied - 1 wrong",
			condition: All(Count(2)),
			resource: []mmap{
				{"id": "1"},
			},
			wantErr: true,
		},
		{
			name:      "should return an error if any condition is not satisfied - 2 wrong",
			condition: All(Count(2), Count(2)),
			resource: []mmap{
				{"id": "1"},
			},
			wantErr: true,
		},
		{
			name:      "should return an error if all conditions are not satisfied - 1 right, 1 wrong",
			condition: All(Count(5), Count(1)),
			resource: []mmap{
				{"id": "1"},
			},
			wantErr: true,
		},
		{
			name: "should return an error when any condition is not satisfied - nested",
			condition: All(
				All(Count(2)),
				All(
					Count(2),
					Count(2),
				),
				All(
					Count(2),
					All(
						Count(1), // this one fails
					),
				),
			),
			resource: []mmap{
				{"id": "1"},
				{"id": "1"},
			},
			wantErr: true,
		},
		{
			name: "should return not return error when all conditions are satisfied - nested",
			condition: All(
				All(Count(2)),
				All(
					Count(2),
					Count(2),
				),
				All(
					Count(2),
					All(
						Count(2),
					),
				),
			),
			resource: []mmap{
				{"id": "1"},
				{"id": "1"},
			},
		},
	}
	testBasicConditions(t, tests)
}

func TestBasicCondition(t *testing.T) {
	cases := []basicConditionTest{
		{
			name: "basic condition test",
			condition: Basic(func(ctx context.Context, resource interface{}) error {
				return nil
			}),
		},
		{
			name: "basic condition test fail",
			condition: Basic(func(ctx context.Context, resource interface{}) error {
				return errors.New("fail")
			}),
			wantErr: true,
		},
	}
	testBasicConditions(t, cases)
}

func TestBasicConditionWithOptions(t *testing.T) {
	b := Basic(
		func(ctx context.Context, resource interface{}) error {
			r := resource.([]mmap)
			if len(r) == 1 {
				return nil
			}
			return errors.New("fail")
		},
		WithBasicDescription("test description"),
		WithBasicFailureMessage(func(resource interface{}, err error) string {
			return "test failure message"
		}),
	)

	require.Equal(t, "test description", b.Description())
	require.Equal(t, "test failure message", b.failMsg(nil, nil))

}
func TestAny(t *testing.T) {
	tests := []basicConditionTest{
		{
			name:      "should not return error if any condition is satisfied - 1 right",
			condition: Any(Count(1)),
			resource: []mmap{
				{"id": "1"},
			},
		},
		{
			name:      "should not return error if any condition is satisfied - 2 right",
			condition: Any(Count(1), Count(1)),
			resource: []mmap{
				{"id": "1"},
			},
		},
		{
			name:      "should return an error if no condition is satisfied - 1 wrong",
			condition: Any(Count(2)),
			resource: []mmap{
				{"id": "1"},
			},
			wantErr: true,
		},
		{
			name:      "should return an error if no condition is satisfied - 2 wrong",
			condition: Any(Count(3), Count(2)),
			resource: []mmap{
				{"id": "1"},
			},
			wantErr: true,
		},
		{
			name:      "should not return an error if any condition is satisfied - 1 right, 1 wrong",
			condition: Any(Count(1), Count(2)),
			resource: []mmap{
				{"id": "1"},
			},
		},
	}
	testBasicConditions(t, tests)
}

func TestExpectedCount(t *testing.T) {
	tests := []basicConditionTest{
		{
			name:      "should not return error if count is satisfied",
			condition: Count(1),
			resource: []mmap{
				{"id": "1"},
			},
		},
		{
			name:      "should return error if count is not satisfied",
			condition: Count(2),
			resource: []mmap{
				{"id": "1"},
			},
			wantErr: true,
		},
		{
			name:      "should  return error if expected type is not a slice, map or arry",
			condition: Count(1),
			resource:  1,
			wantErr:   true,
		},
	}

	testBasicConditions(t, tests)
}
func TestCombo(t *testing.T) {
	tests := []basicConditionTest{
		{
			name: "combo - nested",
			condition: Any(
				Any(Count(2)), //false
				Any(
					Count(2), // false
					Count(2), // false
				), // false
				All(
					Count(1), // true
					Any(
						Count(2), // false
						Count(1), // true
					), // true
				), // true
			), // true
			resource: []mmap{
				{"id": "1"},
			},
		},
		{
			name: "should return error for combo",
			condition: Any(
				Any(Count(2)), //false
				All(
					Count(2),       // false
					Count(1),       // true
					GreaterThan(1), // false
				), // false
				Any(
					Count(2), // false
					Any(
						Count(2), // false
					), // false
					All(
						GreaterThan(1), // false
						Count(1),       // true
					), // false
				), // false
			), // false
			resource: []mmap{
				{"id": "1"},
			},
			wantErr: true,
		},
	}
	testBasicConditions(t, tests)
}

func TestExtendAllAndWithMoreConditions(t *testing.T) {
	firstAll := All(Count(1))
	secondAll := firstAll.And(Count(1))

	require.NotEqual(t, firstAll, secondAll)
}

func TestAllWithCache(t *testing.T) {
	callCount := 0
	c := All(
		CountComparator(func(c int) bool {
			callCount++
			return c == 1
		}, ""), // this will allways succeed
		Count(0), // this will allways fail
	).WithCaching()
	resource := []mmap{{"id": "1"}}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	require.Equal(t, 0, callCount)
	require.Error(t, c.IsSatisfiedBy(ctx, resource))
	require.Equal(t, 1, callCount)

	require.Error(t, c.IsSatisfiedBy(ctx, resource))
	require.Equal(t, 1, callCount)
}

func testBasicConditions(t *testing.T, tt []basicConditionTest) {
	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.condition.Id())
			require.NotEmpty(t, tt.condition.Description())
			err := tt.condition.IsSatisfiedBy(context.Background(), tt.resource)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
