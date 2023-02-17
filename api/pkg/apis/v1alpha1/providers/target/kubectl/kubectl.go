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
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
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

var sLog = logger.NewLogger("coa.runtime")

var (
	decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
)

type KubectlTargetProviderConfig struct {
	Name       string `json:"name"`
	ConfigType string `json:"configType,omitempty"`
	ConfigData string `json:"configData,omitempty"`
	Context    string `json:"context,omitempty"`
	InCluster  bool   `json:"inCluster"`
}

type KubectlTargetProvider struct {
	Config          KubectlTargetProviderConfig
	Context         *contexts.ManagerContext
	Client          *kubernetes.Clientset
	DynamicClient   dynamic.Interface
	DiscoveryClient *discovery.DiscoveryClient
	Mapper          *restmapper.DeferredDiscoveryRESTMapper
	RESTConfig      *rest.Config
}

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

func (i *KubectlTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := KubectlTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func (i *KubectlTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Kubectl Target Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("~~~ Kubectl Target Provider ~~~ : Init()")

	updateConfig, err := toKubectlTargetProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : expected KubectlTargetProviderConfig: %+v", err)
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
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "bytes":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and bytes", v1alpha2.BadConfig)
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
			return err
		}
	}
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
		return err
	}
	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
		return err
	}
	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
		return err
	}
	i.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(kConfig)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Kubectl Target Provider ~~~ : %+v", err)
		return err
	}
	i.Mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(i.DiscoveryClient))
	i.RESTConfig = kConfig

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func toKubectlTargetProviderConfig(config providers.IProviderConfig) (KubectlTargetProviderConfig, error) {
	ret := KubectlTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *KubectlTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Kubectl Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Infof("~~~ Kubectl Target Provider ~~~ : getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	desired := deployment.GetComponentSlice()

	ret := make([]model.ComponentSpec, 0)
	for _, component := range desired {
		if component.Type == "yaml.k8s" {
			if v, ok := component.Properties["yaml.url"]; ok {
				chanMes, chanErr := readYamlFromUrl(v)
				stop := false
				for !stop {
					select {
					case dataBytes, ok := <-chanMes:
						if !ok {
							err := errors.New("failed to receive from data channel")
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return nil, err
						}
						obj, dr, err := i.buildDynamicResourceClient(dataBytes)
						if err != nil {
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to build a new dynamic client: +%v", err)
							return nil, err
						}

						_, err = dr.Get(ctx, obj.GetName(), metav1.GetOptions{})
						if err != nil {
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to read object: +%v", err)
							return nil, err
						}
						ret = append(ret, component)
						stop = true //we do early stop as soon as we found the first resource. we may wnat to support different strategy in the future
					case err, ok := <-chanErr:
						if !ok {
							err = errors.New("failed to receive from error channel")
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return nil, err
						}
						if err == io.EOF {
							stop = true
						} else {
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply Yaml: +%v", err)
							return nil, err
						}
					}
				}
			} else {
				err := errors.New("component doesn't have yaml.url property")
				observ_utils.CloseSpanWithError(span, err)
				sLog.Error("~~~ Kubectl Target Provider ~~~ : component doesn't have yaml.url property")
				return nil, err
			}
		}
	}

	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}
func (i *KubectlTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	return !model.SlicesCover(desired, current)
}
func (i *KubectlTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	return model.SlicesAny(desired, current)
}

func (i *KubectlTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan("Kubectl Target Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	sLog.Infof("~~~ Kubectl Target Provider ~~~ : deleting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := deployment.GetComponentSlice()

	for _, component := range components {
		if component.Type == "yaml.k8s" {
			if v, ok := component.Properties["yaml.url"]; ok {
				chanMes, chanErr := readYamlFromUrl(v)
				stop := false
				for !stop {
					select {
					case dataBytes, ok := <-chanMes:
						if !ok {
							err := errors.New("failed to receive from data channel")
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return err
						}
						obj, dr, err := i.buildDynamicResourceClient(dataBytes)
						if err != nil {
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to build a new dynamic client: +%v", err)
							return err
						}
						err = dr.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
						if err != nil {
							if !kerrors.IsNotFound(err) {
								observ_utils.CloseSpanWithError(span, err)
								sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to delete Yaml: +%v", err)
								return err
							}
						}
					case err, ok := <-chanErr:
						if !ok {
							err = errors.New("failed to receive from error channel")
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return err
						}
						if err == io.EOF {
							stop = true
						} else {
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply Yaml: +%v", err)
							return err
						}
					}
				}
			} else {
				err := errors.New("component doesn't have yaml.url property")
				observ_utils.CloseSpanWithError(span, err)
				sLog.Error("~~~ Kubectl Target Provider ~~~ : component doesn't have yaml.url property")
				return err
			}
		}
	}

	//TODO: Should we remove empty namespaces?
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func (i *KubectlTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec) error {
	_, span := observability.StartSpan("Kubectl Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("~~~ Kubectl Target Provider ~~~ : applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := deployment.GetComponentSlice()

	for _, component := range components {
		if component.Type == "yaml.k8s" {
			if v, ok := component.Properties["yaml.url"]; ok {
				chanMes, chanErr := readYamlFromUrl(v)
				stop := false
				for !stop {
					select {
					case dataBytes, ok := <-chanMes:
						if !ok {
							err := errors.New("failed to receive from data channel")
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return err
						}
						obj, dr, err := i.buildDynamicResourceClient(dataBytes)
						if err != nil {
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to build a new dynamic client: +%v", err)
							return err
						}
						_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, dataBytes, metav1.PatchOptions{
							FieldManager: "application/apply-patch",
						})
						if err != nil {
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply Yaml: +%v", err)
							return err
						}
					case err, ok := <-chanErr:
						if !ok {
							err = errors.New("failed to receive from error channel")
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : +%v", err)
							return err
						}
						if err == io.EOF {
							stop = true
						} else {
							observ_utils.CloseSpanWithError(span, err)
							sLog.Error("~~~ Kubectl Target Provider ~~~ : failed to apply Yaml: +%v", err)
							return err
						}
					}
				}
			} else {
				err := errors.New("component doesn't have yaml.url property")
				observ_utils.CloseSpanWithError(span, err)
				sLog.Error("~~~ Kubectl Target Provider ~~~ : component doesn't have yaml.url property")
				return err
			}
		}
	}

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func readYamlFromUrl(url string) (<-chan []byte, <-chan error) {
	var (
		chanErr   = make(chan error)
		chanBytes = make(chan []byte)
	)
	go func() {
		response, err := http.Get(url)
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

func (i *KubectlTargetProvider) buildDynamicResourceClient(data []byte) (obj *unstructured.Unstructured, dr dynamic.ResourceInterface, err error) {
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
		dr = i.DynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = i.DynamicClient.Resource(mapping.Resource)
	}
	return obj, dr, nil
}
