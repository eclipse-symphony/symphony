/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package stage

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	symproviders "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/remote"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type StageManager struct {
	managers.Manager
	StateProvider states.IStateProvider
	apiClient     utils.ApiClient
}

type StageResult struct {
	Outputs map[string]interface{}
	Site    string
	Error   error
}

type StageTaskResult struct {
	Outputs  map[string]interface{}
	Error    error
	TaskName string
}

type StageTaskProcessor struct {
	TaskResults map[string]interface{}
	TaskErrors  []error
	ErrorCount  int
}

// TaskHandler defines the interface for handling individual tasks
type TaskHandler interface {
	HandleTask(ctx context.Context, task model.TaskSpec, inputs map[string]interface{}, siteName string) (map[string]interface{}, error)
}

// CampaignTaskHandler implements TaskHandler for campaign-specific task processing
type CampaignTaskHandler struct {
	manager     *StageManager
	triggerData v1alpha2.ActivationData
	triggers    map[string]interface{}
}

func NewCampaignTaskHandler(manager *StageManager, triggerData v1alpha2.ActivationData, triggers map[string]interface{}) *CampaignTaskHandler {
	return &CampaignTaskHandler{
		manager:     manager,
		triggerData: triggerData,
		triggers:    triggers,
	}
}

func structToMap(v any) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func mapToStruct(m map[string]interface{}, dst any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func (h *CampaignTaskHandler) HandleTask(ctx context.Context, task model.TaskSpec, inputs map[string]interface{}, siteName string) (map[string]interface{}, error) {
	// Create task-specific inputs
	taskInputs := utils.MergeCollection_StringAny(inputs, task.Inputs)

	// Re-populate __target from task.Target
	if task.Target != "" {
		taskInputs["__target"] = task.Target
	}
	// Trace value on task inputs
	for k, v := range taskInputs {
		var val interface{}
		val, err := h.manager.traceValue(ctx, v, h.triggerData.Namespace, taskInputs, h.triggers, h.triggerData.Outputs)
		if err != nil {
			return nil, err
		}
		taskInputs[k] = val
	}

	if task.Config != nil {
		configMap, err := structToMap(task.Config)
		if err != nil {
			return nil, err
		}
		for k, v := range configMap {
			var val interface{}
			val, err := h.manager.traceValue(ctx, v, h.triggerData.Namespace, taskInputs, h.triggers, h.triggerData.Outputs)
			if err != nil {
				return nil, err
			}
			configMap[k] = val
		}
		err = mapToStruct(configMap, &task.Config)
		if err != nil {
			return nil, err
		}
	}

	// Create task provider
	factory := symproviders.SymphonyProviderFactory{}
	taskProvider, err := factory.CreateProvider(task.Provider, task.Config)
	if err != nil {
		return nil, err
	}

	if taskProvider == nil {
		err = v1alpha2.COAError{
			State:   v1alpha2.BadConfig,
			Message: fmt.Sprintf("task provider %s is not found, skipping task %s for site %s", task.Provider, task.Name, siteName),
		}
		return nil, err
	}

	// Non-stage provider is not allowed in tasks
	if _, ok := taskProvider.(stage.IStageProvider); !ok {
		err = v1alpha2.COAError{
			State:   v1alpha2.BadConfig,
			Message: fmt.Sprintf("non-stage provider cannot be used with tasks, skipping task %s for site %s", task.Name, siteName),
		}
		return nil, err
	}

	// Remote provider is not allowed in tasks
	if _, ok := taskProvider.(*remote.RemoteStageProvider); ok {
		err = v1alpha2.COAError{
			State:   v1alpha2.BadConfig,
			Message: fmt.Sprintf("remote stage provider cannot be used with tasks, skipping task %s for site %s", task.Name, siteName),
		}
		return nil, err
	}

	if _, ok := taskProvider.(contexts.IWithManagerContext); ok {
		taskProvider.(contexts.IWithManagerContext).SetContext(h.manager.Manager.Context)
	}

	// Process task
	outputs, _, err := taskProvider.(stage.IStageProvider).Process(ctx, *h.manager.Manager.Context, taskInputs)
	if err != nil {
		return nil, err
	}

	return outputs, nil
}

// TaskProcessor defines the interface for processing tasks
type TaskProcessor interface {
	Process(ctx context.Context, tasks []model.TaskSpec, inputs map[string]interface{}, handler TaskHandler, errorAction model.ErrorAction, concurrency int) (map[string]interface{}, error)
}

// GoRoutineTaskProcessor implements TaskProcessor using goroutines
type GoRoutineTaskProcessor struct {
	manager *StageManager
	ctx     context.Context
}

func NewGoRoutineTaskProcessor(manager *StageManager, ctx context.Context) *GoRoutineTaskProcessor {
	return &GoRoutineTaskProcessor{
		manager: manager,
		ctx:     ctx,
	}
}

func shouldStopDispatch(errorAction model.ErrorAction, errorCount int) bool {
	switch errorAction.Mode {
	case model.ErrorActionMode_StopOnAnyFailure:
		return errorCount > 0
	case model.ErrorActionMode_StopOnNFailures:
		return errorCount > errorAction.MaxToleratedFailures
	default:
		return false
	}
}

func (p *GoRoutineTaskProcessor) Process(ctx context.Context, tasks []model.TaskSpec, inputs map[string]interface{}, handler TaskHandler, errorAction model.ErrorAction, concurrency int, siteName string) (map[string]interface{}, error) {
	// after introducing parallel workflow, we will have two phases to execute the stage.
	// 1. if stage.Provider is defined, that means we have a prepare stage to be executed before all tasks.
	// 2. if stage.tasks is defined, that means we have several tasks to be executed in parallel, concurrency will be controlled by taskOption.concurrency
	//    each task will be executed in parallel and the results will be collected in a channel.
	//    for each stage.task, we will create a goroutine to execute the task.
	//        input will be stage.input + task.input
	//		  output will be aggregated to stage.output.task.name
	//    stage.status will be determined by prepare-stage and tasks's status + taskOption.errorAction
	// TODO: remote stage provider cannot be worked with parallel tasks, so we will not support remote stage provider for now.
	// TODO: we cannot run remote stage provider in tasks.
	// TODO: In future, we will combine two parallel ways - remote stage provider (cross-cluster) and tasks (in-cluster).
	// TODO: provide ways to parse task.Input accroding to prepare provider's output.

	if len(tasks) == 0 {
		return make(map[string]interface{}), nil
	}

	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	taskQueue := make(chan model.TaskSpec)
	taskResultsChan := make(chan StageTaskResult)

	// Track workers and result processor
	taskWaitGroup := sync.WaitGroup{}
	done := make(chan struct{})

	// State
	taskProcessor := &StageTaskProcessor{
		TaskResults: make(map[string]interface{}),
		TaskErrors:  []error{},
		ErrorCount:  0,
	}

	// Start worker pool
	for i := 0; i < concurrency; i++ {
		taskWaitGroup.Add(1)
		go func() {
			defer taskWaitGroup.Done()
			for task := range taskQueue {
				outputs, err := handler.HandleTask(taskCtx, task, inputs, siteName)
				if err != nil {
					outputs = carryOutPutsToErrorStatus(outputs, err, "")
				}
				select {
				case taskResultsChan <- StageTaskResult{
					TaskName: task.Name,
					Outputs:  outputs,
					Error:    err,
				}:
				case <-taskCtx.Done():
					return
				}
			}
		}()
	}

	// Result monitor and task dispatcher
	go func() {
		defer close(done)

		taskIndex := 0
		pendingTasks := 0

		// Initially dispatch up to `concurrency` tasks
		for i := 0; i < concurrency && i < len(tasks); i++ {
			taskQueue <- tasks[taskIndex]
			taskIndex++
			pendingTasks++
		}

		for pendingTasks > 0 {
			select {
			case result := <-taskResultsChan:
				pendingTasks--

				// Track result
				if result.Error != nil {
					taskProcessor.TaskErrors = append(taskProcessor.TaskErrors, result.Error)
					taskProcessor.ErrorCount++
				}
				taskProcessor.TaskResults[result.TaskName] = result.Outputs

				// Check if we're allowed to dispatch more
				if shouldStopDispatch(errorAction, taskProcessor.ErrorCount) {
					continue
				}

				// Dispatch next task if available
				if taskIndex < len(tasks) {
					taskQueue <- tasks[taskIndex]
					taskIndex++
					pendingTasks++
				}

			case <-taskCtx.Done():
				return
			}
		}
	}()

	// Wait for result monitor to finish dispatching
	<-done

	// No more tasks will be sent
	close(taskQueue)

	// Now wait for workers to finish
	taskWaitGroup.Wait()

	// Close the results channel
	close(taskResultsChan)

	// Final error decision
	if taskProcessor.ErrorCount > 0 {
		switch errorAction.Mode {
		case model.ErrorActionMode_StopOnAnyFailure:
			log.WarnfCtx(ctx, " M (Stage): task errors: %s", utils.ToJsonString(taskProcessor.TaskErrors))
			return taskProcessor.TaskResults, v1alpha2.COAError{
				State:   v1alpha2.InternalError,
				Message: fmt.Sprintf("Task failed %d times, error action is stop on any failure", taskProcessor.ErrorCount),
			}
		case model.ErrorActionMode_StopOnNFailures:
			log.WarnfCtx(ctx, " M (Stage): task errors: %s", utils.ToJsonString(taskProcessor.TaskErrors))
			if taskProcessor.ErrorCount > errorAction.MaxToleratedFailures {
				return taskProcessor.TaskResults, v1alpha2.COAError{
					State:   v1alpha2.InternalError,
					Message: fmt.Sprintf("Task failed %d times, exceeded maximum tolerated failures (%d)", taskProcessor.ErrorCount, errorAction.MaxToleratedFailures),
				}
			}
		default:
			log.WarnfCtx(ctx, " M (Stage): unknown error mode, continuing")
			log.WarnfCtx(ctx, " M (Stage): task errors: %s", utils.ToJsonString(taskProcessor.TaskErrors))
		}
	}

	return taskProcessor.TaskResults, nil
}

func (s *StageManager) processTasks(ctx context.Context, currentStage model.StageSpec, inputCopy map[string]interface{}, triggerData v1alpha2.ActivationData, triggers map[string]interface{}, siteName string) (map[string]interface{}, error) {
	if len(currentStage.Tasks) == 0 {
		return make(map[string]interface{}), nil
	}

	log.InfofCtx(ctx, " M (Stage): processing tasks for site %s", inputCopy["__site"])

	// Create task processor and handler
	processor := NewGoRoutineTaskProcessor(s, ctx)
	handler := NewCampaignTaskHandler(s, triggerData, triggers)

	// Process tasks using the processor
	return processor.Process(ctx, currentStage.Tasks, inputCopy, handler, currentStage.TaskOption.ErrorAction, currentStage.TaskOption.Concurrency, siteName)
}

func (t *StageResult) GetError() error {
	if t.Error != nil {
		return t.Error
	}
	if v, ok := t.Outputs["status"]; ok {
		switch sv := v.(type) {
		case v1alpha2.State:
			break
		case float64:
			state := v1alpha2.State(int(sv))
			stateValue := reflect.ValueOf(state)
			if stateValue.Type() != reflect.TypeOf(v1alpha2.State(0)) {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %v", sv), v1alpha2.InternalError)
			}
			t.Outputs["status"] = state
		case int:
			state := v1alpha2.State(sv)
			stateValue := reflect.ValueOf(state)
			if stateValue.Type() != reflect.TypeOf(v1alpha2.State(0)) {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %d", sv), v1alpha2.InternalError)
			}
			t.Outputs["status"] = state
		case string:
			vInt, err := strconv.ParseInt(sv, 10, 32)
			if err != nil {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %s", sv), v1alpha2.InternalError)
			}
			state := v1alpha2.State(vInt)
			stateValue := reflect.ValueOf(state)
			if stateValue.Type() != reflect.TypeOf(v1alpha2.State(0)) {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %d", vInt), v1alpha2.InternalError)
			}
			t.Outputs["status"] = state
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %v", v), v1alpha2.InternalError)
		}

		if t.Outputs["status"] != v1alpha2.OK {
			if v, ok := t.Outputs["error"]; ok {
				return v1alpha2.NewCOAError(nil, utils.FormatAsString(v), t.Outputs["status"].(v1alpha2.State))
			} else {
				return v1alpha2.NewCOAError(nil, "stage returned unsuccessful status without an error", v1alpha2.InternalError)
			}
		}
	}
	return nil
}

type PendingTask struct {
	Sites         []string                          `json:"sites"`
	OutputContext map[string]map[string]interface{} `json:"outputContext,omitempty"`
}

func (s *StageManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	// TODO: change volatileStateProvider to persistentStateProvider
	stateprovider, err := managers.GetVolatileStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	s.apiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
	return nil
}
func (s *StageManager) Enabled() bool {
	return s.Config.Properties["poll.enabled"] == "true"
}
func (s *StageManager) Poll() []error {
	return nil
}
func (s *StageManager) Reconcil() []error {
	return nil
}
func (s *StageManager) ResumeStage(ctx context.Context, status model.StageStatus, cam model.CampaignSpec) (*v1alpha2.ActivationData, error) {
	log.InfofCtx(ctx, " M (Stage): ResumeStage: %v\n", status)
	campaign, ok := status.Outputs["__campaign"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: campaign (%v) is not valid from output", status.Outputs["__campaign"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: campaign is not valid", v1alpha2.BadRequest)
	}
	activation, ok := status.Outputs["__activation"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: activation (%v) is not valid from output", status.Outputs["__activation"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: activation is not valid", v1alpha2.BadRequest)
	}
	activationGeneration, ok := status.Outputs["__activationGeneration"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: activationGeneration (%v) is not valid from output", status.Outputs["__activationGeneration"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: activationGeneration is not valid", v1alpha2.BadRequest)
	}
	site, ok := status.Outputs["__site"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: site (%v) is not valid from output", status.Outputs["__site"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: site is not valid", v1alpha2.BadRequest)
	}
	stage, ok := status.Outputs["__stage"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: stage (%v) is not valid from output", status.Outputs["__stage"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: stage is not valid", v1alpha2.BadRequest)
	}
	namespace, ok := status.Outputs["__namespace"].(string)
	if !ok {
		namespace = "default"
	}

	entry, err := s.StateProvider.Get(ctx, states.GetRequest{
		ID: fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
		Metadata: map[string]interface{}{
			"namespace": namespace,
		},
	})
	if err != nil {
		return nil, err
	}
	jData, _ := json.Marshal(entry.Body)
	var p PendingTask
	err = json.Unmarshal(jData, &p)
	if err == nil {
		// //find site in p.Sites
		// found := false
		// for _, s := range p.Sites {
		// 	if s == site {
		// 		found = true
		// 		break
		// 	}
		// }
		// if !found {
		// 	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("site %s is not found in pending task", site)
		// }
		//remove site from p.Sites
		newSites := make([]string, 0)
		for _, s := range p.Sites {
			if s != site {
				newSites = append(newSites, s)
			}
		}
		if len(newSites) == 0 {
			log.InfofCtx(ctx, " M (Stage): ResumeStage: all sites are done for activation %s stage %s. Check if we need to move to next stage", activation, stage)
			err := s.StateProvider.Delete(ctx, states.DeleteRequest{
				ID: fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
				Metadata: map[string]interface{}{
					"namespace": namespace,
				},
			})
			if err != nil {
				return nil, err
			}
			//find the next stage
			if cam.SelfDriving {
				outputs := p.OutputContext
				if outputs == nil {
					outputs = make(map[string]map[string]interface{})
				}
				outputs[stage] = status.Outputs
				nextStage := ""
				if currentStage, ok := cam.Stages[stage]; ok {
					parser := utils.NewParser(currentStage.StageSelector)
					eCtx := s.VendorContext.EvaluationContext.Clone()
					eCtx.Context = ctx
					eCtx.Namespace = namespace
					activationState, err := s.apiClient.GetActivation(
						ctx,
						activation,
						namespace,
						s.VendorContext.SiteInfo.CurrentSite.Username,
						s.VendorContext.SiteInfo.CurrentSite.Password,
					)
					if err == nil && activationState.Spec != nil {
						eCtx.Triggers = activationState.Spec.Inputs
					}
					eCtx.Inputs = status.Inputs
					log.DebugfCtx(ctx, " M (Stage): ResumeStage evaluation inputs: %v", eCtx.Inputs)
					if eCtx.Inputs != nil {
						if v, ok := eCtx.Inputs["context"]; ok {
							eCtx.Value = v
						}
					}
					eCtx.Outputs = outputs
					val, err := parser.Eval(*eCtx)
					if err != nil {
						return nil, err
					}

					sVal := ""
					if val != nil {
						sVal = utils.FormatAsString(val)
					}
					if sVal != "" {
						if _, ok := cam.Stages[sVal]; ok {
							nextStage = sVal
						} else {
							return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", sVal), v1alpha2.InternalError)
						}
					}

				}
				if nextStage != "" {
					activationData := &v1alpha2.ActivationData{
						Campaign:             campaign,
						Activation:           activation,
						ActivationGeneration: activationGeneration,
						Stage:                nextStage,
						Inputs:               status.Inputs,
						Provider:             cam.Stages[nextStage].Provider,
						Config:               cam.Stages[nextStage].Config,
						Outputs:              outputs,
						TriggeringStage:      stage,
						Schedule:             cam.Stages[nextStage].Schedule,
						Namespace:            namespace,
						Proxy:                cam.Stages[nextStage].Proxy,
					}
					log.InfofCtx(ctx, " M (Stage): Activating next stage: %s\n", activationData.Stage)
					return activationData, nil
				} else {
					log.InfoCtx(ctx, " M (Stage): No next stage found\n")
					return nil, nil
				}
			}
			return nil, nil
		} else {
			log.InfoCtx(ctx, " M (Stage): ResumeStage: updating pending sites %v for activation %s stage %s", newSites, activation, stage)
			p.Sites = newSites
			// TODO: clean up the remote job status entry for multi-site
			_, err := s.StateProvider.Upsert(ctx, states.UpsertRequest{
				Value: states.StateEntry{
					ID:   fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
					Body: p,
				},
				Metadata: map[string]interface{}{
					"namespace": namespace,
				},
			})
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, v1alpha2.NewCOAError(err, "invalid pending task", v1alpha2.InternalError)
	}

	return nil, nil
}
func (s *StageManager) HandleDirectTriggerEvent(ctx context.Context, triggerData v1alpha2.ActivationData) model.StageStatus {
	ctx, span := observability.StartSpan("Stage Manager", ctx, &map[string]string{
		"method": "HandleDirectTriggerEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, " M (Stage): HandleDirectTriggerEvent for campaign %s, activation %s, stage %s", triggerData.Campaign, triggerData.Activation, triggerData.Stage)

	status := model.StageStatus{
		Stage:     "",
		NextStage: "",
		Outputs: map[string]interface{}{
			"__campaign":             triggerData.Campaign,
			"__namespace":            triggerData.Namespace,
			"__activation":           triggerData.Activation,
			"__activationGeneration": triggerData.ActivationGeneration,
			"__stage":                triggerData.Stage,
			"__previousStage":        triggerData.TriggeringStage,
			"__site":                 s.VendorContext.SiteInfo.SiteId,
		},
		Status:        v1alpha2.Untouched,
		StatusMessage: v1alpha2.Untouched.String(),
		ErrorMessage:  "",
		IsActive:      true,
	}
	var provider providers.IProvider
	factory := symproviders.SymphonyProviderFactory{}
	provider, err = factory.CreateProvider(triggerData.Provider, triggerData.Config)
	if err != nil {
		status.Status = v1alpha2.InternalError
		status.StatusMessage = v1alpha2.InternalError.String()
		status.ErrorMessage = err.Error()
		status.IsActive = false
		log.ErrorfCtx(ctx, " M (Stage): failed to create provider: %v", err)
		return status
	}
	if provider == nil {
		status.Status = v1alpha2.BadRequest
		status.StatusMessage = v1alpha2.BadRequest.String()
		status.ErrorMessage = fmt.Sprintf("provider %s is not found", triggerData.Provider)
		status.IsActive = false
		log.ErrorfCtx(ctx, " M (Stage): failed to create provider: %v", err)
		return status
	}

	if _, ok := provider.(contexts.IWithManagerContext); ok {
		provider.(contexts.IWithManagerContext).SetContext(s.Manager.Context)
	} else {
		log.ErrorfCtx(ctx, " M (Stage): provider %s does not implement IWithManagerContext", triggerData.Provider)
	}

	isRemote := false
	if _, ok := provider.(*remote.RemoteStageProvider); ok {
		isRemote = true
		provider.(*remote.RemoteStageProvider).SetOutputsContext(triggerData.Outputs)
	}

	if triggerData.Schedule != "" && !isRemote {
		log.InfofCtx(ctx, " M (Stage): send schedule event and pause stage %s for site %s", triggerData.Stage, s.VendorContext.SiteInfo.SiteId)
		s.Context.Publish("schedule", v1alpha2.Event{
			Body:    triggerData,
			Context: ctx,
		})
		status.Outputs["status"] = v1alpha2.Paused
		status.Status = v1alpha2.Paused
		status.StatusMessage = v1alpha2.Paused.String()
		status.IsActive = false
		return status
	}

	var outputs map[string]interface{}
	if triggerData.Proxy != nil {
		proxyProvider, err := factory.CreateProvider(triggerData.Proxy.Provider, nil)
		if err != nil {
			status.Status = v1alpha2.InternalError
			status.ErrorMessage = err.Error()
			status.IsActive = false
			return status
		}
		if _, ok := proxyProvider.(contexts.IWithManagerContext); ok {
			proxyProvider.(contexts.IWithManagerContext).SetContext(s.Manager.Context)
		}
		outputs, _, err = proxyProvider.(stage.IProxyStageProvider).Process(ctx, *s.Manager.Context, triggerData)
	} else {
		outputs, _, err = provider.(stage.IStageProvider).Process(ctx, *s.Manager.Context, triggerData.Inputs)
	}

	result := StageResult{
		Outputs: outputs,
		Error:   err,
		Site:    s.VendorContext.SiteInfo.SiteId,
	}

	err = result.GetError()
	if err != nil {
		status.Status = v1alpha2.InternalError
		status.StatusMessage = v1alpha2.InternalError.String()
		status.ErrorMessage = err.Error()
		status.IsActive = false
		status.Outputs = carryOutPutsToErrorStatus(outputs, err, "")
		result.Outputs = carryOutPutsToErrorStatus(outputs, err, "")
		return status
	}

	// Merge outputs, provider output overwrite status.Outputs for the same key
	for k, v := range outputs {
		status.Outputs[k] = v
	}

	status.Outputs["status"] = v1alpha2.OK
	status.Status = v1alpha2.Done
	status.StatusMessage = v1alpha2.Done.String()
	status.IsActive = false

	return status
}
func carryOutPutsToErrorStatus(outputs map[string]interface{}, err error, siteOrTarget string) map[string]interface{} {
	ret := make(map[string]interface{})
	statusKey := "status"
	if siteOrTarget != "" {
		statusKey = fmt.Sprintf("%s.%s", statusKey, siteOrTarget)
	}
	errorKey := "error"
	if siteOrTarget != "" {
		errorKey = fmt.Sprintf("%s.%s", errorKey, siteOrTarget)
	}
	for k, v := range outputs {
		ret[k] = v
	}
	// always override status and error
	if cErr, ok := err.(v1alpha2.COAError); ok {
		ret[statusKey] = cErr.State
	} else if apiError, ok := err.(utils.APIError); ok {
		ret[statusKey] = int(apiError.Code)
	} else {
		ret[statusKey] = v1alpha2.InternalError
	}

	ret[errorKey] = err.Error()
	return ret
}

func (s *StageManager) HandleTriggerEvent(ctx context.Context, campaign model.CampaignSpec, triggerData v1alpha2.ActivationData) (model.StageStatus, *v1alpha2.ActivationData) {
	ctx, span := observability.StartSpan("Stage Manager", ctx, &map[string]string{
		"method": "HandleTriggerEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, " M (Stage): HandleTriggerEvent for campaign %s, activation %s, stage %s, selfDriving %s", triggerData.Campaign, triggerData.Activation, triggerData.Stage, campaign.SelfDriving)
	status := model.StageStatus{
		Stage:         triggerData.Stage,
		NextStage:     "",
		Outputs:       map[string]interface{}{},
		Status:        v1alpha2.Untouched,
		StatusMessage: v1alpha2.Untouched.String(),
		ErrorMessage:  "",
		IsActive:      true,
	}
	var activationData *v1alpha2.ActivationData
	if currentStage, ok := campaign.Stages[triggerData.Stage]; ok {
		sites := make([]string, 0)
		// 1. According to campaign.Contexts, find out which sites will be executed
		if currentStage.Contexts != "" {
			log.InfofCtx(ctx, " M (Stage): evaluating context %s", currentStage.Contexts)
			parser := utils.NewParser(currentStage.Contexts)

			eCtx := s.VendorContext.EvaluationContext.Clone()
			eCtx.Context = ctx
			eCtx.Namespace = triggerData.Namespace
			eCtx.Triggers = triggerData.Inputs
			eCtx.Inputs = currentStage.Inputs
			if eCtx.Inputs != nil {
				if v, ok := eCtx.Inputs["context"]; ok {
					eCtx.Value = v
				}
			}
			eCtx.Outputs = triggerData.Outputs
			var val interface{}
			val, err = parser.Eval(*eCtx)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to evaluate context: %v", err)
				return status, activationData
			}
			if valStringList, ok := val.([]string); ok {
				sites = valStringList
			} else if valList, ok := val.([]interface{}); ok {
				for _, v := range valList {
					sites = append(sites, utils.FormatAsString(v))
				}
			} else if valString, ok := val.(string); ok {
				sites = append(sites, valString)
			} else {
				status.Status = v1alpha2.BadConfig
				status.StatusMessage = v1alpha2.BadConfig.String()
				status.ErrorMessage = fmt.Sprintf("invalid context %s", currentStage.Contexts)
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): invalid context: %v", currentStage.Contexts)
				return status, activationData
			}
			log.InfofCtx(ctx, " M (Stage): evaluated context %s to %v", currentStage.Contexts, sites)
		} else {
			sites = append(sites, s.VendorContext.SiteInfo.SiteId)
		}
		log.DebugfCtx(ctx, " M (Stage): HandleTriggerEvent for campaign %s, activation %s, stage %s, executed on sites {%s}", triggerData.Campaign, triggerData.Activation, triggerData.Stage, strings.Join(sites, ", "))

		// 2. According to triggerData.Inputs and currentStage.Inputs, together with default inputs to generate the runtime inputs
		triggers := triggerData.Inputs
		if triggers == nil {
			triggers = make(map[string]interface{})
		}
		inputs := currentStage.Inputs
		if inputs == nil {
			inputs = make(map[string]interface{})
		}

		// inject default inputs
		inputs["__campaign"] = triggerData.Campaign
		inputs["__namespace"] = triggerData.Namespace
		inputs["__activation"] = triggerData.Activation
		inputs["__stage"] = triggerData.Stage
		inputs["__activationGeneration"] = triggerData.ActivationGeneration
		inputs["__previousStage"] = triggerData.TriggeringStage
		inputs["__site"] = s.VendorContext.SiteInfo.SiteId
		if triggerData.Schedule != "" {
			inputs["__schedule"] = triggerData.Schedule
		}
		inputs["__target"] = currentStage.Target

		for k, v := range inputs {
			var val interface{}
			val, err = s.traceValue(ctx, v, triggerData.Namespace, inputs, triggers, triggerData.Outputs)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to evaluate input: %v", err)
				return status, activationData
			}
			inputs[k] = val
		}

		if triggerData.Outputs != nil {
			if v, ok := triggerData.Outputs[triggerData.Stage]; ok {
				if vs, ok := v["__state"]; ok {
					inputs["__state"] = vs
				}
			}
		}

		// 3. Snapshot the inputs from #2
		snapshotInputs := utils.DeepCopyCollection(inputs)

		// 4. Duplicate inputs from #2, then iterate campaign.task.inputs
		inputsWithStageAndTasks := utils.DeepCopyCollectionWithPrefixExclude(inputs, "__")
		if len(currentStage.Tasks) > 0 {
			for _, task := range currentStage.Tasks {
				for k, v := range task.Inputs {
					if task.Name == "" {
						log.ErrorfCtx(ctx, " M (Stage): task name is empty, cannot process inputs for task: %v", task)
						status.Status = v1alpha2.BadConfig
						status.StatusMessage = v1alpha2.BadConfig.String()
						status.ErrorMessage = fmt.Sprintf("task name is empty for task: %v", task)
						status.IsActive = false
						return status, activationData
					}
					if _, ok := inputsWithStageAndTasks[task.Name]; !ok {
						inputsWithStageAndTasks[task.Name] = make(map[string]interface{})
					}
					inputsWithStageAndTasks[task.Name].(map[string]interface{})[k] = v
				}
			}
		}
		status.Inputs = inputsWithStageAndTasks

		// 5. If campaign.provider exists, initialize a provider
		var provider providers.IProvider
		if triggerData.Provider != "" {

			if triggerData.Config != nil {
				//evaluate config expressions
				configMap, err := structToMap(triggerData.Config)
				if err != nil {
					status.Status = v1alpha2.InternalError
					status.StatusMessage = v1alpha2.InternalError.String()
					status.ErrorMessage = err.Error()
					status.IsActive = false
					log.ErrorfCtx(ctx, " M (Stage): failed to convert config to map: %v", err)
					return status, activationData
				}
				for k, v := range configMap {
					var val interface{}
					val, err = s.traceValue(ctx, v, triggerData.Namespace, snapshotInputs, triggers, triggerData.Outputs)
					if err != nil {
						status.Status = v1alpha2.InternalError
						status.StatusMessage = v1alpha2.InternalError.String()
						status.ErrorMessage = err.Error()
						status.IsActive = false
						log.ErrorfCtx(ctx, " M (Stage): failed to evaluate config: %v", err)
						return status, activationData
					}
					configMap[k] = val
				}
				err = mapToStruct(configMap, &triggerData.Config)
				if err != nil {
					status.Status = v1alpha2.InternalError
					status.StatusMessage = v1alpha2.InternalError.String()
					status.ErrorMessage = err.Error()
					status.IsActive = false
					log.ErrorfCtx(ctx, " M (Stage): failed to convert config map to struct: %v", err)
					return status, activationData
				}
			}
			factory := symproviders.SymphonyProviderFactory{}
			provider, err = factory.CreateProvider(triggerData.Provider, triggerData.Config)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to create provider: %v", err)
				return status, activationData
			}
			if provider == nil {
				status.Status = v1alpha2.BadRequest
				status.StatusMessage = v1alpha2.BadRequest.String()
				status.ErrorMessage = fmt.Sprintf("provider %s is not found", triggerData.Provider)
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to create provider: %v", err)
				return status, activationData
			}

			if _, ok := provider.(contexts.IWithManagerContext); ok {
				provider.(contexts.IWithManagerContext).SetContext(s.Manager.Context)
			} else {
				log.ErrorfCtx(ctx, " M (Stage): provider %s does not implement IWithManagerContext", triggerData.Provider)
			}
		}

		// 6. Iterate all sites, start multiple go routines
		numTasks := len(sites)
		waitGroup := sync.WaitGroup{}
		results := make(chan StageResult, numTasks)
		pauseRequested := false

		for _, site := range sites {
			waitGroup.Add(1)
			go func(wg *sync.WaitGroup, site string, results chan<- StageResult) {
				defer wg.Done()
				inputCopy := make(map[string]interface{})
				for k, v := range snapshotInputs {
					inputCopy[k] = v
				}
				inputCopy["__site"] = site

				for k, v := range inputCopy {
					var val interface{}
					val, err = s.traceValue(ctx, v, triggerData.Namespace, inputCopy, triggers, triggerData.Outputs)
					if err != nil {
						status.Status = v1alpha2.InternalError
						status.StatusMessage = v1alpha2.InternalError.String()
						status.ErrorMessage = err.Error()
						status.IsActive = false
						log.ErrorfCtx(ctx, " M (Stage): failed to evaluate input: %v", err)
						results <- StageResult{
							Outputs: nil,
							Error:   err,
							Site:    site,
						}
						return
					}
					inputCopy[k] = val
				}

				var remoteStageProviderDefined bool = false
				var allOutputs map[string]interface{} = make(map[string]interface{})

				if provider != nil {
					if remoteStageProvider, ok := provider.(*remote.RemoteStageProvider); ok {
						remoteStageProviderDefined = true
						remoteStageProvider.SetOutputsContext(triggerData.Outputs)
					}
				}

				if triggerData.Schedule != "" {
					log.InfofCtx(ctx, " M (Stage): send schedule event and pause stage %s for site %s", triggerData.Stage, site)
					s.Context.Publish("schedule", v1alpha2.Event{
						Body:    triggerData,
						Context: ctx,
					})
					pauseRequested = true
					results <- StageResult{
						Outputs: nil,
						Error:   nil,
						Site:    site,
					}
					return
				}

				// 6.1. If triggerData.provider exists, follow current flow to process, collect the output.
				if provider != nil {
					var outputs map[string]interface{}
					var pause bool
					var iErr error = nil
					if triggerData.Proxy != nil {
						factory := symproviders.SymphonyProviderFactory{}
						proxyProvider, pErr := factory.CreateProvider(triggerData.Proxy.Provider, nil)
						if err != nil {
							results <- StageResult{
								Outputs: nil,
								Error:   pErr,
								Site:    site,
							}
							return
						}
						if _, ok := proxyProvider.(contexts.IWithManagerContext); ok {
							proxyProvider.(contexts.IWithManagerContext).SetContext(s.Manager.Context)
						}
						for k, v := range inputCopy {
							triggerData.Inputs[k] = v
						}
						outputs, pause, iErr = proxyProvider.(stage.IProxyStageProvider).Process(ctx, *s.Manager.Context, triggerData)
					} else {
						outputs, pause, iErr = provider.(stage.IStageProvider).Process(ctx, *s.Manager.Context, inputCopy)
					}
					if iErr != nil {
						log.ErrorfCtx(ctx, " M (Stage): failed to process stage %s for site %s: %v", triggerData.Stage, site, iErr)
						results <- StageResult{
							Outputs: nil,
							Error:   iErr,
							Site:    site,
						}
						return
					}
					if pause {
						log.InfofCtx(ctx, " M (Stage): stage %s in activation %s for site %s get paused result from stage provider", triggerData.Stage, triggerData.Activation, site)
						pauseRequested = true
					}
					allOutputs = utils.MergeCollection_StringAny(allOutputs, outputs)
				}

				// 6.2 & 6.3 If currentStage.task exists, process tasks with concurrency
				if len(currentStage.Tasks) > 0 {
					if remoteStageProviderDefined {
						log.ErrorfCtx(ctx, " M (Stage): remote stage provider cannot be used with parallel tasks, skipping tasks execution for site %s", site)
						results <- StageResult{
							Outputs: allOutputs,
							Error:   nil,
							Site:    site,
						}
						return
					}

					taskResults, err := s.processTasks(ctx, currentStage, inputCopy, triggerData, triggers, site)
					// Merge task results with allOutputs
					allOutputs = utils.MergeCollection_StringAny(allOutputs, taskResults)

					if err != nil {
						results <- StageResult{
							Outputs: allOutputs,
							Error:   err,
							Site:    site,
						}
						return
					}
				}

				// 7. Merge results and return
				results <- StageResult{
					Outputs: allOutputs,
					Error:   err,
					Site:    site,
				}
			}(&waitGroup, site, results)
		}

		waitGroup.Wait()
		close(results)
		// DO NOT REMOVE THIS COMMENT
		// gofail: var afterProvider string

		outputs := make(map[string]interface{})
		hasStageError := false
		for result := range results {
			err = result.GetError()

			if err != nil {
				// Check if error is either an *apierrors.StatusError or a v1alpha2.COAError with status < 500

				// Set the common part regardless of the error type
				site := result.Site
				if result.Site == s.Context.SiteInfo.SiteId {
					site = ""
				}
				status.Outputs = carryOutPutsToErrorStatus(result.Outputs, err, site)
				result.Outputs = carryOutPutsToErrorStatus(result.Outputs, err, site)
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = fmt.Sprintf("%s: %s", result.Site, err.Error())
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to process stage %s for site %s outputs: %v", triggerData.Stage, site, err)
				hasStageError = true

			}
			for k, v := range result.Outputs {
				if result.Site == s.Context.SiteInfo.SiteId {
					outputs[k] = v
				} else {
					outputs[fmt.Sprintf("%s.%s", result.Site, k)] = v
				}
			}
			if result.Site == s.Context.SiteInfo.SiteId {
				if _, ok := result.Outputs["status"]; !ok {
					outputs["status"] = v1alpha2.OK
				}
			} else {
				key := fmt.Sprintf("%s.status", result.Site)
				if _, ok := result.Outputs[key]; !ok {
					outputs[key] = v1alpha2.Untouched
				}
			}
		}

		for k, v := range outputs {
			if !(strings.HasPrefix(k, "__") || strings.HasPrefix(k, "header.")) {
				status.Outputs[k] = v
			}
		}
		if triggerData.Outputs == nil {
			triggerData.Outputs = make(map[string]map[string]interface{})
		}
		triggerData.Outputs[triggerData.Stage] = outputs

		// If stage is paused, save the pending task and return paused status
		if pauseRequested {
			pendingTask := PendingTask{
				Sites:         sites,
				OutputContext: triggerData.Outputs,
			}
			_, err = s.StateProvider.Upsert(ctx, states.UpsertRequest{
				Value: states.StateEntry{
					ID:   fmt.Sprintf("%s-%s-%s", triggerData.Campaign, triggerData.Activation, triggerData.ActivationGeneration),
					Body: pendingTask,
				},
				Metadata: map[string]interface{}{
					"namespace": triggerData.Namespace,
				},
			})
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to save pending task: %v", err)
				return status, activationData
			}
			status.Status = v1alpha2.Paused
			status.StatusMessage = v1alpha2.Paused.String()
			status.IsActive = false
			return status, activationData
		}

		if campaign.SelfDriving {
			parser := utils.NewParser(currentStage.StageSelector)
			eCtx := s.VendorContext.EvaluationContext.Clone()
			eCtx.Context = ctx
			eCtx.Namespace = triggerData.Namespace
			eCtx.Triggers = triggerData.Inputs
			eCtx.Inputs = currentStage.Inputs
			if eCtx.Inputs != nil {
				if v, ok := eCtx.Inputs["context"]; ok {
					eCtx.Value = v
				}
			}
			eCtx.Outputs = triggerData.Outputs
			var val interface{}
			val, err = parser.Eval(*eCtx)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to evaluate stage selector: %v", err)
				return status, activationData
			}

			nextStageName := ""
			if val != nil {
				nextStageName = utils.FormatAsString(val)
			}

			log.InfofCtx(ctx, " M (Stage): stage %s finished. has error? %t", triggerData.Stage, hasStageError)
			if nextStageName != "" {
				if nextStage, ok := campaign.Stages[nextStageName]; ok || hasStageError {
					if !hasStageError || nextStage.HandleErrors {
						status.NextStage = nextStageName
						activationData = &v1alpha2.ActivationData{
							Campaign:             triggerData.Campaign,
							Activation:           triggerData.Activation,
							ActivationGeneration: triggerData.ActivationGeneration,
							Stage:                nextStageName,
							Inputs:               triggerData.Inputs,
							Outputs:              triggerData.Outputs,
							Provider:             nextStage.Provider,
							Config:               nextStage.Config,
							TriggeringStage:      triggerData.Stage,
							Schedule:             nextStage.Schedule,
							Namespace:            triggerData.Namespace,
							Proxy:                nextStage.Proxy,
						}
						s.setStageStatus(&status, nextStageName, v1alpha2.Done, "")
						return status, activationData
					} else {
						s.setStageStatus(&status, "", v1alpha2.InternalError, fmt.Sprintf("stage %s failed", triggerData.Stage))
						log.ErrorfCtx(ctx, " M (Stage): failed to process stage outputs: %v", status.ErrorMessage)
						return status, activationData
					}
				} else {
					s.setStageStatus(&status, "", v1alpha2.BadRequest, fmt.Sprintf("stage %s is not found", nextStageName))
					log.ErrorfCtx(ctx, " M (Stage): failed to find next stage: %v", err)
					return status, activationData
				}
			}
			// sVal is empty, no next stage
			if hasStageError {
				s.setStageStatus(&status, "", v1alpha2.InternalError, fmt.Sprintf("stage %s failed", triggerData.Stage))
				log.ErrorfCtx(ctx, " M (Stage): failed to process stage outputs: %v", status.ErrorMessage)
				return status, activationData
			}
			s.setStageStatus(&status, nextStageName, v1alpha2.Done, "")
			log.InfofCtx(ctx, " M (Stage): stage %s is done", triggerData.Stage)
			return status, activationData
		} else {
			// Not self-driving, no next stage
			if hasStageError {
				s.setStageStatus(&status, "", v1alpha2.InternalError, fmt.Sprintf("stage %s failed", triggerData.Stage))
				log.ErrorfCtx(ctx, " M (Stage): failed to process stage outputs: %v", status.ErrorMessage)
				return status, activationData
			}
			s.setStageStatus(&status, "", v1alpha2.Done, "")
			log.InfofCtx(ctx, " M (Stage): stage %s is done (no next stage)", triggerData.Stage)
			return status, activationData
		}
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", triggerData.Stage), v1alpha2.BadRequest)
	s.setStageStatus(&status, "", v1alpha2.InternalError, err.Error())
	log.ErrorfCtx(ctx, " M (Stage): failed to find stage: %v", err)
	return status, activationData
}

