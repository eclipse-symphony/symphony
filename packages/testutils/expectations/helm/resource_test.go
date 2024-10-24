package helm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/conditions"
	"github.com/eclipse-symphony/symphony/packages/testutils/conditions/jq"
	"github.com/eclipse-symphony/symphony/packages/testutils/internal"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

type (
	mockLister struct {
		count      int
		shoudError bool
	}
)

var (
	_           ListRunner = mockLister{}
	testTimeout            = time.Millisecond * 5
)

func TestNoErrorForNewResourceValidator(t *testing.T) {
	e, err := NewExpectation("name", "*",
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithDescription("description"),
	)

	require.NoError(t, err)
	require.NotEmpty(t, e.Description())
	require.NotEmpty(t, e.Id())

}

func TestErrorForNewResourceWithIncorrectConfig(t *testing.T) {
	_, err := NewExpectation("", "") // no pattern or namespace
	require.Error(t, err)
}

func TestErrorForNewResourceWithInvalidNameValidator(t *testing.T) {
	_, err := NewExpectation("( ", "namespaace")

	require.Error(t, err)
}

func TestShouldPanicForInvalidName(t *testing.T) {
	require.Panics(t, func() {
		MustNew("( ", "namespaace")
	})
}

func TestShouldErrorWhenListFails(t *testing.T) {
	e, err := NewExpectation("name", "namespace", WithListClientBuilder(getMockListerBuilder(0, true)))

	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	err = e.Verify(ctx)

	require.Error(t, err)
}

