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

package k8s

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/coa/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var sLog = logger.NewLogger("coa.runtime")

type K8sStateProviderConfig struct {
	Name       string `json:"name"`
	ConfigType string `json:"configType,omitempty"`
	ConfigData string `json:"configData,omitempty"`
	Context    string `json:"context,omitempty"`
	InCluster  bool   `json:"inCluster"`
}

type K8sStateProvider struct {
	Config        K8sStateProviderConfig
	Context       *contexts.ManagerContext
	DynamicClient dynamic.Interface
}

func K8sStateProviderConfigFromMap(properties map[string]string) (K8sStateProviderConfig, error) {
	ret := K8sStateProviderConfig{}
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
	if ret.ConfigType == "" {
		ret.ConfigType = "path"
	}
	if v, ok := properties["inCluster"]; ok {
		val := v
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'inCluster' setting of K8s state provider", v1alpha2.BadConfig)
			}
			ret.InCluster = bVal
		}
	}
	return ret, nil
}

func (i *K8sStateProvider) InitWithMap(properties map[string]string) error {
	config, err := K8sStateProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (i *K8sStateProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("K8s State Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Debug("  P (K8s State): initialize")

	updateConfig, err := toK8sStateProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (K8s State): expected KubectlTargetProviderConfig: %+v", err)
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
					sLog.Errorf("  P (K8s State): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "bytes":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (K8s State): %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (K8s State): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and bytes", v1alpha2.BadConfig)
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (K8s State): %+v", err)
			return err
		}
	}
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (K8s State): %+v", err)
		return err
	}
	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (K8s State): %+v", err)
		return err
	}

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func toK8sStateProviderConfig(config providers.IProviderConfig) (K8sStateProviderConfig, error) {
	ret := K8sStateProviderConfig{}
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

func (s *K8sStateProvider) SetContext(ctx *contexts.ManagerContext) error {
	s.Context = ctx
	return nil
}

func (s *K8sStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	ctx, span := observability.StartSpan("K8s State Provider", context.Background(), &map[string]string{
		"method": "Upsert",
	})
	sLog.Info("  P (K8s State): upsert state")

	scope := model.ReadProperty(entry.Metadata, "scope", nil)
	group := model.ReadProperty(entry.Metadata, "group", nil)
	version := model.ReadProperty(entry.Metadata, "version", nil)
	resource := model.ReadProperty(entry.Metadata, "resource", nil)

	if scope == "" {
		scope = "default"
	}

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	j, _ := json.Marshal(entry.Value.Body)

	item, err := s.DynamicClient.Resource(resourceId).Namespace(scope).Get(ctx, entry.Value.ID, metav1.GetOptions{})
	if err != nil {
		// TODO: check if not-found error
		template := model.ReadProperty(entry.Metadata, "template", &model.ValueInjections{
			TargetId:     entry.Value.ID,
			SolutionId:   entry.Value.ID, //TODO: This is not very nice. Maybe change ValueInjection to include a generic ID?
			InstanceId:   entry.Value.ID,
			ActivationId: entry.Value.ID,
			CampaignId:   entry.Value.ID,
			CatalogId:    entry.Value.ID,
			DeviceId:     entry.Value.ID,
			ModelId:      entry.Value.ID,
			SkillId:      entry.Value.ID,
		})
		var unc *unstructured.Unstructured
		err = json.Unmarshal([]byte(template), &unc)
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (K8s State): failed to deserialize template: %v", err)
			return "", err
		}
		var dict map[string]interface{}
		err = json.Unmarshal(j, &dict)
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (K8s State): failed to get object: %v", err)
			return "", err
		}
		unc.Object["spec"] = dict["spec"]
		_, err = s.DynamicClient.Resource(resourceId).Namespace(scope).Create(ctx, unc, metav1.CreateOptions{})
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (K8s State): failed to create object: %v", err)
			return "", err
		}
		//Note: state is ignored for new object
	} else {
		j, _ := json.Marshal(entry.Value.Body)
		var dict map[string]interface{}
		err = json.Unmarshal(j, &dict)
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (K8s State): failed to unmarshal object: %v", err)
			return "", err
		}
		if v, ok := dict["spec"]; ok {
			item.Object["spec"] = v

			_, err = s.DynamicClient.Resource(resourceId).Namespace(scope).Update(ctx, item, metav1.UpdateOptions{})
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (K8s State): failed to update object: %v", err)
				return "", err
			}
		}
		if v, ok := dict["status"]; ok {
			status := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": group + "/" + version,
					"kind":       "Status",
					"metadata": map[string]interface{}{
						"name": entry.Value.ID,
					},
					"status": v.(map[string]interface{}),
				},
			}
			status.SetResourceVersion(item.GetResourceVersion())
			_, err = s.DynamicClient.Resource(resourceId).Namespace(scope).UpdateStatus(context.Background(), status, v1.UpdateOptions{})
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (K8s State): failed to update object status: %v", err)
				return "", err
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return entry.Value.ID, nil
}

