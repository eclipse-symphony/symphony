/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package kubectl

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils/metahelper"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/oliveagle/jsonpath"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
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
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

const (
	kubectl     = "kubectl"
	timeout     = "5m"
	interval    = "5s"
	initialWait = "1m"

	providerName = "P (Kubectl Target)"
	loggerName   = "providers.target.kubectl"
)

type (
	// KubectlTargetProviderConfig is the configuration for the kubectl target provider
	KubectlTargetProviderConfig struct {
		Name       string `json:"name,omitempty"`
		ConfigType string `json:"configType,omitempty"`
		ConfigData string `json:"configData,omitempty"`
		Context    string `json:"context,omitempty"`
		InCluster  bool   `json:"inCluster"`
	}

	// KubectlTargetProvider is the kubectl target provider
	KubectlTargetProvider struct {
		Config          KubectlTargetProviderConfig
		Context         *contexts.ManagerContext
		Client          kubernetes.Interface
		DynamicClient   dynamic.Interface
		DiscoveryClient *discovery.DiscoveryClient
		Mapper          *restmapper.DeferredDiscoveryRESTMapper
		RESTConfig      *rest.Config
		MetaPopulator   metahelper.MetaPopulator
	}

	// StatusProbe is the expected resource status property
	StatusProbe struct {
		SucceededValues  []string `json:"succeededValues,omitempty"`
		FailedValues     []string `json:"failedValues,omitempty"`
		StatusPath       string   `json:"statusPath,omitempty"`
		ErrorMessagePath string   `json:"errorMessagePath,omitempty"`
		Timeout          string   `json:"timeout,omitempty"`
		Interval         string   `json:"interval,omitempty"`
		InitialWait      string   `json:"initialWait,omitempty"`
	}
)

// KubectlTargetProviderConfigFromMap converts a map to a KubectlTargetProviderConfig
func KubectlTargetProviderConfigFromMap(properties map[string]string) (KubectlTargetProviderConfig, error) {
	ret := KubectlTargetProviderConfig{}
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

// InitWithMap initializes the kubectl target provider with a map
func (i *KubectlTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := KubectlTargetProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (Kubectl Target): expected KubectlTargetProviderConfig: %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init", providerName), v1alpha2.InitFailed)
	}

	return i.Init(config)
}

func (s *KubectlTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

// Init initializes the kubectl target provider
func (i *KubectlTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"Kubectl Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfoCtx(ctx, "  P (Kubectl Target): Init()")

	updateConfig, err := toKubectlTargetProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): expected KubectlTargetProviderConfig - %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to convert to KubectlTargetProviderConfig", providerName), v1alpha2.InitFailed)
		return err
	}

	i.Config = updateConfig
	var kConfig *rest.Config
	kConfig, err = i.getKubernetesConfig(ctx)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to get the cluster config: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get kubernetes config", providerName), v1alpha2.InitFailed)
		return err
	}

	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to create a new clientset: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create kubernetes client", providerName), v1alpha2.InitFailed)
		return err
	}

	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to create a dynamic client: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create dynamic client", providerName), v1alpha2.InitFailed)
		return err
	}

	i.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to create a discovery client: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create discovery client", providerName), v1alpha2.InitFailed)
		return err
	}

	i.MetaPopulator, err = metahelper.NewMetaPopulator(metahelper.WithDefaultPopulators())
	if err != nil {
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create meta populator", providerName), v1alpha2.InitFailed)
		sLog.ErrorCtx(ctx, err)
		return err
	}

	i.Mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(i.DiscoveryClient))
	i.RESTConfig = kConfig

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorCtx(ctx, err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init metrics", providerName), v1alpha2.InitFailed)
			}
		}
	})
	return err
}

func (i *KubectlTargetProvider) getKubernetesConfig(ctx context.Context) (*rest.Config, error) {
	if i.Config.InCluster {
		return rest.InClusterConfig()
	}

	switch i.Config.ConfigType {
	case "path":
		return i.getConfigFromPath(ctx)
	case "inline":
		return i.getConfigFromInline(ctx)
	default:
		err := v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and inline", v1alpha2.BadConfig)
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to get the cluster config: %+v", err)
		return nil, err
	}
}

