/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package materialize

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var maLock sync.Mutex
var mLog = logger.NewLogger("coa.runtime")

type MaterializeStageProviderConfig struct {
	BaseUrl  string `json:"baseUrl"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type MaterializeStageProvider struct {
	Config  MaterializeStageProviderConfig
	Context *contexts.ManagerContext
}

func (s *MaterializeStageProvider) Init(config providers.IProviderConfig) error {
	maLock.Lock()
	defer maLock.Unlock()
	mockConfig, err := toMaterializeStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	return nil
}
func (s *MaterializeStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toMaterializeStageProviderConfig(config providers.IProviderConfig) (MaterializeStageProviderConfig, error) {
	ret := MaterializeStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *MaterializeStageProvider) InitWithMap(properties map[string]string) error {
	config, err := MaterialieStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MaterializeStageProviderConfigFromVendorMap(properties map[string]string) (MaterializeStageProviderConfig, error) {
	ret := make(map[string]string)
	for k, v := range properties {
		if strings.HasPrefix(k, "wait.") {
			ret[k[5:]] = v
		}
	}
	return MaterialieStageProviderConfigFromMap(ret)
}
func MaterialieStageProviderConfigFromMap(properties map[string]string) (MaterializeStageProviderConfig, error) {
	ret := MaterializeStageProviderConfig{}
	baseUrl, err := utils.GetString(properties, "baseUrl")
	if err != nil {
		return ret, err
	}
	ret.BaseUrl = baseUrl
	if ret.BaseUrl == "" {
		return ret, v1alpha2.NewCOAError(nil, "baseUrl is required", v1alpha2.BadConfig)
	}
	user, err := utils.GetString(properties, "user")
	if err != nil {
		return ret, err
	}
	ret.User = user
	if ret.User == "" {
		return ret, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
	}
	password, err := utils.GetString(properties, "password")
	if err != nil {
		return ret, err
	}
	ret.Password = password
	return ret, nil
}
func (i *MaterializeStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Materialize Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	mLog.Info("  P (Materialize Processor): processing inputs")
	outputs := make(map[string]interface{})

	objects := inputs["names"].([]interface{})
	prefixedNames := make([]string, len(objects))
	for i, object := range objects {
		if s, ok := inputs["__origin"]; ok {
			prefixedNames[i] = fmt.Sprintf("%s-%s", s, object.(string))
		} else {
			prefixedNames[i] = object.(string)
		}
	}
	namespace := stage.GetNamespace(inputs)
	if namespace == "" {
		namespace = "default"
	}

	mLog.Debugf("  P (Materialize Processor): masterialize %v in namespace %s", prefixedNames, namespace)

	var catalogs []model.CatalogState
	catalogs, err = utils.GetCatalogs(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password, namespace)

	if err != nil {
		return outputs, false, err
	}
	creationCount := 0
	for _, catalog := range catalogs {
		for _, object := range prefixedNames {
			if catalog.ObjectMeta.Name == object {
				objectData, _ := json.Marshal(catalog.Spec.Properties) //TODO: handle errors
				name := catalog.ObjectMeta.Name
				if s, ok := inputs["__origin"]; ok {
					name = strings.TrimPrefix(catalog.ObjectMeta.Name, fmt.Sprintf("%s-", s))
				}
				switch catalog.Spec.Type {
				case "instance":
					var instanceState model.InstanceState
					err = json.Unmarshal(objectData, &instanceState)
					if err != nil {
						mLog.Errorf("Failed to unmarshal instance state for catalog %s: %s", name, err.Error())
						return outputs, false, err
					}
					// If inner instace defines a display name, use it as the name
					if instanceState.Spec.DisplayName != "" {
						instanceState.ObjectMeta.Name = instanceState.Spec.DisplayName
					}
					instanceState.ObjectMeta = updateObjectMeta(instanceState.ObjectMeta, inputs, name)
					objectData, _ := json.Marshal(instanceState)
					mLog.Debugf("  P (Materialize Processor): materialize instance %v to namespace %s", instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace)
					err = utils.CreateInstance(ctx, i.Config.BaseUrl, instanceState.ObjectMeta.Name, i.Config.User, i.Config.Password, objectData, instanceState.ObjectMeta.Namespace)
					if err != nil {
						mLog.Errorf("Failed to create instance %s: %s", name, err.Error())
						return outputs, false, err
					}
					creationCount++
				case "solution":
					var solutionState model.SolutionState
					err = json.Unmarshal(objectData, &solutionState)
					if err != nil {
						mLog.Errorf("Failed to unmarshal solution state for catalog %s: %s: %s", name, err.Error())
						return outputs, false, err
					}
					// If inner solution defines a display name, use it as the name
					if solutionState.Spec.DisplayName != "" {
						solutionState.ObjectMeta.Name = solutionState.Spec.DisplayName
					}
					solutionState.ObjectMeta = updateObjectMeta(solutionState.ObjectMeta, inputs, name)
					objectData, _ := json.Marshal(solutionState)
					mLog.Debugf("  P (Materialize Processor): materialize solution %v to namespace %s", solutionState.ObjectMeta.Name, solutionState.ObjectMeta.Namespace)
					err = utils.UpsertSolution(ctx, i.Config.BaseUrl, solutionState.ObjectMeta.Name, i.Config.User, i.Config.Password, objectData, solutionState.ObjectMeta.Namespace)
					if err != nil {
						mLog.Errorf("Failed to create solution %s: %s", name, err.Error())
						return outputs, false, err
					}
					creationCount++
				case "target":
					var targetState model.TargetState
					err = json.Unmarshal(objectData, &targetState)
					if err != nil {
						mLog.Errorf("Failed to unmarshal target state for catalog %s: %s", name, err.Error())
						return outputs, false, err
					}
					// If inner target defines a display name, use it as the name
					if targetState.Spec.DisplayName != "" {
						targetState.ObjectMeta.Name = targetState.Spec.DisplayName
					}
					targetState.ObjectMeta = updateObjectMeta(targetState.ObjectMeta, inputs, name)
					objectData, _ := json.Marshal(targetState)
					mLog.Debugf("  P (Materialize Processor): materialize target %v to namespace %s", targetState.ObjectMeta.Name, targetState.ObjectMeta.Namespace)
					err = utils.CreateTarget(ctx, i.Config.BaseUrl, targetState.ObjectMeta.Name, i.Config.User, i.Config.Password, objectData, targetState.ObjectMeta.Namespace)
					if err != nil {
						mLog.Errorf("Failed to create target %s: %s", name, err.Error())
						return outputs, false, err
					}
					creationCount++
				default:
					// Check wrapped catalog structure and extract wrapped catalog name
					var catalogState model.CatalogState
					err = json.Unmarshal(objectData, &catalogState)
					if err != nil {
						mLog.Errorf("Failed to unmarshal catalog state for catalog %s: %s", name, err.Error())
						return outputs, false, err
					}
					catalogState.ObjectMeta = updateObjectMeta(catalogState.ObjectMeta, inputs, name)
					objectData, _ := json.Marshal(catalogState)
					mLog.Debugf("  P (Materialize Processor): materialize catalog %v to namespace %s", catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
					err = utils.UpsertCatalog(ctx, i.Config.BaseUrl, catalogState.ObjectMeta.Name, i.Config.User, i.Config.Password, objectData)
					if err != nil {
						mLog.Errorf("Failed to create catalog %s: %s", catalogState.ObjectMeta.Name, err.Error())
						return outputs, false, err
					}
					creationCount++
				}
			}
		}
	}
	if creationCount < len(objects) {
		err = v1alpha2.NewCOAError(nil, "failed to create all objects", v1alpha2.InternalError)
		return outputs, false, err
	}
	return outputs, false, nil
}

func updateObjectMeta(objectMeta model.ObjectMeta, inputs map[string]interface{}, catalogName string) model.ObjectMeta {
	if objectMeta.Name == "" {
		// use the same name as catalog wrapping it if not provided
		objectMeta.Name = catalogName
	}
	// stage inputs override objectMeta namespace
	if s := stage.GetNamespace(inputs); s != "" {
		objectMeta.Namespace = s
	} else if objectMeta.Namespace == "" {
		objectMeta.Namespace = "default"
	}
	return objectMeta
}
