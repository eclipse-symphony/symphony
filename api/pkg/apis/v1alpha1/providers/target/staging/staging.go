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

package staging

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var sLog = logger.NewLogger("coa.runtime")

var (
	decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
)

type StagingTargetProviderConfig struct {
	Name           string `json:"name"`
	ConfigType     string `json:"configType,omitempty"`
	ConfigData     string `json:"configData,omitempty"`
	Context        string `json:"context,omitempty"`
	InCluster      bool   `json:"inCluster"`
	TargetName     string `json:"targetName"`
	SingleSolution bool   `json:"singleSolution,omitempty"`
}

type StagingTargetProvider struct {
	Config        StagingTargetProviderConfig
	Context       *contexts.ManagerContext
	DynamicClient dynamic.Interface
}

func KubectlTargetProviderConfigFromMap(properties map[string]string) (StagingTargetProviderConfig, error) {
	ret := StagingTargetProviderConfig{}
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
	if v, ok := properties["targetName"]; ok {
		ret.TargetName = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "invalid staging provider config, exptected 'targetName'", v1alpha2.BadConfig)
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
	if v, ok := properties["singleSolution"]; ok {
		val := v
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'singleSolution' setting of kubectl provider", v1alpha2.BadConfig)
			}
			ret.SingleSolution = bVal
		}
	} else {
		ret.SingleSolution = true
	}
	return ret, nil
}

func (i *StagingTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := KubectlTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func (i *StagingTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Staging Target Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("  P (Staging Target): Init()")

	updateConfig, err := toStagingTargetProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Staging Target): expected StagingTargetProviderConfig: %+v", err)
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
					sLog.Errorf("  P (Staging Target): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "bytes":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Staging Target): %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Staging Target): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and bytes", v1alpha2.BadConfig)
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Staging Target): %+v", err)
			return err
		}
	}
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Staging Target): %+v", err)
		return err
	}
	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Staging Target):  %+v", err)
		return err
	}

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func toStagingTargetProviderConfig(config providers.IProviderConfig) (StagingTargetProviderConfig, error) {
	ret := StagingTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

type loalTypeMeta struct {
	Kind       string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
}
type localTarget struct {
	loalTypeMeta `json:",inline"`
	Spec         model.TargetSpec `json:"spec"`
}

func (i *StagingTargetProvider) getTarget(ctx context.Context, scope string) (model.TargetSpec, error) {
	resourceId := schema.GroupVersionResource{
		Group:    model.FabricGroup,
		Version:  "v1",
		Resource: "targets",
	}
	item, err := i.DynamicClient.Resource(resourceId).Namespace(scope).Get(ctx, i.Config.TargetName, metav1.GetOptions{})
	if err != nil {
		return model.TargetSpec{}, err
	}
	target := localTarget{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &target)
	if err != nil {
		return model.TargetSpec{}, err
	}
	return target.Spec, nil
}
func (i *StagingTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Staging Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Infof("  P (Staging Target): getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	scope := deployment.Instance.Scope
	if scope == "" {
		scope = "default"
	}
	target, err := i.getTarget(ctx, scope)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Staging Target): failed to get target: %v", err)
		return nil, err
	}

	observ_utils.CloseSpanWithError(span, nil)
	return target.Components, nil
}
func (i *StagingTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	_, span := observability.StartSpan("Staging Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("  P (Staging Target): applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	err := i.GetValidationRule(ctx).Validate([]model.ComponentSpec{}) //this provider doesn't handle any components	TODO: is this right?
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return nil, err
	}
	if isDryRun {
		observ_utils.CloseSpanWithError(span, nil)
		return nil, nil
	}
	ret := step.PrepareResultMap()

	scope := deployment.Instance.Scope
	if scope == "" {
		scope = "default"
	}

	resourceId := schema.GroupVersionResource{
		Group:    model.FabricGroup,
		Version:  "v1",
		Resource: "targets",
	}
	item, err := i.DynamicClient.Resource(resourceId).Namespace(scope).Get(ctx, i.Config.TargetName, metav1.GetOptions{})
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Staging Target): failed to get unstructed target: %v", err)
		return ret, err
	}
	target := localTarget{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &target)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Staging Target): failed to get target: %v", err)
		return ret, err
	}

	components := step.GetUpdatedComponents()
	if len(components) > 0 {
		//TODO: multiple solutions?
		if target.Spec.Metadata == nil {
			target.Spec.Metadata = make(map[string]string)
		}
		target.Spec.Metadata["__solution"] = deployment.SolutionName

		if i.Config.SingleSolution {
			target.Spec.Components = components
			for i, _ := range target.Spec.Components {
				if !strings.HasPrefix(target.Spec.Components[i].Type, "staged:") {
					target.Spec.Components[i].Type = "staged:" + target.Spec.Components[i].Type
				}
			}
		} else {
			for i, component := range components {
				found := false
				for j, c := range target.Spec.Components {
					if c.Name == component.Name {
						found = true
						target.Spec.Components[j] = components[i]
						if !strings.HasPrefix(target.Spec.Components[j].Type, "staged:") {
							target.Spec.Components[j].Type = "staged:" + target.Spec.Components[j].Type //the stage prefix avoids the component be picked up by the instance role
						}
						break
					}
				}
				if !found {
					target.Spec.Components = append(target.Spec.Components, component)
					target.Spec.Components[len(target.Spec.Components)-1].Type = "staged:" + component.Type
				}
			}
		}
		j, _ := json.Marshal(target)
		err = json.Unmarshal(j, &item.Object)
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Staging Target): failed to update unstructed target: %v", err)
			return ret, err
		}
		_, err = i.DynamicClient.Resource(resourceId).Namespace(scope).Update(ctx, item, metav1.UpdateOptions{})
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Staging Target): failed to update target: %v", err)
			return ret, err
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		//TODO: multiple solutions?
		if target.Spec.Metadata != nil {
			delete(target.Spec.Metadata, "__solution")
		}

		if i.Config.SingleSolution {
			target.Spec.Components = make([]model.ComponentSpec, 0)
		} else {
			for i := len(target.Spec.Components) - 1; i >= 0; i-- {
				for _, c := range components {
					if c.Name == target.Spec.Components[i].Name {
						target.Spec.Components = append(target.Spec.Components[:i], target.Spec.Components[i+1:]...)
						break
					}
				}
			}
		}

		j, _ := json.Marshal(target)
		err = json.Unmarshal(j, &item.Object)
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Staging Target): failed to update unstructed target: %v", err)
			return ret, err
		}
		_, err = i.DynamicClient.Resource(resourceId).Namespace(scope).Update(ctx, item, metav1.UpdateOptions{})
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Staging Target): failed to update target: %v", err)
			return ret, err
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}

func (*StagingTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}
