/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package ingress

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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
	// IngressTargetProviderConfig is the configuration for the ingress target provider
	IngressTargetProviderConfig struct {
		Name       string `json:"name,omitempty"`
		ConfigType string `json:"configType,omitempty"`
		ConfigData string `json:"configData,omitempty"`
		Context    string `json:"context,omitempty"`
		InCluster  bool   `json:"inCluster"`
	}

	// IngressTargetProvider is the kubectl target provider
	IngressTargetProvider struct {
		Config          IngressTargetProviderConfig
		Context         *contexts.ManagerContext
		Client          kubernetes.Interface
		DynamicClient   dynamic.Interface
		DiscoveryClient *discovery.DiscoveryClient
		Mapper          *restmapper.DeferredDiscoveryRESTMapper
		RESTConfig      *rest.Config
	}
)

// IngressTargetProviderConfigFromMap converts a map to a IngressTargetProviderConfig
func IngressTargetProviderConfigFromMap(properties map[string]string) (IngressTargetProviderConfig, error) {
	ret := IngressTargetProviderConfig{}
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
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'inCluster' setting of ingress provider", v1alpha2.BadConfig)
			}
			ret.InCluster = bVal
		}
	}
	return ret, nil
}

func (s *IngressTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

// InitWithMap initializes the ingress target provider with a map
func (i *IngressTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := IngressTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}

	return i.Init(config)
}

// Init initializes the ingress target provider
func (i *IngressTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan(
		"Ingress Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	sLog.Info("  P (Ingress Target): Init()")

	updateConfig, err := toIngressTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Ingress Target): expected IngressTargetProviderConfig - %+v", err)
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
					sLog.Errorf("  P (Ingress Target): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "inline":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.Errorf("  P (Ingress Target):  %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.Errorf("  P (Ingress Target): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and inline", v1alpha2.BadConfig)
			sLog.Errorf("  P (Ingress Target): %+v", err)
			return err
		}
	}
	if err != nil {
		sLog.Errorf("  P (Ingress Target): failed to get the cluster config: %+v", err)
		return err
	}

	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (Ingress Target): failed to create a new clientset: %+v", err)
		return err
	}

	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (Ingress Target): failed to create a dynamic client: %+v", err)
		return err
	}

	i.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (Ingress Target): failed to create a discovery client: %+v", err)
		return err
	}

	i.Mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(i.DiscoveryClient))
	i.RESTConfig = kConfig
	return nil
}

// toIngressTargetProviderConfig converts a generic IProviderConfig to a IngressTargetProviderConfig
func toIngressTargetProviderConfig(config providers.IProviderConfig) (IngressTargetProviderConfig, error) {
	ret := IngressTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}

// Get gets the artifacts for a ingress
func (i *IngressTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"Ingress Target Provider",
		ctx, &map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Ingress Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		var obj *networkingv1.Ingress
		obj, err = i.Client.NetworkingV1().Ingresses(deployment.Instance.Scope).Get(ctx, component.Component.Name, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				sLog.Infof("  P (Ingress Target): resource not found: %v, traceId: %s", err, span.SpanContext().TraceID().String())
				continue
			}
			sLog.Errorf("  P (Ingress Target): failed to read object: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return nil, err
		}
		component.Component.Properties = make(map[string]interface{})

		component.Component.Properties["rules"] = obj.Spec.Rules
		ret = append(ret, component.Component)
	}

	return ret, nil
}

