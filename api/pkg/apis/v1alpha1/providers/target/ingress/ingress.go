/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
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

const (
	loggerName   = "providers.target.ingress"
	providerName = "P (Ingress Target)"
	ingress      = "ingress"
)

var (
	decUnstructured          = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
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
		sLog.Errorf("  P (Ingress Target): expected IngressTargetProviderConfig: %+v", err)
		return err
	}

	return i.Init(config)
}

// Init initializes the ingress target provider
func (i *IngressTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"Ingress Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfoCtx(ctx, "  P (Ingress Target): Init()")

	updateConfig, err := toIngressTargetProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): expected IngressTargetProviderConfig - %+v", err)
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
					sLog.ErrorfCtx(ctx, "  P (Ingress Target): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "inline":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Ingress Target):  %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.ErrorfCtx(ctx, "  P (Ingress Target): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and inline", v1alpha2.BadConfig)
			sLog.ErrorfCtx(ctx, "  P (Ingress Target): %+v", err)
			return err
		}
	}
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to get the cluster config: %+v", err)
		return err
	}

	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to create a new clientset: %+v", err)
		return err
	}

	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to create a dynamic client: %+v", err)
		return err
	}

	i.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to create a discovery client: %+v", err)
		return err
	}

	i.Mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(i.DiscoveryClient))
	i.RESTConfig = kConfig

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to create metrics: %+v", err)
			}
		}
	})

	return err
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
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Ingress Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		var obj *networkingv1.Ingress
		obj, err = i.Client.NetworkingV1().Ingresses(deployment.Instance.Spec.Scope).Get(ctx, component.Component.Name, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				sLog.InfofCtx(ctx, "  P (Ingress Target): resource %s not found: %v", component.Component.Name, err)
				continue
			}
			sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to read object: %+v", err)
			return nil, err
		}
		component.Component.Properties = make(map[string]interface{})

		component.Component.Properties["rules"] = obj.Spec.Rules
		sLog.InfofCtx(ctx, "  P (Ingress Target): append component: %s", component.Component.Name)
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
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Ingress Target):  applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := utils.GetFunctionName()
	applyTime := time.Now().UTC()
	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to validate components: %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			ingress,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.UpdateOperationType,
			v1alpha2.ValidateFailed.String(),
		)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: the rule validation failed", providerName), v1alpha2.ValidateFailed)
		return nil, err
	}
	if isDryRun {
		sLog.DebugCtx(ctx, "  P (Ingress Target): dryRun is enabled, skipping apply")
		return nil, nil
	}

	ret := step.PrepareResultMap()
	components = step.GetUpdatedComponents()
	if len(components) > 0 {
		sLog.InfofCtx(ctx, "  P (Ingress Target): get updated components: count - %d", len(components))
		for _, component := range components {
			if component.Type == "ingress" {
				newIngress := &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      component.Name,
						Namespace: deployment.Instance.Spec.Scope,
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
						sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to unmarshal ingress rules: %+v", err)
						err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: parse ingress rules failed", providerName), v1alpha2.BadConfig)
						providerOperationMetrics.ProviderOperationErrors(
							ingress,
							functionName,
							metrics.ApplyOperation,
							metrics.UpdateOperationType,
							v1alpha2.BadConfig.String(),
						)
						return ret, err
					}
					newIngress.Spec.Rules = rules
				}

				if v, ok := component.Properties["ingressClassName"]; ok {
					s, ok := v.(string)
					if ok {
						newIngress.Spec.IngressClassName = &s
					} else {
						sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to convert ingress class name: %+v", v)
						err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to convert ingress class name", providerName), v1alpha2.BadConfig)
						providerOperationMetrics.ProviderOperationErrors(
							ingress,
							functionName,
							metrics.ApplyOperation,
							metrics.UpdateOperationType,
							v1alpha2.BadConfig.String(),
						)
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

				i.ensureNamespace(ctx, deployment.Instance.Spec.Scope)
				err = i.applyIngress(ctx, newIngress, deployment.Instance.Spec.Scope)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to apply ingress: %+v", err)
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to apply ingress", providerName), v1alpha2.IngressApplyFailed)
					providerOperationMetrics.ProviderOperationErrors(
						ingress,
						functionName,
						metrics.ApplyOperation,
						metrics.UpdateOperationType,
						v1alpha2.IngressApplyFailed.String(),
					)
					return ret, err
				}
			}
		}
		providerOperationMetrics.ProviderOperationLatency(
			applyTime,
			ingress,
			functionName,
			metrics.ApplyOperation,
			metrics.UpdateOperationType,
		)
	}
	deleteTime := time.Now().UTC()
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		sLog.InfofCtx(ctx, "  P (Ingress Target): get deleted components: count - %d", len(components))
		for _, component := range components {
			if component.Type == "ingress" {
				err = i.deleteIngress(ctx, component.Name, deployment.Instance.Spec.Scope)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to delete ingress: %+v", err)
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to delete ingress", providerName), v1alpha2.IngressApplyFailed)
					providerOperationMetrics.ProviderOperationErrors(
						ingress,
						functionName,
						metrics.ApplyOperation,
						metrics.DeleteOperationType,
						v1alpha2.IngressApplyFailed.String(),
					)
					return ret, err
				}
			}
		}
		providerOperationMetrics.ProviderOperationLatency(
			deleteTime,
			ingress,
			functionName,
			metrics.ApplyOperation,
			metrics.DeleteOperationType,
		)
	}
	return ret, nil
}

