package kube

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/conditions"
	"github.com/eclipse-symphony/symphony/packages/testutils/conditions/jq"
	"github.com/eclipse-symphony/symphony/packages/testutils/helpers"
	"github.com/eclipse-symphony/symphony/packages/testutils/internal"
	gomega "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
)

var (
	testResources = []runtime.Object{
		internal.Pod("test-1", "namespace-1"),
		internal.Pod("test-2", "namespace-2"),
		internal.Pod("different-1", "namespace-1"),
		internal.Pod("different-2", "namespace-2"),
		internal.Resource("config-1", "namespace-1", helpers.GVK("", "v1", "ConfigMap")),
		internal.Resource("config-2", "namespace-2", helpers.GVK("", "v1", "ConfigMap")),
		internal.Target("test-1", "namespace-1"),
		internal.OutOfSyncResource("test-2", "namespace-2", helpers.TargetGVK),
		internal.Namespace("namespace-1"),
		internal.Namespace("namespace-2"),
	}
	testScheme        = getScheme()
	testDynamicClient = fakeDynamic.NewSimpleDynamicClient(testScheme, testResources...)
	testDiscovery     = generateTestDiscoveryClient()
	testTimeout       = time.Millisecond * 5
)

func getScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	s.AddKnownTypes(
		schema.GroupVersion{Group: "", Version: "v1"},
		&corev1.Pod{},
		&corev1.PodList{},
		&corev1.Namespace{},
		&corev1.ConfigMap{},
		&corev1.ConfigMapList{},
		&corev1.NamespaceList{},
	)
	s.AddKnownTypes(
		helpers.TargetGVK.GroupVersion(),
		&unstructured.Unstructured{},
		&unstructured.UnstructuredList{},
	)
	return s
}

func generateTestDiscoveryClient() *fakeDiscovery.FakeDiscovery {
	f := &fakeDiscovery.FakeDiscovery{
		Fake: &testDynamicClient.Fake,
	}
	f.Resources = internal.GenerateTestApiResourceList()
	return f
}

func testDynamicClientBuilder() (dynamic.Interface, error) {
	return testDynamicClient, nil
}

func testDiscoveryClientBuilder() (discovery.DiscoveryInterface, error) {
	return testDiscovery, nil
}

func TestSuccess(t *testing.T) {
	r, err := Resource("test", "*", helpers.GVK("", "v1", "Pod"))
	require.NoError(t, err)
	require.NotEmpty(t, r.Description())
	require.NotEmpty(t, r.Id())
}

func TestSuccessAlternateDescription(t *testing.T) {
	r, err := Resource("test", "*", helpers.GVK("", "v1", "Pod"), WithDescription("alternate"))
	require.NoError(t, err)
	require.Equal(t, "alternate", r.Description())
}

func TestFailOnInvalidPattern(t *testing.T) {
	_, err := Resource("test(", "namespace-1", helpers.GVK("", "v1", "Pod")) // invalid pattern
	require.Error(t, err)
}

func TestMustSucceed(t *testing.T) {
	require.NotPanics(t, func() {
		Must(Resource("test", "*", helpers.GVK("", "v1", "Pod")))
	})
}

func TestMustPanics(t *testing.T) {
	require.Panics(t, func() {
		Must(Resource("test (", "", helpers.GVK("", "v1", "Pod"))) // invalid pattern
	})
}

