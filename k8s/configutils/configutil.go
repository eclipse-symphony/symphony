package configutils

import (
	"context"

	configv1 "gopls-workspace/apis/config/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func GetValidationPoilicies() (map[string][]configv1.ValidationPolicy, error) {
	// home := homedir.HomeDir()
	// // use the current context in kubeconfig
	// config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	// if err != nil {
	// 	panic(err.Error())
	// }

	// // create the clientset
	// clientset, err := kubernetes.NewForConfig(config)

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	configMap, err := clientset.CoreV1().ConfigMaps("symphony-k8s-system").Get(context.Background(), "symphony-k8s-manager-config", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var myConfig configv1.ProjectConfig
	data := configMap.Data["controller_manager_config.yaml"]
	err = yaml.Unmarshal([]byte(data), &myConfig)
	if err != nil {
		return nil, err
	}

	return myConfig.ValidationPolicies, nil
}

func CheckValidationPack(myName string, myValue, validationType string, pack []configv1.ValidationStruct) (string, error) {
	if validationType == "unique" {
		for _, p := range pack {
			if p.Field == myValue {
				if myName != p.Name {
					return myValue, nil
				}
			}
		}
	}
	return "", nil
}