func TestShouldFindMatchingReleaseFromExactName(t *testing.T) {
	e, err := NewExpectation("release-0", "namespace",
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithReleaseCondition(jq.MustNew(`.chart.metadata.name`, jq.WithValue("chart-0"))),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestShouldFailOnMatchButFalseReleaseCondition(t *testing.T) {
	e, err := NewExpectation("release-0", "namespace",
		WithLogger(func(format string, args ...interface{}) {}),
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithReleaseCondition(jq.MustNew(`.chart.metadata.name`, jq.WithValue("wrong"))),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestShouldFailOnMatchButFalseValueCondition(t *testing.T) {
	e, err := NewExpectation("release-0", "namespace",
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithValueCondition(jq.MustNew(`.foo`, jq.WithValue("baz"))),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestShouldFindMatchingReleaseFromExactNameAndExtraConditions(t *testing.T) {
	e, err := NewExpectation("release-0", "namespace",
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithReleaseCondition(conditions.All(
			jq.Equality(`.chart.metadata.name`, "chart-0"),
			jq.Equality(`.chart.metadata.version`, "x.y.z"),
		)),
		WithReleaseListCondition(conditions.Count(1)),
		WithValueCondition(conditions.All(
			jq.Equality(`.foo`, "bar"),
			jq.Equality(`.baz`, "qux"),
		)),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestShouldNotFindMatchingReleaseWithUnexactName(t *testing.T) {
	e, err := NewExpectation("release", "namespace", // No release with name "release"
		WithListClientBuilder(getMockListerBuilder(1, false)),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}
func TestShouldMatchAbsentRelease(t *testing.T) {
	e := MustNewAbsent("some-non-existent-release", "*", // No release with name "some-non-existent-release" in any namespace
		WithListClientBuilder(getMockListerBuilder(1, false)),
	)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestShouldFindMultipleReleases(t *testing.T) {
	e, err := NewExpectation("release-.+", "namespace", // regex to match all releases in namespace
		WithListClientBuilder(getMockListerBuilder(5, false)),
		WithReleaseListCondition(
			conditions.Count(5),
		),
		WithValueListCondition(
			conditions.Count(5),
		),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestShouldFindMultipleReleasesButFailForIncorrectReleaseCondtion(t *testing.T) {
	e, err := NewExpectation("release-.+", "namespace", // regex to match all releases in namespace
		WithListClientBuilder(getMockListerBuilder(10, false)),
		WithReleaseListCondition(
			conditions.Count(5), // correct count is 10

		),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestShouldFindMultipleReleasesButFailForIncorrectValueCondtion(t *testing.T) {
	e, err := NewExpectation("release-.+", "namespace", // regex to match all releases in namespace
		WithListClientBuilder(getMockListerBuilder(10, false)),
		WithValueListCondition(
			conditions.Count(5), // correct count is 10
		),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestVerifyPasses(t *testing.T) {
	e := MustNew("release-0", "namespace",
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithReleaseCondition(jq.MustNew(`.chart.metadata.name`, jq.WithValue("chart-0"))),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	require.NoError(t, e.Verify(ctx))
}

func TestVerifyFailsWhenContextEnds(t *testing.T) {
	e := MustNew("release-0", "namespace",
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithReleaseCondition(jq.MustNew(`.chart.metadata.name`, jq.WithValue("wrong"))),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	require.Error(t, e.Verify(ctx))
}

func TestWhenListBuilderInitializeFails(t *testing.T) {
	ex, err := NewExpectation("release-0", "namespace",
		WithListClientBuilder(func() (ListRunner, error) {
			return nil, fmt.Errorf("some error")
		}),
		WithReleaseCondition(jq.MustNew(`.chart.metadata.name`, jq.WithValue("wrong"))),
	)

	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	require.Error(t, ex.Verify(ctx))
}

func TestWorksWithGomegaAssertion(t *testing.T) {
	mt := internal.NewMockT()
	g := gomega.NewGomegaWithT(mt)
	e, err := NewExpectation("release-0", "namespace",
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithReleaseCondition(conditions.All(
			jq.Equality(`.chart.metadata.name`, "chart-0"),
			jq.Equality(`.chart.metadata.version`, "x.y.z"),
		)), // false runs the condiition on individual resources
		WithReleaseListCondition(conditions.Count(1)),
		WithValueCondition(conditions.All(
			jq.Equality(`.foo`, "bar"),
			jq.Equality(`.baz`, "qux"),
		)),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	mt.On("Helper").Return()
	g.Eventually(e.AsGomegaSubject()).WithContext(ctx).Should(e.ToGomegaMatcher())
	mt.AssertExpectations(t)

	mt.On("Fatalf", mock.Anything, mock.Anything).Return()
	g.Eventually(e.AsGomegaSubject()).WithContext(ctx).ShouldNot(e.ToGomegaMatcher())
	mt.AssertExpectations(t)
}

func TestWorksWithGomegaAssertionFailingCondition(t *testing.T) {
	mt := internal.NewMockT()
	g := gomega.NewGomegaWithT(mt)
	e, err := NewExpectation("release-0", "namespace",
		WithListClientBuilder(getMockListerBuilder(1, false)),
		WithReleaseCondition(conditions.All(
			jq.Equality(`.chart.metadata.name`, "chart-0"),
			jq.Equality(`.chart.metadata.version`, "x.y.z"),
		)),
		WithReleaseListCondition(conditions.Count(1)),
		WithValueCondition(conditions.All(
			jq.Equality(`.foo`, "wrong"), // wrong value
			jq.Equality(`.baz`, "qux"),
		)),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	mt.On("Helper").Return()
	g.Eventually(e.AsGomegaSubject()).WithContext(ctx).ShouldNot(e.ToGomegaMatcher())
	mt.AssertExpectations(t)

	mt.On("Fatalf", mock.Anything, mock.Anything).Return()
	g.Eventually(e.AsGomegaSubject()).WithContext(ctx).Should(e.ToGomegaMatcher())
	mt.AssertExpectations(t)
}

func (ml mockLister) Run() ([]*release.Release, error) {
	if ml.shoudError {
		return nil, fmt.Errorf("some error")
	}
	releases := make([]*release.Release, ml.count)
	for i := 0; i < ml.count; i++ {
		releases[i] = &release.Release{
			Name: fmt.Sprintf("release-%d", i),
			Config: map[string]interface{}{
				"foo": "bar",
				"baz": "qux",
			},
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Name:    fmt.Sprintf("chart-%d", i),
					Version: "x.y.z",
				},
			},
		}
	}
	return releases, nil
}

func getMockListerBuilder(count int, shouldError bool) func() (ListRunner, error) {
	return func() (ListRunner, error) {
		return mockLister{
			count:      count,
			shoudError: shouldError,
		}, nil
	}
}