func (s *K8sStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	var entities []states.StateEntry

	ctx, span := observability.StartSpan("K8s State Provider", context.Background(), &map[string]string{
		"method": "List",
	})
	sLog.Info("  P (K8s State): list state")

	scope := model.ReadProperty(request.Metadata, "scope", nil)
	group := model.ReadProperty(request.Metadata, "group", nil)
	version := model.ReadProperty(request.Metadata, "version", nil)
	resource := model.ReadProperty(request.Metadata, "resource", nil)

	if scope == "" {
		scope = "default"
	}

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	items, err := s.DynamicClient.Resource(resourceId).Namespace(scope).List(ctx, metav1.ListOptions{})
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (K8s State): failed to list objects: %v", err)
		return nil, "", err
	}
	for _, v := range items.Items {
		generation := v.GetGeneration()
		entry := states.StateEntry{
			ETag: strconv.FormatInt(generation, 10),
			ID:   v.GetName(),
			Body: map[string]interface{}{
				"spec":   v.Object["spec"],
				"status": v.Object["status"],
			},
		}
		entities = append(entities, entry)
	}
	observ_utils.CloseSpanWithError(span, nil)
	return entities, "", nil
}

func (s *K8sStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	ctx, span := observability.StartSpan("K8s State Provider", context.Background(), &map[string]string{
		"method": "Delete",
	})
	sLog.Info("  P (K8s State): delete state")

	scope := model.ReadProperty(request.Metadata, "scope", nil)
	group := model.ReadProperty(request.Metadata, "group", nil)
	version := model.ReadProperty(request.Metadata, "version", nil)
	resource := model.ReadProperty(request.Metadata, "resource", nil)

	if scope == "" {
		scope = "default"
	}

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	err := s.DynamicClient.Resource(resourceId).Namespace(scope).Delete(ctx, request.ID, metav1.DeleteOptions{})
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (K8s State): failed to delete objects: %v", err)
		return err
	}
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func (s *K8sStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	ctx, span := observability.StartSpan("K8s State Provider", context.Background(), &map[string]string{
		"method": "Get",
	})
	sLog.Info("  P (K8s State): get state")

	scope := model.ReadProperty(request.Metadata, "scope", nil)
	group := model.ReadProperty(request.Metadata, "group", nil)
	version := model.ReadProperty(request.Metadata, "version", nil)
	resource := model.ReadProperty(request.Metadata, "resource", nil)

	if scope == "" {
		scope = "default"
	}

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	item, err := s.DynamicClient.Resource(resourceId).Namespace(scope).Get(ctx, request.ID, metav1.GetOptions{})
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (K8s State): failed to get objects: %v", err)
		return states.StateEntry{}, err
	}
	generation := item.GetGeneration()
	ret := states.StateEntry{
		ID:   request.ID,
		ETag: strconv.FormatInt(generation, 10),
		Body: map[string]interface{}{
			"spec":   item.Object["spec"],
			"status": item.Object["status"],
		},
	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}