func (i *KubectlTargetProvider) getConfigFromPath(ctx context.Context) (*rest.Config, error) {
	if i.Config.ConfigData == "" {
		home := homedir.HomeDir()
		if home == "" {
			err := v1alpha2.NewCOAError(nil, "can't locate home directory to read default kubernetes config file. To run in cluster, set inCluster to true", v1alpha2.BadConfig)
			sLog.ErrorfCtx(ctx, "  P (Kubectl Target): %+v", err)
			return nil, err
		}
		i.Config.ConfigData = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
}

func (i *KubectlTargetProvider) getConfigFromInline(ctx context.Context) (*rest.Config, error) {
	if i.Config.ConfigData == "" {
		err := v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): %+v", err)
		return nil, err
	}
	return clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
}

// toKubectlTargetProviderConfig converts a generic IProviderConfig to a KubectlTargetProviderConfig
func toKubectlTargetProviderConfig(config providers.IProviderConfig) (KubectlTargetProviderConfig, error) {
	ret := KubectlTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}

// Get gets the artifacts for a deployment
func (i *KubectlTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"Kubectl Target Provider",
		ctx, &map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Kubectl Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		if v, ok := component.Component.Properties["yaml"].(string); ok {
			chanMes, chanErr := readYaml(v)
			stop := false
			for !stop {
				select {
				case dataBytes, ok := <-chanMes:
					if !ok {
						sLog.ErrorfCtx(ctx, "  P (Kubectl Target): %+v", err)
						err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to receive from data channel when reading yaml property", providerName), v1alpha2.GetComponentSpecFailed)
						return nil, err
					}

					_, err = i.getCustomResource(ctx, dataBytes, deployment.Instance.Spec.Scope)
					if err != nil {
						if kerrors.IsNotFound(err) {
							sLog.InfofCtx(ctx, "  P (Kubectl Target): custom resource not found: %+v", err)
							continue
						}
						sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to read object: %+v", err)
						err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get custom resource from data bytes in yaml property", providerName), v1alpha2.GetComponentSpecFailed)
						return nil, err
					}

					sLog.InfofCtx(ctx, "  P (Kubectl Target): append component: %s", component.Component.Name)
					ret = append(ret, component.Component)
					stop = true //we do early stop as soon as we found the first resource. we may want to support different strategy in the future

				case err, ok := <-chanErr:
					if !ok {
						err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to receive from err channel when reading yaml property", providerName), v1alpha2.GetComponentSpecFailed)
						sLog.ErrorfCtx(ctx, "  P (Kubectl Target): %+v", err)
						return nil, err
					}

					if err == io.EOF {
						stop = true
					} else {
						sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to apply Yaml: %+v", err)
						err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to read yaml property", providerName), v1alpha2.GetComponentSpecFailed)
						return nil, err
					}
				}
			}
		} else if component.Component.Properties["resource"] != nil {
			var dataBytes []byte
			dataBytes, err = json.Marshal(component.Component.Properties["resource"])
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to get deployment bytes from component: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get data bytes from component resource property", providerName), v1alpha2.GetComponentSpecFailed)
				return nil, err
			}

			_, err = i.getCustomResource(ctx, dataBytes, deployment.Instance.Spec.Scope)
			if err != nil {
				if kerrors.IsNotFound(err) {
					sLog.InfofCtx(ctx, "  P (Kubectl Target): custom resource not found: %+v", err)
					continue
				}
				sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to read object: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get custom resource from data bytes in component resource property", providerName), v1alpha2.GetComponentSpecFailed)
				return nil, err
			}
			sLog.InfofCtx(ctx, "  P (Kubectl Target): append component: %s", component.Component.Name)
			ret = append(ret, component.Component)
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: component doesn't have yaml or resource property", providerName), v1alpha2.GetComponentSpecFailed)
			sLog.ErrorCtx(ctx, "  P (Kubectl Target): component doesn't have yaml or resource property")
			return nil, err
		}
	}

	return ret, nil
}

