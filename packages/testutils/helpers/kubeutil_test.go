package helpers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/internal"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/discovery"
	fakedisc "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	fakedyn "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	fakekube "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type (
	kubeutilTestCase struct {
		name        string
		configError bool
		builderErr  bool
		wantErr     bool
	}
)

var (
	matrix = []kubeutilTestCase{
		{
			name:        "should return error if config builder fails and dynamic builder fails",
			configError: true,
			builderErr:  true,
			wantErr:     true,
		},
		{
			name:        "should return error if config builder fails",
			configError: true,
			wantErr:     true,
		},
		{
			name:       "should return error if dynamic builder fails",
			builderErr: true,
			wantErr:    true,
		},
		{
			name: "should return dynamic client",
		},
	}
	testTimeout = 1 * time.Millisecond
)

func setDynBuilder(shouldError bool) {
	dynamicBuilder = func(*rest.Config) (dynamic.Interface, error) {
		if shouldError {
			return nil, errors.New("error")
		}
		return &fakedyn.FakeDynamicClient{}, nil
	}
}

func setConfigBuilder(shouldError bool) {
	configGetter = func(string, string) (*rest.Config, error) {
		if shouldError {
			return nil, errors.New("error")
		}
		return &rest.Config{}, nil
	}
}

func setDiscoveryBuilder(shouldError bool) {
	discoveryBuilder = func(*rest.Config) (discovery.DiscoveryInterface, error) {
		if shouldError {
			return nil, errors.New("error")
		}
		return &fakedisc.FakeDiscovery{}, nil
	}
}

func setKubeBuilder(shouldError bool) {
	kubernetesBuilder = func(*rest.Config) (kubernetes.Interface, error) {
		if shouldError {
			return nil, errors.New("error")
		}
		fakekube.NewSimpleClientset()
		return &fakekube.Clientset{}, nil
	}
}

func TestKubernetes(t *testing.T) {
	for _, tt := range matrix {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(cleanup)
			setKubeBuilder(tt.builderErr)
			setConfigBuilder(tt.configError)
			_, err := KubeClient()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDiscovery(t *testing.T) {
	for _, tt := range matrix {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(cleanup)
			setDiscoveryBuilder(tt.builderErr)
			setConfigBuilder(tt.configError)
			_, err := DiscoveryClient()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDynamic(t *testing.T) {
	for _, tt := range matrix {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(cleanup)
			setDynBuilder(tt.builderErr)
			setConfigBuilder(tt.configError)
			_, err := DynamicClient()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestRestConfig_Success(t *testing.T) {
	t.Cleanup(cleanup)
	setConfigBuilder(false)
	_, err := RestConfig()
	require.NoError(t, err)
}
func TestRestConfig_ReturnsFromCache(t *testing.T) {
	t.Cleanup(cleanup)
	configInst = &rest.Config{}
	setConfigBuilder(false)
	returned, err := RestConfig()
	require.NoError(t, err)
	require.Equal(t, configInst, returned)
}

func TestRestConfig_Faile(t *testing.T) {
	t.Cleanup(cleanup)
	setConfigBuilder(true)
	_, err := RestConfig()
	require.Error(t, err)
}

func TestEnsureNamespaceSucceedsWhenNamespaceAlreadyExist(t *testing.T) {
	t.Cleanup(cleanup)
	client := fakekube.NewSimpleClientset(internal.Namespace("test")) // initializes the fake client with the namespace object
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	err := EnsureNamespace(ctx, client, "test")
	require.NoError(t, err)
}

func TestEnsureNamespaceSuccessWhenNamespaceDoesntExist(t *testing.T) {
	t.Cleanup(cleanup)
	client := fakekube.NewSimpleClientset()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	err := EnsureNamespace(ctx, client, "test")
	require.NoError(t, err)
}

func cleanup() {
	configInst = nil
}
