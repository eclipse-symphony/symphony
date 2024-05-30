// testhelpers contains helpers for tests
package helpers

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	configInst      *rest.Config
	configInstMutex sync.Mutex

	TargetGVK      = GVK("fabric.symphony", "v1", "Target")
	InstanceGVK    = GVK("solution.symphony", "v1", "Instance")
	SolutionGVK    = GVK("solution.symphony", "v1", "Solution")
	ConfigMapGVK   = GVK("", "v1", "ConfigMap")
	PodGVK         = GVK("", "v1", "Pod")
	NamespaceGVK   = GVK("", "v1", "Namespace")
	ClusterRoleGVK = GVK("rbac.authorization.k8s.io", "v1", "ClusterRole")

	configGetter   func(string, string) (*rest.Config, error)    = clientcmd.BuildConfigFromFlags
	dynamicBuilder func(*rest.Config) (dynamic.Interface, error) = func(config *rest.Config) (dynamic.Interface, error) {
		return dynamic.NewForConfig(config)
	}
	discoveryBuilder func(*rest.Config) (discovery.DiscoveryInterface, error) = func(config *rest.Config) (discovery.DiscoveryInterface, error) {
		return discovery.NewDiscoveryClientForConfig(config)
	}
	kubernetesBuilder func(*rest.Config) (kubernetes.Interface, error) = func(config *rest.Config) (kubernetes.Interface, error) {
		return kubernetes.NewForConfig(config)
	}
)

// KubeClient returns the kubectl client from the default kube config
func KubeClient() (kubernetes.Interface, error) {
	config, err := RestConfig()
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetesBuilder(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// DiscoveryClient returns the discovery client from the default kube config
func DiscoveryClient() (discovery.DiscoveryInterface, error) {
	config, err := RestConfig()
	if err != nil {
		return nil, err
	}

	// create the clientset
	client, err := discoveryBuilder(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// DynamicClient returns the dynamic client from the default kube config
func DynamicClient() (dynamic.Interface, error) {
	config, err := RestConfig()
	if err != nil {
		return nil, err
	}

	// create the clientset
	client, err := dynamicBuilder(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// RestConfig returns the default kube config
func RestConfig() (*rest.Config, error) {
	configInstMutex.Lock()
	defer configInstMutex.Unlock()
	var err error
	if configInst != nil {
		return configInst, nil
	}
	homeDir, _ := os.UserHomeDir()
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")

	configInst, err = configGetter("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return configInst, nil
}

// EnsureNamespace ensures that the namespace exists. If it does not exist, it creates it.
func EnsureNamespace(ctx context.Context, client kubernetes.Interface, namespace string) error {
	_, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if kerrors.IsNotFound(err) {
		_, err = client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}

	} else {
		return err
	}

	return nil
}

// GVK creates a GroupVersionKind
func GVK(group, version, kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
}
