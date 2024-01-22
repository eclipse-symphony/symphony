/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
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

func (s *K8sStateProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *K8sStateProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("K8s State Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Debug("  P (K8s State): initialize")

	updateConfig, err := toK8sStateProviderConfig(config)
	if err != nil {
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
					sLog.Errorf("  P (K8s State): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "bytes":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.Errorf("  P (K8s State): %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.Errorf("  P (K8s State): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and bytes", v1alpha2.BadConfig)
			sLog.Errorf("  P (K8s State): %+v", err)
			return err
		}
	}
	if err != nil {
		sLog.Errorf("  P (K8s State): %+v", err)
		return err
	}
	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.Errorf("  P (K8s State): %+v", err)
		return err
	}

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

func (s *K8sStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	ctx, span := observability.StartSpan("K8s State Provider", ctx, &map[string]string{
		"method": "Upsert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

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

	if entry.Value.ID == "" {
		err := v1alpha2.NewCOAError(nil, "found invalid request ID", v1alpha2.InternalError)
		return "", err
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
			SiteId:       entry.Value.ID,
		})
		var unc *unstructured.Unstructured
		err = json.Unmarshal([]byte(template), &unc)
		if err != nil {
			sLog.Errorf("  P (K8s State): failed to deserialize template: %v", err)
			return "", err
		}
		var dict map[string]interface{}
		err = json.Unmarshal(j, &dict)
		if err != nil {
			sLog.Errorf("  P (K8s State): failed to get object: %v", err)
			return "", err
		}
		unc.Object["spec"] = dict["spec"]
		_, err = s.DynamicClient.Resource(resourceId).Namespace(scope).Create(ctx, unc, metav1.CreateOptions{})
		if err != nil {
			sLog.Errorf("  P (K8s State): failed to create object: %v", err)
			return "", err
		}
		//Note: state is ignored for new object
	} else {
		j, _ := json.Marshal(entry.Value.Body)
		var dict map[string]interface{}
		err = json.Unmarshal(j, &dict)
		if err != nil {
			sLog.Errorf("  P (K8s State): failed to unmarshal object: %v", err)
			return "", err
		}
		if v, ok := dict["spec"]; ok {
			item.Object["spec"] = v

			_, err = s.DynamicClient.Resource(resourceId).Namespace(scope).Update(ctx, item, metav1.UpdateOptions{})
			if err != nil {
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
			_, err = s.DynamicClient.Resource(resourceId).Namespace(scope).UpdateStatus(ctx, status, v1.UpdateOptions{})
			if err != nil {
				sLog.Errorf("  P (K8s State): failed to update object status: %v", err)
				return "", err
			}
		}
	}
	return entry.Value.ID, nil
}

func (s *K8sStateProvider) ListAllNamespaces(ctx context.Context, version string) ([]string, error) {
	namespaceResource := schema.GroupVersionResource{Group: "", Version: version, Resource: "namespaces"}
	namespaces, err := s.DynamicClient.Resource(namespaceResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	ret := []string{}
	for _, namespace := range namespaces.Items {
		ret = append(ret, namespace.GetName())
	}
	return ret, err
}

func (s *K8sStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	var entities []states.StateEntry

	ctx, span := observability.StartSpan("K8s State Provider", ctx, &map[string]string{
		"method": "List",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (K8s State): list state")

	scope := model.ReadProperty(request.Metadata, "scope", nil)
	group := model.ReadProperty(request.Metadata, "group", nil)
	version := model.ReadProperty(request.Metadata, "version", nil)
	resource := model.ReadProperty(request.Metadata, "resource", nil)

	var namespaces []string
	if scope == "" {
		ret, err := s.ListAllNamespaces(ctx, version)
		if err != nil {
			sLog.Errorf("  P (K8s State): failed to list namespaces: %v", err)
			return nil, "", err
		}
		namespaces = ret
	} else {
		namespaces = []string{scope}
	}
	for _, namespace := range namespaces {
		resourceId := schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resource,
		}
		items, err := s.DynamicClient.Resource(resourceId).Namespace(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			sLog.Errorf("  P (K8s State): failed to list objects in namespace %s: %v ", namespace, err)
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
					"scope":  namespace,
				},
			}
			entities = append(entities, entry)
		}
	}
	return entities, "", nil
}

func (s *K8sStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	ctx, span := observability.StartSpan("K8s State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (K8s State): delete state")

	scope := model.ReadProperty(request.Metadata, "scope", nil)
	group := model.ReadProperty(request.Metadata, "group", nil)
	version := model.ReadProperty(request.Metadata, "version", nil)
	resource := model.ReadProperty(request.Metadata, "resource", nil)

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	if scope == "" {
		scope = "default"
	}

	if request.ID == "" {
		err := v1alpha2.NewCOAError(nil, "found invalid request ID", v1alpha2.InternalError)
		return err
	}

	err = s.DynamicClient.Resource(resourceId).Namespace(scope).Delete(ctx, request.ID, metav1.DeleteOptions{})
	if err != nil {
		sLog.Errorf("  P (K8s State): failed to delete objects: %v", err)
		return err
	}
	return nil
}

func (s *K8sStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	ctx, span := observability.StartSpan("K8s State Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

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

	if request.ID == "" {
		err := v1alpha2.NewCOAError(nil, "found invalid request ID", v1alpha2.InternalError)
		return states.StateEntry{}, err
	}

	item, err := s.DynamicClient.Resource(resourceId).Namespace(scope).Get(ctx, request.ID, metav1.GetOptions{})
	if err != nil {
		coaError := v1alpha2.NewCOAError(err, "failed to get object", v1alpha2.InternalError)
		//check if not found
		if k8s_errors.IsNotFound(err) {
			coaError.State = v1alpha2.NotFound
		}
		sLog.Errorf("  P (K8s State %v", coaError.Error())
		return states.StateEntry{}, coaError
	}
	generation := item.GetGeneration()
	ret := states.StateEntry{
		ID:   request.ID,
		ETag: strconv.FormatInt(generation, 10),
		Body: map[string]interface{}{
			"spec":   item.Object["spec"],
			"status": item.Object["status"],
			"scope":  scope,
		},
	}
	return ret, nil
}

// Implmeement the IConfigProvider interface
func (s *K8sStateProvider) Read(object string, field string) (string, error) {
	obj, err := s.Get(context.TODO(), states.GetRequest{
		ID: object,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "catalogs",
		},
	})
	if err != nil {
		return "", err
	}
	if v, ok := obj.Body.(map[string]interface{})["spec"]; ok {
		spec := v.(map[string]interface{})
		if v, ok := spec["properties"]; ok {
			properties := v.(map[string]interface{})
			if v, ok := properties[field]; ok {
				return v.(string), nil
			} else {
				return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration catalog '%s'", field, object), v1alpha2.NotFound)
			}
		} else {
			return "", v1alpha2.NewCOAError(nil, "properties not found", v1alpha2.NotFound)
		}
	}
	return "", v1alpha2.NewCOAError(nil, "spec not found", v1alpha2.NotFound)
}

func (s *K8sStateProvider) ReadObject(object string) (map[string]string, error) {
	obj, err := s.Get(context.TODO(), states.GetRequest{
		ID: object,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "catalogs",
		},
	})
	if err != nil {
		return nil, err
	}
	if v, ok := obj.Body.(map[string]interface{})["spec"]; ok {
		spec := v.(map[string]interface{})
		if v, ok := spec["properties"]; ok {
			properties := v.(map[string]interface{})
			ret := map[string]string{}
			for k, v := range properties {
				ret[k] = v.(string)
			}
			return ret, nil
		} else {
			return nil, v1alpha2.NewCOAError(nil, "properties not found", v1alpha2.NotFound)
		}
	}
	return nil, v1alpha2.NewCOAError(nil, "spec not found", v1alpha2.NotFound)
}

func (s *K8sStateProvider) Set(object string, field string, value string, scope string) error {
	obj, err := s.Get(context.TODO(), states.GetRequest{
		ID: object,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "catalogs",
			"scope":    scope,
		},
	})
	if err != nil {
		return err
	}
	if v, ok := obj.Body.(map[string]interface{})["spec"]; ok {
		spec := v.(map[string]interface{})
		if v, ok := spec["properties"]; ok {
			properties := v.(map[string]interface{})
			properties[field] = value
			_, err := s.Upsert(context.TODO(), states.UpsertRequest{
				Value: obj,
				Metadata: map[string]string{
					"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Catalog", "metadata": {"name": "${{$catalog()}}"}}`, model.FederationGroup),
					"scope":    scope,
					"group":    model.FederationGroup,
					"version":  "v1",
					"resource": "catalogs",
				},
			})
			return err
		} else {
			return v1alpha2.NewCOAError(nil, "properties not found", v1alpha2.NotFound)
		}
	}
	return v1alpha2.NewCOAError(nil, "spec not found", v1alpha2.NotFound)
}
func (s *K8sStateProvider) SetObject(object string, values map[string]string, scope string) error {
	obj, err := s.Get(context.TODO(), states.GetRequest{
		ID: object,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "catalogs",
			"scope":    scope,
		},
	})
	if err != nil {
		return err
	}
	if v, ok := obj.Body.(map[string]interface{})["spec"]; ok {
		spec := v.(map[string]interface{})
		if v, ok := spec["properties"]; ok {
			properties := v.(map[string]interface{})
			for k, v := range values {
				properties[k] = v
			}
			_, err := s.Upsert(context.TODO(), states.UpsertRequest{
				Value: obj,
				Metadata: map[string]string{
					"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Catalog", "metadata": {"name": "${{$catalog()}}"}}`, model.FederationGroup),
					"scope":    scope,
					"group":    model.FederationGroup,
					"version":  "v1",
					"resource": "catalogs",
				},
			})
			return err
		} else {
			return v1alpha2.NewCOAError(nil, "properties not found", v1alpha2.NotFound)
		}
	}
	return v1alpha2.NewCOAError(nil, "spec not found", v1alpha2.NotFound)
}
func (s *K8sStateProvider) Remove(object string, field string, scope string) error {
	obj, err := s.Get(context.TODO(), states.GetRequest{
		ID: object,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "catalogs",
			"scope":    scope,
		},
	})
	if err != nil {
		return err
	}
	if v, ok := obj.Body.(map[string]interface{})["spec"]; ok {
		spec := v.(map[string]interface{})
		if v, ok := spec["properties"]; ok {
			properties := v.(map[string]interface{})
			delete(properties, field)
			_, err := s.Upsert(context.TODO(), states.UpsertRequest{
				Value: obj,
				Metadata: map[string]string{
					"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Catalog", "metadata": {"name": "${{$catalog()}}"}}`, model.FederationGroup),
					"scope":    scope,
					"group":    model.FederationGroup,
					"version":  "v1",
					"resource": "catalogs",
				},
			})
			return err
		} else {
			return v1alpha2.NewCOAError(nil, "properties not found", v1alpha2.NotFound)
		}
	}
	return v1alpha2.NewCOAError(nil, "spec not found", v1alpha2.NotFound)
}
func (s *K8sStateProvider) RemoveObject(object string, scope string) error {
	return s.Delete(context.TODO(), states.DeleteRequest{
		ID: object,
		Metadata: map[string]string{
			"scope":    scope,
			"group":    model.FederationGroup,
			"version":  "v1",
			"resource": "catalogs",
		},
	})
}