// Apply applies the deployment artifacts
func (i *KubectlTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan(
		"Kubectl Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Kubectl Target):  applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := utils.GetFunctionName()
	applyTime := time.Now().UTC()
	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		providerOperationMetrics.ProviderOperationErrors(
			kubectl,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.UpdateOperationType,
			v1alpha2.ValidateFailed.String(),
		)

		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to validate components, error: %v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: the rule validation failed", providerName), v1alpha2.ValidateFailed)
		return nil, err
	}
	if isDryRun {
		sLog.DebugfCtx(ctx, "  P (Kubectl Target): dryRun is enabled,, skipping apply")
		return nil, nil
	}

	ret := step.PrepareResultMap()
	components = step.GetUpdatedComponents()
	if len(components) > 0 {
		sLog.InfofCtx(ctx, "  P (Kubectl Target): get updated components: count - %d", len(components))
		for _, component := range components {
			applyComponentTime := time.Now().UTC()
			if component.Type == "yaml.k8s" {
				if v, ok := component.Properties["yaml"].(string); ok {
					chanMes, chanErr := readYaml(v)
					stop := false
					for !stop {
						select {
						case dataBytes, ok := <-chanMes:
							if !ok {
								err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to receive from data channel when reading yaml property", providerName), v1alpha2.ReadYamlFailed)
								sLog.ErrorfCtx(ctx, "  P (Kubectl Target): %+v", err)
								ret[component.Name] = model.ComponentResultSpec{
									Status:  v1alpha2.UpdateFailed,
									Message: err.Error(),
								}
								providerOperationMetrics.ProviderOperationErrors(
									kubectl,
									functionName,
									metrics.ReceiveDataChannelOperation,
									metrics.GetOperationType,
									v1alpha2.ReadYamlFailed.String(),
								)
								return ret, err
							}

							i.ensureNamespace(ctx, deployment.Instance.Spec.Scope)
							err = i.applyCustomResource(ctx, dataBytes, deployment.Instance.Spec.Scope, deployment.Instance)
							if err != nil {
								sLog.ErrorfCtx(ctx, "  P (Kubectl Target):  failed to apply Yaml: %+v", err)
								err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to apply Yaml", providerName), v1alpha2.ApplyYamlFailed)
								ret[component.Name] = model.ComponentResultSpec{
									Status:  v1alpha2.UpdateFailed,
									Message: err.Error(),
								}
								providerOperationMetrics.ProviderOperationErrors(
									kubectl,
									functionName,
									metrics.ApplyYamlOperation,
									metrics.UpdateOperationType,
									v1alpha2.ApplyYamlFailed.String(),
								)

								return ret, err
							}

							ret[component.Name] = model.ComponentResultSpec{
								Status:  v1alpha2.Updated,
								Message: fmt.Sprintf("No error. %s has been updated", component.Name),
							}

						case err, ok := <-chanErr:
							if !ok {
								err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to receive from error channel when reading yaml property", providerName), v1alpha2.ReadYamlFailed)
								sLog.ErrorfCtx(ctx, "  P (Kubectl Target): %+v", err)

								ret[component.Name] = model.ComponentResultSpec{
									Status:  v1alpha2.UpdateFailed,
									Message: err.Error(),
								}
								providerOperationMetrics.ProviderOperationErrors(
									kubectl,
									functionName,
									metrics.ReceiveErrorChannelOperation,
									metrics.UpdateOperationType,
									v1alpha2.ReadYamlFailed.String(),
								)
								return ret, err
							}

							if err == io.EOF {
								stop = true
							} else {
								sLog.ErrorfCtx(ctx, "  P (Kubectl Target):  failed to apply Yaml: %+v", err)
								err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to apply Yaml", providerName), v1alpha2.ApplyYamlFailed)

								ret[component.Name] = model.ComponentResultSpec{
									Status:  v1alpha2.UpdateFailed,
									Message: err.Error(),
								}
								providerOperationMetrics.ProviderOperationErrors(
									kubectl,
									functionName,
									metrics.ApplyYamlOperation,
									metrics.UpdateOperationType,
									v1alpha2.ApplyYamlFailed.String(),
								)
								return ret, err
							}
						}
					}
				} else if component.Properties["resource"] != nil {
					var dataBytes []byte
					dataBytes, err = json.Marshal(component.Properties["resource"])
					if err != nil {
						sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to convert resource data to bytes: %+v", err)
						err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to convert resource data to bytes", providerName), v1alpha2.ReadResourcePropertyFailed)
						ret[component.Name] = model.ComponentResultSpec{
							Status:  v1alpha2.UpdateFailed,
							Message: err.Error(),
						}
						providerOperationMetrics.ProviderOperationErrors(
							kubectl,
							functionName,
							metrics.ConvertResourceDataBytesOperation,
							metrics.UpdateOperationType,
							v1alpha2.ReadResourcePropertyFailed.String(),
						)

						return ret, err
					}

					i.ensureNamespace(ctx, deployment.Instance.Spec.Scope)
					err = i.applyCustomResource(ctx, dataBytes, deployment.Instance.Spec.Scope, deployment.Instance)
					if err != nil {
						sLog.ErrorfCtx(ctx, "  P (Kubectl Target):  failed to apply custom resource: %+v", err)
						err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to apply custom resource", providerName), v1alpha2.ApplyResourceFailed)
						ret[component.Name] = model.ComponentResultSpec{
							Status:  v1alpha2.UpdateFailed,
							Message: err.Error(),
						}
						providerOperationMetrics.ProviderOperationErrors(
							kubectl,
							functionName,
							metrics.ApplyCustomResource,
							metrics.UpdateOperationType,
							v1alpha2.ApplyResourceFailed.String(),
						)

						return ret, err
					}

					// check the resource status
					if component.Properties["statusProbe"] != nil {
						//check the status propbe property
						statusProbe, err := toStatusProbe(component.Properties["statusProbe"])
						if err != nil {
							sLog.ErrorfCtx(ctx, "Status property is not correctly defined: +%v", err)
						}
						resourceStatus, err := i.checkResourceStatus(ctx, dataBytes, deployment.Instance.Spec.Scope, statusProbe, component.Name)
						if err != nil {
							sLog.ErrorfCtx(ctx, "Failed to check resource status: +%v", err)
						}
						ret[component.Name] = resourceStatus
					} else {
						ret[component.Name] = model.ComponentResultSpec{
							Status:  v1alpha2.Updated,
							Message: fmt.Sprintf("No error. %s has been updated", component.Name),
						}
					}

				} else {
					err := v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: component doesn't have yaml or resource property", providerName), v1alpha2.YamlResourcePropertyNotFound)
					sLog.ErrorfCtx(ctx, "  P (Kubectl Target):  component doesn't have yaml property or resource property, error: %+v", err)

					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.UpdateFailed,
						Message: err.Error(),
					}
					providerOperationMetrics.ProviderOperationErrors(
						kubectl,
						functionName,
						metrics.ApplyOperation,
						metrics.UpdateOperationType,
						v1alpha2.YamlResourcePropertyNotFound.String(),
					)
					return ret, err
				}
			}

			providerOperationMetrics.ProviderOperationLatency(
				applyComponentTime,
				kubectl,
				functionName,
				metrics.ApplyOperation,
				metrics.UpdateOperationType,
			)
		}
	}

	providerOperationMetrics.ProviderOperationLatency(
		applyTime,
		kubectl,
		functionName,
		metrics.ApplyOperation,
		metrics.UpdateOperationType,
	)

	deleteTime := time.Now().UTC()
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		sLog.InfofCtx(ctx, "  P (Kubectl Target): get deleted components: count - %d", len(components))
		for _, component := range components {
			deleteComponentTime := time.Now().UTC()
			if component.Type == "yaml.k8s" {
				if v, ok := component.Properties["yaml"].(string); ok {
					chanMes, chanErr := readYaml(v)
					stop := false
					for !stop {
						select {
						case dataBytes, ok := <-chanMes:
							if !ok {
								err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to receive from data channel when reading yaml property", providerName), v1alpha2.ReadYamlFailed)
								sLog.ErrorfCtx(ctx, "  P (Kubectl Target):  %+v", err)

								ret[component.Name] = model.ComponentResultSpec{
									Status:  v1alpha2.DeleteFailed,
									Message: err.Error(),
								}
								providerOperationMetrics.ProviderOperationErrors(
									kubectl,
									functionName,
									metrics.ReceiveDataChannelOperation,
									metrics.DeleteOperationType,
									v1alpha2.ReadYamlFailed.String(),
								)
								return ret, err
							}

							err = i.deleteCustomResource(ctx, dataBytes, deployment.Instance.Spec.Scope)
							if err != nil && !kerrors.IsNotFound(err) {
								sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to read object: %+v", err)
								err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to delete object from yaml property", providerName), v1alpha2.DeleteYamlFailed)

								ret[component.Name] = model.ComponentResultSpec{
									Status:  v1alpha2.DeleteFailed,
									Message: err.Error(),
								}
								providerOperationMetrics.ProviderOperationErrors(
									kubectl,
									functionName,
									metrics.ObjectOperation,
									metrics.DeleteOperationType,
									v1alpha2.DeleteYamlFailed.String(),
								)

								return ret, err
							}

							ret[component.Name] = model.ComponentResultSpec{
								Status:  v1alpha2.Deleted,
								Message: "",
							}

						case err, ok := <-chanErr:
							if !ok {
								err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to receive from err channel when reading yaml property", providerName), v1alpha2.ReadYamlFailed)
								sLog.ErrorfCtx(ctx, "  P (Kubectl Target): %+v", err)

								ret[component.Name] = model.ComponentResultSpec{
									Status:  v1alpha2.DeleteFailed,
									Message: err.Error(),
								}
								providerOperationMetrics.ProviderOperationErrors(
									kubectl,
									functionName,
									metrics.ReceiveErrorChannelOperation,
									metrics.DeleteOperationType,
									v1alpha2.ReadYamlFailed.String(),
								)
								return ret, err
							}

							if err == io.EOF {
								stop = true
							} else {
								sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to remove resource: %+v", err)
								err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to delete object from yaml property", providerName), v1alpha2.DeleteYamlFailed)

								ret[component.Name] = model.ComponentResultSpec{
									Status:  v1alpha2.DeleteFailed,
									Message: err.Error(),
								}
								providerOperationMetrics.ProviderOperationErrors(
									kubectl,
									functionName,
									metrics.ResourceOperation,
									metrics.DeleteOperationType,
									v1alpha2.DeleteYamlFailed.String(),
								)
								return ret, err
							}
						}
					}
				} else if component.Properties["resource"] != nil {
					var dataBytes []byte
					dataBytes, err = json.Marshal(component.Properties["resource"])
					if err != nil {
						sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to convert resource data to bytes: %+v", err)
						err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to convert resource data to bytes", providerName), v1alpha2.ReadResourcePropertyFailed)

						ret[component.Name] = model.ComponentResultSpec{
							Status:  v1alpha2.DeleteFailed,
							Message: err.Error(),
						}
						providerOperationMetrics.ProviderOperationErrors(
							kubectl,
							functionName,
							metrics.ConvertResourceDataBytesOperation,
							metrics.DeleteOperationType,
							v1alpha2.ReadResourcePropertyFailed.String(),
						)
						return ret, err
					}

					err = i.deleteCustomResource(ctx, dataBytes, deployment.Instance.Spec.Scope)
					if err != nil && !kerrors.IsNotFound(err) {
						sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to delete custom resource: %+v", err)
						err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to delete custom resource", providerName), v1alpha2.DeleteResourceFailed)

						ret[component.Name] = model.ComponentResultSpec{
							Status:  v1alpha2.DeleteFailed,
							Message: err.Error(),
						}
						providerOperationMetrics.ProviderOperationErrors(
							kubectl,
							functionName,
							metrics.ApplyCustomResource,
							metrics.DeleteOperationType,
							v1alpha2.DeleteResourceFailed.String(),
						)
						return ret, err
					}

					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.Deleted,
						Message: "",
					}

				} else {
					err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: component doesn't have yaml or resource property", providerName), v1alpha2.DeleteFailed)
					sLog.ErrorCtx(ctx, "  P (Kubectl Target): component doesn't have yaml property or resource property")
					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.DeleteFailed,
						Message: err.Error(),
					}
					providerOperationMetrics.ProviderOperationErrors(
						kubectl,
						functionName,
						metrics.ApplyOperation,
						metrics.DeleteOperationType,
						v1alpha2.YamlResourcePropertyNotFound.String(),
					)
					return ret, err
				}
			}

			providerOperationMetrics.ProviderOperationLatency(
				deleteComponentTime,
				kubectl,
				functionName,
				metrics.ApplyOperation,
				metrics.DeleteOperationType,
			)
		}
	}

	providerOperationMetrics.ProviderOperationLatency(
		deleteTime,
		kubectl,
		functionName,
		metrics.ApplyOperation,
		metrics.DeleteOperationType,
	)
	providerOperationMetrics.ProviderOperationLatency(
		applyTime,
		kubectl,
		functionName,
		metrics.ApplyOperation,
		metrics.UpdateOperationType,
	)

	return ret, nil
}

