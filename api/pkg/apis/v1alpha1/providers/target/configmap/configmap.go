/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package configmap

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	decUnstructured          = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	loggerName               = "providers.target.configmap"
	providerName             = "P (ConfigMap Target)"
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
	configmap                = "configmap"
)

type (
	// ConfigMapTargetProviderConfig is the configuration for the kubectl target provider
	ConfigMapTargetProviderConfig struct {
		Name       string `json:"name,omitempty"`
		ConfigType string `json:"configType,omitempty"`
		ConfigData string `json:"configData,omitempty"`
		Context    string `json:"context,omitempty"`
		InCluster  bool   `json:"inCluster"`
	}

	// ConfigMapTargetProvider is the kubectl target provider
	ConfigMapTargetProvider struct {
		Config          ConfigMapTargetProviderConfig
		Context         *contexts.ManagerContext
		Client          kubernetes.Interface
		DynamicClient   dynamic.Interface
		DiscoveryClient *discovery.DiscoveryClient
		Mapper          *restmapper.DeferredDiscoveryRESTMapper
		RESTConfig      *rest.Config
	}
)

// ConfigMapTargetProviderConfigFromMap converts a map to a ConfigMapTargetProviderConfig
func ConfigMapTargetProviderConfigFromMap(properties map[string]string) (ConfigMapTargetProviderConfig, error) {
	ret := ConfigMapTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["configType"]; ok {
		ret.ConfigType = v
	}
	if v, ok := properties["configData"]; ok {
		ret.ConfigData = v
	}
	if v, ok := properties["context"]; ok {
		ret.Context = v
	}
	if v, ok := properties["inCluster"]; ok {
		val := v
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'inCluster' setting of kubectl provider", v1alpha2.BadConfig)
			}
			ret.InCluster = bVal
		}
	}
	return ret, nil
}

func (s *ConfigMapTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

// InitWithMap initializes the configmap target provider with a map
func (i *ConfigMapTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := ConfigMapTargetProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (ConfigMap Target): expected ConfigMapTargetProviderConfig %+v", err)
		return err
	}

	return i.Init(config)
}

// Init initializes the configmap target provider
func (i *ConfigMapTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"ConfigMap Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)
	sLog.InfoCtx(ctx, "  P (ConfigMap Target): Init()")

	updateConfig, err := toConfigMapTargetProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): expected ConfigMapTargetProviderConfig - %+v", err)
		return err
	}

	i.Config = updateConfig
	var kConfig *rest.Config
	if i.Config.InCluster {
		kConfig, err = rest.InClusterConfig()
	} else {
		switch i.Config.ConfigType {
		case "path":
			if i.Config.ConfigData == "" {
				if home := homedir.HomeDir(); home != "" {
					i.Config.ConfigData = filepath.Join(home, ".kube", "config")
				} else {
					err = v1alpha2.NewCOAError(nil, "can't locate home direction to read default kubernetes config file, to run in cluster, set inCluster config setting to true", v1alpha2.BadConfig)
					sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "inline":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to read kube config: %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and inline", v1alpha2.BadConfig)
			sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): %+v", err)
			return err
		}
	}
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to get the cluster config: %+v", err)
		return err
	}

	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to create a new clientset: %+v", err)
		return err
	}

	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to create a dynamic client: %+v", err)
		return err
	}

	i.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to create a discovery client: %+v", err)
		return err
	}

	i.Mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(i.DiscoveryClient))
	i.RESTConfig = kConfig

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to create metrics: %+v", err)
			}
		}
	})

	return err
}

// toConfigMapTargetProviderConfig converts a generic IProviderConfig to a ConfigMapTargetProviderConfig
func toConfigMapTargetProviderConfig(config providers.IProviderConfig) (ConfigMapTargetProviderConfig, error) {
	ret := ConfigMapTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}

// Get gets the artifacts for a configmap
func (i *ConfigMapTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"ConfigMap Target Provider",
		ctx, &map[string]string{
			"method": "Get",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)
	sLog.InfofCtx(ctx, "  P (ConfigMap Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		var obj *corev1.ConfigMap
		obj, err = i.Client.CoreV1().ConfigMaps(deployment.Instance.Spec.Scope).Get(ctx, component.Component.Name, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				sLog.InfofCtx(ctx, "  P (ConfigMap Target): resource not found: %s", err)
				continue
			}
			sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to read object: %+v", err)
			return nil, err
		}
		component.Component.Properties = make(map[string]interface{})
		for key, value := range obj.Data {
			var data interface{}
			err = json.Unmarshal([]byte(value), &data)
			if err == nil {
				component.Component.Properties[key] = data
			} else {
				component.Component.Properties[key] = value
			}
		}
		ret = append(ret, component.Component)
	}

	return ret, nil
}

