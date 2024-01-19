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
	"io"
	"net/http"
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
	decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	sLog            = logger.NewLogger("coa.runtime")
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
		return err
	}

	return i.Init(config)
}

func (s *KubectlTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

// Init initializes the kubectl target provider
func (i *KubectlTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan(
		"Kubectl Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	sLog.Info("  P (Kubectl Target): Init()")

	updateConfig, err := toKubectlTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Kubectl Target): expected KubectlTargetProviderConfig - %+v")
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
					sLog.Errorf("  P (Kubectl Target): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "inline":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.Errorf("  P (Kubectl Target): failed to get RESTconfg: %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.Errorf("  P (Kubectl Target): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and inline", v1alpha2.BadConfig)
			sLog.Errorf("  P (Kubectl Target): %+v", err)
			return err
		}
	}
	if err != nil {
		sLog.Errorf("  P (Kubectl Target): failed to get the cluster config: %+v", err)
		return err
	}

	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (Kubectl Target): failed to create a new clientset: %+v", err)
		return err
	}

	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (Kubectl Target): failed to create a dynamic client: %+v", err)
		return err
	}

	i.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (Kubectl Target): failed to create a discovery client: %+v", err)
		return err
	}

	i.Mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(i.DiscoveryClient))
	i.RESTConfig = kConfig
	return nil
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
	sLog.Infof("  P (Kubectl Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		if v, ok := component.Component.Properties["yaml"].(string); ok {
			chanMes, chanErr := readYaml(v)
			stop := false
			for !stop {
				select {
				case dataBytes, ok := <-chanMes:
					if !ok {
						err = errors.New("failed to receive from data channel")
						sLog.Error("  P (Kubectl Target): +%v, traceId: %s", err, span.SpanContext().TraceID().String())
						return nil, err
					}

					_, err = i.getCustomResource(ctx, dataBytes, deployment.Instance.Scope)
					if err != nil {
						if kerrors.IsNotFound(err) {
							sLog.Infof("  P (Kubectl Target): resource not found: %s, traceId: %s", err, span.SpanContext().TraceID().String())
							continue
						}
						sLog.Error("  P (Kubectl Target): failed to read object: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
						return nil, err
					}

					ret = append(ret, component.Component)
					stop = true //we do early stop as soon as we found the first resource. we may want to support different strategy in the future

				case err, ok := <-chanErr:
					if !ok {
						err = errors.New("failed to receive from error channel")
						sLog.Error("  P (Kubectl Target): +%v, traceId: %s", err, span.SpanContext().TraceID().String())
						return nil, err
					}

					if err == io.EOF {
						stop = true
					} else {
						sLog.Error("  P (Kubectl Target): failed to apply Yaml: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
						return nil, err
					}
				}
			}
		} else if component.Component.Properties["resource"] != nil {
			var dataBytes []byte
			dataBytes, err = json.Marshal(component.Component.Properties["resource"])
			if err != nil {
				sLog.Errorf("  P (Kubectl Target): failed to get deployment bytes from component: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
				return nil, err
			}

			_, err = i.getCustomResource(ctx, dataBytes, deployment.Instance.Scope)
			if err != nil {
				if kerrors.IsNotFound(err) {
					sLog.Infof("  P (Kubectl Target): resource not found: %v, traceId: %s", err, span.SpanContext().TraceID().String())
					continue
				}
				sLog.Errorf("  P (Kubectl Target): failed to read object: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
				return nil, err
			}

			ret = append(ret, component.Component)

		} else {
			err = errors.New("component doesn't have yaml or resource property")
			sLog.Errorf("  P (Kubectl Target): component doesn't have yaml or resource property, traceId: %s", span.SpanContext().TraceID().String())
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
	sLog.Infof("  P (Kubectl Target):  applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.Errorf(" P (Kubectl Target): failed to validate components, error: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	if isDryRun {
		return nil, nil
	}

	ret := step.PrepareResultMap()
	components = step.GetUpdatedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Type == "yaml.k8s" {
				if v, ok := component.Properties["yaml"].(string); ok {
					chanMes, chanErr := readYaml(v)
					stop := false
					for !stop {
						select {
						case dataBytes, ok := <-chanMes:
							if !ok {
								err = errors.New("failed to receive from data channel")
								sLog.Error("  P (Kubectl Target):  +%v, traceId: %s", err, span.SpanContext().TraceID().String())
								return ret, err
							}

							i.ensureNamespace(ctx, deployment.Instance.Scope)
							err = i.applyCustomResource(ctx, dataBytes, deployment.Instance.Scope)
							if err != nil {
								sLog.Error("  P (Kubectl Target):  failed to apply Yaml: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
								return ret, err
							}

						case err, ok := <-chanErr:
							if !ok {
								err = errors.New("failed to receive from error channel")
								sLog.Error("  P (Kubectl Target):  +%v, traceId: %s, traceId: %s", err, span.SpanContext().TraceID().String())
								return ret, err
							}

							if err == io.EOF {
								stop = true
							} else {
								sLog.Error("  P (Kubectl Target):  failed to apply Yaml: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
								return ret, err
							}
						}
					}
				} else if component.Properties["resource"] != nil {
					var dataBytes []byte
					dataBytes, err = json.Marshal(component.Properties["resource"])
					if err != nil {
						sLog.Error("  P (Kubectl Target): failed to convert resource data to bytes: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
						return ret, err
					}

					i.ensureNamespace(ctx, deployment.Instance.Scope)
					err = i.applyCustomResource(ctx, dataBytes, deployment.Instance.Scope)
					if err != nil {
						sLog.Error("  P (Kubectl Target):  failed to apply custom resource: +%v, traceId: %s", err, err, span.SpanContext().TraceID().String())
						return ret, err
					}

				} else {
					err = errors.New("component doesn't have yaml property or resource property")
					sLog.Errorf("  P (Kubectl Target):  component doesn't have yaml property or resource property, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
			}
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Type == "yaml.k8s" {
				if v, ok := component.Properties["yaml"].(string); ok {
					chanMes, chanErr := readYaml(v)
					stop := false
					for !stop {
						select {
						case dataBytes, ok := <-chanMes:
							if !ok {
								err = errors.New("failed to receive from data channel")
								sLog.Errorf("  P (Kubectl Target):  +%v, traceId: %s", err, span.SpanContext().TraceID().String())
								return ret, err
							}

							err = i.deleteCustomResource(ctx, dataBytes, deployment.Instance.Scope)
							if err != nil {
								sLog.Errorf("  P (Kubectl Target): failed to read object: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
								return ret, err
							}

						case err, ok := <-chanErr:
							if !ok {
								err = errors.New("failed to receive from error channel")
								sLog.Errorf("  P (Kubectl Target): +%v, traceId: %s", err, span.SpanContext().TraceID().String())
								return ret, err
							}

							if err == io.EOF {
								stop = true
							} else {
								sLog.Errorf("  P (Kubectl Target): failed to remove resource: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
								return ret, err
							}
						}
					}
				} else if component.Properties["resource"] != nil {
					var dataBytes []byte
					dataBytes, err = json.Marshal(component.Properties["resource"])
					if err != nil {
						sLog.Errorf("  P (Kubectl Target): failed to convert resource data to bytes: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
						return ret, err
					}

					err = i.deleteCustomResource(ctx, dataBytes, deployment.Instance.Scope)
					if err != nil {
						sLog.Errorf("  P (Kubectl Target): failed to delete custom resource: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
						return ret, err
					}

				} else {
					err = errors.New("component doesn't have yaml property or resource property")
					sLog.Errorf("  P (Kubectl Target): component doesn't have yaml property or resource property, traceId: %s", span.SpanContext().TraceID().String())
					return ret, err
				}
			}
		}
	}
	return ret, nil
}

// ensureNamespace ensures that the namespace exists
func (k *KubectlTargetProvider) ensureNamespace(ctx context.Context, namespace string) error {
	_, span := observability.StartSpan(
		"Kubectl Target Provider",
		ctx,
		&map[string]string{
			"method": "ensureNamespace",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Kubectl Target): ensureNamespace %s, traceId: %s", namespace, span.SpanContext().TraceID().String())

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
			sLog.Errorf("  P (Kubectl Target): failed to create namespace: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
			return err
		}

	} else {
		sLog.Errorf("  P (Kubectl Target): failed to get namespace: +%v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}

	return nil
}

// GetValidationRule returns validation rule for the provider
func (*KubectlTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{},
		OptionalProperties:    []string{"yaml", "resource"},
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
func (i KubectlTargetProvider) buildDynamicResourceClient(data []byte, scope string) (obj *unstructured.Unstructured, dr dynamic.ResourceInterface, err error) {
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
		obj.SetNamespace(scope)
		dr = i.DynamicClient.Resource(mapping.Resource).Namespace(scope)
	} else {
		// for cluster-wide resources
		dr = i.DynamicClient.Resource(mapping.Resource)
	}

	return obj, dr, nil
}

// getCustomResource gets a custom resource from a byte array
func (i *KubectlTargetProvider) getCustomResource(ctx context.Context, dataBytes []byte, scope string) (*unstructured.Unstructured, error) {
	obj, dr, err := i.buildDynamicResourceClient(dataBytes, scope)
	if err != nil {
		sLog.Error("  P (Kubectl Target): failed to build a new dynamic client: +%v", err)
		return nil, err
	}

	obj, err = dr.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		sLog.Error("  P (Kubectl Target): failed to read object: +%v", err)
		return nil, err
	}

	return obj, nil
}

// deleteCustomResource deletes a custom resource from a byte array
func (i *KubectlTargetProvider) deleteCustomResource(ctx context.Context, dataBytes []byte, scope string) error {
	obj, dr, err := i.buildDynamicResourceClient(dataBytes, scope)
	if err != nil {
		sLog.Error("  P (Kubectl Target): failed to build a new dynamic client: +%v", err)
		return err
	}

	err = dr.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.Error("  P (Kubectl Target): failed to delete Yaml: +%v", err)
			return err
		}
	}

	return nil
}

// applyCustomResource applies a custom resource from a byte array
func (i *KubectlTargetProvider) applyCustomResource(ctx context.Context, dataBytes []byte, scope string) error {
	obj, dr, err := i.buildDynamicResourceClient(dataBytes, scope)
	if err != nil {
		sLog.Error("  P (Kubectl Target): failed to build a new dynamic client: +%v", err)
		return err
	}

	// Check if the object exists
	existing, err := dr.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.Error("  P (Kubectl Target): failed to read object: +%v", err)
			return err
		}
		// Create the object
		_, err = dr.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			sLog.Error("  P (Kubectl Target): failed to create Yaml: +%v", err)
			return err
		}
		return nil
	}

	// Update the object
	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = dr.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		sLog.Error("  P (Kubectl Target): failed to apply Yaml: +%v", err)
		return err
	}

	return nil
}