// checkResourceStatus checks the status of the resource
func (k *KubectlTargetProvider) checkResourceStatus(ctx context.Context, dataBytes []byte, namespace string, status *StatusProbe, componentName string) (model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan(
		"Kubectl Target Provider",
		ctx,
		&map[string]string{
			"method": "checkResourceStatus",
		},
	)
	var err error
	result := model.ComponentResultSpec{}
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)

	//add initial wait before checking the resource status
	if status.InitialWait == "" {
		status.InitialWait = initialWait
	}

	waitTime, _ := time.ParseDuration(status.InitialWait)
	time.Sleep(waitTime)
	if status.Timeout == "" {
		status.Timeout = timeout
	}

	timeout, _ := time.ParseDuration(status.Timeout)
	if status.Interval == "" {
		status.Interval = interval
	}

	interval, _ := time.ParseDuration(status.Interval)

	// set default values for succeeded and failed values
	if status.SucceededValues == nil {
		status.SucceededValues = []string{"Succeeded"}
	}

	if status.FailedValues == nil {
		status.FailedValues = []string{"Failed"}
	}

	if namespace == "" {
		namespace = constants.DefaultScope
	}

	resource, err := k.getCustomResource(ctx, dataBytes, namespace)
	if err != nil {
		return result, err
	}

	context, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	compiledPath, err := jsonpath.Compile(status.StatusPath)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to parse the the status path: +%v", err)
	}

	errorPath, err := jsonpath.Compile(status.ErrorMessagePath)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to parse the error message path: +%v", err)
	}

	for {
		select {
		case <-context.Done():
			err := v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to get the status of the resource within the timeout period", providerName), v1alpha2.CheckResourceStatusFailed)
			result = model.ComponentResultSpec{
				Status:  v1alpha2.UpdateFailed,
				Message: err.Error(),
			}
			return result, err
		case <-time.After(interval):
			// checks if the status of the resource is the same as the status in the CRD status probe property
			resourceStatus, err := compiledPath.Lookup(resource.Object)
			sLog.InfofCtx(ctx, "Checking the resource status: %v", resourceStatus)
			if err != nil {
				// if the path is not found then continue to wait for the status
				sLog.ErrorfCtx(ctx, "Warning - waiting for reosurce to be created: +%v", err)
			}

			if resourceStatus != nil {
				// check for succeeded values
				for _, succeededValue := range status.SucceededValues {
					sLog.InfofCtx(ctx, "Checking the resource status for succeededValue: %v", succeededValue)
					if resourceStatus.(string) == succeededValue {
						result = model.ComponentResultSpec{
							Status:  v1alpha2.Updated,
							Message: fmt.Sprintf("No error. %s has been updated", componentName),
						}
						return result, nil
					}
				}
				// check for failed values
				for _, failedValue := range status.FailedValues {
					sLog.InfofCtx(ctx, "Checking the resource status for failedValue: %v", failedValue)
					if resourceStatus.(string) == failedValue {
						// get the error message from the resource
						errorMessage, err := errorPath.Lookup(resource.Object)
						if err != nil {
							errorMessage = "failed to apply custom resource"
						}

						result = model.ComponentResultSpec{
							Status:  v1alpha2.UpdateFailed,
							Message: fmt.Sprintf("%s: %s", providerName, errorMessage.(string)),
						}
						return result, nil
					}
				}
			}
		}
	}
}