// ensureNamespace ensures that the namespace exists
func (k *IngressTargetProvider) ensureNamespace(ctx context.Context, namespace string) error {
	ctx, span := observability.StartSpan(
		"Ingress Target Provider",
		ctx,
		&map[string]string{
			"method": "ensureNamespace",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Ingress Target): ensureNamespace %s", namespace)

	_, err = k.Client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if kerrors.IsNotFound(err) {
		sLog.InfofCtx(ctx, "  P (Ingress Target): start to create namespace %s", namespace)
		utils.EmitUserAuditsLogs(ctx, "  P (Ingress Target): Start to create namespace - %s", namespace)
		_, err = k.Client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
		if err != nil && !kerrors.IsAlreadyExists(err) {
			sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to create namespace: %+v", err)
			return err
		}
	} else {
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to get namespace: %+v", err)
		return err
	}

	return nil
}

// GetValidationRule returns validation rule for the provider
func (*IngressTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
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
			ChangeDetectionMetadata: []model.PropertyDesc{
				{
					Name: "annotations.*", //react to all annotation changes
				},
			},
		},
	}
}

// deleteConfigMap deletes a configmap
func (i *IngressTargetProvider) deleteIngress(ctx context.Context, name string, namespace string) error {
	_, span := observability.StartSpan(
		"Ingress Target Provider",
		ctx,
		&map[string]string{
			"method": "deleteIngress",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Ingress Target): deleteIngress name %s, namespace %s", name, namespace)

	utils.EmitUserAuditsLogs(ctx, "  P (Ingress Target): Start to delete ingress - %s", name)
	err = i.Client.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to delete ingress: %+v", err)
			return err
		} else {
			sLog.InfoCtx(ctx, "  P (Ingress Target): ingress %s is not found in the namespace %s", name, namespace)
		}
	}
	return nil
}

// applyCustomResource applies a custom resource from a byte array
func (i *IngressTargetProvider) applyIngress(ctx context.Context, ingress *networkingv1.Ingress, namespace string) error {
	ctx, span := observability.StartSpan(
		"Ingress Target Provider",
		ctx,
		&map[string]string{
			"method": "applyIngress",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Ingress Target): applyIngress namespace %s, name %s", namespace, ingress.Name)

	existingIngress, err := i.Client.NetworkingV1().Ingresses(namespace).Get(ctx, ingress.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			sLog.InfofCtx(ctx, "  P (Ingress Target): resource not found: %v", err)
			utils.EmitUserAuditsLogs(ctx, "  P (Ingress Target): Start to create ingress - %s", ingress.Name)
			_, err = i.Client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to create ingress: %+v", err)
				return err
			}
			return nil
		}
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to read object %s in the namespace %s: %+v", ingress.Name, namespace, err)
		return err
	}

	sLog.InfofCtx(ctx, "  P (Ingress Target): resource exists and start to update ingress %s in the namespace", ingress.Name, namespace)
	existingIngress.Spec.Rules = ingress.Spec.Rules
	if ingress.ObjectMeta.Annotations != nil {
		existingIngress.ObjectMeta.Annotations = ingress.ObjectMeta.Annotations
	}
	utils.EmitUserAuditsLogs(ctx, "  P (Ingress Target): Start to update ingress - %s", ingress.Name)
	_, err = i.Client.NetworkingV1().Ingresses(namespace).Update(ctx, existingIngress, metav1.UpdateOptions{})
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Ingress Target): failed to update ingress: %+v", err)
		return err
	}
	return nil
}
