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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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
	decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	sLog            = logger.NewLogger("coa.runtime")
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
	_, span := observability.StartSpan(
		"ConfigMap Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)
	sLog.Info("  P (ConfigMap Target): Init()")

	updateConfig, err := toConfigMapTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (ConfigMap Target): expected ConfigMapTargetProviderConfig - %+v", err)
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
					sLog.Errorf("  P (ConfigMap Target): %+v, traceId: %s", err, span.SpanContext().TraceID().String())
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "inline":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.Errorf("  P (ConfigMap Target): failed to read kube config: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.Errorf("  P (ConfigMap Target): %+v, traceId: %s", err, span.SpanContext().TraceID().String())
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and inline", v1alpha2.BadConfig)
			sLog.Errorf("  P (ConfigMap Target): %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return err
		}
	}
	if err != nil {
		sLog.Errorf("  P (ConfigMap Target): failed to get the cluster config: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (ConfigMap Target): failed to create a new clientset: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (ConfigMap Target): failed to create a dynamic client: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	i.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (ConfigMap Target): failed to create a discovery client: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	i.Mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(i.DiscoveryClient))
	i.RESTConfig = kConfig
	return nil
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
	sLog.Infof("  P (ConfigMap Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		var obj *corev1.ConfigMap
		obj, err = i.Client.CoreV1().ConfigMaps(deployment.Instance.Scope).Get(ctx, component.Component.Name, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				sLog.Infof("  P (ConfigMap Target): resource not found: %s, traceId: %s", err, span.SpanContext().TraceID().String())
				continue
			}
			sLog.Error("  P (ConfigMap Target): failed to read object: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
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

	sLog.Infof("  P (ConfigMap Target):  applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
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
						Namespace: deployment.Instance.Scope,
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
				i.ensureNamespace(ctx, deployment.Instance.Scope)
				err = i.applyConfigMap(ctx, newConfigMap, deployment.Instance.Scope)
				if err != nil {
					sLog.Error("  P (ConfigMap Target): failed to apply configmap: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
			}
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Type == "config" {
				err = i.deleteConfigMap(ctx, component.Name, deployment.Instance.Scope)
				if err != nil {
					sLog.Error("  P (ConfigMap Target): failed to delete configmap: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
			}
		}
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
	sLog.Infof("  P (ConfigMap Target):  ensureNamespace namespace - %s, traceId: %s", namespace, span.SpanContext().TraceID().String())

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
		if err != nil {
			sLog.Error("  P (ConfigMap Target): failed to create namespace: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
			return err
		}
	} else {
		sLog.Error("  P (ConfigMap Target): failed to get namespace: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	return nil
}

// GetValidationRule returns validation rule for the provider
func (*ConfigMapTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
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
	}
}

// deleteConfigMap deletes a configmap
func (i *ConfigMapTargetProvider) deleteConfigMap(ctx context.Context, name string, scope string) error {
	ctx, span := observability.StartSpan(
		"ConfigMap Target Provider",
		ctx,
		&map[string]string{
			"method": "deleteConfigMap",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (ConfigMap Target):  deleteConfigMap name %s, scope: %s, traceId: %s", name, scope, span.SpanContext().TraceID().String())

	err = i.Client.CoreV1().ConfigMaps(scope).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.Error("  P (Kubectl Target): failed to delete configmap: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
			return err
		}
	}
	return nil
}

// applyCustomResource applies a custom resource from a byte array
func (i *ConfigMapTargetProvider) applyConfigMap(ctx context.Context, config *corev1.ConfigMap, scope string) error {
	ctx, span := observability.StartSpan(
		"ConfigMap Target Provider",
		ctx,
		&map[string]string{
			"method": "applyConfigMap",
		},
	)
	var err error = nil
	defer utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (ConfigMap Target):  applyConfigMap scope: %s, traceId: %s", scope, span.SpanContext().TraceID().String())

	existingConfigMap, err := i.Client.CoreV1().ConfigMaps(scope).Get(ctx, config.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			sLog.Infof("  P (ConfigMap Target): resource not found: %s", err)
			_, err = i.Client.CoreV1().ConfigMaps(scope).Create(ctx, config, metav1.CreateOptions{})
			if err != nil {
				sLog.Error("  P (ConfigMap Target): failed to create configmap: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
				return err
			}
			return nil
		}
		sLog.Error("  P (ConfigMap Target): failed to read object: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	existingConfigMap.Data = config.Data

	_, err = i.Client.CoreV1().ConfigMaps(scope).Update(ctx, existingConfigMap, metav1.UpdateOptions{})
	if err != nil {
		sLog.Error("  P (ConfigMap Target): failed to update configmap: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}
	return nil
}