// ensureNamespace ensures that the namespace exists
func (k *KubectlTargetProvider) ensureNamespace(ctx context.Context, namespace string) error {
	ctx, span := observability.StartSpan(
		"Kubectl Target Provider",
		ctx,
		&map[string]string{
			"method": "ensureNamespace",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Kubectl Target): ensureNamespace %s", namespace)

	if namespace == "" || namespace == "default" {
		return nil
	}

	_, err = k.Client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if kerrors.IsNotFound(err) {
		observ_utils.EmitUserAuditsLogs(ctx, "  P (Kubectl Target): Start to create namespace - %s", namespace)
		_, err = k.Client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
		if err != nil && !kerrors.IsAlreadyExists(err) {
			sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to create namespace: %+v", err)
			return err
		}

	} else {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to get namespace: %+v")
		return err
	}

	return nil
}

// GetValidationRule returns validation rule for the provider
func (*KubectlTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{},
			OptionalProperties:    []string{"yaml", "resource"},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "yaml", IgnoreCase: false, SkipIfMissing: true},
				{Name: "resource", IgnoreCase: false, SkipIfMissing: true},
			},
		},
	}
}

// ReadYaml reads yaml from url
func readYaml(yaml string) (<-chan []byte, <-chan error) {
	var (
		chanErr   = make(chan error)
		chanBytes = make(chan []byte)
	)
	go func() {
		response, err := http.Get(yaml)
		if err != nil {
			chanErr <- err
			return
		}
		defer response.Body.Close()

		data, err := io.ReadAll(response.Body)
		if err != nil {
			chanErr <- err
			return
		}

		multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
		for {
			buf, err := multidocReader.Read()
			if err != nil {
				chanErr <- err
				return
			}

			chanBytes <- buf
		}

	}()
	return chanBytes, chanErr
}

