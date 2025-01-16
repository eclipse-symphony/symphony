/*
* Copyright (c) Microsoft Corporation.
* Licensed under the MIT license.
* SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/staging"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type SolutionVendor struct {
	vendors.Vendor
	SolutionManager *solution.SolutionManager
	PlanManager     *PlanManager
	StageManager    *stage.StageManager
	StagingManager  *staging.StagingManager
}

func NewPlanManager() *PlanManager {
	return &PlanManager{}
}

var apiOperationMetrics *metrics.Metrics

func (o *SolutionVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Solution",
		Producer: "Microsoft",
	}
}

func (e *SolutionVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*stage.StageManager); ok {
			e.StageManager = c
		}
		if c, ok := m.(*staging.StagingManager); ok {
			e.StagingManager = c
		}
		if c, ok := m.(*solution.SolutionManager); ok {
			e.SolutionManager = c
		}
	}
	e.PlanManager = NewPlanManager()
	if e.StageManager == nil {
		return v1alpha2.NewCOAError(nil, "stage manager is not supplied", v1alpha2.MissingConfig)
	}
	if e.StagingManager == nil {
		return v1alpha2.NewCOAError(nil, "staging manager is not supplied", v1alpha2.MissingConfig)
	}
	if e.SolutionManager == nil {
		return v1alpha2.NewCOAError(nil, "solution manager is not supplied", v1alpha2.MissingConfig)
	}
	e.Vendor.Context.Subscribe(DeploymentStepTopic, v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			log.InfoCtx(ctx, "V(Solution): subscribe deployment-step and begin to apply step ")
			// get data
			err := e.handleDeploymentStep(ctx, event)
			if err != nil {
				log.ErrorfCtx(ctx, "V(StageVendor): Failed to handle deployment plan: %v", err)
				// release lock
				var stepEnvelope StepEnvelope
				jData, _ := json.Marshal(event.Body)
				json.Unmarshal(jData, &stepEnvelope)
				lockName := api_utils.GenerateKeyLockName(stepEnvelope.PlanState.Namespace, stepEnvelope.PlanState.Deployment.Instance.ObjectMeta.Name)
				e.SolutionManager.KeyLockProvider.UnLock(lockName)
				return err
			}
			return err
		},
		Group: "Solution-vendor",
	})
	e.Vendor.Context.Subscribe(DeploymentPlanTopic, v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}

			log.InfoCtx(ctx, "V(StageVendor): Begin to execute deployment-plan")
			err := e.handleDeploymentPlan(ctx, event)
			if err != nil {
				log.ErrorfCtx(ctx, "V(StageVendor): Failed to handle deployment plan: %v", err)
				// release lock
				var planEnvelope PlanEnvelope
				jData, _ := json.Marshal(event.Body)
				json.Unmarshal(jData, &planEnvelope)
				lockName := api_utils.GenerateKeyLockName(planEnvelope.Namespace, planEnvelope.Deployment.Instance.ObjectMeta.Name)
				e.SolutionManager.KeyLockProvider.UnLock(lockName)
				return err
			}
			return err
		},
		Group: "stage-vendor",
	})
	e.Vendor.Context.Subscribe(CollectStepResultTopic, v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := event.Context
			if ctx == nil {
				ctx = context.TODO()
			}
			err := e.handleStepResult(ctx, event)
			if err != nil {
				log.ErrorfCtx(ctx, "V(Solution): Failed to handle step result: %v", err)
				return err
			}
			return err
		},
		Group: "stage-vendor",
	})
	return nil
}
func (e *SolutionVendor) handleDeploymentPlan(ctx context.Context, event v1alpha2.Event) error {
	var planEnvelope PlanEnvelope
	jData, _ := json.Marshal(event.Body)
	err := json.Unmarshal(jData, &planEnvelope)
	if err != nil {
		log.ErrorCtx(ctx, "failed to unmarshal plan envelope :%v", err)
		return err
	}

	planState := e.createPlanState(ctx, planEnvelope)
	lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	e.SolutionManager.KeyLockProvider.TryLock(lockName)
	log.InfoCtx(ctx, "begin to save summary for %s", planEnvelope.PlanId)
	if err := e.SaveSummaryInfo(ctx, planState, model.SummaryStateRunning); err != nil {
		// lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
		// e.UnlockObject(ctx, lockName)
		return err
	}
	if planState.isCompleted() {
		return e.handlePlanComplete(ctx, planState)

	}
	for stepId, step := range planEnvelope.Plan.Steps {
		switch planEnvelope.Phase {
		case PhaseGet:
			log.InfoCtx(ctx, "phase get begin deployment %+v", planEnvelope.Deployment)
			if err := e.publishDeploymentStep(ctx, stepId, planState, planEnvelope.Remove, planState.Steps[stepId]); err != nil {
				return err
			}
		case PhaseApply:
			planState.Summary.PlannedDeployment += len(step.Components)
		}
	}
	// for i, step := range planEnvelope.Plan.Steps {
	switch planEnvelope.Phase {
	case PhaseApply:
		// planState.Summary.PlannedDeployment += len(planEnvelope.Plan.Steps[0].Components)
		log.InfoCtx(ctx, "V(Solution): publish deployment step id %s step %+v", 0, planEnvelope.Plan.Steps[0].Role)
		if err := e.publishDeploymentStep(ctx, 0, planState, planEnvelope.Remove, planState.Steps[0]); err != nil {
			return err
		}
	}
	// }
	log.InfoCtx(ctx, "V(Solution): store plan id %s in map %+v", planEnvelope.PlanId)
	e.PlanManager.Plans.Store(planEnvelope.PlanId, planState)
	return nil
}
func (e *SolutionVendor) publishDeploymentStep(ctx context.Context, stepId int, planState *PlanState, remove bool, step model.DeploymentStep) error {
	log.InfoCtx(ctx, "V(StageVendor): publish deployment step for PlanId %s StepId %s", planState.PlanId, stepId)
	if err := e.Vendor.Context.Publish(DeploymentStepTopic, v1alpha2.Event{
		Body: StepEnvelope{
			Step:      step,
			StepId:    stepId,
			Remove:    remove,
			PlanState: planState,
		},
		Context: ctx,
	}); err != nil {
		// log.InfoCtx(ctx, "unlock3")
		// lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
		// e.UnlockObject(ctx, lockName)
		log.InfoCtx(ctx, "V(StageVendor): publish deployment step failed PlanId %s, stepId %s", planState.PlanId, stepId)
		return err
	}
	return nil
}
func (e *SolutionVendor) publishStepResult(ctx context.Context, target string, planId string, stepId int, Error error, getResult []model.ComponentSpec, applyResult map[string]model.ComponentResultSpec) error {
	errorString := ""
	if Error != nil {
		errorString = Error.Error()
	}
	return e.Vendor.Context.Publish(CollectStepResultTopic, v1alpha2.Event{
		Body: StepResult{
			Target:      target,
			PlanId:      planId,
			StepId:      stepId,
			GetResult:   getResult,
			ApplyResult: applyResult,
			Timestamp:   time.Now(),
			Error:       errorString,
		},
	})
}

// create inital plan state
func (e *SolutionVendor) createPlanState(ctx context.Context, planEnvelope PlanEnvelope) *PlanState {
	return &PlanState{
		PlanId:     planEnvelope.PlanId,
		StartTime:  time.Now(),
		TotalSteps: len(planEnvelope.Plan.Steps),
		Phase:      planEnvelope.Phase,
		Summary: model.SummarySpec{
			TargetResults:       make(map[string]model.TargetResultSpec),
			TargetCount:         len(planEnvelope.Deployment.Targets),
			SuccessCount:        0,
			AllAssignedDeployed: true,
			JobID:               planEnvelope.Deployment.JobID,
			IsRemoval:           planEnvelope.Remove,
		},
		PreviousDesiredState: planEnvelope.PreviousDesiredState,
		CompletedSteps:       0,
		MergedState:          planEnvelope.MergedState,
		Deployment:           planEnvelope.Deployment,
		Namespace:            planEnvelope.Namespace,
		Remove:               planEnvelope.Remove,
		TargetResult:         make(map[string]int),
		CurrentState:         planEnvelope.CurrentState,
		StepStates:           make([]StepState, len(planEnvelope.Plan.Steps)),
		Steps:                planEnvelope.Plan.Steps,
	}
}

// saveStepResult updates the plan state with the step result and saves the summary.
func (e *SolutionVendor) saveStepResult(ctx context.Context, planState *PlanState, stepResult StepResult) error {
	// Log the update of plan state with the step result
	log.InfoCtx(ctx, "V(Solution): Update plan state %v with step result %v phase %s", planState, stepResult, planState.Phase)
	planState.CompletedSteps++
	lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	e.SolutionManager.KeyLockProvider.TryLock(lockName)
	switch planState.Phase {
	case PhaseGet:
		// Update the GetResult for the specific step
		planState.StepStates[stepResult.StepId].GetResult = stepResult.GetResult
	case PhaseApply:
		if stepResult.Error != "" {
			// Handle error case and update the target result status and message
			targetResultStatus := fmt.Sprintf("%s Failed", deploymentTypeMap[planState.Remove])
			targetResultMessage := fmt.Sprintf("Failed to create provider %s, err: %s", deploymentTypeMap[planState.Remove], stepResult.Error)
			targetResultSpec := model.TargetResultSpec{Status: targetResultStatus, Message: targetResultMessage, ComponentResults: stepResult.ApplyResult}
			planState.Summary.UpdateTargetResult(stepResult.Target, targetResultSpec)
			planState.Summary.AllAssignedDeployed = false
			for _, ret := range stepResult.ApplyResult {
				if (!planState.Remove && ret.Status == v1alpha2.Updated) || (planState.Remove && ret.Status == v1alpha2.Deleted) {
					planState.Summary.CurrentDeployed++
				}
			}
			if planState.TargetResult[stepResult.Target] == 1 || planState.TargetResult[stepResult.Target] == 0 {
				planState.TargetResult[stepResult.Target] = -1
				planState.Summary.SuccessCount -= planState.TargetResult[stepResult.Target]
			}
			return e.handleAllPlanCompletetion(ctx, planState)
		} else {
			// Handle success case and update the target result status and message
			targetResultSpec := model.TargetResultSpec{Status: "OK", Message: "", ComponentResults: stepResult.ApplyResult}
			planState.Summary.UpdateTargetResult(stepResult.Target, targetResultSpec)
			log.InfoCtx(ctx, "Update plan state target spec %v", targetResultSpec)
			planState.Summary.CurrentDeployed += len(stepResult.ApplyResult)
			if planState.TargetResult[stepResult.Target] == 0 {
				planState.TargetResult[stepResult.Target] = 1
				planState.Summary.SuccessCount++
			}
			// publish next step execute event
			if stepResult.StepId != planState.TotalSteps-1 {
				log.InfoCtx(ctx, "V(Solution): publish deployment step id %s step %+v", stepResult.StepId+1, planState.Steps[stepResult.StepId+1].Role)
				if err := e.publishDeploymentStep(ctx, stepResult.StepId+1, planState, planState.Remove, planState.Steps[stepResult.StepId+1]); err != nil {
					log.InfoCtx(ctx, "V(Solution): publish deployment step failed PlanId %s, stepId %s", planState.PlanId, 0)
					return err
				}
			}

		}

		// If no components are deployed, set success count to target count
		if planState.Summary.CurrentDeployed == 0 && planState.Summary.AllAssignedDeployed {
			planState.Summary.SuccessCount = planState.Summary.TargetCount
		}

		// Save the summary information
		log.InfoCtx(ctx, "begin to save summary for %s", planState.Deployment.Instance.ObjectMeta.Name)
		if err := e.SaveSummaryInfo(ctx, planState, model.SummaryStateRunning); err != nil {
			log.ErrorfCtx(ctx, "Failed to save summary progress: %v", err)

		}
	}

	// Store the updated plan state
	e.PlanManager.Plans.Store(planState.PlanId, planState)

	// Check if all steps are completed and handle plan completion
	if planState.isCompleted() {
		err := e.handlePlanComplete(ctx, planState)
		if err != nil {
			log.InfoCtx(ctx, "V(Solution): handle plan Complete failed %+v", err)
			log.InfoCtx(ctx, "unlock2")
			e.SolutionManager.CleanupHeartbeat(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Namespace, planState.Remove)
			lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
			e.UnlockObject(ctx, lockName)
		}
		return err
	}

	return nil
}

// handleGetPlanCompletetion handles the completion of the get plan phase.
func (e *SolutionVendor) handleGetPlanCompletetion(ctx context.Context, planState *PlanState) error {
	// Collect result
	log.InfoCtx(ctx, "V(Solution): Begin to get current state %v", planState)
	Plan, planState, err := e.threeStateMerge(ctx, planState)
	if err != nil {
		log.ErrorfCtx(ctx, "V(Solution): Failed to merge states: %v", err)
		return err
	}
	e.Vendor.Context.Publish(DeploymentPlanTopic, v1alpha2.Event{
		Metadata: map[string]string{
			"Id": planState.Deployment.JobID,
		},
		Body: PlanEnvelope{
			Plan:                 Plan,
			Deployment:           planState.Deployment,
			MergedState:          planState.MergedState,
			CurrentState:         planState.CurrentState,
			PreviousDesiredState: planState.PreviousDesiredState,
			PlanId:               planState.PlanId,
			Remove:               planState.Remove,
			Namespace:            planState.Namespace,
			Phase:                PhaseApply,
		},
		Context: ctx,
	})
	return nil
}

// handlePlanComplete handles the completion of a plan and updates its status.
func (e *SolutionVendor) handlePlanComplete(ctx context.Context, planState *PlanState) error {
	log.InfoCtx(ctx, "V(Solution): Plan state %s is completed %v", planState.Phase, planState)
	if !planState.Summary.AllAssignedDeployed {
		planState.Status = "failed"
	}
	log.InfoCtx(ctx, "V(Solution): Plan state is completed %v", planState.Summary.AllAssignedDeployed)
	switch planState.Phase {
	case PhaseGet:
		if err := e.handleGetPlanCompletetion(ctx, planState); err != nil {
			e.PlanManager.DeletePlan(planState.PlanId)
			// lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
			// e.UnlockObject(ctx, lockName)
			// log.ErrorfCtx(ctx, "V(Solution): Failed to handle get plan completion: %v", err)
			return err
		}
	case PhaseApply:
		if err := e.handleAllPlanCompletetion(ctx, planState); err != nil {
			// lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
			// e.UnlockObject(ctx, lockName)
			// log.ErrorfCtx(ctx, "V(Solution): Failed to handle apply plan completion: %v", err)
			return err
		}
		e.PlanManager.DeletePlan(planState.PlanId)
	}

	return nil
}

// handleStepResult processes the event and updates the plan state accordingly.
func (e *SolutionVendor) handleStepResult(ctx context.Context, event v1alpha2.Event) error {
	var stepResult StepResult
	// Marshal the event body to JSON
	jData, _ := json.Marshal(event.Body)
	log.InfofCtx(ctx, "Received event body: %s", string(jData))

	// Unmarshal the JSON data into stepResult
	if err := json.Unmarshal(jData, &stepResult); err != nil {
		log.ErrorfCtx(ctx, "Failed to unmarshal step result: %v", err)
		return err
	}

	planId := stepResult.PlanId
	// Load the plan state object from the PlanManager
	planStateObj, exists := e.PlanManager.Plans.Load(planId)
	if !exists {
		// log.ErrorCtx(ctx, "Plan not found: %s", planId)
		// e.UnlockObject(ctx, api_utils.GenerateKeyLockName(stepResult.Namespace, stepResult.PlanId))
		return fmt.Errorf("Plan not found: %s", planId)
	}
	planState := planStateObj.(*PlanState)

	// Update the plan state in the map and save the summary
	if err := e.saveStepResult(ctx, planState, stepResult); err != nil {
		log.ErrorCtx(ctx, "Failed to handle step result: %v", err)
		// lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
		// log.InfoCtx(ctx, "unlock1")
		// e.UnlockObject(ctx, lockName)
		return err
	}
	return nil
}

func (o *SolutionVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "solution"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost, fasthttp.MethodGet, fasthttp.MethodDelete},
			Route:   route + "/instances", //this route is to support ITargetProvider interface via a proxy provider
			Version: o.Version,
			Handler: o.onApplyDeployment,
		},
		{
			Methods:    []string{fasthttp.MethodPost},
			Route:      route + "/reconcile",
			Version:    o.Version,
			Parameters: []string{"delete?"},
			Handler:    o.onReconcile,
		},
		{
			Methods: []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:   route + "/queue",
			Version: o.Version,
			Handler: o.onQueue,
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/tasks",
			Version: o.Version,
			Handler: o.onGetRequest,
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/task/getResult",
			Version: o.Version,
			Handler: o.onGetResponse,
		},
	}
}
func (e *SolutionVendor) handleDeploymentStep(ctx context.Context, event v1alpha2.Event) error {
	var stepEnvelope StepEnvelope
	jData, err := json.Marshal(event.Body)
	if err != nil {
		log.ErrorfCtx(ctx, "V (Solution): failed to unmarshal event body: %v", err)
		return err
	}
	if err := json.Unmarshal(jData, &stepEnvelope); err != nil {
		log.ErrorfCtx(ctx, "V (Solution): failed to unmarshal step envelope: %v", err)
		return err
	}
	if stepEnvelope.Step.Role == "container" {
		stepEnvelope.Step.Role = "instance"
	}
	switch stepEnvelope.PlanState.Phase {
	case PhaseGet:
		return e.handlePhaseGet(ctx, stepEnvelope)
	case PhaseApply:
		return e.handlePhaseApply(ctx, stepEnvelope)
	}
	return nil
}
func findAgentFromDeploymentState(deployment model.DeploymentSpec, targetName string) bool {
	// find targt component
	targetSpec := deployment.Targets[targetName]
	log.Info("compare between state and target name %s, %+v", targetName, targetSpec)
	for _, component := range targetSpec.Spec.Components {
		log.Info("compare between state and target name %+v, %s", component, component.Name)
		if component.Type == "remote-agent" {
			log.Info("It is remote call ")
			return true
		} else {
			log.Info(" it is not remote call target Name %s", targetName)
		}
	}
	return false
}

func (e *SolutionVendor) handlePhaseGet(ctx context.Context, stepEnvelope StepEnvelope) error {
	if findAgentFromDeploymentState(stepEnvelope.PlanState.Deployment, stepEnvelope.Step.Target) {
		return e.enqueueProviderGetRequest(ctx, stepEnvelope)
	}
	return e.getProviderAndExecute(ctx, stepEnvelope)
}
func (e *SolutionVendor) enqueueProviderGetRequest(ctx context.Context, stepEnvelope StepEnvelope) error {
	operationId := uuid.New().String()
	providerGetRequest := &ProviderGetRequest{
		AgentRequest: AgentRequest{
			OperationID: operationId,
			Provider:    stepEnvelope.Step.Role,
			Action:      string(PhaseGet),
		},
		References: stepEnvelope.Step.Components,
		Deployment: stepEnvelope.PlanState.Deployment,
	}

	log.InfoCtx(ctx, "V(Solution): Enqueue get message %s-%s %+v ", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace, providerGetRequest)
	messageID, err := e.StagingManager.QueueProvider.Enqueue(fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace), providerGetRequest)
	if err != nil {
		log.ErrorCtx(ctx, "V(Solution): Error in enqueue message %s", fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace))
		return err
	}
	err = e.upsertOperationState(ctx, operationId, stepEnvelope.StepId, stepEnvelope.PlanState.PlanId, stepEnvelope.Step.Target, stepEnvelope.PlanState.Phase, stepEnvelope.PlanState.Namespace, stepEnvelope.Remove, messageID)
	if err != nil {
		log.ErrorCtx(ctx, "V(Solution) Error in insert operation Id %s", operationId)
		return e.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{})
	}
	return err
}

func (e *SolutionVendor) getProviderAndExecute(ctx context.Context, stepEnvelope StepEnvelope) error {
	provider, err := e.SolutionManager.GetTargetProviderForStep(stepEnvelope.Step.Target, stepEnvelope.Step.Role, stepEnvelope.PlanState.Deployment, stepEnvelope.PlanState.PreviousDesiredState)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to create provider & Failed to save summary progress: %v", err)
		return e.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{})
	}
	dep := stepEnvelope.PlanState.Deployment
	dep.ActiveTarget = stepEnvelope.Step.Target
	getResult, stepError := (provider.(tgt.ITargetProvider)).Get(ctx, dep, stepEnvelope.Step.Components)
	if stepError != nil {
		log.ErrorCtx(ctx, "V(Solution) Error in get target current states %+v", stepError)
		return e.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{})
	}
	return e.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, getResult, map[string]model.ComponentResultSpec{})
}

func (e *SolutionVendor) handlePhaseApply(ctx context.Context, stepEnvelope StepEnvelope) error {
	if findAgentFromDeploymentState(stepEnvelope.PlanState.Deployment, stepEnvelope.Step.Target) {
		return e.enqueueProviderApplyRequest(ctx, stepEnvelope)
	}
	return e.applyProviderAndExecute(ctx, stepEnvelope)
}

func (e *SolutionVendor) enqueueProviderApplyRequest(ctx context.Context, stepEnvelope StepEnvelope) error {
	operationId := uuid.New().String()
	providApplyRequest := &ProviderApplyRequest{
		AgentRequest: AgentRequest{
			OperationID: operationId,
			Provider:    stepEnvelope.Step.Role,
			Action:      string(PhaseApply),
		},
		Deployment: stepEnvelope.PlanState.Deployment,
		Step:       stepEnvelope.Step,
		IsDryRun:   stepEnvelope.PlanState.Deployment.IsDryRun,
	}
	messageId, err := e.StagingManager.QueueProvider.Enqueue(fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace), providApplyRequest)
	if err != nil {
		return err
	}
	log.InfoCtx(ctx, "V(Solution): Enqueue apply message %s-%s %+v ", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace, providApplyRequest)
	err = e.upsertOperationState(ctx, operationId, stepEnvelope.StepId, stepEnvelope.PlanState.PlanId, stepEnvelope.Step.Target, stepEnvelope.PlanState.Phase, stepEnvelope.PlanState.Namespace, stepEnvelope.Remove, messageId)
	if err != nil {
		log.ErrorCtx(ctx, "error in insert operation Id %s", operationId)
		return e.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{})
	}
	return err
}

func (e *SolutionVendor) applyProviderAndExecute(ctx context.Context, stepEnvelope StepEnvelope) error {
	// get provider todo : is dry run
	provider, err := e.SolutionManager.GetTargetProviderForStep(stepEnvelope.Step.Target, stepEnvelope.Step.Role, stepEnvelope.PlanState.Deployment, stepEnvelope.PlanState.PreviousDesiredState)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to create provider & Failed to save summary progress: %v", err)
		return e.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{})
	}
	previousDesiredState := stepEnvelope.PlanState.PreviousDesiredState
	currentState := stepEnvelope.PlanState.CurrentState
	step := stepEnvelope.Step
	if previousDesiredState != nil {
		testState := solution.MergeDeploymentStates(&previousDesiredState.State, currentState)
		if e.SolutionManager.CanSkipStep(ctx, step, step.Target, provider.(tgt.ITargetProvider), previousDesiredState.State.Components, testState) {
			log.InfofCtx(ctx, " M (Solution): skipping step with role %s on target %s", step.Role, step.Target)
			return e.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, nil, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{})
		}
	}
	componentResults, stepError := (provider.(tgt.ITargetProvider)).Apply(ctx, stepEnvelope.PlanState.Deployment, stepEnvelope.Step, stepEnvelope.PlanState.Deployment.IsDryRun)
	return e.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, stepError, []model.ComponentSpec{}, componentResults)
}

func (e *SolutionVendor) onQueue(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onQueue",
	})
	defer span.End()
	instance := request.Parameters["instance"]
	sLog.InfofCtx(rContext, "V (Solution): onQueue, method: %s, %s", request.Method, instance)

	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onQueue-GET", rContext, nil)
		defer span.End()
		instance := request.Parameters["instance"]

		if instance == "" {
			sLog.ErrorCtx(ctx, "V (Solution): onQueue failed - 400 instance parameter is not found")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - instance parameter is not found\"}"),
				ContentType: "application/json",
			})
		}
		summary, err := e.SolutionManager.GetSummary(ctx, instance, namespace)
		data, _ := json.Marshal(summary)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onQueue failed - %s", err.Error())
			if utils.IsNotFound(err) {
				errorMsg := fmt.Sprintf("instance '%s' is not found in namespace %s", instance, namespace)
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.NotFound,
					Body:  []byte(errorMsg),
				})
			} else {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
					Body:  data,
				})
			}
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        data,
			ContentType: "application/json",
		})
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onQueue-POST", rContext, nil)
		defer span.End()

		// DO NOT REMOVE THIS COMMENT
		// gofail: var onQueueError string

		instance := request.Parameters["instance"]
		delete := request.Parameters["delete"]
		objectType := request.Parameters["objectType"]
		target := request.Parameters["target"]

		if objectType == "" { // For backward compatibility
			objectType = "instance"
		}

		if target == "true" {
			objectType = "target"
		}

		if objectType == "deployment" {
			deployment, err := model.ToDeployment(request.Body)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State:       v1alpha2.DeserializeError,
					ContentType: "application/json",
					Body:        []byte(fmt.Sprintf(`{"result":"%s"}`, err.Error())),
				})
			}
			instance = deployment.Instance.ObjectMeta.Name
		}

		if instance == "" {
			sLog.ErrorCtx(ctx, "V (Solution): onQueue failed - 400 instance parameter is not found")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - instance parameter is not found\"}"),
				ContentType: "application/json",
			})
		}
		action := v1alpha2.JobUpdate
		if delete == "true" {
			action = v1alpha2.JobDelete
		}
		e.Vendor.Context.Publish("job", v1alpha2.Event{
			Metadata: map[string]string{
				"objectType": objectType,
				"namespace":  namespace,
			},
			Body: v1alpha2.JobData{
				Id:     instance,
				Scope:  namespace,
				Action: action,
				Data:   request.Body,
			},
			Context: ctx,
		})
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte("{\"result\":\"200 - instance reconcilation job accepted\"}"),
			ContentType: "application/json",
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onQueue-DELETE", rContext, nil)
		defer span.End()
		instance := request.Parameters["instance"]

		if instance == "" {
			sLog.ErrorCtx(ctx, "V (Solution): onQueue failed - 400 instance parameter is not found")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - instance parameter is not found\"}"),
				ContentType: "application/json",
			})
		}

		err := e.SolutionManager.DeleteSummary(ctx, instance, namespace)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onQueue DeleteSummary failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			ContentType: "application/json",
		})
	}
	sLog.ErrorCtx(rContext, "V (Solution): onQueue failed - 405 method not allowed")
	return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	})
}

func (e *SolutionVendor) onReconcile(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onReconcile",
	})
	defer span.End()

	sLog.InfofCtx(rContext, "V (Solution): onReconcile, method: %s", request.Method)
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}
	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onReconcile-POST", rContext, nil)
		defer span.End()
		var deployment model.DeploymentSpec
		err := json.Unmarshal(request.Body, &deployment)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onReconcile failed POST - unmarshal request %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		lockName := api_utils.GenerateKeyLockName(namespace, deployment.Instance.ObjectMeta.Name)
		// if !e.SolutionManager.KeyLockProvider.TryLock(api_utils.GenerateKeyLockName(namespace, deployment.Instance.ObjectMeta.Name)) {
		// 	log.Info("can not get lock %s", lockName)
		// }
		e.SolutionManager.KeyLockProvider.Lock(lockName)
		log.InfoCtx(ctx, "lock succeed %s", lockName)
		delete := request.Parameters["delete"]
		remove := delete == "true"
		targetName := ""
		if request.Metadata != nil {
			if v, ok := request.Metadata["active-target"]; ok {
				targetName = v
			}
		}
		log.InfoCtx(ctx, "get deployment %+v", deployment)
		log.InfofCtx(ctx, " M (Solution): reconciling deployment.InstanceName: %s, deployment.SolutionName: %s, remove: %t, namespace: %s, targetName: %s, generation: %s, jobID: %s",
			deployment.Instance.ObjectMeta.Name,
			deployment.SolutionName,
			remove,
			namespace,
			targetName,
			deployment.Generation,
			deployment.JobID)
		previousDesiredState := e.SolutionManager.GetPreviousState(ctx, deployment.Instance.ObjectMeta.Name, namespace)
		// create new deployment state
		var state model.DeploymentState
		state, err = solution.NewDeploymentState(deployment)
		if err != nil {
			log.InfoCtx(ctx, "unlock5")
			e.UnlockObject(ctx, lockName)
			log.ErrorfCtx(ctx, " M (Solution): failed to create manager state for deployment: %+v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.MethodNotAllowed,
				Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
				ContentType: "application/json",
			})
		}
		// save summary
		summary := model.SummarySpec{
			TargetResults:       make(map[string]model.TargetResultSpec),
			TargetCount:         len(deployment.Targets),
			SuccessCount:        0,
			AllAssignedDeployed: false,
			JobID:               deployment.JobID,
		}
		data, _ := json.Marshal(summary)
		err = e.SolutionManager.SaveSummary(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, model.SummaryStateRunning, namespace)
		if err != nil {
			log.InfoCtx(ctx, "unlock6")
			e.UnlockObject(ctx, lockName)
			log.ErrorfCtx(ctx, " M (Solution): failed to create manager state for deployment: %+v", err)

			e.SolutionManager.ConcludeSummary(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.InternalError,
				Body:        []byte(fmt.Sprintf("{\"result\":\"500 - M (Solution): failed to save summary: %+v\"}", err)),
				ContentType: "application/json",
			})
		}

		stopCh := make(chan struct{})
		defer close(stopCh)
		go e.SolutionManager.SendHeartbeat(ctx, deployment.Instance.ObjectMeta.Name, namespace, remove, stopCh)

		// get the components count for the deployment
		componentCount := len(deployment.Solution.Spec.Components)
		apiOperationMetrics.ApiComponentCount(
			componentCount,
			metrics.ReconcileOperation,
			metrics.UpdateOperationType,
		)

		if e.SolutionManager.VendorContext != nil && e.SolutionManager.VendorContext.EvaluationContext != nil {
			context := e.SolutionManager.VendorContext.EvaluationContext.Clone()
			context.DeploymentSpec = deployment
			context.Value = deployment
			context.Component = ""
			context.Namespace = namespace
			context.Context = ctx
			deployment, err = api_utils.EvaluateDeployment(*context)
			if err != nil {
				if remove {
					log.InfofCtx(ctx, " M (Solution): skipped failure to evaluate deployment spec: %+v", err)
				} else {
					summary.SummaryMessage = "failed to evaluate deployment spec: " + err.Error()
					log.ErrorfCtx(ctx, " M (Solution): failed to evaluate deployment spec: %+v", err)
					e.SolutionManager.ConcludeSummary(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
					return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
						State:       v1alpha2.InternalError,
						Body:        []byte(fmt.Sprintf("{\"result\":\"500 - M (Solution): failed to evaluate deployment spec: %+v\"}", err)),
						ContentType: "application/json",
					})
				}
			}

		}
		// e.SolutionManager.KeyLockProvider.Lock(api_utils.GenerateKeyLockName(namespace, deployment.Instance.ObjectMeta.Name))
		// Generate new deployment plan for deployment
		initalPlan, err := solution.PlanForDeployment(deployment, state)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  data,
			})
		}

		// remove no use steps
		var stepList []model.DeploymentStep
		for _, step := range initalPlan.Steps {
			if e.SolutionManager.IsTarget && !api_utils.ContainsString(e.SolutionManager.TargetNames, step.Target) {
				continue
			}
			if targetName != "" && targetName != step.Target {
				continue
			}
			stepList = append(stepList, step)
		}
		initalPlan.Steps = stepList
		log.InfoCtx(ctx, "publish topic for object %s", deployment.Instance.ObjectMeta.Name)
		e.Vendor.Context.Publish(DeploymentPlanTopic, v1alpha2.Event{
			Metadata: map[string]string{
				"Id": deployment.JobID,
			},
			Body: PlanEnvelope{
				Plan:                 initalPlan,
				Deployment:           deployment,
				MergedState:          model.DeploymentState{},
				PreviousDesiredState: previousDesiredState,
				PlanId:               uuid.New().String(),
				Remove:               delete == "true",
				Namespace:            namespace,
				Phase:                PhaseGet,
			},
			Context: ctx,
		})

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        data,
			ContentType: "application/json",
		})
	}
	sLog.ErrorCtx(rContext, "V (Solution): onReconcile failed - 405 method not allowed")
	return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	})
}

// onGetRequest handles the get request from the remote agent.
func (e *SolutionVendor) onGetRequest(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onGetRequest",
	})
	defer span.End()
	var agentRequest AgentRequest
	sLog.InfoCtx(ctx, "V(Solution): get request from remote agent")
	target := request.Parameters["target"]
	namespace := request.Parameters["namespace"]
	getAll, exists := request.Parameters["getAll"]

	if exists && getAll == "true" {
		// Logic to handle getALL parameter
		sLog.InfoCtx(ctx, "V(Solution): getALL request from remote agent %+v", agentRequest)
		return e.getTaskFromQueue(ctx, target, namespace, true)
	}
	return e.getTaskFromQueue(ctx, target, namespace, false)
}

// onGetResponse handles the get response from the remote agent.
func (e *SolutionVendor) onGetResponse(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onGetResponse",
	})
	defer span.End()

	var asyncResult AsyncResult
	err := json.Unmarshal(request.Body, &asyncResult)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V(Solution): onGetResponse failed - %s", err.Error())
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	sLog.InfoCtx(ctx, "V(Solution): get async result from remote agent %+v", asyncResult)
	return e.handleRemoteAgentExecuteResult(ctx, asyncResult)
}

// handleRemoteAgentExecuteResult handles the execution result from the remote agent.
func (e *SolutionVendor) handleRemoteAgentExecuteResult(ctx context.Context, asyncResult AsyncResult) v1alpha2.COAResponse {
	// Get operation ID
	operationId := asyncResult.OperationID
	// Get related info from redis - todo: timeout
	log.InfoCtx(ctx, "V(SolutionVendor): handle remote agent request %+v", asyncResult)
	operationBody, err := e.getOperationState(ctx, operationId)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V(SolutionVendor): onGetResponse failed - %s", err.Error())
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	queueName := fmt.Sprintf("%s-%s", operationBody.Target, operationBody.NameSpace)
	switch operationBody.Action {
	case PhaseGet:
		// Send to step result
		var response []model.ComponentSpec
		err := json.Unmarshal(asyncResult.Body, &response)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		e.publishStepResult(ctx, operationBody.Target, operationBody.PlanId, operationBody.StepId, asyncResult.Error, response, map[string]model.ComponentResultSpec{})
		deleteRequest := states.DeleteRequest{
			ID: operationId,
		}

		err = e.StagingManager.StateProvider.Delete(ctx, deleteRequest)
		if err != nil {
			return v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"405 - delete operation Id failed\"}"),
				ContentType: "application/json",
			}
		}
		// delete from queue

		e.StagingManager.QueueProvider.RemoveFromQueue(queueName, operationBody.MessageId)
		return v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte("{\"result\":\"200 - handle async result successfully\"}"),
			ContentType: "application/json",
		}
	case PhaseApply:
		var response map[string]model.ComponentResultSpec
		err := json.Unmarshal(asyncResult.Body, &response)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		e.publishStepResult(ctx, operationBody.Target, operationBody.PlanId, operationBody.StepId, asyncResult.Error, []model.ComponentSpec{}, response)
		deleteRequest := states.DeleteRequest{
			ID: operationId,
		}
		err = e.StagingManager.StateProvider.Delete(ctx, deleteRequest)
		// delete from queue
		e.StagingManager.QueueProvider.RemoveFromQueue(queueName, operationBody.MessageId)
		if err != nil {
			return v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"delete operation Id failed\"}"),
				ContentType: "application/json",
			}
		}
		return v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte("{\"result\":\"200 - get response successfully\"}"),
			ContentType: "application/json",
		}
	}
	return v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
}

// getTaskFromQueue retrieves a task from the queue for the specified target and namespace.
func (e *SolutionVendor) getTaskFromQueue(ctx context.Context, target string, namespace string, fromBegining bool) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doGetFromQueue",
	})
	queueName := fmt.Sprintf("%s-%s", target, namespace)
	sLog.InfoCtx(ctx, "V(SolutionVendor): getFromQueue %s queue length %s", queueName)
	defer span.End()
	var queueElement interface{}
	var err error
	if fromBegining {
		queueElement, err = e.StagingManager.QueueProvider.PeekFromBegining(queueName)
	} else {
		queueElement, err = e.StagingManager.QueueProvider.Peek(queueName)
	}
	if err != nil {
		sLog.ErrorfCtx(ctx, "V(SolutionVendor): getQueue failed - %s", err.Error())
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	data, _ := json.Marshal(queueElement)
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}

// upsertOperationState upserts the operation state for the specified parameters.
func (e *SolutionVendor) upsertOperationState(ctx context.Context, operationId string, stepId int, planId string, target string, action JobPhase, namespace string, remove bool, messageId string) error {
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: operationId,
			Body: map[string]interface{}{
				"StepId":    stepId,
				"PlanId":    planId,
				"Target":    target,
				"Action":    action,
				"namespace": namespace,
				"Remove":    remove,
				"MessageId": messageId,
			}},
	}
	_, err := e.StagingManager.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

// getOperationState retrieves the operation state for the specified operation ID.
func (e *SolutionVendor) getOperationState(ctx context.Context, operationId string) (OperationBody, error) {
	getRequest := states.GetRequest{
		ID: operationId,
	}
	var entry states.StateEntry
	entry, err := e.StagingManager.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return OperationBody{}, err
	}
	var ret OperationBody
	ret, err = e.getOperationBody(entry.Body)
	if err != nil {
		log.ErrorfCtx(ctx, "V(SolutionVendor): Failed to convert to operation state for %s", operationId)
		return OperationBody{}, err
	}
	return ret, err
}

// getOperationBody converts the body to an OperationBody.
func (e *SolutionVendor) getOperationBody(body interface{}) (OperationBody, error) {
	var operationBody OperationBody
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &operationBody)
	if err != nil {
		return OperationBody{}, err
	}
	return operationBody, nil
}

func (e *SolutionVendor) onApplyDeployment(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onApplyDeployment",
	})
	defer span.End()

	sLog.InfofCtx(rContext, "V (Solution): onApplyDeployment %s", request.Method)
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}
	targetName := ""
	if request.Metadata != nil {
		if v, ok := request.Metadata["active-target"]; ok {
			targetName = v
		}
	}
	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("Apply Deployment", rContext, nil)
		defer span.End()
		deployment := new(model.DeploymentSpec)
		err := json.Unmarshal(request.Body, &deployment)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onApplyDeployment failed - %s", err.Error())
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := e.doDeploy(ctx, *deployment, namespace, targetName)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("Get Components", rContext, nil)
		defer span.End()
		deployment := new(model.DeploymentSpec)
		err := json.Unmarshal(request.Body, &deployment)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onApplyDeployment failed - %s", err.Error())
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := e.doGet(ctx, *deployment, targetName)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("Delete Components", rContext, nil)
		defer span.End()
		var deployment model.DeploymentSpec
		err := json.Unmarshal(request.Body, &deployment)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onApplyDeployment failed - %s", err.Error())
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := e.doRemove(ctx, deployment, namespace, targetName)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	}
	sLog.ErrorCtx(rContext, "V (Solution): onApplyDeployment failed - 405 method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (e *SolutionVendor) doGet(ctx context.Context, deployment model.DeploymentSpec, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doGet",
	})
	defer span.End()
	sLog.InfoCtx(ctx, "V (Solution): doGet")

	_, components, err := e.SolutionManager.Get(ctx, deployment, targetName)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (Solution): doGet failed - %s", err.Error())
		response := v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
		observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
		return response
	}
	data, _ := json.Marshal(components)
	response := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
	return response
}
func (e *SolutionVendor) doDeploy(ctx context.Context, deployment model.DeploymentSpec, namespace string, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doDeploy",
	})
	defer span.End()
	sLog.InfoCtx(ctx, "V (Solution): doDeploy")
	summary, err := e.SolutionManager.Reconcile(ctx, deployment, false, namespace, targetName)
	data, _ := json.Marshal(summary)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (Solution): doDeploy failed - %s", err.Error())
		response := v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  data,
		}
		observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
		return response
	}
	response := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
	return response
}
func (e *SolutionVendor) doRemove(ctx context.Context, deployment model.DeploymentSpec, namespace string, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doRemove",
	})
	defer span.End()

	sLog.InfoCtx(ctx, "V (Solution): doRemove")
	summary, err := e.SolutionManager.Reconcile(ctx, deployment, true, namespace, targetName)
	data, _ := json.Marshal(summary)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (Solution): doRemove failed - %s", err.Error())
		response := v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  data,
		}
		observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
		return response
	}
	response := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
	return response
}

// threeStateMerge merges the current, previous, and desired states to create a deployment plan.
func (e *SolutionVendor) threeStateMerge(ctx context.Context, planState *PlanState) (model.DeploymentPlan, *PlanState, error) {
	currentState := model.DeploymentState{}
	currentState.TargetComponent = make(map[string]string)

	for _, StepState := range planState.StepStates {
		for _, c := range StepState.GetResult {
			key := fmt.Sprintf("%s::%s", c.Name, StepState.Target)
			role := c.Type
			if role == "" {
				role = "instance"
			}
			log.InfoCtx(ctx, "V(Solution): Store key value in current key: %s value: %s", key, role)
			currentState.TargetComponent[key] = role
		}
	}
	log.InfoCtx(ctx, "V(Solution): Compute current state %v for Plan ID: %s", currentState, planState.PlanId)
	planState.CurrentState = currentState
	previousDesiredState := e.SolutionManager.GetPreviousState(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Namespace)
	log.InfoCtx(ctx, "V(Solution): Get previous desired state %+v", previousDesiredState)
	planState.PreviousDesiredState = previousDesiredState
	var currentDesiredState model.DeploymentState
	currentDesiredState, err := solution.NewDeploymentState(planState.Deployment)
	if err != nil {
		log.ErrorfCtx(ctx, "V(Solution): Failed to get current desired state: %+v", err)
		return model.DeploymentPlan{}, &PlanState{}, err
	}
	log.InfoCtx(ctx, "V(Solution): Get current desired state %+v", currentDesiredState)
	desiredState := currentDesiredState
	if previousDesiredState != nil {
		desiredState = solution.MergeDeploymentStates(&previousDesiredState.State, currentDesiredState)
	}
	log.InfoCtx(ctx, "V(Solution): Get desired state %+v", desiredState)
	if planState.Remove {
		desiredState.MarkRemoveAll()
		log.InfoCtx(ctx, "V(Solution): After remove desired state %+v", desiredState)
	}

	mergedState := solution.MergeDeploymentStates(&currentState, desiredState)
	planState.MergedState = mergedState
	log.InfoCtx(ctx, "get merged state %+v", mergedState)
	Plan, err := solution.PlanForDeployment(planState.Deployment, mergedState)
	if err != nil {
		return model.DeploymentPlan{}, &PlanState{}, err
	}
	e.PlanManager.Plans.Store(planState.PlanId, planState)
	log.InfoCtx(ctx, "V(Solution): Begin to publish topic to deployment plan %v merged state %v get plan %v", planState, mergedState, Plan)
	return Plan, planState, nil
}

func (e *SolutionVendor) UnlockObject(ctx context.Context, lockName string) {
	e.SolutionManager.KeyLockProvider.TryLock(lockName)
	log.InfoCtx(ctx, "unlock %s", lockName)
	e.SolutionManager.KeyLockProvider.UnLock(lockName)
}
func (e *SolutionVendor) SaveSummaryInfo(ctx context.Context, planState *PlanState, state model.SummaryState) error {
	return e.SolutionManager.SaveSummary(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Deployment.Generation, planState.Deployment.Hash, planState.Summary, state, planState.Namespace)
}

func (e *SolutionVendor) handleAllPlanCompletetion(ctx context.Context, planState *PlanState) error {
	log.InfofCtx(ctx, "handle plan completetion:begin to handle plan completetion %v", planState)
	if err := e.SaveSummaryInfo(ctx, planState, model.SummaryStateDone); err != nil {
		return err
	}
	// update summary
	log.InfoCtx(ctx, "begin to save summary for %s", planState.Deployment.Instance.ObjectMeta.Name)
	planState.MergedState.ClearAllRemoved()
	log.InfoCtx(ctx, "if it is dry run %+v", planState.Deployment.IsDryRun)
	log.InfoCtx(ctx, "get dep %+v", planState.Deployment)
	if !planState.Deployment.IsDryRun {
		if len(planState.MergedState.TargetComponent) == 0 && planState.Remove {
			log.DebugfCtx(ctx, " M (Solution): no assigned components to manage, deleting state")
			e.SolutionManager.StateProvider.Delete(ctx, states.DeleteRequest{
				ID: planState.Deployment.Instance.ObjectMeta.Name,
				Metadata: map[string]interface{}{
					"namespace": planState.Namespace,
					"group":     model.SolutionGroup,
					"version":   "v1",
					"resource":  DeploymentState,
				},
			})
		} else {
			log.InfoCtx(ctx, "begin to save state %s", planState.Deployment.Instance.ObjectMeta.Name)
			log.InfoCtx(ctx, "begin to save state deployment%s", planState.Deployment)
			log.InfoCtx(ctx, "begin to save state deployment%s", planState.MergedState)
			e.SolutionManager.StateProvider.Upsert(ctx, states.UpsertRequest{
				Value: states.StateEntry{
					ID: planState.Deployment.Instance.ObjectMeta.Name,
					Body: solution.SolutionManagerDeploymentState{
						Spec:  planState.Deployment,
						State: planState.MergedState,
					},
				},
				Metadata: map[string]interface{}{
					"namespace": planState.Namespace,
					"group":     model.SolutionGroup,
					"version":   "v1",
					"resource":  DeploymentState,
				},
			})
		}
	}
	if planState.Deployment.IsDryRun {
		planState.Summary.SuccessCount = 0
	}
	if err := e.SolutionManager.ConcludeSummary(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Deployment.Generation, planState.Deployment.Hash, planState.Summary, planState.Namespace); err != nil {
		return err
	}
	log.InfoCtx(ctx, "final unlock %s", planState.Deployment.Instance.ObjectMeta.Name)
	e.SolutionManager.CleanupHeartbeat(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Namespace, planState.Remove)
	e.UnlockObject(ctx, api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name))
	return nil
}

func (p *PlanState) IsExpired() bool {
	log.Info("time now")
	log.Info("time expired")
	return time.Now().After(p.ExpireTime)
}

func (p *PlanState) isCompleted() bool {
	return p.CompletedSteps == p.TotalSteps
}
func (pm *PlanManager) GetPlan(planId string) (*PlanState, bool) {
	if value, ok := pm.Plans.Load(planId); ok {
		return value.(*PlanState), true
	}
	return nil, false
}
func (pm *PlanManager) DeletePlan(planId string) {
	pm.Plans.Delete(planId)
}
