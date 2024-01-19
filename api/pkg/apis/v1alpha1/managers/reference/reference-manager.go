/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package reference

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/oliveagle/jsonpath"
)

var log = logger.NewLogger("coa.runtime")

type ReferenceManager struct {
	managers.Manager
	ReferenceProviders map[string]reference.IReferenceProvider
	StateProvider      states.IStateProvider
	Reporter           reporter.IReporter
	CacheLifespan      uint64
}

type CachedItem struct {
	Created time.Time
	Item    interface{}
}

func (s *ReferenceManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		log.Errorf("M (Reference): failed to initialize manager %+v", err)
		return err
	}

	stateProvider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateProvider
	} else {
		log.Errorf("M (Reference): failed to get state provider %+v", err)
		return err
	}

	reportProvider, err := managers.GetReporter(config, providers)
	if err == nil {
		s.Reporter = reportProvider
	} else {
		log.Errorf("M (Reference): failed to get reporter %+v", err)
		return err
	}

	ctx := contexts.ManagerContext{}
	err = ctx.Init(context, nil)
	if err != nil {
		log.Errorf("M (Reference): failed to initialize manager context %+v", err)
		return err
	}

	s.CacheLifespan = 60
	if val, ok := config.Properties["cacheLifespan"]; ok {
		if i, err := strconv.ParseUint(val, 10, 32); err == nil {
			s.CacheLifespan = i
		}
	}

	s.Context = &ctx

	s.ReferenceProviders = make(map[string]reference.IReferenceProvider)

	for _, p := range providers {
		if kp, ok := p.(reference.IReferenceProvider); ok {
			s.ReferenceProviders[kp.ReferenceType()] = kp
			s.ReferenceProviders[kp.ReferenceType()].SetContext(s.Context)
		}
	}

	s.StateProvider.SetContext(s.Context)
	return nil
}

func (s *ReferenceManager) GetExt(refType string, namespace string, id1 string, group1 string, kind1 string, version1 string, id2 string, group2 string, kind2 string, version2 string, iteration string, alias string) ([]byte, error) {
	log.Infof("M (Reference): GetExt id1 - %s, id2 - %s, group2 - %s", id1, id2, group2)

	if group2 != "download" {
		data1, err := s.Get(refType, id1, namespace, group1, kind1, version1, "", "")
		if err != nil {
			log.Errorf("M (Reference): failed to get %s: %+v", id1, err)
			return nil, err
		}
		data2, err := s.Get(refType, id2, namespace, group2, kind2, version2, "", "")
		if err != nil {
			log.Errorf("M (Reference): failed to get %s: %+v", id2, err)
			return nil, err
		}
		return fillParameters(data1, data2, id1, alias)
	} else {
		data1, err := s.Get(refType, id1, namespace, group1, kind1, version1, "", "")
		if err != nil {
			log.Errorf("M (Reference): failed to get %s: %+v", id1, err)
			return nil, err
		}
		obj := make(map[string]interface{}, 0)
		err = json.Unmarshal(data1, &obj)
		if err != nil {
			log.Errorf("M (Reference): failed to unmarshall %s object: %+v", id1, err)
			return nil, err
		}
		var specData []byte
		if v, ok := obj["spec"]; ok {
			specData, err = json.Marshal(v)
			if err != nil {
				log.Errorf("M (Reference): failed to unmarshall %s spec: %+v", id1, err)
				return nil, err
			}
		} else {
			log.Errorf("M (Reference): %s spec property not found", id1)
			return nil, v1alpha2.NewCOAError(nil, "resolved object doesn't contain a 'spec' property", v1alpha2.InternalError)
		}

		model := model.ModelSpec{}
		err = json.Unmarshal(specData, &model)
		if err != nil {
			log.Errorf("M (Reference): failed to unmarshall %s object spec: %+v", id1, err)
			return nil, err
		}
		modelType := safeRead("model.type", model.Properties)
		if modelType != "customvision" {
			log.Errorf("M (Reference): failed to unmarshall %s object spec:", id1)
			return nil, v1alpha2.NewCOAError(nil, "only 'customvision' model type is supported", v1alpha2.InternalError)
		}
		modelProject := safeRead("model.project", model.Properties)
		if modelProject == "" {
			log.Errorf("M (Reference): failed to read %s model.project property", id1)
			return nil, v1alpha2.NewCOAError(nil, "property 'model.project' is not found", v1alpha2.InternalError)
		}
		modelEndpoint := safeRead("model.endpoint", model.Properties)
		if modelEndpoint == "" {
			log.Errorf("M (Reference): failed to read %s model.endpoint property", id1)
			return nil, v1alpha2.NewCOAError(nil, "property 'model.endpoint' is not found", v1alpha2.InternalError)
		}
		modelVersions := make(map[string]string)
		for k, v := range model.Properties {
			if strings.HasPrefix(k, "model.version.") {
				modelVersions[k] = v
			}
		}
		if len(modelVersions) == 0 {
			log.Errorf("M (Reference): failed to read %s model.version property", id1)
			return nil, v1alpha2.NewCOAError(nil, "no model version are found", v1alpha2.InternalError)
		}
		selection := ""
		if iteration == "latest" {
			selection = findLatest(modelVersions)
		} else {
			if v, ok := modelVersions["model.version."+iteration]; ok {
				selection = v
			}
		}
		if selection == "" {
			log.Errorf("M (Reference): failed to read %s model.version property", id1)
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("requested version 'model.version.%s' is not found", iteration), v1alpha2.InternalError)
		}

		downloadData, err := s.Get("v1alpha2.CustomVision", modelProject, modelEndpoint, kind2, version2, selection, "", "")
		if err != nil {
			log.Errorf("M (Reference): failed to get %s: %+v", modelProject, err)
			return nil, err
		}
		return downloadData, nil
	}
}
func findLatest(dict map[string]string) string {
	largest := 0
	ret := ""
	for k, v := range dict {
		vk := k[14:]
		i, err := strconv.Atoi(vk)
		if err == nil && i >= largest {
			largest = i
			ret = v
		}
	}
	return ret
}

