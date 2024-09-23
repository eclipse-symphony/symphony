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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
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
	ctx, span := observability.StartSpan("K8s State Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.Debug("  P (K8s State): initialize")

	updateConfig, err := toK8sStateProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8s State): expected KubectlTargetProviderConfig: %+v", err)
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
					sLog.ErrorfCtx(ctx, "  P (K8s State): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "bytes":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (K8s State): %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.ErrorfCtx(ctx, "  P (K8s State): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and bytes", v1alpha2.BadConfig)
			sLog.ErrorfCtx(ctx, "  P (K8s State): %+v", err)
			return err
		}
	}
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8s State): %+v", err)
		return err
	}
	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8s State): %+v", err)
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	namespace := model.ReadPropertyCompat(entry.Metadata, "namespace", nil)
	group := model.ReadPropertyCompat(entry.Metadata, "group", nil)
	version := model.ReadPropertyCompat(entry.Metadata, "version", nil)
	resource := model.ReadPropertyCompat(entry.Metadata, "resource", nil)
	kind := model.ReadPropertyCompat(entry.Metadata, "kind", nil)

	if namespace == "" {
		namespace = "default"
	}
	sLog.DebugfCtx(ctx, "  P (K8s State): upsert state %s in namespace %s", entry.Value.ID, namespace)

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	if entry.Value.ID == "" {
		sLog.ErrorfCtx(ctx, "  P (K8s State): found invalid request ID")
		err := v1alpha2.NewCOAError(nil, "found invalid request ID", v1alpha2.BadRequest)
		return "", err
	}

	j, _ := json.Marshal(entry.Value.Body)
	var item *unstructured.Unstructured
	item, err = s.DynamicClient.Resource(resourceId).Namespace(namespace).Get(ctx, entry.Value.ID, metav1.GetOptions{})
	if err != nil {
		template := fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "%s", "metadata": {}}`, group, kind)
		var unc *unstructured.Unstructured
		err = json.Unmarshal([]byte(template), &unc)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (K8s State): failed to deserialize template: %v", err)
			return "", v1alpha2.NewCOAError(err, "failed to upsert state because of bad template", v1alpha2.BadRequest)
		}
		var dict map[string]interface{}
		err = json.Unmarshal(j, &dict)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (K8s State): failed to get object: %v", err)
			return "", v1alpha2.NewCOAError(err, "failed to upsert state because of bad inputs", v1alpha2.BadRequest)
		}
		unc.Object["spec"] = dict["spec"]
		metaJson, _ := json.Marshal(dict["metadata"])
		var metadata metav1.ObjectMeta
		err = json.Unmarshal(metaJson, &metadata)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (K8s State): failed to get object: %v", err)
			return "", v1alpha2.NewCOAError(err, "failed to upsert state because of bad inputs", v1alpha2.BadRequest)
		}
		unc.SetName(metadata.Name)
		unc.SetNamespace(metadata.Namespace)
		unc.SetLabels(metadata.Labels)
		unc.SetAnnotations(metadata.Annotations)

		_, err = s.DynamicClient.Resource(resourceId).Namespace(namespace).Create(ctx, unc, metav1.CreateOptions{})
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (K8s State): failed to create object: %v", err)
			return "", err
		}
		//Note: state is ignored for new object
	} else {
		j, _ := json.Marshal(entry.Value.Body)
		var dict map[string]interface{}
		err = json.Unmarshal(j, &dict)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (K8s State): failed to unmarshal object: %v", err)
			return "", v1alpha2.NewCOAError(err, "failed to upsert state because failed to unmarshal object", v1alpha2.BadRequest)
		}
		if v, ok := dict["metadata"]; ok {
			metaJson, _ := json.Marshal(v)
			var metadata model.ObjectMeta
			err = json.Unmarshal(metaJson, &metadata)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (K8s State): failed to unmarshal object metadata: %v", err)
				return "", v1alpha2.NewCOAError(err, "failed to upsert state because failed to unmarshal object metadata", v1alpha2.BadRequest)
			}
			item.SetName(metadata.Name)
			item.SetNamespace(metadata.Namespace)
			item.SetLabels(metadata.Labels)
			item.SetAnnotations(metadata.Annotations)
		}
		getResourceVersion := false
		if v, ok := dict["spec"]; ok && !entry.Options.UpdateStatusOnly {
			item.Object["spec"] = v

			_, err = s.DynamicClient.Resource(resourceId).Namespace(namespace).Update(ctx, item, metav1.UpdateOptions{})
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (K8s State): failed to update object: %v", err)
				return "", err
			}
			getResourceVersion = true
		}
		if v, ok := dict["status"]; ok {
			if getResourceVersion {
				// Get latest resource version in case the the object spec is also updated
				item, err = s.DynamicClient.Resource(resourceId).Namespace(namespace).Get(ctx, entry.Value.ID, metav1.GetOptions{})
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (K8s State): failed to get object when trying to update status: %v", err)
					return "", err
				}
			}

			if vMap, ok := v.(map[string]interface{}); ok {
				status := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": group + "/" + version,
						"kind":       "Status",
						"metadata": map[string]interface{}{
							"name": entry.Value.ID,
						},
						"status": vMap,
					},
				}
				status.SetResourceVersion(item.GetResourceVersion())
				_, err = s.DynamicClient.Resource(resourceId).Namespace(namespace).UpdateStatus(ctx, status, v1.UpdateOptions{})
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (K8s State): failed to update object status: %v", err)
					return "", err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "status field is not a valid map[string]interface{}", v1alpha2.BadRequest)
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	namespace := model.ReadPropertyCompat(request.Metadata, "namespace", nil)
	group := model.ReadPropertyCompat(request.Metadata, "group", nil)
	version := model.ReadPropertyCompat(request.Metadata, "version", nil)
	resource := model.ReadPropertyCompat(request.Metadata, "resource", nil)

	sLog.InfofCtx(ctx, "  P (K8s State): list state for %s.%s in namespace %s", resource, group, namespace)

	var namespaces []string
	if namespace == "" {
		ret, err := s.ListAllNamespaces(ctx, version)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (K8s State): failed to list namespaces: %v", err)
			return nil, "", err
		}
		namespaces = ret
	} else {
		namespaces = []string{namespace}
	}
	for _, namespace := range namespaces {
		resourceId := schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resource,
		}
		options := metav1.ListOptions{}
		filterValue := ""
		switch request.FilterType {
		case "label":
			labelSelector := request.FilterValue
			options = metav1.ListOptions{
				LabelSelector: labelSelector,
			}
		case "field":
			fieldSelector := request.FilterValue
			options = metav1.ListOptions{
				FieldSelector: fieldSelector,
			}
		case "spec":
			filterValue = request.FilterValue
		case "status":
			filterValue = request.FilterValue
		case "":
			//no filter
		default:
			sLog.ErrorfCtx(ctx, "  P (K8s State): invalid filter type: %s", request.FilterType)
			return nil, "", v1alpha2.NewCOAError(nil, "invalid filter type", v1alpha2.BadRequest)
		}
		items, err := s.DynamicClient.Resource(resourceId).Namespace(namespace).List(ctx, options)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (K8s State): failed to list objects in namespace %s: %v ", namespace, err)
			return nil, "", err
		}
		for _, v := range items.Items {

			if filterValue != "" {
				switch request.FilterType {
				case "spec":
					var dict map[string]interface{}
					j, _ := json.Marshal(v.Object["spec"])
					err = json.Unmarshal(j, &dict)
					if err != nil {
						sLog.ErrorfCtx(ctx, "  P (K8s State): failed to unmarshal object spec: %v", err)
						return nil, "", v1alpha2.NewCOAError(err, "failed to upsert state because failed to unmarshal object spec", v1alpha2.BadRequest)
					}
					if v, e := utils.JsonPathQuery(dict, filterValue); e != nil || v == nil {
						continue
					}
				case "status":
					if v.Object["status"] != nil {
						var dict map[string]interface{}
						j, _ := json.Marshal(v.Object["status"])
						err = json.Unmarshal(j, &dict)
						if err != nil {
							sLog.ErrorfCtx(ctx, "  P (K8s State): failed to unmarshal object status: %v", err)
							return nil, "", v1alpha2.NewCOAError(err, "failed to upsert state because failed to unmarshal object status", v1alpha2.BadRequest)
						}
						if v, e := utils.JsonPathQuery(dict, filterValue); e != nil || v == nil {
							continue
						}
					}
				}
			}

			generation := v.GetGeneration()
			metadata := model.ObjectMeta{
				Name:        v.GetName(),
				Namespace:   v.GetNamespace(),
				Labels:      v.GetLabels(),
				Annotations: v.GetAnnotations(),
			}
			entry := states.StateEntry{
				ETag: strconv.FormatInt(generation, 10),
				ID:   v.GetName(),
				Body: map[string]interface{}{
					"spec":     v.Object["spec"],
					"status":   v.Object["status"],
					"metadata": metadata,
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	namespace := model.ReadPropertyCompat(request.Metadata, "namespace", nil)
	group := model.ReadPropertyCompat(request.Metadata, "group", nil)
	version := model.ReadPropertyCompat(request.Metadata, "version", nil)
	resource := model.ReadPropertyCompat(request.Metadata, "resource", nil)

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	if namespace == "" {
		namespace = "default"
	}
	sLog.InfofCtx(ctx, "  P (K8s State): delete state %s in namespace %s", request.ID, namespace)

	if request.ID == "" {
		sLog.ErrorfCtx(ctx, "  P (K8s State): found invalid request ID")
		err := v1alpha2.NewCOAError(nil, "found invalid request ID", v1alpha2.BadRequest)
		return err
	}

	err = s.DynamicClient.Resource(resourceId).Namespace(namespace).Delete(ctx, request.ID, metav1.DeleteOptions{})
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8s State): failed to delete objects: %v", err)
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	namespace := model.ReadPropertyCompat(request.Metadata, "namespace", nil)
	group := model.ReadPropertyCompat(request.Metadata, "group", nil)
	version := model.ReadPropertyCompat(request.Metadata, "version", nil)
	resource := model.ReadPropertyCompat(request.Metadata, "resource", nil)

	if namespace == "" {
		namespace = "default"
	}

	sLog.InfofCtx(ctx, "  P (K8s State): get state %s in namespace %s", request.ID, namespace)

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	if request.ID == "" {
		sLog.ErrorfCtx(ctx, "  P (K8s State): found invalid request ID")
		err := v1alpha2.NewCOAError(nil, "found invalid request ID", v1alpha2.BadRequest)
		return states.StateEntry{}, err
	}

	item, err := s.DynamicClient.Resource(resourceId).Namespace(namespace).Get(ctx, request.ID, metav1.GetOptions{})
	if err != nil {
		coaError := v1alpha2.NewCOAError(err, "failed to get object", v1alpha2.InternalError)
		//check if not found
		if k8s_errors.IsNotFound(err) {
			coaError.State = v1alpha2.NotFound
		}
		sLog.ErrorfCtx(ctx, "  P (K8s State) %v", coaError.Error())
		return states.StateEntry{}, coaError
	}
	generation := item.GetGeneration()

	metadata := model.ObjectMeta{
		Name:        item.GetName(),
		Namespace:   item.GetNamespace(),
		Labels:      item.GetLabels(),
		Annotations: item.GetAnnotations(),
	}

	ret := states.StateEntry{
		ID:   request.ID,
		ETag: strconv.FormatInt(generation, 10),
		Body: map[string]interface{}{
			"spec":     item.Object["spec"],
			"status":   item.Object["status"],
			"metadata": metadata,
		},
	}
	return ret, nil
}