// BuildDynamicResourceClient builds a new dynamic client
func (i KubectlTargetProvider) buildDynamicResourceClient(data []byte, namespace string) (obj *unstructured.Unstructured, dr dynamic.ResourceInterface, err error) {
	// Decode YAML manifest into unstructured.Unstructured
	obj = &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(data, nil, obj)
	if err != nil {
		return obj, dr, err
	}

	// Find GVR
	mapping, err := i.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return obj, dr, err
	}

	i.DynamicClient, err = dynamic.NewForConfig(i.RESTConfig)
	if err != nil {
		return obj, dr, err
	}

	// Obtain REST interface for the GVR
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		obj.SetNamespace(namespace)
		dr = i.DynamicClient.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		dr = i.DynamicClient.Resource(mapping.Resource)
	}

	return obj, dr, nil
}

// getCustomResource gets a custom resource from a byte array
func (i *KubectlTargetProvider) getCustomResource(ctx context.Context, dataBytes []byte, namespace string) (*unstructured.Unstructured, error) {
	obj, dr, err := i.buildDynamicResourceClient(dataBytes, namespace)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to build a new dynamic client: %+v", err)
		return nil, err
	}

	obj, err = dr.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to read object: %+v", err)
		return nil, err
	}

	return obj, nil
}

// deleteCustomResource deletes a custom resource from a byte array
func (i *KubectlTargetProvider) deleteCustomResource(ctx context.Context, dataBytes []byte, namespace string) error {
	obj, dr, err := i.buildDynamicResourceClient(dataBytes, namespace)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to build a new dynamic client: %+v", err)
		return err
	}

	observ_utils.EmitUserAuditsLogs(ctx, "  P (Kubectl Target): Start to delete object - %s", obj.GetName())
	err = dr.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to delete Yaml: %+v", err)
			return err
		}
	}

	return nil
}

