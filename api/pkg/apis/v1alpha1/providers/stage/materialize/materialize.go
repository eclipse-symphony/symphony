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
	"time"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	loggerName   = "providers.stage.materialize"
	providerName = "P (Materialize Stage)"
	materialize  = "materialize"
)

var (
	maLock                   sync.Mutex
	mLog                     = logger.NewLogger(loggerName)
	once                     sync.Once
	providerOperationMetrics *metrics.Metrics
)

type MaterializeStageProviderConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type MaterializeStageProvider struct {
	Config    MaterializeStageProviderConfig
	Context   *contexts.ManagerContext
	ApiClient api_utils.ApiClient
}

func (s *MaterializeStageProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("[Stage] Materialize Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	maLock.Lock()
	defer maLock.Unlock()
	var mockConfig MaterializeStageProviderConfig
	mockConfig, err = toMaterializeStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	s.ApiClient, err = api_utils.GetApiClient()
	if err != nil {
		return err
	}
	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				mLog.ErrorfCtx(ctx, "  P (Materialize Stage): failed to create metrics: %+v", err)
			}
		}
	})
	return err
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
	if api_utils.ShouldUseUserCreds() {
		user, err := api_utils.GetString(properties, "user")
		if err != nil {
			return ret, err
		}
		ret.User = user
		if ret.User == "" {
			return ret, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		}
		password, err := api_utils.GetString(properties, "password")
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	mLog.InfoCtx(ctx, "  P (Materialize Processor): processing inputs")
	processTime := time.Now().UTC()
	functionName := observ_utils.GetFunctionName()
	defer providerOperationMetrics.ProviderOperationLatency(
		processTime,
		materialize,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)

	outputs := make(map[string]interface{})

	objects, ok := inputs["names"].([]interface{})
	if !ok {
		err = v1alpha2.NewCOAError(nil, "input names is not a valid list", v1alpha2.BadRequest)
		providerOperationMetrics.ProviderOperationErrors(
			materialize,
			functionName,
			metrics.ProcessOperation,
			metrics.ValidateOperationType,
			v1alpha2.BadConfig.String(),
		)
		return outputs, false, err
	}
	prefixedNames := make([]string, len(objects))
	for i, object := range objects {
		objString, ok := object.(string)
		if !ok {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("input name is not a valid string: %v", objects), v1alpha2.BadRequest)
			providerOperationMetrics.ProviderOperationErrors(
				materialize,
				functionName,
				metrics.ProcessOperation,
				metrics.ValidateOperationType,
				v1alpha2.BadConfig.String(),
			)
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

	mLog.DebugfCtx(ctx, "  P (Materialize Processor): materialize %v in namespace %s", prefixedNames, namespace)

	// Fail fast check
	catalogs := []model.CatalogState{}
	errorMessage := "Failed to get all catalogs: "
	anyCatalogInvalid := false
	for _, objectRef := range prefixedNames {
		objectName := api_utils.ConvertReferenceToObjectName(objectRef)
		catalog, err := i.ApiClient.GetCatalog(ctx, objectName, namespace, i.Config.User, i.Config.Password)
		if err != nil {
			errorMessage = fmt.Sprintf("%s %s(reason: %s).", errorMessage, objectRef, err.Error())
			anyCatalogInvalid = anyCatalogInvalid || api_utils.IsNotFound(err)
			continue
		}
		if !checkCatalog(&catalog) {
			errorMessage = fmt.Sprintf("%s %s(reason: catalog doesn't have a valid Spec.CatalogType).", errorMessage, objectRef)
			anyCatalogInvalid = true
			continue
		}
		catalogs = append(catalogs, catalog)
	}
	if len(catalogs) < len(prefixedNames) {
		mLog.ErrorCtx(ctx, errorMessage)
		providerOperationMetrics.ProviderOperationErrors(
			materialize,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.CatalogsGetFailed.String(),
		)
		if anyCatalogInvalid {
			return outputs, false, v1alpha2.NewCOAError(nil, errorMessage, v1alpha2.BadRequest)
		} else {
			return outputs, false, v1alpha2.NewCOAError(nil, errorMessage, v1alpha2.InternalError)
		}
	}

	createdObjectList := make(map[string]bool, 0)
	for _, catalog := range catalogs {
		label_key := os.Getenv("LABEL_KEY")
		label_value := os.Getenv("LABEL_VALUE")
		annotation_name := os.Getenv("ANNOTATION_KEY")
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
				mLog.ErrorfCtx(ctx, "Failed to unmarshal instance state for catalog %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidInstanceCatalog.String(),
				)
				return outputs, false, err
			}

			if instanceState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "Instance name is empty: catalog - %s", name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidInstanceCatalog.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty instance name: catalog - %s", name), v1alpha2.BadRequest)
			}

			instanceState.ObjectMeta = updateObjectMeta(instanceState.ObjectMeta, inputs)
			objectData, _ := json.Marshal(instanceState)
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): materialize instance %v to namespace %s", instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to materialize instance %v to namespace %s", instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace)
			err = i.ApiClient.CreateInstance(ctx, instanceState.ObjectMeta.Name, objectData, instanceState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to create instance %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateInstanceFromCatalogFailed.String(),
				)
				return outputs, false, err
			}
			createdObjectList[catalog.ObjectMeta.Name] = true
		case "solution":
			var solutionState model.SolutionState
			err = json.Unmarshal(objectData, &solutionState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal solution state for catalog %s: %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidSolutionCatalog.String(),
				)
				return outputs, false, err
			}

			if solutionState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "Solution name is empty: catalog - %s", name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidSolutionCatalog.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty solution name: catalog - %s", name), v1alpha2.BadRequest)
			}

			solutionName := solutionState.ObjectMeta.Name
			parts := strings.Split(solutionName, constants.ReferenceSeparator)
			if len(parts) == 2 {
				solutionState.Spec.RootResource = parts[0]
				solutionState.Spec.Version = parts[1]
			} else {
				mLog.ErrorfCtx(ctx, "Solution name is invalid: solution - %s, catalog - %s", solutionName, name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidSolutionCatalog.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid solution name: catalog - %s", name), v1alpha2.BadRequest)
			}

			if annotation_name != "" {
				solutionState.ObjectMeta.UpdateAnnotation(annotation_name, parts[1])
			}
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): check solution contains %v, namespace %s", solutionState.Spec.RootResource, namespace)
			_, err := i.ApiClient.GetSolutionContainer(ctx, solutionState.Spec.RootResource, namespace, i.Config.User, i.Config.Password)
			if err != nil && strings.Contains(err.Error(), v1alpha2.NotFound.String()) {
				mLog.DebugfCtx(ctx, "Solution container %s doesn't exist: %s", solutionState.Spec.RootResource, err.Error())
				solutionContainerState := model.SolutionContainerState{ObjectMeta: model.ObjectMeta{Name: solutionState.Spec.RootResource, Namespace: namespace, Labels: solutionState.ObjectMeta.Labels}}
				containerObjectData, _ := json.Marshal(solutionContainerState)
				observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to create solution container %v in namespace %s", solutionState.Spec.RootResource, namespace)
				err = i.ApiClient.CreateSolutionContainer(ctx, solutionState.Spec.RootResource, containerObjectData, namespace, i.Config.User, i.Config.Password)
				if err != nil {
					mLog.ErrorfCtx(ctx, "Failed to create solution container %s: %s", solutionState.Spec.RootResource, err.Error())
					providerOperationMetrics.ProviderOperationErrors(
						materialize,
						functionName,
						metrics.ProcessOperation,
						metrics.RunOperationType,
						v1alpha2.ParentObjectCreateFailed.String(),
					)
					return outputs, false, err
				}
			} else if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to get solution container %s: %s", solutionState.Spec.RootResource, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.ParentObjectMissing.String(),
				)
				return outputs, false, err
			}

			solutionState.ObjectMeta = updateObjectMeta(solutionState.ObjectMeta, inputs)
			objectData, _ := json.Marshal(solutionState)
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): materialize solution %v to namespace %s", solutionState.ObjectMeta.Name, solutionState.ObjectMeta.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to materialize solution %v to namespace %s", solutionState.ObjectMeta.Name, solutionState.ObjectMeta.Namespace)
			err = i.ApiClient.UpsertSolution(ctx, solutionState.ObjectMeta.Name, objectData, solutionState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to create solution %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateSolutionFromCatalogFailed.String(),
				)
				return outputs, false, err
			}
			createdObjectList[catalog.ObjectMeta.Name] = true
		case "target":
			var targetState model.TargetState
			err = json.Unmarshal(objectData, &targetState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal target state for catalog %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidTargetCatalog.String(),
				)
				return outputs, false, err
			}

			if targetState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "Target name is empty: catalog - %s", name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidTargetCatalog.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty target name: catalog - %s", name), v1alpha2.BadRequest)
			}

			targetState.ObjectMeta = updateObjectMeta(targetState.ObjectMeta, inputs)
			objectData, _ := json.Marshal(targetState)
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): materialize target %v to namespace %s", targetState.ObjectMeta.Name, targetState.ObjectMeta.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to materialize target %v to namespace %s", targetState.ObjectMeta.Name, targetState.ObjectMeta.Namespace)
			err = i.ApiClient.CreateTarget(ctx, targetState.ObjectMeta.Name, objectData, targetState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to create target %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateTargetFromCatalogFailed.String(),
				)
				return outputs, false, err
			}
			createdObjectList[catalog.ObjectMeta.Name] = true
		default:
			// Check wrapped catalog structure and extract wrapped catalog name
			var catalogState model.CatalogState
			err = json.Unmarshal(objectData, &catalogState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal catalog state for catalog %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidCatalogCatalog.String(),
				)
				return outputs, false, err
			}

			if catalogState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "Catalog name is empty %s", name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidCatalogCatalog.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty catalog name: %s", name), v1alpha2.BadRequest)
			}

			catalogName := catalogState.ObjectMeta.Name
			parts := strings.Split(catalogName, constants.ReferenceSeparator)
			if len(parts) == 2 {
				catalogState.Spec.RootResource = parts[0]
				catalogState.Spec.Version = parts[1]
			} else {
				mLog.ErrorfCtx(ctx, "Catalog name is invalid: catalog - %s, parent catalog - %s", catalogName, name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidCatalogCatalog.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid catalog name: catalog - %s", name), v1alpha2.BadRequest)
			}

			if annotation_name != "" {
				catalogState.ObjectMeta.UpdateAnnotation(annotation_name, parts[1])
			}
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): check catalog contains %v, namespace %s", catalogState.Spec.RootResource, namespace)
			_, err := i.ApiClient.GetCatalogContainer(ctx, catalogState.Spec.RootResource, namespace, i.Config.User, i.Config.Password)
			if err != nil && strings.Contains(err.Error(), v1alpha2.NotFound.String()) {
				mLog.DebugfCtx(ctx, "Catalog container %s doesn't exist: %s", catalogState.Spec.RootResource, err.Error())
				catalogContainerState := model.CatalogContainerState{ObjectMeta: model.ObjectMeta{Name: catalogState.Spec.RootResource, Namespace: namespace, Labels: catalogState.ObjectMeta.Labels}}
				containerObjectData, _ := json.Marshal(catalogContainerState)
				observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to create catalog container %v in namespace %s", catalogState.Spec.RootResource, namespace)
				err = i.ApiClient.CreateCatalogContainer(ctx, catalogState.Spec.RootResource, containerObjectData, namespace, i.Config.User, i.Config.Password)
				if err != nil {
					mLog.ErrorfCtx(ctx, "Failed to create catalog container %s: %s", catalogState.Spec.RootResource, err.Error())
					providerOperationMetrics.ProviderOperationErrors(
						materialize,
						functionName,
						metrics.ProcessOperation,
						metrics.RunOperationType,
						v1alpha2.ParentObjectCreateFailed.String(),
					)
					return outputs, false, err
				}
			} else if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to get catalog container %s: %s", catalogState.Spec.RootResource, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.ParentObjectMissing.String(),
				)
				return outputs, false, err
			}

			catalogState.ObjectMeta = updateObjectMeta(catalogState.ObjectMeta, inputs)
			objectData, _ := json.Marshal(catalogState)
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): materialize catalog %v to namespace %s", catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to materialize catalog %v to namespace %s", catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
			err = i.ApiClient.UpsertCatalog(ctx, catalogState.ObjectMeta.Name, objectData, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to create catalog %s: %s", catalogState.ObjectMeta.Name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateCatalogFromCatalogFailed.String(),
				)
				return outputs, false, err
			}
			createdObjectList[catalog.ObjectMeta.Name] = true
		}
	}
	if len(createdObjectList) < len(objects) {
		errorMessage := "failed to create all objects:"
		for _, catalog := range catalogs {
			if _, ok := createdObjectList[catalog.ObjectMeta.Name]; !ok {
				errorMessage = fmt.Sprintf("%s %s", errorMessage, catalog.ObjectMeta.Name)
			}
		}
		err = v1alpha2.NewCOAError(nil, errorMessage, v1alpha2.InternalError)
		providerOperationMetrics.ProviderOperationErrors(
			materialize,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.MaterializeBatchFailed.String(),
		)
		return outputs, false, err
	}

	return outputs, false, nil
}

func updateObjectMeta(objectMeta model.ObjectMeta, inputs map[string]interface{}) model.ObjectMeta {
	if strings.Contains(objectMeta.Name, constants.ReferenceSeparator) {
		objectMeta.Name = strings.ReplaceAll(objectMeta.Name, constants.ReferenceSeparator, constants.ResourceSeperator)
	}
	// stage inputs override objectMeta namespace
	if s := stage.GetNamespace(inputs); s != "" {
		objectMeta.Namespace = s
	} else if objectMeta.Namespace == "" {
		objectMeta.Namespace = "default"
	}
	return objectMeta
}

func checkCatalog(catalog *model.CatalogState) bool {
	if catalog.Spec == nil {
		return false
	}
	if catalog.Spec.CatalogType == "instance" || catalog.Spec.CatalogType == "solution" || catalog.Spec.CatalogType == "target" || catalog.Spec.CatalogType == "catalog" {
		return true
	}
	return false
}