func TestFindsSinglePodInNamespace(t *testing.T) {
	e := Must(Resource("test-1", "namespace-1", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithCondition(
			jq.MustNew(".spec.containers[0].name", jq.WithValue("test-1")),
		),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestFindsSinglePodInNamespaceWithFailingCondition(t *testing.T) {
	e := Must(Resource("test-1", "namespace-1", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithCondition(
			jq.MustNew(".spec.containers[0].name", jq.WithValue("wrong")),
		),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestExpectedCountSuccess(t *testing.T) {
	e := Must(Resource("test.+", "*", helpers.GVK("", "v1", "Pod"), // finds all pods starting with "test" in all namespaces
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithListCondition(conditions.Count(2)), // correct number
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestClusterLevelResourceSuccess(t *testing.T) {
	e := Must(Resource("namespace-.*", "*", helpers.GVK("", "v1", "Namespace"), // finds a namespace at the cluster level
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithListCondition(conditions.Count(2)),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestExpectedCountFail(t *testing.T) {
	e := Must(Resource("test.+", "*", helpers.GVK("", "v1", "Pod"), // finds all pods starting with "test" in all namespaces
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithListCondition(conditions.Count(0)), // wrong number
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestAbsentPodSuccess(t *testing.T) {
	e := Must(AbsentPod("nonexistent", "*", // tries to find a pod that doesn't exist
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestAbsentPodFail(t *testing.T) {
	e := Must(AbsentPod("test-1", "namespace-1", // tries to find a pod that shouldn't exist but it does
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestAbsentResourceSuccess(t *testing.T) {
	expect, err := AbsentResource("nonexistent", "*", helpers.TargetGVK,
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder))

	e := Must(expect, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	err = e.Verify(ctx)
	require.NoError(t, err)
}

func TestAbsentResourceFail(t *testing.T) {
	e := Must(AbsentResource("test-1", "namespace-1", helpers.TargetGVK, // tries to find a target that shouldn't exist but it does
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestAbsentResourcePassWithRedundantConditions(t *testing.T) {
	e := Must(AbsentResource("non-existent", "namespace-1", helpers.TargetGVK, // resource doesn't exist
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithListCondition(conditions.Count(0)), // redundant condition
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestUnknownResourceFail(t *testing.T) {
	e := Must(Resource("test-1", "namespace-1", helpers.GVK("random.group", "v1", "unknown"), // tries to find a Resource type that is not known
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestFailOnNoNamespaceForNamespacedResource(t *testing.T) {
	e := Must(Pod("test-1", "", // tries to find a Resource type that is not known
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestShouldFailWhenDiscoveryFails(t *testing.T) {
	ex, err := Resource("test-1", "namespace-1", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(func() (discovery.DiscoveryInterface, error) {
			return nil, errors.New("discovery failed")
		}),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	require.Error(t, ex.Verify(ctx))
}

func TestShouldFailWhenDynamicFails(t *testing.T) {
	ex, err := Resource("test-1", "namespace-1", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(func() (dynamic.Interface, error) {
			return nil, errors.New("dynamic failed")
		}),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	require.Error(t, ex.Verify(ctx))
}

func TestCanUseCustomLogger(t *testing.T) {
	called := false
	logger := func(format string, args ...interface{}) {
		called = true
	}
	e := Must(Resource("test-1", "namespace-1", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithLogger(logger),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
	require.True(t, called)
}

func TestCanUseCustomTickInterval(t *testing.T) {
	e := Must(Resource("test-1", "namespace-1", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithTick(time.Millisecond),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
	require.Equal(t, time.Millisecond, e.tick)
}

func TestAnnotationMathSuccesfulOnResource(t *testing.T) {
	e := Must(Resource("test-1", "namespace-1", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithCondition(
			NewAnnotationMatchCondition("test-annotation", "test-annotation-value"),
		),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}
func TestCombinedSuccessfullConditions(t *testing.T) {
	e := Must(Resource("test-.*", "*", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithCondition(conditions.All(
			NewAnnotationMatchCondition("test-annotation", "test-annotation-value"),
			NewAnnotationMatchCondition("management.azure.com/operationId", "test-operation-id"),
		)),
		WithListCondition(conditions.Count(2)),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestOperationIdTargetSuccess(t *testing.T) {
	e := Must(Resource("test-1", "namespace-1", helpers.GVK("fabric.symphony", "v1", "Target"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithCondition(OperationIdMatchCondition),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.NoError(t, e.Verify(ctx))
}

func TestOperationIdTargetFail(t *testing.T) {
	e := Must(Resource("test-2", "namespace-2", helpers.GVK("fabric.symphony", "v1", "Target"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithCondition(OperationIdMatchCondition),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	require.Error(t, e.Verify(ctx))
}

func TestCommonCommonConstructors(t *testing.T) {
	_, err := Target("test-1", "namespace-1")
	require.NoError(t, err)

	_, err = AbsentTarget("test-1", "namespace-1")
	require.NoError(t, err)

	_, err = Instance("test-1", "namespace-1")
	require.NoError(t, err)

	_, err = AbsentInstance("test-1", "namespace-1")
	require.NoError(t, err)

	_, err = Solution("test-1", "namespace-1")
	require.NoError(t, err)

	_, err = AbsentSolution("test-1", "namespace-1")
	require.NoError(t, err)
}

func TestShouldWorkWithGomegaAssersions(t *testing.T) {
	mt := internal.NewMockT()
	g := gomega.NewWithT(mt)
	e := Must(Resource("test-.*", "*", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithCondition(conditions.All(
			NewAnnotationMatchCondition("test-annotation", "test-annotation-value"),
			NewAnnotationMatchCondition("management.azure.com/operationId", "test-operation-id"),
		)),
		WithListCondition(conditions.Count(2)),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	mt.On("Helper").Return()
	g.Eventually(e.AsGomegaSubject()).WithContext(ctx).Should(e.ToGomegaMatcher())
	mt.AssertExpectations(t)

	ctx, cancel = context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	mt.On("Fatalf", mock.Anything, mock.Anything).Return()
	g.Eventually(e.AsGomegaSubject()).WithContext(ctx).ShouldNot(e.ToGomegaMatcher())
	mt.AssertExpectations(t)
}

func TestShouldWorkWithInvertedGomegaAssersions(t *testing.T) {
	mt := internal.NewMockT()
	g := gomega.NewWithT(mt)
	e := Must(Resource("test-.*", "*", helpers.GVK("", "v1", "Pod"),
		WithDynamicClientBuilder(testDynamicClientBuilder),
		WithDiscoveryClientBuilder(testDiscoveryClientBuilder),
		WithCondition(conditions.All(
			NewAnnotationMatchCondition("test-annotation", "test-annotation-value"),
			NewAnnotationMatchCondition("management.azure.com/operationId", "test-operation-id"),
			jq.Equality(".spec.containers[0].name", "wrong"),
		)),
		WithListCondition(conditions.Count(2)),
	))
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	mt.On("Helper").Return()
	failingCall := mt.On("Fatalf", mock.Anything, mock.Anything).Return()
	g.Eventually(e.AsGomegaSubject()).WithContext(ctx).Should(e.ToGomegaMatcher())
	mt.AssertExpectations(t)

	failingCall.Unset()
	g.Eventually(e.AsGomegaSubject()).WithContext(ctx).ShouldNot(e.ToGomegaMatcher())
	mt.AssertExpectations(t)
}
