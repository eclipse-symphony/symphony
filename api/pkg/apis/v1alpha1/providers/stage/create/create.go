/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package create

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	loggerName   = "providers.stage.create"
	providerName = "P (Create Stage)"
	create       = "create"
)

var (
	msLock                   sync.Mutex
	mLog                     = logger.NewLogger(loggerName)
	once                     sync.Once
	providerOperationMetrics *metrics.Metrics
	label_key                = os.Getenv("LABEL_KEY")
	label_value              = os.Getenv("LABEL_VALUE")
	annotation_name          = os.Getenv("ANNOTATION_KEY")
)

type CreateStageProviderConfig struct {
	User         string `json:"user"`
	Password     string `json:"password"`
	WaitCount    int    `json:"wait.count,omitempty"`
	WaitInterval int    `json:"wait.interval,omitempty"`
}

type CreateStageProvider struct {
	Config    CreateStageProviderConfig
	Context   *contexts.ManagerContext
	ApiClient api_utils.ApiClient
}

const (
	RemoveAction = "remove"
	CreateAction = "create"
)

func (s *CreateStageProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("[Stage] Create Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	msLock.Lock()
	defer msLock.Unlock()
	var mockConfig CreateStageProviderConfig
	mockConfig, err = toSymphonyStageProviderConfig(config)
	if err != nil {
		return err
	}
	if mockConfig.WaitInterval == 0 {
		mockConfig.WaitInterval = 20
	}

	if mockConfig.WaitCount == 0 {
		mockConfig.WaitCount = 31 * 60 / mockConfig.WaitInterval
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
				mLog.ErrorfCtx(ctx, "  P (Create Stage): failed to create metrics: %+v", err)
			}
		}
	})
	return err
}
func (s *CreateStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toSymphonyStageProviderConfig(config providers.IProviderConfig) (CreateStageProviderConfig, error) {
	ret := CreateStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = utils.UnmarshalJson(data, &ret)
	return ret, err
}
func (i *CreateStageProvider) InitWithMap(properties map[string]string) error {
	config, err := SymphonyStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func SymphonyStageProviderConfigFromMap(properties map[string]string) (CreateStageProviderConfig, error) {
	ret := CreateStageProviderConfig{}
	if api_utils.ShouldUseUserCreds() {
		user, err := api_utils.GetString(properties, "user")
		if err != nil {
			return ret, err
		}
		ret.User = user
		if ret.User == "" && !api_utils.ShouldUseSATokens() {
			return ret, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		}
		password, err := api_utils.GetString(properties, "password")
		ret.Password = password
		if err != nil {
			return ret, err
		}
	}
	waitStr, err := api_utils.GetString(properties, "wait.count")
	if err != nil {
		return ret, err
	}
	waitCount, err := strconv.Atoi(waitStr)
	if err != nil {
		return ret, v1alpha2.NewCOAError(err, "wait.count must be an integer", v1alpha2.BadConfig)
	}
	ret.WaitCount = waitCount
	waitStr, err = api_utils.GetString(properties, "wait.interval")
	if err != nil {
		return ret, err
	}
	waitInterval, err := strconv.Atoi(waitStr)
	if err != nil {
		return ret, v1alpha2.NewCOAError(err, "wait.interval must be an integer", v1alpha2.BadConfig)
	}
	ret.WaitInterval = waitInterval
	if waitCount <= 0 {
		waitCount = 1
	}
	return ret, nil
}
func (i *CreateStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Create provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	mLog.InfofCtx(ctx, "  P (Create Stage) process started")
	processTime := time.Now().UTC()
	functionName := observ_utils.GetFunctionName()
	defer providerOperationMetrics.ProviderOperationLatency(
		processTime,
		create,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)

	outputs := make(map[string]interface{})
	objectType := stage.ReadInputString(inputs, "objectType")
	objectName := stage.ReadInputString(inputs, "objectName")
	action := stage.ReadInputString(inputs, "action")
	object := inputs["object"]
	var objectData []byte
	if object != nil {
		objectData, _ = json.Marshal(object)
	}
	objectName = api_utils.ConvertReferenceToObjectName(objectName)
	objectNamespace := stage.GetNamespace(inputs)
	if objectNamespace == "" {
		objectNamespace = "default"
	}
	switch objectType {
	case "solution":
		if strings.EqualFold(action, RemoveAction) {
			solutionName := api_utils.ConvertReferenceToObjectName(objectName)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Create Stage): Start to delete solution name %s namespace %s", solutionName, objectNamespace)
			err = i.ApiClient.DeleteSolution(ctx, solutionName, objectNamespace, i.Config.User, i.Config.Password)
			if err != nil {
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.DeleteSolutionFailed.String(),
				)
				mLog.ErrorfCtx(ctx, "  P (Create Stage) process failed, failed to delete solution: %+v", err)
				return nil, false, err
			}
			outputs["objectType"] = objectType
			outputs["objectName"] = objectName
			return outputs, false, nil
		} else if strings.EqualFold(action, CreateAction) {
			objectName := stage.ReadInputString(inputs, "objectName")
			solutionName := api_utils.ConvertReferenceToObjectName(objectName)
			var solutionState model.SolutionState
			err = utils2.UnmarshalJson(objectData, &solutionState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal solution state for input %s: %s", objectName, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidSolutionCatalog.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid embeded solution in inputs %s", objectName), v1alpha2.BadRequest)
			}

			solutionState.ObjectMeta.Namespace = objectNamespace
			solutionState.ObjectMeta.Name = solutionName
			parts := strings.Split(objectName, constants.ReferenceSeparator)
			if len(parts) == 2 {
				solutionState.Spec.RootResource = parts[0]
				solutionState.Spec.Version = parts[1]
			} else {
				mLog.ErrorfCtx(ctx, "Solution name is invalid: solution - %s.", objectName)
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InvalidSolutionCatalog.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid solution name: %s", objectName), v1alpha2.BadRequest)
			}

			if label_key != "" && label_value != "" {
				// Check if labels exists within metadata, if not create it
				labels := solutionState.ObjectMeta.Labels
				if labels == nil {
					labels = make(map[string]string)
					solutionState.ObjectMeta.Labels = labels
				}
				// Add the label
				labels[label_key] = label_value
			}
			if annotation_name != "" {
				solutionState.ObjectMeta.UpdateAnnotation(annotation_name, parts[1])
			}
			mLog.DebugfCtx(ctx, "  P (Create Processor): check solution contains %v, namespace %s", solutionState.Spec.RootResource, objectNamespace)
			_, err := i.ApiClient.GetSolutionContainer(ctx, solutionState.Spec.RootResource, objectNamespace, i.Config.User, i.Config.Password)
			if err != nil && api_utils.IsNotFound(err) {
				mLog.DebugfCtx(ctx, "Solution container %s doesn't exist: %s", solutionState.Spec.RootResource, err.Error())
				solutionContainerState := model.SolutionContainerState{ObjectMeta: model.ObjectMeta{Name: solutionState.Spec.RootResource, Namespace: objectNamespace, Labels: solutionState.ObjectMeta.Labels}}

				// Set the owner reference
				target := stage.ReadInputString(inputs, "__target")
				target = api_utils.GetInstanceTargetName(target)
				ownerReference, err := api_utils.GetSolutionContainerOwnerReferences(i.ApiClient, ctx, target, objectNamespace, i.Config.User, i.Config.Password)
				if err != nil {
					mLog.ErrorfCtx(ctx, "Failed to get owner reference for solution %s: %s", objectName, err.Error())
					providerOperationMetrics.ProviderOperationErrors(
						create,
						functionName,
						metrics.ProcessOperation,
						metrics.RunOperationType,
						v1alpha2.CreateSolutionFailed.String(),
					)
					return outputs, false, err
				}
				if ownerReference != nil {
					solutionContainerState.ObjectMeta.OwnerReferences = ownerReference
				}
				containerObjectData, _ := json.Marshal(solutionContainerState)
				observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to create solution container %v in namespace %s", solutionState.Spec.RootResource, objectNamespace)
				err = i.ApiClient.CreateSolutionContainer(ctx, solutionState.Spec.RootResource, containerObjectData, objectNamespace, i.Config.User, i.Config.Password)
				if err != nil {
					mLog.ErrorfCtx(ctx, "Failed to create solution container %s: %s", solutionState.Spec.RootResource, err.Error())
					providerOperationMetrics.ProviderOperationErrors(
						create,
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
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.ParentObjectMissing.String(),
				)
				return outputs, false, err
			}

			objectData, _ := json.Marshal(solutionState)
			mLog.DebugfCtx(ctx, "  P (Materialize Processor): materialize solution %v to namespace %s", solutionState.ObjectMeta.Name, solutionState.ObjectMeta.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Materialize Processor): Start to materialize solution %v to namespace %s", solutionState.ObjectMeta.Name, solutionState.ObjectMeta.Namespace)
			err = i.ApiClient.UpsertSolution(ctx, solutionState.ObjectMeta.Name, objectData, solutionState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to create solution %s: %s", solutionState.ObjectMeta.Name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateSolutionFromCatalogFailed.String(),
				)
				return outputs, false, err
			}
			outputs["objectType"] = objectType
			outputs["objectName"] = solutionName
			return outputs, false, nil
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Unsupported action: %s", action), v1alpha2.BadRequest)
			providerOperationMetrics.ProviderOperationErrors(
				create,
				functionName,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				v1alpha2.UnsupportedAction.String(),
			)
			mLog.ErrorfCtx(ctx, "  P (Create Stage) process failed, error: %+v", err)
			return nil, false, err
		}
	case "instance":
		if strings.EqualFold(action, RemoveAction) {
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Create Stage): Start to delete instance name %s namespace %s", objectName, objectNamespace)
			err = i.ApiClient.DeleteInstance(ctx, objectName, objectNamespace, i.Config.User, i.Config.Password)
			if err != nil {
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.DeleteInstanceFailed.String(),
				)
				mLog.ErrorfCtx(ctx, "  P (Create Stage) process failed, failed to delete instance: %+v", err)
				return nil, false, err
			}
			for ic := 0; ic < i.Config.WaitCount; ic++ {
				remainings := api_utils.FilterIncompleteDelete(ctx, &i.ApiClient, objectNamespace, []string{objectName}, true, i.Config.User, i.Config.Password)
				if len(remainings) == 0 {
					return outputs, false, nil
				}
				mLog.InfofCtx(ctx, "  P (Create Stage) process: Waiting for instance deletion: %+v", remainings)
				time.Sleep(time.Duration(i.Config.WaitInterval) * time.Second)
			}
			providerOperationMetrics.ProviderOperationErrors(
				create,
				functionName,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				v1alpha2.DeploymentNotReached.String(),
			)
			err = v1alpha2.NewCOAError(nil, "Instance deletion reconcile timeout", v1alpha2.InternalError)
			mLog.ErrorfCtx(ctx, "  P (Create Stage) process failed, error: %+v", err)
			return nil, false, err
		} else if strings.EqualFold(action, CreateAction) {
			var instanceState model.InstanceState
			err = utils.UnmarshalJson(objectData, &instanceState)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to unmarshal instance state %s: %s", objectName, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateInstanceFailed.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid objectdata: %v", objectData), v1alpha2.BadRequest)
			}

			target := stage.ReadInputString(inputs, "__target")
			if target != "" {
				if instanceState.Spec == nil {
					instanceState.Spec = &model.InstanceSpec{}
				}
				instanceState.Spec.Target = model.TargetSelector{
					Name: target,
				}
			}

			// Get the solution and container name
			solutionContainer, solutionVersion := api_utils.GetSolutionAndContainerName(instanceState.Spec.Solution)
			if solutionContainer == "" || solutionVersion == "" {
				mLog.ErrorfCtx(ctx, "Invalid solution name: instance - %s", objectName)
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateInstanceFailed.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid solution name: instance - %s", objectName), v1alpha2.BadRequest)
			}
			// Get instanceName
			instanceName := api_utils.GetInstanceName(solutionContainer, objectName)

			// Set the owner reference
			ownerReference, err := api_utils.GetInstanceOwnerReferences(i.ApiClient, ctx, solutionContainer, objectNamespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to get owner reference for instance %s: %s", instanceName, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateInstanceFailed.String(),
				)
				return outputs, false, err
			}
			if ownerReference != nil {
				instanceState.ObjectMeta.OwnerReferences = ownerReference
			}

			// Set the labels
			if label_key != "" && label_value != "" {
				// Check if labels exists within metadata, if not create it
				labels := instanceState.ObjectMeta.Labels
				if labels == nil {
					labels = make(map[string]string)
					instanceState.ObjectMeta.Labels = labels
				}
				// Add the label
				labels[label_key] = label_value
			}
			if annotation_name != "" {
				instanceState.ObjectMeta.UpdateAnnotation(annotation_name, objectName)
			}

			// Add annotation for Private.Edge provider resources
			anno := api_utils.GenerateSystemDataAnnotations(ctx, instanceState.ObjectMeta.Annotations, target)
			instanceState.ObjectMeta.Annotations = anno

			// TODO: azure build flag
			// TODO: also update in materialize stage provider
			operationIdKey := api_utils.GenerateOperationId()
			if operationIdKey != "" {
				if instanceState.ObjectMeta.Annotations == nil {
					instanceState.ObjectMeta.Annotations = make(map[string]string)
				}
				instanceState.ObjectMeta.Annotations[constants.AzureOperationIdKey] = operationIdKey
				mLog.InfofCtx(ctx, "  P (Create Stage): update %s annotation: %s to %s", instanceName, constants.AzureOperationIdKey, instanceState.ObjectMeta.Annotations[constants.AzureOperationIdKey])
			}
			instanceState.ObjectMeta.Namespace = objectNamespace
			instanceState.ObjectMeta.Name = instanceName

			if instanceState.ObjectMeta.Name == "" {
				mLog.ErrorfCtx(ctx, "Instance name is empty: - %s", instanceName)
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateInstanceFailed.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty instance name: - %s", instanceName), v1alpha2.BadRequest)
			}

			// get instance first to get the previous jobid before update
			previousJobId := "-1"
			ret, err := i.ApiClient.GetInstance(ctx, instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil && !api_utils.IsNotFound(err) {
				mLog.ErrorfCtx(ctx, "Failed to get instance %s: %s", instanceState.ObjectMeta.Name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					create,
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

			objectData, _ := json.Marshal(instanceState)
			mLog.InfofCtx(ctx, "  P (Create Stage): creating instance %s with state: %s", instanceName, objectData)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Create Stage): Start to create instance name %s namespace %s", instanceName, objectNamespace)
			err = i.ApiClient.CreateInstance(ctx, instanceName, objectData, objectNamespace, i.Config.User, i.Config.Password)
			if err != nil {
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateInstanceFailed.String(),
				)
				mLog.ErrorfCtx(ctx, "  P (Create Stage) process failed, failed to create instance: %+v", err)
				if apiError, ok := err.(api_utils.APIError); ok {
					mLog.InfofCtx(ctx, "  P (Create Stage): This is an webhook error with status: %d, message: %v", apiError.Code, apiError)
				}

				return nil, false, err
			}

			// check guid after instance created
			ret, err = i.ApiClient.GetInstance(ctx, instanceState.ObjectMeta.Name, instanceState.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				mLog.ErrorfCtx(ctx, "Failed to get instance %s: %s", instanceState.ObjectMeta.Name, err.Error())
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.InstanceGetFailed.String(),
				)
				return outputs, false, err
			}
			summaryId := ret.ObjectMeta.GetSummaryId()
			if summaryId == "" {
				mLog.ErrorfCtx(ctx, "Instance GUID is empty: - %s", instanceName)
				providerOperationMetrics.ProviderOperationErrors(
					create,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CreateInstanceFailed.String(),
				)
				return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("Empty instance guid: - %s", instanceName), v1alpha2.BadRequest)
			}
			var remaining []api_utils.ObjectInfo
			var failed []api_utils.FailedDeployment
			for ic := 0; ic < i.Config.WaitCount; ic++ {
				obj := api_utils.ObjectInfo{
					Name:         instanceName,
					SummaryId:    summaryId,
					SummaryJobId: previousJobId,
				}
				remaining, failed = api_utils.FilterIncompleteDeploymentUsingSummary(ctx, &i.ApiClient, objectNamespace, []api_utils.ObjectInfo{obj}, true, i.Config.User, i.Config.Password)
				if len(remaining) == 0 {
					outputs["objectType"] = objectType
					outputs["objectName"] = instanceName
					mLog.InfofCtx(ctx, "  P (Create Stage) process completed with fail count is %d", len(failed))

					outputs["failedDeploymentCount"] = len(failed)
					outputs["status"] = 200
					if len(failed) > 0 {
						outputs["error"] = failed[0].Message
						outputs["status"] = 400
					}
					return outputs, false, nil
				}
				time.Sleep(time.Duration(i.Config.WaitInterval) * time.Second)
			}
			providerOperationMetrics.ProviderOperationErrors(
				create,
				functionName,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				v1alpha2.DeploymentNotReached.String(),
			)
			mLog.ErrorfCtx(ctx, "  P (Create Stage) Instance creation reconcile timeout.")
			outputs["objectType"] = objectType
			outputs["objectName"] = instanceName
			outputs["failedDeploymentCount"] = len(remaining)
			outputs["error"] = fmt.Sprintf("Instance creation reconcile timeout after %d seconds", i.Config.WaitCount*i.Config.WaitInterval)
			outputs["status"] = 400
			return outputs, false, nil
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Unsupported action: %s", action), v1alpha2.BadRequest)
			providerOperationMetrics.ProviderOperationErrors(
				create,
				functionName,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				v1alpha2.UnsupportedAction.String(),
			)
			mLog.ErrorfCtx(ctx, "  P (Create Stage) process failed, error: %+v", err)
			return nil, false, err
		}
	default:
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Unsupported object type: %s", objectType), v1alpha2.BadRequest)
		mLog.ErrorfCtx(ctx, "  P (Create Stage) process failed, error: %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			create,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.InvalidObjectType.String(),
		)
		return nil, false, err
	}
}
