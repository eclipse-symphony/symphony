package internal

import (
	"testing"

	"github.com/eclipse-symphony/symphony/packages/testutils/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPod(t *testing.T) {
	p := Pod("test-pod", "test-namespace")
	require.Equal(t, "test-pod", p.GetName())
	require.Equal(t, "test-namespace", p.GetNamespace())
	require.Equal(t, helpers.PodGVK, p.GroupVersionKind())
}

func TestTarget(t *testing.T) {
	tgt := Target("test", "test-namespace")
	require.Equal(t, "test", tgt.GetName())
	require.Equal(t, "test-namespace", tgt.GetNamespace())
	require.Equal(t, helpers.TargetGVK, tgt.GroupVersionKind())
}
func TestOutOfSyncTarget(t *testing.T) {
	tgt := OutOfSyncResource("test", "test-namespace", helpers.TargetGVK)
	require.Equal(t, "test", tgt.GetName())
	require.Equal(t, "test-namespace", tgt.GetNamespace())
	require.Equal(t, helpers.TargetGVK, tgt.GroupVersionKind())
}

func TestApiResourceListGenerator(t *testing.T) {
	arl := GenerateTestApiResourceList()
	require.Len(t, arl, 2)
}

func TestResource(t *testing.T) {
	r := Resource("test", "test-namespace", helpers.InstanceGVK)
	require.Equal(t, "test", r.GetName())
	require.Equal(t, "test-namespace", r.GetNamespace())
	require.Equal(t, helpers.InstanceGVK, r.GroupVersionKind())
}

func TestNamespace(t *testing.T) {
	ns := Namespace("test")
	require.Equal(t, "test", ns.GetName())
	require.Equal(t, helpers.NamespaceGVK, ns.GroupVersionKind())
}

func TestMockT(t *testing.T) {
	m := NewMockT()
	m.On("Helper")
	m.On("Errorf", mock.Anything, mock.Anything)
	m.On("Fatalf", mock.Anything, mock.Anything)
	m.Helper()
	m.Errorf("test")
	m.Fatalf("test")
	m.AssertExpectations(t)
}
