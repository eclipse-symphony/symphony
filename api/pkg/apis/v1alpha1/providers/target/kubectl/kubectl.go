/*
   MIT License

   Copyright (c) Microsoft Corporation.

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   SOFTWARE

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

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
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
		Client          *kubernetes.Clientset
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

// Init initializes the kubectl target provider
func (i *KubectlTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan(
		"Kubectl Target Provider",
		context.Background(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Info("~~~ Kubectl Target Provider ~~~ : Init()")

	updateConfig, err := toKubectlTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : expected KubectlTargetProviderConfig: %+v", err)
		return err
	}

	i.Config = updateConfig
	var kConfig *rest.Config
	if i.Config.InCluster {
		kConfig, err = rest.InClusterConfig()
	} else {
		switch i.Config.ConfigType {
		case "path", "inline":
			if i.Config.ConfigData == "" {
				if home := homedir.HomeDir(); home != "" {
					i.Config.ConfigData = filepath.Join(home, ".kube", "config")
				} else {
					err = v1alpha2.NewCOAError(nil, "can't locate home direction to read default kubernetes config file, to run in cluster, set inCluster config setting to true", v1alpha2.BadConfig)
					sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
					return err
				}
			}

			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)

		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and inline", v1alpha2.BadConfig)
			sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
			return err
		}
	}
	if err != nil {
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
		return err
	}

	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
		return err
	}

	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
		return err
	}

	i.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(kConfig)
	if err != nil {
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
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
func (i *KubectlTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"Kubectl Target Provider",
		ctx, &map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Infof("~~~ Kubectl Target Provider ~~~ : getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := deployment.GetComponentSlice()
	ret := make([]model.ComponentSpec, 0)
	for _, component := range components {
		if v, ok := component.Properties["yaml"].(string); ok {
			chanMes, chanErr := readYaml(v)
			stop := false
			for !stop {
				select {
				case dataBytes, ok := <-chanMes:
					if !ok {
						err = errors.New("failed to receive from data channel")
						sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
						return nil, err
					}

					_, err = i.getCustomResource(ctx, dataBytes, deployment.Instance.Scope)
					if err != nil {
						if kerrors.IsNotFound(err) {
							sLog.Infof("~~~ Kubectl Target Provider ~~~ : resource not found: %s", err)
							continue
						}
						sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to read object: +%v", err)
						return nil, err
					}

					ret = append(ret, component)
					stop = true //we do early stop as soon as we found the first resource. we may want to support different strategy in the future

				case err, ok := <-chanErr:
					if !ok {
						err = errors.New("failed to receive from error channel")
						sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
						return nil, err
					}

					if err == io.EOF {
						stop = true
					} else {
						sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply Yaml: +%v", err)
						return nil, err
					}
				}
			}
		} else if component.Properties["resource"] != nil {
			var dataBytes []byte
			dataBytes, err = json.Marshal(component.Properties["resource"])
			if err != nil {
				sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to get deployment bytes from component: +%v", err)
				return nil, err
			}

			_, err = i.getCustomResource(ctx, dataBytes, deployment.Instance.Scope)
			if err != nil {
				if kerrors.IsNotFound(err) {
					sLog.Infof("~~~ Kubectl Target Provider ~~~ : resource not found: %s", err)
					continue
				}
				sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to read object: +%v", err)
				return nil, err
			}

			ret = append(ret, component)

		} else {
			err = errors.New("component doesn't have yaml or resource property")
			sLog.Error("~~~ Kubectl Target Provider ~~~ : component doesn't have yaml or resource property")
			return nil, err
		}
	}

	return ret, nil
}

// NeedsUpdate checks if the current artifacts need to be updated
func (i *KubectlTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	// With symphony's current implementation, we don't have a reliable way of checking if the current artifacts need to be updated
	// so we always return true and delegate the responsibility of updating (if needed) to the kubernetes api server
	return true
}

// NeedsRemove checks if the current artifacts need to be removed
func (i *KubectlTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	// With symphony's current implementation, we don't have a reliable way of checking if the current artifacts need to be removed
	// so we always return false and delegate the responsibility of removing (if needed) to the kubernetes api server

	return true
}

// Remove removes the current artifacts
func (i *KubectlTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan(
		"Kubectl Target Provider",
		ctx,
		&map[string]string{
			"method": "Remove",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Infof("~~~ Kubectl Target Provider ~~~ : deleting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := deployment.GetComponentSlice()
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
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return err
						}

						err = i.deleteCustomResource(ctx, dataBytes, deployment.Instance.Scope)
						if err != nil {
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to read object: +%v", err)
							return err
						}

					case err, ok := <-chanErr:
						if !ok {
							err = errors.New("failed to receive from error channel")
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return err
						}

						if err == io.EOF {
							stop = true
						} else {
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to remove resource: +%v", err)
							return err
						}
					}
				}
			} else if component.Properties["resource"] != nil {
				var dataBytes []byte
				dataBytes, err = json.Marshal(component.Properties["resource"])
				if err != nil {
					sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to convert resource data to bytes: +%v", err)
					return err
				}

				err = i.deleteCustomResource(ctx, dataBytes, deployment.Instance.Scope)
				if err != nil {
					sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to delete custom resource: +%v", err)
					return err
				}

			} else {
				err = errors.New("component doesn't have yaml property or resource property")
				sLog.Error("~~~ Kubectl Target Provider ~~~ : component doesn't have yaml property or resource property")
				return err
			}
		}
	}

	//TODO: Should we remove empty namespaces?
	return nil
}

// Apply applies the deployment artifacts
func (i *KubectlTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, isDryRun bool) error {
	_, span := observability.StartSpan(
		"Kubectl Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Infof("~~~ Kubectl Target Provider ~~~ : applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := deployment.GetComponentSlice()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		return err
	}

	if isDryRun {
		return nil
	}

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
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return err
						}

						err = i.applyCustomResource(ctx, dataBytes, deployment.Instance.Scope)
						if err != nil {
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply Yaml: +%v", err)
							return err
						}

					case err, ok := <-chanErr:
						if !ok {
							err = errors.New("failed to receive from error channel")
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return err
						}

						if err == io.EOF {
							stop = true
						} else {
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply Yaml: +%v", err)
							return err
						}
					}
				}
			} else if component.Properties["resource"] != nil {
				var dataBytes []byte
				dataBytes, err = json.Marshal(component.Properties["resource"])
				if err != nil {
					sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to convert resource data to bytes: +%v", err)
					return err
				}

				err = i.applyCustomResource(ctx, dataBytes, deployment.Instance.Scope)
				if err != nil {
					sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply custom resource: +%v", err)
					return err
				}

			} else {
				err := errors.New("component doesn't have yaml property or resource property")
				sLog.Error("~~~ Kubectl Target Provider ~~~ : component doesn't have yaml property or resource property")
				return err
			}
		}
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
		// TODO: Do we really want to accept the namespace from the component? typically this component is part
		// of a deployment and the namespace should be set by the deployment
		namespace := obj.GetNamespace()
		if namespace == "" {
			namespace = "default"
		}
		if namespace != scope {
			return obj, dr, fmt.Errorf("namespace %s does not match scope %s", namespace, scope)
		}
		dr = i.DynamicClient.Resource(mapping.Resource).Namespace(namespace)
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
		sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to build a new dynamic client: +%v", err)
		return nil, err
	}

	obj, err = dr.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to read object: +%v", err)
		return nil, err
	}

	return obj, nil
}

// deleteCustomResource deletes a custom resource from a byte array
func (i *KubectlTargetProvider) deleteCustomResource(ctx context.Context, dataBytes []byte, scope string) error {
	obj, dr, err := i.buildDynamicResourceClient(dataBytes, scope)
	if err != nil {
		sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to build a new dynamic client: +%v", err)
		return err
	}

	err = dr.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to delete Yaml: +%v", err)
			return err
		}
	}

	return nil
}

// applyCustomResource applies a custom resource from a byte array
func (i *KubectlTargetProvider) applyCustomResource(ctx context.Context, dataBytes []byte, scope string) error {
	obj, dr, err := i.buildDynamicResourceClient(dataBytes, scope)
	if err != nil {
		sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to build a new dynamic client: +%v", err)
		return err
	}
	// Check if the object exists
	_, err = dr.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to read object: +%v", err)
			return err
		}
		// Create the object
		_, err = dr.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to create Yaml: +%v", err)
			return err
		}
		return nil
	}

	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, dataBytes, metav1.PatchOptions{
		FieldManager: "application/apply-patch",
	})
	if err != nil {
		sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply Yaml: +%v", err)
		return err
	}

	return nil
}