// Apply applies the ingress artifacts
func (i *IngressTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan(
		"Ingress Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Ingress Target):  applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		return nil, err
	}
	if isDryRun {
		return nil, nil
	}

	ret := step.PrepareResultMap()
	components = step.GetUpdatedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Type == "ingress" {
				newIngress := &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      component.Name,
						Namespace: deployment.Instance.Scope,
					},
					Spec: networkingv1.IngressSpec{
						Rules: make([]networkingv1.IngressRule, 0),
					},
				}
				if v, ok := component.Properties["rules"]; ok {
					jData, _ := json.Marshal(v)
					var rules []networkingv1.IngressRule
					err = json.Unmarshal(jData, &rules)
					if err != nil {
						sLog.Errorf("  P (Ingress Target): failed to unmarshal ingress: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
						return ret, err
					}
					newIngress.Spec.Rules = rules
				}

				if v, ok := component.Properties["ingressClassName"]; ok {
					s, ok := v.(string)
					if ok {
						newIngress.Spec.IngressClassName = &s
					} else {
						sLog.Errorf("  P (Ingress Target): failed to convert ingress class name: %+v, traceId: %s", v, span.SpanContext().TraceID().String())
						return ret, err
					}
				}

				for k, v := range component.Metadata {
					if strings.HasPrefix(k, "annotations.") {
						if newIngress.ObjectMeta.Annotations == nil {
							newIngress.ObjectMeta.Annotations = make(map[string]string)
						}
						newIngress.ObjectMeta.Annotations[k[12:]] = v
					}
				}

				i.ensureNamespace(ctx, deployment.Instance.Scope)
				err = i.applyIngress(ctx, newIngress, deployment.Instance.Scope)
				if err != nil {
					sLog.Errorf("  P (Ingress Target): failed to apply ingress: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
			}
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Type == "ingress" {
				err = i.deleteIngress(ctx, component.Name, deployment.Instance.Scope)
				if err != nil {
					sLog.Errorf("  P (Ingress Target): failed to delete ingress: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
			}
		}
	}
	return ret, nil
}

// ensureNamespace ensures that the namespace exists
func (k *IngressTargetProvider) ensureNamespace(ctx context.Context, namespace string) error {
	_, span := observability.StartSpan(
		"Ingress Target Provider",
		ctx,
		&map[string]string{
			"method": "ensureNamespace",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Ingress Target): ensureNamespace %s, traceId: %s", namespace, span.SpanContext().TraceID().String())

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
			sLog.Errorf("  P (Ingress Target): failed to create namespace: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return err
		}
	} else {
		sLog.Errorf("  P (Ingress Target): failed to get namespace: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	return nil
}

// GetValidationRule returns validation rule for the provider
func (*IngressTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
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
		ChangeDetectionMetadata: []model.PropertyDesc{
			{
				Name: "annotations.*", //react to all annotation changes
			},
		},
	}
}

// deleteConfigMap deletes a configmap
func (i *IngressTargetProvider) deleteIngress(ctx context.Context, name string, scope string) error {
	_, span := observability.StartSpan(
		"Ingress Target Provider",
		ctx,
		&map[string]string{
			"method": "deleteIngress",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Ingress Target): deleteIngress name %s, scope %s, traceId: %s", name, scope, span.SpanContext().TraceID().String())

	err = i.Client.NetworkingV1().Ingresses(scope).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.Errorf("  P (Ingress Target): failed to delete ingress: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return err
		}
	}
	return nil
}

// applyCustomResource applies a custom resource from a byte array
func (i *IngressTargetProvider) applyIngress(ctx context.Context, ingress *networkingv1.Ingress, scope string) error {
	_, span := observability.StartSpan(
		"Ingress Target Provider",
		ctx,
		&map[string]string{
			"method": "applyIngress",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Ingress Target): applyIngress scope %s, name %s, traceId: %s", scope, ingress.Name, span.SpanContext().TraceID().String())

	existingIngress, err := i.Client.NetworkingV1().Ingresses(scope).Get(ctx, ingress.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			sLog.Infof("  P (Ingress Target): resource not found: %v, traceId: %s", err, span.SpanContext().TraceID().String())
			_, err = i.Client.NetworkingV1().Ingresses(scope).Create(ctx, ingress, metav1.CreateOptions{})
			if err != nil {
				sLog.Errorf("  P (Ingress Target): failed to create ingress: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
				return err
			}
			return nil
		}
		sLog.Errorf("  P (Ingress Target): failed to read object: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	existingIngress.Spec.Rules = ingress.Spec.Rules
	if ingress.ObjectMeta.Annotations != nil {
		existingIngress.ObjectMeta.Annotations = ingress.ObjectMeta.Annotations
	}
	_, err = i.Client.NetworkingV1().Ingresses(scope).Update(ctx, existingIngress, metav1.UpdateOptions{})
	if err != nil {
		sLog.Errorf("  P (Ingress Target): failed to update ingress: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}
	return nil
}
