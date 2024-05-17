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
	User     string `json:"user"`
	Password string `json:"password"`
}

type MaterializeStageProvider struct {
	Config    MaterializeStageProviderConfig
	Context   *contexts.ManagerContext
	ApiClient utils.ApiClient
}

func (s *MaterializeStageProvider) Init(config providers.IProviderConfig) error {
	maLock.Lock()
	defer maLock.Unlock()
	mockConfig, err := toMaterializeStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	s.ApiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
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
	if utils.ShouldUseUserCreds() {
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
	}
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

	objects, ok := inputs["names"].([]interface{})
	if !ok {
		err = v1alpha2.NewCOAError(nil, "input names is not a valid list", v1alpha2.BadRequest)
		return outputs, false, err
	}
	prefixedNames := make([]string, len(objects))
	for i, object := range objects {
		objString, ok := object.(string)
		if !ok {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("input name is not a valid string: %v", objects), v1alpha2.BadRequest)
			return outputs, false, err
		}
		if s, ok := inputs["__origin"]; ok {
			prefixedNames[i] = fmt.Sprintf("%s-%s", s, objString)
		} else {
			prefixedNames[i] = objString
		}
	}
	namespace := stage.GetNamespace(inputs)
	if namespace == "" {
		namespace = "default"
	}

	mLog.Debugf("  P (Materialize Processor): masterialize %v in namespace %s", prefixedNames, namespace)

	var catalogs []model.CatalogState
	catalogs, err = i.ApiClient.GetCatalogs(ctx, namespace, i.Config.User, i.Config.Password)

	if err != nil {
		return outputs, false, err
	}
	creationCount := 0
	for _, catalog := range catalogs {
		for _, object := range prefixedNames {
			objectName := object
			mLog.Debugf("  P (Materialize Processor): >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> objectName %v ", objectName)

			if strings.Contains(objectName, ":") {
				objectName = strings.ReplaceAll(objectName, ":", "-")
			}
			if catalog.ObjectMeta.Name == objectName {
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

					if instanceState.ObjectMeta.Name == "" {
						mLog.Errorf("Instance name is empty %s", name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty instance name: %s", name), v1alpha2.BadRequest)
					}

					instanceName := object
					if instanceState.ObjectMeta.Name != name {
						instanceName = instanceState.ObjectMeta.Name
						var rootResource string
						var version string
						parts := strings.Split(instanceName, ":")
						if len(parts) == 2 {
							rootResource = parts[0]
							version = parts[1]
						} else {
							mLog.Errorf("Instance name is invalid %s", instanceName)
							return outputs, false, err
						}
						if (instanceState.Spec.RootResource == "" || instanceState.Spec.Version == "") && rootResource != "" && version != "" {
							instanceState.Spec.RootResource = rootResource
							instanceState.Spec.Version = version
						}
					}

					instanceState.ObjectMeta = updateObjectMeta(instanceState.ObjectMeta, inputs, name)
					objectData, _ := json.Marshal(instanceState)
					mLog.Debugf("  P (Materialize Processor): materialize instance %v to namespace %s", instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace)
					err = i.ApiClient.CreateInstance(ctx, instanceName, objectData, instanceState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
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

					if solutionState.ObjectMeta.Name == "" {
						mLog.Errorf("Solution name is empty %s", name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty solution name: %s", name), v1alpha2.BadRequest)
					}

					solutionName := object
					if solutionState.ObjectMeta.Name != name {
						solutionName = solutionState.ObjectMeta.Name
						var rootResource string
						var version string
						parts := strings.Split(solutionName, ":")
						if len(parts) == 2 {
							rootResource = parts[0]
							version = parts[1]
						} else {
							mLog.Errorf("Instance name is invalid %s", solutionName)
							return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid instance name: %s", name), v1alpha2.BadRequest)
						}
						if (solutionState.Spec.RootResource == "" || solutionState.Spec.Version == "") && rootResource != "" && version != "" {
							solutionState.Spec.RootResource = rootResource
							solutionState.Spec.Version = version
						}
					}

					solutionState.ObjectMeta = updateObjectMeta(solutionState.ObjectMeta, inputs, name)
					objectData, _ := json.Marshal(solutionState)
					mLog.Debugf("  P (Materialize Processor): materialize solution %v to namespace %s", solutionState.ObjectMeta.Name, solutionState.ObjectMeta.Namespace)
					err = i.ApiClient.UpsertSolution(ctx, solutionName, objectData, solutionState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
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

					if targetState.ObjectMeta.Name == "" {
						mLog.Errorf("Target name is empty %s", name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty target name: %s", name), v1alpha2.BadRequest)
					}

					targetName := object
					if targetState.ObjectMeta.Name != name {
						targetName = targetState.ObjectMeta.Name
						var rootResource string
						var version string
						parts := strings.Split(targetName, ":")
						if len(parts) == 2 {
							rootResource = parts[0]
							version = parts[1]
						} else {
							mLog.Errorf("Instance name is invalid %s", targetName)
							return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid target name: %s", name), v1alpha2.BadRequest)
						}
						if (targetState.Spec.RootResource == "" || targetState.Spec.Version == "") && rootResource != "" && version != "" {
							targetState.Spec.RootResource = rootResource
							targetState.Spec.Version = version
						}
					}

					targetState.ObjectMeta = updateObjectMeta(targetState.ObjectMeta, inputs, name)
					objectData, _ := json.Marshal(targetState)
					mLog.Debugf("  P (Materialize Processor): materialize target %v to namespace %s", targetState.ObjectMeta.Name, targetState.ObjectMeta.Namespace)
					err = i.ApiClient.CreateTarget(ctx, targetName, objectData, targetState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
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

					if catalogState.ObjectMeta.Name == "" {
						mLog.Errorf("Catalog name is empty %s", name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty catalog name: %s", name), v1alpha2.BadRequest)
					}

					catalogName := object
					if catalogState.ObjectMeta.Name != name {
						catalogName = catalogState.ObjectMeta.Name
						var rootResource string
						var version string
						parts := strings.Split(catalogName, ":")
						if len(parts) == 2 {
							rootResource = parts[0]
							version = parts[1]
						} else {
							mLog.Errorf("Catalog name is invalid %s", catalogName)
							return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid catalog name: %s", name), v1alpha2.BadRequest)
						}
						if (catalogState.Spec.RootResource == "" || catalogState.Spec.Version == "") && rootResource != "" && version != "" {
							catalogState.Spec.RootResource = rootResource
							catalogState.Spec.Version = version
						}
					}

					catalogState.ObjectMeta = updateObjectMeta(catalogState.ObjectMeta, inputs, name)
					objectData, _ := json.Marshal(catalogState)
					mLog.Debugf("  P (Materialize Processor): materialize catalog %v to namespace %s", catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
					mLog.Debugf("  P (Materialize Processor): >>>>>>>>..>>> debug material, %s, %s", catalogState.ObjectMeta.Name, name)

					err = i.ApiClient.UpsertCatalog(ctx, catalogName, objectData, i.Config.User, i.Config.Password)
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
	if strings.Contains(objectMeta.Name, ":") {
		objectMeta.Name = strings.ReplaceAll(objectMeta.Name, ":", "-")
	}
	// stage inputs override objectMeta namespace
	if s := stage.GetNamespace(inputs); s != "" {
		objectMeta.Namespace = s
	} else if objectMeta.Namespace == "" {
		objectMeta.Namespace = "default"
	}
	return objectMeta
}
