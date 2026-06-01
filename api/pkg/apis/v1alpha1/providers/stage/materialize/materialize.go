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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
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
	// TODO: this config is only available for k8s mode right now. Will support them in standalone later
	WaitForDeployment   bool          `json:"waitForDeployment"`
	WaitTimeout         string        `json:"waitTimeout"`
	WaitTimeoutDuration time.Duration `json:"-"` // this is not a json field
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
	err = utils2.UnmarshalJson(data, &ret)
	if err != nil {
		return ret, err
	}
	if ret.WaitForDeployment {
		if ret.WaitTimeout != "" {
			ret.WaitTimeoutDuration, err = time.ParseDuration(ret.WaitTimeout)
			if err != nil {
				return ret, err
			}
		} else {
			ret.WaitTimeoutDuration = 5 * time.Minute
		}
	}
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

func setLabels(meta *model.ObjectMeta) {
	label_key := os.Getenv("LABEL_KEY")
	label_value := os.Getenv("LABEL_VALUE")
	if label_key != "" && label_value != "" {
		if meta.Labels == nil {
			meta.Labels = make(map[string]string)
		}
		meta.Labels[label_key] = label_value
	}
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
	catalogversions := []model.CatalogVersionState{}
	errorMessage := "Failed to get all catalogversions: "
	anyCatalogVersionInvalid := false
	for _, objectRef := range prefixedNames {
		objectName := api_utils.ConvertReferenceToObjectName(objectRef)
		catalogversion, err := i.ApiClient.GetCatalogVersion(ctx, objectName, namespace, i.Config.User, i.Config.Password)
		if err != nil {
			errorMessage = fmt.Sprintf("%s %s(reason: %s).", errorMessage, objectRef, err.Error())
			anyCatalogVersionInvalid = anyCatalogVersionInvalid || api_utils.IsNotFound(err)
			continue
		}
		if !checkCatalogVersion(&catalogversion) {
			errorMessage = fmt.Sprintf("%s %s(reason: catalogversion doesn't have a valid Spec.CatalogType).", errorMessage, objectRef)
			anyCatalogVersionInvalid = true
			continue
		}
		catalogversions = append(catalogversions, catalogversion)
	}
	if len(catalogversions) < len(prefixedNames) {
		mLog.ErrorCtx(ctx, errorMessage)
		providerOperationMetrics.ProviderOperationErrors(
			materialize,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.CatalogVersionsGetFailed.String(),
		)
		if anyCatalogVersionInvalid {
			return outputs, false, v1alpha2.NewCOAError(nil, errorMessage, v1alpha2.BadRequest)
		} else {
			return outputs, false, v1alpha2.NewCOAError(nil, errorMessage, v1alpha2.InternalError)
		}
	}

	createdObjectList := make(map[string]bool, 0)
	instanceList := make([]utils.ObjectInfo, 0)
	targetList := make([]utils.ObjectInfo, 0)

	annotation_name := os.Getenv("ANNOTATION_KEY")

	for _, catalogversion := range catalogversions {
		objectData, _ := json.Marshal(catalogversion.Spec.Properties) //TODO: handle errors
		name := catalogversion.ObjectMeta.Name
		if s, ok := inputs["__origin"]; ok {
			name = strings.TrimPrefix(catalogversion.ObjectMeta.Name, fmt.Sprintf("%s-", s))
		}
		switch catalogversion.Spec.CatalogType {
		case "instance":
			var instanceState model.InstanceState
			err = utils2.UnmarshalJson(objectData, &instanceState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal instance state for catalogversion %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidInstanceCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid embeded instance in catalogversion %s", name), v1alpha2.BadRequest)
			}

			if instanceState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "Instance name is empty: catalogversion - %s", name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidInstanceCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty instance name: catalogversion - %s", name), v1alpha2.BadRequest)
			}

			operationIdKey := api_utils.GenerateOperationId()
			if operationIdKey != "" {
				if instanceState.ObjectMeta.Annotations == nil {
					instanceState.ObjectMeta.Annotations = make(map[string]string)
				}
				instanceState.ObjectMeta.Annotations[constants.AzureOperationIdKey] = operationIdKey
				mLog.InfofCtx(ctx, "  P (Materialize Processor): update %s annotation: %s to %s", instanceState.ObjectMeta.Name, constants.AzureOperationIdKey, instanceState.ObjectMeta.Annotations[constants.AzureOperationIdKey])
			}

			setLabels(&instanceState.ObjectMeta)

			instanceState.ObjectMeta = updateObjectMeta(instanceState.ObjectMeta, inputs)
			previousJobId := "-1"
			ret, err := i.ApiClient.GetInstance(ctx, instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil && !api_utils.IsNotFound(err) {
				mLog.ErrorfCtx(ctx, "Failed to get instance %s: %s", instanceState.ObjectMeta.Name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InstanceGetFailed.String(),
				)
				return outputs, false, err
			}
			if err == nil {
				// Get the previous job ID if instance exists
				previousJobId = ret.ObjectMeta.GetSummaryJobId()
			}
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): previous jobid is %s for instance: %s", previousJobId, instanceState.ObjectMeta.Name)

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
					v1alpha2.CreateInstanceFromCatalogVersionFailed.String(),
				)
				return outputs, false, err
			}
			// Get instance guid
			ret, err = i.ApiClient.GetInstance(ctx, instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to get instance %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InstanceGetFailed.String(),
				)
				return outputs, false, err
			}
			summaryId := ret.ObjectMeta.GetSummaryId()
			if summaryId == "" {
				mLog.ErrorfCtx(ctx, "Instance GUID is empty: - %s", instanceState.ObjectMeta.Name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateInstanceFailed.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty instance guid: - %s", instanceState.ObjectMeta.Name), v1alpha2.BadRequest)
			}
			instanceList = append(instanceList, utils.ObjectInfo{Name: ret.ObjectMeta.Name, SummaryId: summaryId, SummaryJobId: previousJobId})
			createdObjectList[catalogversion.ObjectMeta.Name] = true
		case "solutionVersion":
			var solutionversionState model.SolutionVersionState
			err = utils2.UnmarshalJson(objectData, &solutionversionState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal solutionversion state for catalogversion %s: %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidSolutionVersionCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid embeded solutionversion in catalogversion %s", name), v1alpha2.BadRequest)
			}

			if solutionversionState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "SolutionVersion name is empty: catalogversion - %s", name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidSolutionVersionCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty solutionversion name: catalogversion - %s", name), v1alpha2.BadRequest)
			}

			solutionversionName := solutionversionState.ObjectMeta.Name
			parts := strings.Split(solutionversionName, constants.ReferenceSeparator)
			if len(parts) == 2 {
				solutionversionState.Spec.RootResource = parts[0]
				solutionversionState.Spec.Version = parts[1]
			} else {
				mLog.ErrorfCtx(ctx, "SolutionVersion name is invalid: solutionversion - %s, catalogversion - %s", solutionversionName, name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidSolutionVersionCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid solutionversion name: catalogversion - %s", name), v1alpha2.BadRequest)
			}

			if annotation_name != "" {
				solutionversionState.ObjectMeta.UpdateAnnotation(annotation_name, parts[1])
			}

			setLabels(&solutionversionState.ObjectMeta)

			mLog.DebugfCtx(ctx, "  P (Materialize Processor): check solutionversion contains %v, namespace %s", solutionversionState.Spec.RootResource, namespace)
			_, err := i.ApiClient.GetSolution(ctx, solutionversionState.Spec.RootResource, namespace, i.Config.User, i.Config.Password)
			if err != nil && api_utils.IsNotFound(err) {
				mLog.DebugfCtx(ctx, "SolutionVersion container %s doesn't exist: %s", solutionversionState.Spec.RootResource, err.Error())
				solutionversionContainerState := model.SolutionState{ObjectMeta: model.ObjectMeta{Name: solutionversionState.Spec.RootResource, Namespace: namespace, Labels: solutionversionState.ObjectMeta.Labels}}
				containerObjectData, _ := json.Marshal(solutionversionContainerState)
				observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to create solutionversion container %v in namespace %s", solutionversionState.Spec.RootResource, namespace)
				err = i.ApiClient.CreateSolution(ctx, solutionversionState.Spec.RootResource, containerObjectData, namespace, i.Config.User, i.Config.Password)
				if err != nil {
					mLog.ErrorfCtx(ctx, "Failed to create solutionversion container %s: %s", solutionversionState.Spec.RootResource, err.Error())
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
				mLog.ErrorfCtx(ctx, "Failed to get solutionversion container %s: %s", solutionversionState.Spec.RootResource, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.ParentObjectMissing.String(),
				)
				return outputs, false, err
			}

			solutionversionState.ObjectMeta = updateObjectMeta(solutionversionState.ObjectMeta, inputs)
			objectData, _ := json.Marshal(solutionversionState)
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): materialize solutionversion %v to namespace %s", solutionversionState.ObjectMeta.Name, solutionversionState.ObjectMeta.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to materialize solutionversion %v to namespace %s", solutionversionState.ObjectMeta.Name, solutionversionState.ObjectMeta.Namespace)
			err = i.ApiClient.UpsertSolutionVersion(ctx, solutionversionState.ObjectMeta.Name, objectData, solutionversionState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to create solutionversion %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateSolutionVersionFromCatalogVersionFailed.String(),
				)
				return outputs, false, err
			}
			createdObjectList[catalogversion.ObjectMeta.Name] = true
		case "target":
			var targetState model.TargetState
			err = utils2.UnmarshalJson(objectData, &targetState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal target state for catalogversion %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidTargetCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid embeded target in catalogversion %s", name), v1alpha2.BadRequest)
			}

			if targetState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "Target name is empty: catalogversion - %s", name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidTargetCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty target name: catalogversion - %s", name), v1alpha2.BadRequest)
			}

			operationIdKey := api_utils.GenerateOperationId()
			if operationIdKey != "" {
				if targetState.ObjectMeta.Annotations == nil {
					targetState.ObjectMeta.Annotations = make(map[string]string)
				}
				targetState.ObjectMeta.Annotations[constants.AzureOperationIdKey] = operationIdKey
				mLog.InfofCtx(ctx, "  P (Materialize Processor): update %s annotation: %s to %s", targetState.ObjectMeta.Name, constants.AzureOperationIdKey, targetState.ObjectMeta.Annotations[constants.AzureOperationIdKey])
			}
			setLabels(&targetState.ObjectMeta)
			targetState.ObjectMeta = updateObjectMeta(targetState.ObjectMeta, inputs)
			previousJobId := "-1"
			ret, err := i.ApiClient.GetTarget(ctx, targetState.ObjectMeta.Name, targetState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil && !api_utils.IsNotFound(err) {
				mLog.ErrorfCtx(ctx, "Failed to get target %s: %s", targetState.ObjectMeta.Name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.TargetGetFailed.String(),
				)
				return outputs, false, err
			}
			if err == nil {
				// Get the previous job ID if target exists
				previousJobId = ret.ObjectMeta.GetSummaryJobId()
			}
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): previous jobid is %s for target: %s", previousJobId, targetState.ObjectMeta.Name)

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
					v1alpha2.CreateTargetFromCatalogVersionFailed.String(),
				)
				return outputs, false, err
			}
			// Get target guid
			ret, err = i.ApiClient.GetTarget(ctx, targetState.ObjectMeta.Name, targetState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to get instance %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.TargetGetFailed.String(),
				)
				return outputs, false, err
			}

			summaryId := ret.ObjectMeta.GetSummaryId()
			if summaryId == "" {
				mLog.ErrorfCtx(ctx, "Target GUID is empty: - %s", targetState.ObjectMeta.Name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.TargetGetFailed.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty target guid: - %s", targetState.ObjectMeta.Name), v1alpha2.BadRequest)
			}
			targetList = append(targetList, utils.ObjectInfo{Name: ret.ObjectMeta.Name, SummaryId: summaryId, SummaryJobId: previousJobId})
			createdObjectList[catalogversion.ObjectMeta.Name] = true
		default:
			// Check wrapped catalogversion structure and extract wrapped catalogversion name
			var catalogversionState model.CatalogVersionState
			err = utils2.UnmarshalJson(objectData, &catalogversionState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal catalogversion state for catalogversion %s: %s", name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidCatalogVersionCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid embeded catalogversion in catalogversion %s", name), v1alpha2.BadRequest)
			}

			if catalogversionState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "CatalogVersion name is empty %s", name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidCatalogVersionCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty catalogversion name: %s", name), v1alpha2.BadRequest)
			}

			catalogversionName := catalogversionState.ObjectMeta.Name
			parts := strings.Split(catalogversionName, constants.ReferenceSeparator)
			if len(parts) == 2 {
				catalogversionState.Spec.RootResource = parts[0]
				catalogversionState.Spec.Version = parts[1]
			} else {
				mLog.ErrorfCtx(ctx, "CatalogVersion name is invalid: catalogversion - %s, parent catalogversion - %s", catalogversionName, name)
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidCatalogVersionCatalogVersion.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid catalogversion name: catalogversion - %s", name), v1alpha2.BadRequest)
			}

			if annotation_name != "" {
				catalogversionState.ObjectMeta.UpdateAnnotation(annotation_name, parts[1])
			}

			setLabels(&catalogversionState.ObjectMeta)

			mLog.DebugfCtx(ctx, "  P (Materialize Processor): check catalogversion contains %v, namespace %s", catalogversionState.Spec.RootResource, namespace)
			_, err := i.ApiClient.GetCatalog(ctx, catalogversionState.Spec.RootResource, namespace, i.Config.User, i.Config.Password)
			if err != nil && api_utils.IsNotFound(err) {
				mLog.DebugfCtx(ctx, "CatalogVersion container %s doesn't exist: %s", catalogversionState.Spec.RootResource, err.Error())
				catalogState := model.CatalogState{ObjectMeta: model.ObjectMeta{Name: catalogversionState.Spec.RootResource, Namespace: namespace, Labels: catalogversionState.ObjectMeta.Labels}}
				containerObjectData, _ := json.Marshal(catalogState)
				observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to create catalogversion container %v in namespace %s", catalogversionState.Spec.RootResource, namespace)
				err = i.ApiClient.CreateCatalog(ctx, catalogversionState.Spec.RootResource, containerObjectData, namespace, i.Config.User, i.Config.Password)
				if err != nil {
					mLog.ErrorfCtx(ctx, "Failed to create catalogversion container %s: %s", catalogversionState.Spec.RootResource, err.Error())
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
				mLog.ErrorfCtx(ctx, "Failed to get catalogversion container %s: %s", catalogversionState.Spec.RootResource, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.ParentObjectMissing.String(),
				)
				return outputs, false, err
			}

			catalogversionState.ObjectMeta = updateObjectMeta(catalogversionState.ObjectMeta, inputs)
			objectData, _ := json.Marshal(catalogversionState)
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): materialize catalogversion %v to namespace %s", catalogversionState.ObjectMeta.Name, catalogversionState.ObjectMeta.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to materialize catalogversion %v to namespace %s", catalogversionState.ObjectMeta.Name, catalogversionState.ObjectMeta.Namespace)
			err = i.ApiClient.UpsertCatalogVersion(ctx, catalogversionState.ObjectMeta.Name, objectData, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to create catalogversion %s: %s", catalogversionState.ObjectMeta.Name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					materialize,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateCatalogVersionFromCatalogVersionFailed.String(),
				)
				return outputs, false, err
			}
			createdObjectList[catalogversion.ObjectMeta.Name] = true
		}
		// DO NOT REMOVE THIS COMMENT
		// gofail: var afterMaterializeOnce bool
	}
	if len(createdObjectList) < len(objects) {
		errorMessage := "failed to create all objects:"
		for _, catalogversion := range catalogversions {
			if _, ok := createdObjectList[catalogversion.ObjectMeta.Name]; !ok {
				errorMessage = fmt.Sprintf("%s %s", errorMessage, catalogversion.ObjectMeta.Name)
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
	// Wait for deployment to finish
	if i.Config.WaitForDeployment {
		outputs["failedDeployment"] = []api_utils.FailedDeployment{}
		timeout := time.After(i.Config.WaitTimeoutDuration)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
	ForLoop:
		for {
			select {
			case <-ticker.C:
				var failed []api_utils.FailedDeployment
				instanceList, failed = api_utils.FilterIncompleteDeploymentUsingSummary(ctx, &i.ApiClient, namespace, instanceList, true, i.Config.User, i.Config.Password)
				outputs["failedDeployment"] = append(outputs["failedDeployment"].([]api_utils.FailedDeployment), failed...)
				targetList, failed = api_utils.FilterIncompleteDeploymentUsingSummary(ctx, &i.ApiClient, namespace, targetList, false, i.Config.User, i.Config.Password)
				outputs["failedDeployment"] = append(outputs["failedDeployment"].([]api_utils.FailedDeployment), failed...)
				if len(instanceList) == 0 && len(targetList) == 0 {
					break ForLoop
				}
				mLog.InfofCtx(ctx, "  P (Materialize Processor): waiting for deployment to finish. Instance: %v, Target: %v", toObjectNameList(instanceList), toObjectNameList(targetList))
			case <-timeout:
				// Timeout, function was not called
				errorMessage := fmt.Sprintf("timeout waiting for deployment to finish. Instance: %v, Target: %v", toObjectNameList(instanceList), toObjectNameList(targetList))
				mLog.ErrorfCtx(ctx, "  P (Materialize Processor): %s", errorMessage)
				return outputs, false, v1alpha2.NewCOAError(nil, errorMessage, v1alpha2.InternalError)
			}
		}
		outputs["failedDeploymentCount"] = len(outputs["failedDeployment"].([]api_utils.FailedDeployment))
		mLog.InfofCtx(ctx, "  P (Materialize Processor): successfully waited for deployment to finish.")
	}
	return outputs, false, nil
}

func toObjectNameList(objects []api_utils.ObjectInfo) []string {
	ret := make([]string, len(objects))
	for i, object := range objects {
		ret[i] = object.Name
	}
	return ret
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

func checkCatalogVersion(catalogversion *model.CatalogVersionState) bool {
	if catalogversion.Spec == nil {
		return false
	}
	if catalogversion.Spec.CatalogType == "instance" || catalogversion.Spec.CatalogType == "solutionVersion" || catalogversion.Spec.CatalogType == "target" || catalogversion.Spec.CatalogType == "catalogVersion" || catalogversion.Spec.CatalogType == "config" {
		return true
	}
	return false
}