// applyCustomResource applies a custom resource from a byte array
func (i *KubectlTargetProvider) applyCustomResource(ctx context.Context, dataBytes []byte, namespace string, instance model.InstanceState) error {
	sLog.ErrorfCtx(ctx, "  P (Kubectl Target): apply custom resource in the namespace: %s", namespace)
	obj, dr, err := i.buildDynamicResourceClient(dataBytes, namespace)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to build a new dynamic client: %+v", err)
		return err
	}

	// Check if the object exists
	existing, err := dr.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to read object: %+v", err)
			return err
		} else {
			sLog.InfofCtx(ctx, "  P (Kubectl Target): object %s not found: %+v", obj.GetName(), err)
		}

		if err = i.MetaPopulator.PopulateMeta(obj, instance); err != nil {
			sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to populate meta: +%v", err)
			return err
		}

		// Create the object
		observ_utils.EmitUserAuditsLogs(ctx, "  P (Kubectl Target): Start to create object - %s", obj.GetName())
		_, err = dr.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to create Yaml: %+v", err)
			return err
		}
		return nil
	}

	if err = i.MetaPopulator.PopulateMeta(obj, instance); err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to populate meta: +%v", err)
		return err
	}
	// Update the object
	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = dr.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Kubectl Target): failed to apply Yaml: %+v", err)
		return err
	}

	return nil
}

// toStatusProbe converts a component status property to a status probe property
func toStatusProbe(status interface{}) (*StatusProbe, error) {
	statusProbe, ok := status.(map[string]interface{})
	if !ok {
		return nil, errors.New("statusProbe property is not present in the component")
	}
	ret := StatusProbe{}
	data, err := json.Marshal(statusProbe)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