func safeRead(key string, dict map[string]string) string {
	if v, ok := dict[key]; ok {
		return v
	}
	return ""
}

func (s *ReferenceManager) Get(refType string, id string, namespace string, group string, kind string, version string, labelSelector string, fieldSelector string) ([]byte, error) {
	var entityId string
	if labelSelector != "" || fieldSelector != "" {
		entityId = fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s", refType, labelSelector, fieldSelector, namespace, group, kind, version)
	} else {
		entityId = fmt.Sprintf("%s-%s-%s-%s-%s-%s", refType, id, namespace, group, kind, version)
	}

	log.Infof("M (Reference): Get entityId- %s", entityId)

	entity, err := s.StateProvider.Get(context.TODO(), states.GetRequest{
		ID: entityId,
	})
	if err == nil {
		data, _ := json.Marshal(entity.Body)
		cachedItem := CachedItem{}
		if err == nil {
			err := json.Unmarshal(data, &cachedItem)
			if err == nil {
				if time.Since(cachedItem.Created).Seconds() <= float64(s.CacheLifespan) {
					cacheData, _ := json.Marshal(cachedItem.Item)
					return cacheData, nil
				}
			}
		}
	}
	var provider reference.IReferenceProvider
	if p, ok := s.ReferenceProviders[refType]; ok {
		provider = p
	} else if len(s.ReferenceProviders) == 1 {
		for _, v := range s.ReferenceProviders {
			provider = v
			break
		}
	} else {
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("reference provider for '%s' is not configured", refType), v1alpha2.InternalError)
	}

	var ref interface{}
	if labelSelector != "" || fieldSelector != "" {
		ref, err = provider.List(labelSelector, fieldSelector, namespace, group, kind, version, refType)
	} else {
		ref, err = provider.Get(id, namespace, group, kind, version, refType)
	}
	if err != nil {
		return nil, err
	}
	s.StateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: entityId,
			Body: CachedItem{
				Created: time.Now(),
				Item:    ref,
			},
		},
	})
	refData, _ := json.Marshal(ref)
	return refData, nil
}

func (s *ReferenceManager) Report(id string, namespace string, group string, kind string, version string, properties map[string]string, overwrite bool) error {
	return s.Reporter.Report(id, namespace, group, kind, version, properties, overwrite)
}
func (s *ReferenceManager) Enabled() bool {
	return s.Config.Properties["poll.enabled"] == "true"
}
func (s *ReferenceManager) Poll() []error {
	return nil
}

func (s *ReferenceManager) Reconcil() []error {
	return nil
}

func fillParameters(data1 []byte, data2 []byte, id string, alias string) ([]byte, error) {
	params1, err := getParameterMap(data1, "", "") //parameters in skill
	if err != nil {
		return nil, err
	}
	params2, err := getParameterMap(data2, id, alias) // parameters in instance
	if err != nil {
		return nil, err
	}
	// for k, _ := range params1 {
	// 	key := id + "." + k
	// 	if alias != "" {
	// 		key = id + "." + alias + "." + k
	// 	}
	// 	if v2, ok := params2[key]; ok {
	// 		params1[k] = v2
	// 	}
	// }
	for k, _ := range params1 {
		if v2, ok := params2[k]; ok {
			params1[k] = v2
		}
	}
	strData := string(data1)
	for k, v := range params1 {
		strData = strings.ReplaceAll(strData, "$param("+k+")", v) //TODO: this needs to use property expression syntax instead of string replaces
	}
	return []byte(strData), nil
}
func getParameterMap(data []byte, skill string, alias string) (map[string]string, error) {
	var obj interface{}
	dict := make(map[string]string)
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	params, err := jsonpath.JsonPathLookup(obj, "$.parameters")
	if err == nil {
		coll := params.(map[string]interface{})
		for k, p := range coll {
			dict[k] = p.(string)
		}
	}
	params, err = jsonpath.JsonPathLookup(obj, "$.spec.parameters")
	if err == nil {
		coll := params.(map[string]interface{})
		for k, p := range coll {
			dict[k] = p.(string)
		}
	}
	if skill != "" && alias != "" {
		params, err = jsonpath.JsonPathLookup(obj, fmt.Sprintf("$.pipelines[?(@.name == '%s' && @.skill == '%s')].parameters", skill, alias))
		if err == nil {
			coll := params.([]interface{})
			for _, p := range coll {
				pk := p.(map[string]interface{})
				for k, v := range pk {
					dict[k] = v.(string)
				}
			}
		}
		params, err = jsonpath.JsonPathLookup(obj, fmt.Sprintf("$.spec.pipelines[?(@.name == '%s' && @.skill == '%s')].parameters", skill, alias))
		if err == nil {
			coll := params.([]interface{})
			for _, p := range coll {
				pk := p.(map[string]interface{})
				for k, v := range pk {
					dict[k] = v.(string)
				}
			}
		}
	}

	return dict, nil
}