// Apply applies the configmap artifacts
func (i *ConfigMapTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan(
		"ConfigMap Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)

	sLog.InfofCtx(ctx, "  P (ConfigMap Target):  applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := utils.GetFunctionName()
	applyTime := time.Now().UTC()
	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		providerOperationMetrics.ProviderOperationErrors(
			configmap,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.UpdateOperationType,
			v1alpha2.ValidateFailed.String(),
		)
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()
	components = step.GetUpdatedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Type == "config" {
				newConfigMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      component.Name,
						Namespace: deployment.Instance.Spec.Scope,
					},
					Data: make(map[string]string),
				}
				for key, value := range component.Properties {
					if v, ok := value.(string); ok {
						newConfigMap.Data[key] = v
					} else {
						jData, _ := json.Marshal(value)
						newConfigMap.Data[key] = string(jData)
					}
				}
				i.ensureNamespace(ctx, deployment.Instance.Spec.Scope)
				err = i.applyConfigMap(ctx, newConfigMap, deployment.Instance.Spec.Scope)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to apply configmap: %+v", err)
					providerOperationMetrics.ProviderOperationErrors(
						configmap,
						functionName,
						metrics.ApplyOperation,
						metrics.UpdateOperationType,
						v1alpha2.ConfigMapApplyFailed.String(),
					)
					return ret, err
				}
			}
		}
		providerOperationMetrics.ProviderOperationLatency(
			applyTime,
			configmap,
			functionName,
			metrics.ApplyOperation,
			metrics.UpdateOperationType,
		)
	}
	deleteTime := time.Now().UTC()
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Type == "config" {
				err = i.deleteConfigMap(ctx, component.Name, deployment.Instance.Spec.Scope)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to delete configmap: %+v", err)
					providerOperationMetrics.ProviderOperationErrors(
						configmap,
						functionName,
						metrics.ApplyOperation,
						metrics.DeleteOperationType,
						v1alpha2.ConfigMapApplyFailed.String(),
					)
					return ret, err
				}
			}
		}
		providerOperationMetrics.ProviderOperationLatency(
			deleteTime,
			configmap,
			functionName,
			metrics.ApplyOperation,
			metrics.DeleteOperationType,
		)
	}
	return ret, nil
}

// ensureNamespace ensures that the namespace exists
func (k *ConfigMapTargetProvider) ensureNamespace(ctx context.Context, namespace string) error {
	ctx, span := observability.StartSpan(
		"ConfigMap Target Provider",
		ctx,
		&map[string]string{
			"method": "ensureNamespace",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)
	sLog.InfofCtx(ctx, "  P (ConfigMap Target):  ensureNamespace namespace - %s", namespace)

	if namespace == "" || namespace == "default" {
		return nil
	}

	_, err = k.Client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if kerrors.IsNotFound(err) {
		_, err = k.Client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
		if err != nil && !kerrors.IsAlreadyExists(err) {
			sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to create namespace: %+v", err)
			return err
		}
	} else {
		sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to get namespace: %+v", err)
		return err
	}

	return nil
}

// GetValidationRule returns validation rule for the provider
func (*ConfigMapTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{
					Name: "*", //react to all property changes
				},
			},
		},
	}
}

// deleteConfigMap deletes a configmap
func (i *ConfigMapTargetProvider) deleteConfigMap(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan(
		"ConfigMap Target Provider",
		ctx,
		&map[string]string{
			"method": "deleteConfigMap",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)
	sLog.InfofCtx(ctx, "  P (ConfigMap Target):  deleteConfigMap name %s, namespace: %s", name, namespace)

	err = i.Client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to delete configmap: %+v", err)
			return err
		}
	}
	return nil
}

// applyCustomResource applies a custom resource from a byte array
func (i *ConfigMapTargetProvider) applyConfigMap(ctx context.Context, config *corev1.ConfigMap, namespace string) error {
	ctx, span := observability.StartSpan(
		"ConfigMap Target Provider",
		ctx,
		&map[string]string{
			"method": "applyConfigMap",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)
	sLog.InfofCtx(ctx, "  P (ConfigMap Target):  applyConfigMap namespace: %s", namespace)

	existingConfigMap, err := i.Client.CoreV1().ConfigMaps(namespace).Get(ctx, config.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			sLog.InfofCtx(ctx, "  P (ConfigMap Target): resource not found: %s", err)
			_, err = i.Client.CoreV1().ConfigMaps(namespace).Create(ctx, config, metav1.CreateOptions{})
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to create configmap: %+v", err)
				return err
			}
			return nil
		}
		sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to read object: %+v", err)
		return err
	}

	existingConfigMap.Data = config.Data

	_, err = i.Client.CoreV1().ConfigMaps(namespace).Update(ctx, existingConfigMap, metav1.UpdateOptions{})
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (ConfigMap Target): failed to update configmap: %+v", err)
		return err
	}
	return nil
}
