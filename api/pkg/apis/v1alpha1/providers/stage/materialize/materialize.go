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
	"os"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
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
	mLog.Infof("  P (Materialize Processor): processing inputs, traceId: %s", span.SpanContext().TraceID().String())

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
			object := api_utils.ReplaceSeperator(object)
			if catalog.ObjectMeta.Name == object {
				label_key := os.Getenv("LABEL_KEY")
				label_value := os.Getenv("LABEL_VALUE")
				if label_key != "" && label_value != "" {
					// Check if metadata exists, if not create it
					metadata, ok := catalog.Spec.Properties["metadata"].(map[string]interface{})
					if !ok {
						metadata = make(map[string]interface{})
						catalog.Spec.Properties["metadata"] = metadata
					}

					// Check if labels exists within metadata, if not create it
					labels, ok := metadata["labels"].(map[string]string)
					if !ok {
						labels = make(map[string]string)
						metadata["labels"] = labels
					}

					// Add the label
					labels[label_key] = label_value
				}
				objectData, _ := json.Marshal(catalog.Spec.Properties) //TODO: handle errors
				name := catalog.ObjectMeta.Name
				if s, ok := inputs["__origin"]; ok {
					name = strings.TrimPrefix(catalog.ObjectMeta.Name, fmt.Sprintf("%s-", s))
				}
				switch catalog.Spec.CatalogType {
				case "instance":
					var instanceState model.InstanceState
					err = json.Unmarshal(objectData, &instanceState)
					if err != nil {
						mLog.Errorf("Failed to unmarshal instance state for catalog %s: %s", name, err.Error())
						return outputs, false, err
					}

					if instanceState.ObjectMeta.Name == "" {
						mLog.Errorf("Instance name is empty: catalog - %s", name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty instance name: catalog - %s", name), v1alpha2.BadRequest)
					}

					instanceState.ObjectMeta = updateObjectMeta(instanceState.ObjectMeta, inputs)
					objectData, _ := json.Marshal(instanceState)
					mLog.Debugf("  P (Materialize Processor): materialize instance %v to namespace %s", instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace)
					err = i.ApiClient.CreateInstance(ctx, instanceState.ObjectMeta.Name, objectData, instanceState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
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
						mLog.Errorf("Solution name is empty: catalog - %s", name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty solution name: catalog - %s", name), v1alpha2.BadRequest)
					}

					solutionName := solutionState.ObjectMeta.Name
					parts := strings.Split(solutionName, ":")
					if len(parts) == 2 {
						solutionState.Spec.RootResource = parts[0]
						solutionState.Spec.Version = parts[1]
					} else {
						mLog.Errorf("Solution name is invalid: solution - %s, catalog - %s", solutionName, name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid solution name: catalog - %s", name), v1alpha2.BadRequest)
					}

					mLog.Debugf("  P (Materialize Processor): check solution contains %v, namespace %s", solutionState.Spec.RootResource, namespace)
					_, err := i.ApiClient.GetSolutionContainer(ctx, solutionState.Spec.RootResource, namespace, i.Config.User, i.Config.Password)
					if err != nil && strings.Contains(err.Error(), constants.NotFound) {
						mLog.Debugf("Solution container %s doesn't exist: %s", solutionState.Spec.RootResource, err.Error())
						solutionContainerState := model.SolutionContainerState{ObjectMeta: model.ObjectMeta{Name: solutionState.Spec.RootResource, Namespace: namespace, Labels: solutionState.ObjectMeta.Labels}}
						containerObjectData, _ := json.Marshal(solutionContainerState)
						err = i.ApiClient.CreateSolutionContainer(ctx, solutionState.Spec.RootResource, containerObjectData, namespace, i.Config.User, i.Config.Password)
						if err != nil {
							mLog.Errorf("Failed to create solution container %s: %s", solutionState.Spec.RootResource, err.Error())
							return outputs, false, err
						}
					} else if err != nil {
						mLog.Errorf("Failed to get solution container %s: %s", solutionState.Spec.RootResource, err.Error())
						return outputs, false, err
					}

					solutionState.ObjectMeta = updateObjectMeta(solutionState.ObjectMeta, inputs)
					objectData, _ := json.Marshal(solutionState)
					mLog.Debugf("  P (Materialize Processor): materialize solution %v to namespace %s", solutionState.ObjectMeta.Name, solutionState.ObjectMeta.Namespace)
					err = i.ApiClient.UpsertSolution(ctx, solutionState.ObjectMeta.Name, objectData, solutionState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
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
						mLog.Errorf("Target name is empty: catalog - %s", name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty target name: catalog - %s", name), v1alpha2.BadRequest)
					}

					targetState.ObjectMeta = updateObjectMeta(targetState.ObjectMeta, inputs)
					objectData, _ := json.Marshal(targetState)
					mLog.Debugf("  P (Materialize Processor): materialize target %v to namespace %s", targetState.ObjectMeta.Name, targetState.ObjectMeta.Namespace)
					err = i.ApiClient.CreateTarget(ctx, targetState.ObjectMeta.Name, objectData, targetState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
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

					catalogName := catalogState.ObjectMeta.Name
					parts := strings.Split(catalogName, ":")
					if len(parts) == 2 {
						catalogState.Spec.RootResource = parts[0]
						catalogState.Spec.Version = parts[1]
					} else {
						mLog.Errorf("Catalog name is invalid: catalog - %s, parent catalog - %s", catalogName, name)
						return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid catalog name: catalog - %s", name), v1alpha2.BadRequest)
					}

					mLog.Debugf("  P (Materialize Processor): check catalog contains %v, namespace %s", catalogState.Spec.RootResource, namespace)
					_, err := i.ApiClient.GetCatalogContainer(ctx, catalogState.Spec.RootResource, namespace, i.Config.User, i.Config.Password)
					if err != nil && strings.Contains(err.Error(), constants.NotFound) {
						mLog.Debugf("Catalog container %s doesn't exist: %s", catalogState.Spec.RootResource, err.Error())
						catalogContainerState := model.CatalogContainerState{ObjectMeta: model.ObjectMeta{Name: catalogState.Spec.RootResource, Namespace: namespace, Labels: catalogState.ObjectMeta.Labels}}
						containerObjectData, _ := json.Marshal(catalogContainerState)
						err = i.ApiClient.CreateCatalogContainer(ctx, catalogState.Spec.RootResource, containerObjectData, namespace, i.Config.User, i.Config.Password)
						if err != nil {
							mLog.Errorf("Failed to create catalog container %s: %s", catalogState.Spec.RootResource, err.Error())
							return outputs, false, err
						}
					} else if err != nil {
						mLog.Errorf("Failed to get catalog container %s: %s", catalogState.Spec.RootResource, err.Error())
						return outputs, false, err
					}

					catalogState.ObjectMeta = updateObjectMeta(catalogState.ObjectMeta, inputs)
					objectData, _ := json.Marshal(catalogState)
					mLog.Debugf("  P (Materialize Processor): materialize catalog %v to namespace %s", catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
					err = i.ApiClient.UpsertCatalog(ctx, catalogState.ObjectMeta.Name, objectData, i.Config.User, i.Config.Password)
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

func updateObjectMeta(objectMeta model.ObjectMeta, inputs map[string]interface{}) model.ObjectMeta {
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
