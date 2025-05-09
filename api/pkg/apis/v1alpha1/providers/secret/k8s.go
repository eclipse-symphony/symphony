package secret

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const loggerName = "providers.secret.k8s"

var sLog = logger.NewLogger(loggerName)

type K8sSecretProviderConfig struct {
	Name       string `json:"name"`
	ConfigType string `json:"configType,omitempty"`
	ConfigData string `json:"configData,omitempty"`
	InCluster  bool   `json:"inCluster"`
}

type K8sInterface interface {
	CoreV1() corev1.CoreV1Interface
}

type K8sSecretProvider struct {
	Clientset K8sInterface
	Config    K8sSecretProviderConfig
}

func K8sSecretProviderConfigFromMap(properties map[string]string) (K8sSecretProviderConfig, error) {
	ret := K8sSecretProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["configType"]; ok {
		ret.ConfigType = v
	}
	if v, ok := properties["configData"]; ok {
		ret.ConfigData = v
	}
	if ret.ConfigType == "" {
		ret.ConfigType = "path"
	}
	if v, ok := properties["inCluster"]; ok {
		val := v
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'inCluster' setting of K8s secret provider", v1alpha2.BadConfig)
			}
			ret.InCluster = bVal
		}
	}
	return ret, nil
}

func (s *K8sSecretProvider) InitWithMap(properties map[string]string) error {
	config, err := K8sSecretProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return s.Init(config)
}

func (s *K8sSecretProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("K8s Secret Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Debug("  P (K8s Secret): initialize")

	updateConfig, err := toK8sStateProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (K8s Secret): expected KubectlTargetProviderConfig: %+v", err)
		return err
	}
	s.Config = updateConfig
	var kConfig *rest.Config
	if s.Config.InCluster {
		kConfig, err = rest.InClusterConfig()
	} else {
		switch s.Config.ConfigType {
		case "path":
			if s.Config.ConfigData == "" {
				if home := homedir.HomeDir(); home != "" {
					s.Config.ConfigData = filepath.Join(home, ".kube", "config")
				} else {
					err = v1alpha2.NewCOAError(nil, "can't locate home direction to read default kubernetes config file, to run in cluster, set inCluster config setting to true", v1alpha2.BadConfig)
					sLog.Errorf("  P (K8s secret): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", s.Config.ConfigData)
		case "bytes":
			if s.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(s.Config.ConfigData))
				if err != nil {
					sLog.Errorf("  P (K8s secret): %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.Errorf("  P (K8s secret): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and bytes", v1alpha2.BadConfig)
			sLog.Errorf("  P (K8s secret): %+v", err)
			return err
		}
	}
	if err != nil {
		sLog.Errorf("  P (K8s secret): %+v", err)
		return err
	}
	s.Clientset, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (K8s secret): %+v", err)
		return err
	}

	return nil
}

func toK8sStateProviderConfig(config providers.IProviderConfig) (K8sSecretProviderConfig, error) {
	ret := K8sSecretProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	if ret.ConfigType == "" {
		ret.ConfigType = "path"
	}
	return ret, err
}

func (s *K8sSecretProvider) Read(ctx context.Context, name string, field string, localContext interface{}) (string, error) {
	// Get the secret
	ctx, span := observability.StartSpan("K8s Secret Provider", ctx, &map[string]string{
		"method": "Read",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	jsonData, err := json.Marshal(localContext)
	sLog.DebugfCtx(ctx, "  P (K8s Secret): read secret %s field %s with context %s", name, field, string(jsonData))
	namespace := utils.GetNamespaceFromContext(localContext)
	secret, err := s.Clientset.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		sLog.ErrorfCtx(ctx, "Error getting secret %s in namespace %s. Error: %s", name, namespace, err.Error())
		return "", err
	}
	// Get the field from the secret data
	value, ok := secret.Data[field]
	if !ok {
		sLog.ErrorfCtx(ctx, "Field %s not found in secret %s", field, name)
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("field %s not found in secret %s", field, name), v1alpha2.MissingConfig)
		return "", err
	}

	return string(value), nil
}
