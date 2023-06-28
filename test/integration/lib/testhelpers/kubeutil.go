// testhelpers contains helpers for tests
package testhelpers

import (
	"flag"
	"path/filepath"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	configInst      *rest.Config
	configInstMutex sync.Mutex
)

// KubeClient returns the kubectl client
func KubeClient() (*kubernetes.Clientset, error) {
	config, err := RestConfig()
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// RestConfig returns the kube config
func RestConfig() (*rest.Config, error) {
	configInstMutex.Lock()
	defer configInstMutex.Unlock()

	if configInst == nil {
		// based on the example from https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster-client-configuration/main.go
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}

		configInst = config
	}

	return configInst, nil
}