func (s *StageManager) setStageStatus(status *model.StageStatus, nextStage string, state v1alpha2.State, errMsg string) {
	status.NextStage = nextStage
	status.Status = state
	status.StatusMessage = state.String()
	status.IsActive = false
	status.ErrorMessage = errMsg
}

func (s *StageManager) traceValue(ctx context.Context, v interface{}, namespace string, inputs map[string]interface{}, triggers map[string]interface{}, outputs map[string]map[string]interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		parser := utils.NewParser(val)
		context := s.Context.VencorContext.EvaluationContext.Clone()
		context.Context = ctx
		context.DeploymentSpec = s.Context.VencorContext.EvaluationContext.DeploymentSpec
		context.Namespace = namespace
		context.Inputs = inputs
		if context.Inputs != nil {
			if v, ok := context.Inputs["context"]; ok {
				context.Value = v
			}
		}
		context.Triggers = triggers
		context.Outputs = outputs
		v, err := parser.Eval(*context)
		if err != nil {
			return "", err
		}
		switch vt := v.(type) {
		case string:
			return vt, nil
		default:
			return s.traceValue(ctx, v, namespace, inputs, triggers, outputs)
		}
	case []interface{}:
		ret := []interface{}{}
		for _, v := range val {
			tv, err := s.traceValue(ctx, v, namespace, inputs, triggers, outputs)
			if err != nil {
				return "", err
			}
			ret = append(ret, tv)
		}
		return ret, nil
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range val {
			tv, err := s.traceValue(ctx, v, namespace, inputs, triggers, outputs)
			if err != nil {
				return "", err
			}
			ret[k] = tv
		}
		return ret, nil
	default:
		return val, nil
	}
}

