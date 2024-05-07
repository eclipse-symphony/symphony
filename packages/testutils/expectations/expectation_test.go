package expectations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type (
	basicConditionTest struct {
		name        string
		expectation types.Expectation
		wantErr     bool
	}
)

func TestAll(t *testing.T) {
	tests := []basicConditionTest{
		{
			name:        "should not return error if all expectations are satisfied - 1 right",
			expectation: All(expectation(true)),
		},
		{
			name:        "should not return error if all expectations are satisfied - 2 right",
			expectation: All(expectation(true), expectation(true)),
		},
		{
			name:        "should return an error if any expectation is not satisfied - 1 wrong",
			expectation: All(expectation(false)),
			wantErr:     true,
		},
		{
			name:        "should return an error if any expectation is not satisfied - 2 wrong",
			expectation: All(expectation(false), expectation(false)),
			wantErr:     true,
		},
		{
			name:        "should return an error if all expectations are not satisfied - 1 right, 1 wrong",
			expectation: All(expectation(true), expectation(false)),
			wantErr:     true,
		},
		{
			name: "should return an error when any expectation is not satisfied - nested",
			expectation: All(
				All(expectation(true)),
				All(
					expectation(true),
					expectation(true),
				),
				All(
					expectation(true),
					All(
						expectation(false), // this one fails
					),
				),
			),
			wantErr: true,
		},
		{
			name: "should return not return error when all expectations are satisfied - nested",
			expectation: All(
				All(expectation(true)),
				All(
					expectation(true),
					expectation(true),
				),
				All(
					expectation(true),
					All(
						expectation(true),
					),
				),
			),
		},
	}
	testBasicConditions(t, tests)
}

func TestAny(t *testing.T) {
	tests := []basicConditionTest{
		{
			name:        "should not return error if any expectation is satisfied - 1 right",
			expectation: Any(expectation(true)),
		},
		{
			name:        "should not return error if any expectation is satisfied - 2 right",
			expectation: Any(expectation(true), expectation(true)),
		},
		{
			name:        "should return an error if no expectation is satisfied - 1 wrong",
			expectation: Any(expectation(false)),
			wantErr:     true,
		},
		{
			name:        "should return an error if no expectation is satisfied - 2 wrong",
			expectation: Any(expectation(false), expectation(false)),
			wantErr:     true,
		},
		{
			name:        "should not return an error if any expectation is satisfied - 1 right, 1 wrong",
			expectation: Any(expectation(false), expectation(true)),
		},
	}
	testBasicConditions(t, tests)
}

func TestCombo(t *testing.T) {
	tests := []basicConditionTest{
		{
			name: "combo - nested",
			expectation: Any(
				Any(expectation(false)), //false
				Any(
					expectation(false), // false
					expectation(false), // false
				), // false
				All(
					expectation(true), // true
					Any(
						expectation(false), // false
						expectation(true),  // true
					), // true
				), // true
			), // true
		},
		{
			name: "should return error for combo",
			expectation: Any(
				Any(expectation(false)), //false
				All(
					expectation(false), // false
					expectation(true),  // true
				), // false
				Any(
					expectation(false), // false
					Any(
						expectation(false), // false
					), // false
				), // false
			), // false
			wantErr: true,
		},
	}
	testBasicConditions(t, tests)
}

func TestAllWithCache(t *testing.T) {
	passingExpectation := expectation(true)
	failingExpectation := expectation(false)
	c := All(
		passingExpectation,
		failingExpectation,
	).WithCaching()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	require.Equal(t, 0, passingExpectation.callexpectations["Verify"])
	require.Equal(t, 0, failingExpectation.callexpectations["Verify"])
	require.Error(t, c.Verify(ctx))

	require.Equal(t, 1, passingExpectation.callexpectations["Verify"])
	require.Equal(t, 1, failingExpectation.callexpectations["Verify"])

	require.Error(t, c.Verify(ctx))

	require.Equal(t, 1, passingExpectation.callexpectations["Verify"])
	require.Equal(t, 2, failingExpectation.callexpectations["Verify"])
}

func testBasicConditions(t *testing.T, tt []basicConditionTest) {
	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.expectation.Id())
			require.NotEmpty(t, tt.expectation.Description())

			err := tt.expectation.Verify(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type mockExpectation struct {
	id               string
	callexpectations map[string]int
	shouldPass       bool
}

// Description implements Expectation.
func (m mockExpectation) Description() string {
	defer m.call("Description")
	return "testing expectation"
}

// Id implements Expectation.
func (m mockExpectation) Id() string {
	defer m.call("Id")
	return m.id
}

// Verify implements Expectation.
func (m mockExpectation) Verify(ctx context.Context) error {
	defer m.call("Verify")
	if !m.shouldPass {
		return errors.New("mock expectation failed")
	}
	return nil
}

func (m mockExpectation) call(n string) {
	m.callexpectations[n]++
}

func expectation(shouldPass bool) mockExpectation {
	return mockExpectation{
		id:               uuid.NewString(),
		shouldPass:       shouldPass,
		callexpectations: make(map[string]int),
	}
}