func (s *StageManager) HandleActivationEvent(ctx context.Context, actData v1alpha2.ActivationData, campaign model.CampaignSpec, activation model.ActivationState) (*v1alpha2.ActivationData, error) {
	stage := actData.Stage
	if _, ok := campaign.Stages[stage]; !ok {
		stage = campaign.FirstStage
	}
	if stage == "" {
		return nil, v1alpha2.NewCOAError(nil, "no stage found", v1alpha2.BadRequest)
	}
	if stageSpec, ok := campaign.Stages[stage]; ok {
		// There could be rare case where a stage is triggered twice and the history may contain two entries.
		// Skip the following check until we have stage dedup
		// if activation.Status != nil && activation.Status.StageHistory != nil && len(activation.Status.StageHistory) != 0 &&
		// 	activation.Status.StageHistory[len(activation.Status.StageHistory)-1].Stage != "" &&
		// 	activation.Status.StageHistory[len(activation.Status.StageHistory)-1].NextStage != stage {
		// 	log.ErrorfCtx(ctx, " M (Stage): current stage is %s, expected next stage is %s, actual next stage is %s",
		// 		activation.Status.StageHistory[len(activation.Status.StageHistory)-1].Stage,
		// 		activation.Status.StageHistory[len(activation.Status.StageHistory)-1].NextStage,
		// 		stage)
		// 	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not the next stage", stage), v1alpha2.BadRequest)
		// }
		return &v1alpha2.ActivationData{
			Campaign:             actData.Campaign,
			Activation:           actData.Activation,
			ActivationGeneration: actData.ActivationGeneration,
			Stage:                stage,
			Inputs:               activation.Spec.Inputs,
			Provider:             stageSpec.Provider,
			Config:               stageSpec.Config,
			TriggeringStage:      stage,
			Schedule:             stageSpec.Schedule,
			Namespace:            actData.Namespace,
			Proxy:                stageSpec.Proxy,
		}, nil
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", stage), v1alpha2.BadRequest)
}
