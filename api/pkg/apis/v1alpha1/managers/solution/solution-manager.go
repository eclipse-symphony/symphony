/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	sp "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers"
	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	config "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/keylock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/queue"
	secret "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/google/uuid"
)

var (
	log                 = logger.NewLogger("coa.runtime")
	apiOperationMetrics *metrics.Metrics
)

const (
	SYMPHONY_AGENT string = "/symphony-agent:"
	ENV_NAME       string = "SYMPHONY_AGENT_ADDRESS"

	// DeploymentType_Update indicates the type of deployment is Update. This is
	// to give a deployment status on Symphony Target deployment.
	DeploymentType_Update string = "Target Update"
	// DeploymentType_Delete indicates the type of deployment is Delete. This is
	// to give a deployment status on Symphony Target deployment.
	DeploymentType_Delete string = "Target Delete"

	Summary         = "Summary"
	DeploymentState = "DeployState"
	DeploymentPlan  = "DeploymentPlan"
	OperationState  = "OperationState"
)

type SolutionManager struct {
	managers.Manager
	TargetProviders map[string]tgt.ITargetProvider
	StateProvider   states.IStateProvider
	ConfigProvider  config.IExtConfigProvider
	SecretProvider  secret.ISecretProvider
	KeyLockProvider keylock.IKeyLockProvider
	QueueProvider   queue.IQueueProvider
	IsTarget        bool
	TargetNames     []string
	ApiClientHttp   api_utils.ApiClient
}

type SolutionManagerDeploymentState struct {
	Spec  model.DeploymentSpec  `json:"spec,omitempty"`
	State model.DeploymentState `json:"state,omitempty"`
}

func (s *SolutionManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	s.TargetProviders = make(map[string]tgt.ITargetProvider)
	for k, v := range providers {
		if p, ok := v.(tgt.ITargetProvider); ok {
			s.TargetProviders[k] = p
		}
	}

	keylockprovider, err := managers.GetKeyLockProvider(config, providers)
	if err == nil {
		s.KeyLockProvider = keylockprovider

	} else {
		return err
	}

	stateprovider, err := managers.GetPersistentStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}

	configProvider, err := managers.GetExtConfigProvider(config, providers)
	if err == nil {
		s.ConfigProvider = configProvider
	} else {
		return err
	}

	secretProvider, err := managers.GetSecretProvider(config, providers)
	if err == nil {
		s.SecretProvider = secretProvider
	} else {
		return err
	}

	if v, ok := config.Properties["isTarget"]; ok {
		b, err := strconv.ParseBool(v)
		if err == nil || b {
			s.IsTarget = b
		}
	}

	targetNames := ""

	if v, ok := config.Properties["targetNames"]; ok {
		targetNames = v
	}
	sTargetName := os.Getenv("SYMPHONY_TARGET_NAME")
	if sTargetName != "" {
		targetNames = sTargetName
	}

	s.TargetNames = strings.Split(targetNames, ",")

	if s.IsTarget {
		if len(s.TargetNames) == 0 {
			return errors.New("target mode is set but target name is not set")
		}
	}

	if apiOperationMetrics == nil {
		apiOperationMetrics, err = metrics.New()
		if err != nil {
			return err
		}
	}
	s.ApiClientHttp, err = api_utils.GetParentApiClient(s.Context.SiteInfo.ParentSite.BaseUrl)
	if err != nil {
		return err
	}
	return nil
}
func (s *SolutionManager) AsyncReconcile(ctx context.Context, deployment model.DeploymentSpec, remove bool, namespace string, targetName string) (model.SummarySpec, error) {
	lockName := api_utils.GenerateKeyLockName(namespace, deployment.Instance.ObjectMeta.Name)
	// if !e.SolutionManager.KeyLockProvider.TryLock(api_utils.GenerateKeyLockName(namespace, deployment.Instance.ObjectMeta.Name)) {
	// 	log.Info("can not get lock %s", lockName)
	// }
	s.KeyLockProvider.Lock(lockName)
	log.InfofCtx(ctx, "lock succeed %s", lockName)

	log.InfoCtx(ctx, "get deployment %+v", deployment)
	log.InfofCtx(ctx, " M (Solution): reconciling deployment.InstanceName: %s, deployment.SolutionName: %s, remove: %t, namespace: %s, targetName: %s, generation: %s, jobID: %s",
		deployment.Instance.ObjectMeta.Name,
		deployment.SolutionName,
		remove,
		namespace,
		targetName,
		deployment.Generation,
		deployment.JobID)
	previousDesiredState := s.GetPreviousState(ctx, deployment.Instance.ObjectMeta.Name, namespace)
	// save summary
	summary := model.SummarySpec{
		TargetResults:       make(map[string]model.TargetResultSpec),
		TargetCount:         len(deployment.Targets),
		SuccessCount:        0,
		AllAssignedDeployed: false,
		JobID:               deployment.JobID,
	}
	// create new deployment state
	var state model.DeploymentState
	state, err := NewDeploymentState(deployment)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to create manager state for deployment: %+v", err)
		s.KeyLockProvider.UnLock(lockName)
		return summary, err
	}
	err = s.CheckJobId(ctx, deployment.JobID, namespace, deployment.Instance.ObjectMeta.Name)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): job id is less than exists for deployment: %+v", err)
		s.KeyLockProvider.UnLock(lockName)
		return summary, err
	}

	// stopCh := make(chan struct{})
	// defer close(stopCh)
	// go e.SolutionManager.SendHeartbeat(ctx, deployment.Instance.ObjectMeta.Name, namespace, remove, stopCh)

	// get the components count for the deployment
	componentCount := len(deployment.Solution.Spec.Components)
	apiOperationMetrics.ApiComponentCount(
		componentCount,
		metrics.ReconcileOperation,
		metrics.UpdateOperationType,
	)

	if s.VendorContext != nil && s.VendorContext.EvaluationContext != nil {
		context := s.VendorContext.EvaluationContext.Clone()
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
				s.concludeSummary(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
				log.InfoCtx(ctx, "unlock7")
				s.KeyLockProvider.UnLock(lockName)
				return summary, err
			}
		}

	}
	// e.SolutionManager.KeyLockProvider.Lock(api_utils.GenerateKeyLockName(namespace, deployment.Instance.ObjectMeta.Name))
	// Generate new deployment plan for deployment
	initalPlan, err := PlanForDeployment(deployment, state)
	if err != nil {
		s.concludeSummary(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
		log.ErrorfCtx(ctx, " M (Solution): failed initalPlan for deployment: %+v", err)
		s.KeyLockProvider.UnLock(lockName)
		return summary, err
	}

	// remove no use steps
	var stepList []model.DeploymentStep
	for _, step := range initalPlan.Steps {
		if s.IsTarget && !api_utils.ContainsString(s.TargetNames, step.Target) {
			continue
		}
		if targetName != "" && targetName != step.Target {
			continue
		}
		stepList = append(stepList, step)
	}
	initalPlan.Steps = stepList
	log.InfoCtx(ctx, "publish topic for object %s", deployment.Instance.ObjectMeta.Name)
	s.VendorContext.Publish(DeploymentPlanTopic, v1alpha2.Event{
		Metadata: map[string]string{
			"Id": deployment.JobID,
		},
		Body: PlanEnvelope{
			Plan:                 initalPlan,
			Deployment:           deployment,
			MergedState:          model.DeploymentState{},
			PreviousDesiredState: previousDesiredState,
			PlanId:               uuid.New().String(),
			Remove:               remove,
			Namespace:            namespace,
			Phase:                PhaseGet,
		},
		Context: ctx,
	})
	return summary, nil
}
func (s *SolutionManager) getPreviousState(ctx context.Context, instance string, namespace string) *SolutionManagerDeploymentState {
	state, err := s.StateProvider.Get(ctx, states.GetRequest{
		ID: instance,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  DeploymentState,
		},
	})
	if err == nil {
		var managerState SolutionManagerDeploymentState
		jData, _ := json.Marshal(state.Body)
		err = utils.UnmarshalJson(jData, &managerState)
		if err == nil {
			return &managerState
		}
	}
	log.InfofCtx(ctx, " M (Solution): failed to get previous state for instance %s in namespace %s: %+v", instance, namespace, err)
	return nil
}

func (s *SolutionManager) GetPreviousState(ctx context.Context, instance string, namespace string) SolutionManagerDeploymentState {
	state, err := s.StateProvider.Get(ctx, states.GetRequest{
		ID: instance,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  DeploymentState,
		},
	})
	if err == nil {
		var managerState SolutionManagerDeploymentState
		jData, _ := json.Marshal(state.Body)
		err = utils.UnmarshalJson(jData, &managerState)
		if err == nil {
			return managerState
		}
	}
	log.InfofCtx(ctx, " M (Solution): failed to get previous state for instance %s in namespace %s: %+v", instance, namespace, err)
	return SolutionManagerDeploymentState{}
}
func (s *SolutionManager) GetSummary(ctx context.Context, key string, namespace string) (model.SummaryResult, error) {
	// lock.Lock()
	// defer lock.Unlock()

	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "GetSummary",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, " M (Solution): get summary, key: %s, namespace: %s", key, namespace)

	var state states.StateEntry
	state, err = s.StateProvider.Get(ctx, states.GetRequest{
		ID: fmt.Sprintf("%s-%s", "summary", key),
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  Summary,
		},
	})
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to get deployment summary[%s]: %+v", key, err)
		return model.SummaryResult{}, err
	}

	var result model.SummaryResult
	jData, _ := json.Marshal(state.Body)
	err = utils.UnmarshalJson(jData, &result)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to deserailze deployment summary[%s]: %+v", key, err)
		return model.SummaryResult{}, err
	}

	return result, nil
}
func (s *SolutionManager) HandleDeploymentPlan(ctx context.Context, event v1alpha2.Event) error {
	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "HandleDeploymentPlan",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	var planEnvelope PlanEnvelope
	jData, _ := json.Marshal(event.Body)
	err = utils.UnmarshalJson(jData, &planEnvelope)
	if err != nil {
		log.ErrorCtx(ctx, "failed to unmarshal plan envelope :%v", err)
		return err
	}
	log.InfoCtx(ctx, "M(Solution): handle deployment plan %s", planEnvelope.PlanId)
	planState := createPlanState(planEnvelope)
	lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	tryLockresult := s.KeyLockProvider.TryLock(lockName)
	log.InfofCtx(ctx, "M (Solution): Try lock result %s", tryLockresult)
	err = s.CheckJobId(ctx, planState.Deployment.JobID, planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	if err != nil {
		s.KeyLockProvider.UnLock(lockName)
		return err
	}
	if err := s.saveSummaryInfo(ctx, planState, model.SummaryStateRunning); err != nil {
		return err
	}
	s.upsertPlanState(ctx, planEnvelope.PlanId, planState)
	if planState.isCompleted() {
		// no step to run
		return s.handlePlanComplete(ctx, planState)

	}
	// for _, step := range planEnvelope.Plan.Steps {
	// 	switch planEnvelope.Phase {
	// 	case PhaseApply:
	// 		planState.Summary.PlannedDeployment += len(step.Components)
	// 	}
	// }
	s.upsertPlanState(ctx, planEnvelope.PlanId, planState)
	switch planEnvelope.Phase {
	case PhaseGet:
		if err := s.publishDeploymentStep(ctx, 0, planState, planEnvelope.Remove, planState.Steps[0]); err != nil {
			log.InfofCtx(ctx, "failed to publish deployment step %s", err)
			// return err
		}
	case PhaseApply:
		for _, step := range planEnvelope.Plan.Steps {
			planState.Summary.PlannedDeployment += len(step.Components)
		}
		if err := s.publishDeploymentStep(ctx, 0, planState, planEnvelope.Remove, planState.Steps[0]); err != nil {
			log.InfofCtx(ctx, "failed to publish deployment step %s", err)
			return err
		}
	}
	log.InfoCtx(ctx, "M(Solution): store plan id %s in map %+v", planEnvelope.PlanId)
	return nil
}

// getOperationBody converts the body to an OperationBody.
func (c *SolutionManager) getOperationBody(body interface{}) (OperationBody, error) {
	var operationBody OperationBody
	bytes, _ := json.Marshal(body)
	err := utils.UnmarshalJson(bytes, &operationBody)
	if err != nil {
		return OperationBody{}, err
	}
	return operationBody, nil
}
func (s *SolutionManager) publishDeploymentStep(ctx context.Context, stepId int, planState PlanState, remove bool, step model.DeploymentStep) error {
	log.InfofCtx(ctx, "M(Solution): publish deployment step for PlanId %s StepId %s", planState.PlanId, stepId)
	if err := s.VendorContext.Publish(DeploymentStepTopic, v1alpha2.Event{
		Body: StepEnvelope{
			Step:      step,
			StepId:    stepId,
			Remove:    remove,
			PlanState: planState,
		},
		Metadata: map[string]string{
			"namespace": planState.Namespace,
		},
		Context: ctx,
	}); err != nil {
		log.InfoCtx(ctx, "M(Solution): publish deployment step failed PlanId %s, stepId %s", planState.PlanId, stepId)
		return err
	}
	return nil
}

// handlePlanComplete handles the completion of a plan and updates its status.
func (s *SolutionManager) handlePlanComplete(ctx context.Context, planState PlanState) error {
	log.InfofCtx(ctx, "M(Solution): Plan state %s is completed %s", planState.Phase, planState.PlanId)
	if !planState.Summary.AllAssignedDeployed {
		planState.Status = "failed"
	}
	switch planState.Phase {
	case PhaseGet:
		if err := s.handleGetPlanCompletetion(ctx, planState); err != nil {
			s.deletePlanState(ctx, planState.PlanId, planState.Namespace)
			return err
		}
	case PhaseApply:
		if err := s.handleAllPlanCompletetion(ctx, planState); err != nil {
			s.deletePlanState(ctx, planState.PlanId, planState.Namespace)
			return err
		}
		s.deletePlanState(ctx, planState.PlanId, planState.Namespace)
	}

	return nil
}

func (s *SolutionManager) handleAllPlanCompletetion(ctx context.Context, planState PlanState) error {
	log.InfofCtx(ctx, "M(Solution): Handle plan completetion:begin to handle plan completetion %v", planState)
	if err := s.saveSummaryInfo(ctx, planState, model.SummaryStateDone); err != nil {
		return err
	}
	// update summary
	planState.MergedState.ClearAllRemoved()
	if !planState.Deployment.IsDryRun {
		if len(planState.MergedState.TargetComponent) == 0 && planState.Remove {
			log.DebugfCtx(ctx, " M (Solution): no assigned components to manage, deleting state")
			s.StateProvider.Delete(ctx, states.DeleteRequest{
				ID: planState.Deployment.Instance.ObjectMeta.Name,
				Metadata: map[string]interface{}{
					"namespace": planState.Namespace,
					"group":     model.SolutionGroup,
					"version":   "v1",
					"resource":  DeploymentState,
				},
			})
		} else {
			s.StateProvider.Upsert(ctx, states.UpsertRequest{
				Value: states.StateEntry{
					ID: planState.Deployment.Instance.ObjectMeta.Name,
					Body: SolutionManagerDeploymentState{
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
	if err := s.concludeSummary(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Deployment.Generation, planState.Deployment.Hash, planState.Summary, planState.Namespace); err != nil {
		return err
	}
	lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	s.KeyLockProvider.UnLock(lockName)
	return nil
}

// threeStateMerge merges the current, previous, and desired states to create a deployment plan.
func (s *SolutionManager) threeStateMerge(ctx context.Context, planState PlanState) (model.DeploymentPlan, PlanState, error) {
	currentState := model.DeploymentState{}
	currentState.TargetComponent = make(map[string]string)

	for _, StepState := range planState.StepStates {
		for _, c := range StepState.GetResult {
			key := fmt.Sprintf("%s::%s", c.Name, StepState.Target)
			role := c.Type
			if role == "" {
				role = "instance"
			}
			currentState.TargetComponent[key] = role
		}
	}
	planState.CurrentState = currentState
	previousDesiredState := s.GetPreviousState(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Namespace)
	planState.PreviousDesiredState = previousDesiredState
	var currentDesiredState model.DeploymentState
	currentDesiredState, err := NewDeploymentState(planState.Deployment)
	if err != nil {
		log.ErrorfCtx(ctx, "M(Solution): Failed to get current desired state: %+v", err)
		return model.DeploymentPlan{}, PlanState{}, err
	}
	desiredState := currentDesiredState
	desiredState = MergeDeploymentStates(previousDesiredState.State, currentDesiredState)
	if planState.Remove {
		desiredState.MarkRemoveAll()
	}
	mergedState := MergeDeploymentStates(currentState, desiredState)
	log.InfofCtx(ctx, "M(Solution): Get Merged state %+v", mergedState)
	planState.MergedState = mergedState
	Plan, err := PlanForDeployment(planState.Deployment, mergedState)
	if err != nil {
		return model.DeploymentPlan{}, PlanState{}, err
	}
	s.upsertPlanState(ctx, planState.PlanId, planState)
	return Plan, planState, nil
}

// handleGetPlanCompletetion handles the completion of the get plan phase.
func (s *SolutionManager) handleGetPlanCompletetion(ctx context.Context, planState PlanState) error {
	// Collect result
	log.InfofCtx(ctx, "M(Solution): Handle get plan completetion:begin to handle get plan completetion %v", planState)
	Plan, planState, err := s.threeStateMerge(ctx, planState)
	if err != nil {
		log.ErrorfCtx(ctx, "M(Solution): Failed to merge states: %v", err)
		return err
	}
	s.VendorContext.Publish(DeploymentPlanTopic, v1alpha2.Event{
		Metadata: map[string]string{
			"Id":        planState.Deployment.JobID,
			"namespace": planState.Namespace,
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

// create inital plan state
func createPlanState(planEnvelope PlanEnvelope) PlanState {
	return PlanState{
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

func (p PlanState) isCompleted() bool {
	return p.CompletedSteps == p.TotalSteps
}

func (s *SolutionManager) HandleDeploymentStep(ctx context.Context, event v1alpha2.Event) error {
	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "HandleDeploymentStep",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	var stepEnvelope StepEnvelope
	jData, err := json.Marshal(event.Body)
	if err != nil {
		log.ErrorfCtx(ctx, "M (Solution): failed to unmarshal event body: %v", err)
		return err
	}
	if err := utils.UnmarshalJson(jData, &stepEnvelope); err != nil {
		log.ErrorfCtx(ctx, "M (Solution): failed to unmarshal step envelope: %v", err)
		return err
	}
	lockName := api_utils.GenerateKeyLockName(stepEnvelope.PlanState.Namespace, stepEnvelope.PlanState.Deployment.Instance.ObjectMeta.Name)
	tryLockresult := s.KeyLockProvider.TryLock(lockName)
	log.InfofCtx(ctx, "M (Solution): Try lock result %s", tryLockresult)
	err = s.CheckJobId(ctx, stepEnvelope.PlanState.Deployment.JobID, stepEnvelope.PlanState.Namespace, stepEnvelope.PlanState.Deployment.Instance.ObjectMeta.Name)
	if err != nil {
		s.KeyLockProvider.UnLock(lockName)
		return err
	}
	if stepEnvelope.Step.Role == "container" {
		stepEnvelope.Step.Role = "instance"
	}
	// Load the plan state object from the PlanManager
	planState, err := s.getPlanState(ctx, stepEnvelope.PlanState.PlanId, stepEnvelope.PlanState.Namespace)
	err = s.CheckJobId(ctx, planState.Deployment.JobID, planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): job id is out of data, step will not be executed: %+v", err)
		s.KeyLockProvider.UnLock(lockName)
		return err
	}

	if err != nil {
		return fmt.Errorf("Plan not found: %s", stepEnvelope.PlanState.PlanId)
	}
	stepEnvelope.PlanState = planState
	switch stepEnvelope.PlanState.Phase {
	case PhaseGet:
		return s.handlePhaseGet(ctx, stepEnvelope)
	case PhaseApply:
		return s.handlePhaseApply(ctx, stepEnvelope)
	}
	return nil
}

func (s *SolutionManager) handlePhaseGet(ctx context.Context, stepEnvelope StepEnvelope) error {
	if stepTargetIsRemoteTarget(stepEnvelope.PlanState.Deployment, stepEnvelope.Step.Target) {
		return s.enqueueProviderGetRequest(ctx, stepEnvelope)
	}
	return s.getProviderAndExecute(ctx, stepEnvelope)
}

func stepTargetIsRemoteTarget(deployment model.DeploymentSpec, targetName string) bool {
	// find targt component
	targetSpec := deployment.Targets[targetName]
	for _, component := range targetSpec.Spec.Components {
		if component.Type == "remote-agent" {
			return true
		}
	}
	return false
}
func (s *SolutionManager) enqueueProviderGetRequest(ctx context.Context, stepEnvelope StepEnvelope) error {
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

	log.InfofCtx(ctx, "M(Solution): Enqueue get message %s-%s %+v ", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace, providerGetRequest)
	messageID, err := s.QueueProvider.Enqueue(fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace), providerGetRequest)
	if err != nil {
		log.ErrorfCtx(ctx, "M(Solution): Error in enqueue message %s", fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace))
		return err
	}
	err = s.upsertOperationState(ctx, operationId, stepEnvelope.StepId, stepEnvelope.PlanState.PlanId, stepEnvelope.Step.Target, stepEnvelope.PlanState.Phase, stepEnvelope.PlanState.Namespace, stepEnvelope.Remove, messageID)
	if err != nil {
		log.ErrorfCtx(ctx, "M(Solution) Error in insert operation Id %s", operationId)
		return s.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{}, stepEnvelope.PlanState.Namespace)
	}
	return err
}

func (s *SolutionManager) getProviderAndExecute(ctx context.Context, stepEnvelope StepEnvelope) error {
	provider, err := s.GetTargetProviderForStep(stepEnvelope.Step.Target, stepEnvelope.Step.Role, stepEnvelope.PlanState.Deployment, &stepEnvelope.PlanState.PreviousDesiredState)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to create provider & Failed to save summary progress: %v", err)
		return s.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{}, stepEnvelope.PlanState.Namespace)
	}
	dep := stepEnvelope.PlanState.Deployment
	dep.ActiveTarget = stepEnvelope.Step.Target
	getResult, stepError := (provider.(tgt.ITargetProvider)).Get(ctx, dep, stepEnvelope.Step.Components)
	if stepError != nil {
		log.ErrorCtx(ctx, "M(Solution) Error in get target current states %+v", stepError)
		return s.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{}, stepEnvelope.PlanState.Namespace)
	}
	return s.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, getResult, map[string]model.ComponentResultSpec{}, stepEnvelope.PlanState.Namespace)
}

func (s *SolutionManager) handlePhaseApply(ctx context.Context, stepEnvelope StepEnvelope) error {
	if stepTargetIsRemoteTarget(stepEnvelope.PlanState.Deployment, stepEnvelope.Step.Target) {
		return s.enqueueProviderApplyRequest(ctx, stepEnvelope)
	}
	return s.applyProviderAndExecute(ctx, stepEnvelope)
}

func (s *SolutionManager) enqueueProviderApplyRequest(ctx context.Context, stepEnvelope StepEnvelope) error {
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
	messageId, err := s.QueueProvider.Enqueue(fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace), providApplyRequest)
	if err != nil {
		log.ErrorfCtx(ctx, "M(Solution): Error in enqueue apply message %s", fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace))
		return err
	}
	log.InfofCtx(ctx, "M(Solution): Enqueue apply message %s-%s %+v ", stepEnvelope.Step.Target, stepEnvelope.PlanState.Namespace, providApplyRequest)
	err = s.upsertOperationState(ctx, operationId, stepEnvelope.StepId, stepEnvelope.PlanState.PlanId, stepEnvelope.Step.Target, stepEnvelope.PlanState.Phase, stepEnvelope.PlanState.Namespace, stepEnvelope.Remove, messageId)
	if err != nil {
		log.ErrorfCtx(ctx, "error in insert operation Id %s", operationId)
		return s.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{}, stepEnvelope.PlanState.Namespace)
	}
	return err
}

func (s *SolutionManager) applyProviderAndExecute(ctx context.Context, stepEnvelope StepEnvelope) error {
	// get provider todo : is dry run
	provider, err := s.GetTargetProviderForStep(stepEnvelope.Step.Target, stepEnvelope.Step.Role, stepEnvelope.PlanState.Deployment, &stepEnvelope.PlanState.PreviousDesiredState)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to create provider: %v", err)
		return s.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, err, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{}, stepEnvelope.PlanState.Namespace)
	}
	previousDesiredState := stepEnvelope.PlanState.PreviousDesiredState
	currentState := stepEnvelope.PlanState.CurrentState
	step := stepEnvelope.Step
	testState := MergeDeploymentStates(previousDesiredState.State, currentState)
	if s.canSkipStep(ctx, step, step.Target, provider.(tgt.ITargetProvider), previousDesiredState.State.Components, testState) {
		log.InfofCtx(ctx, " M (Solution): skipping step with role %s on target %s", step.Role, step.Target)
		return s.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, nil, []model.ComponentSpec{}, map[string]model.ComponentResultSpec{}, stepEnvelope.PlanState.Namespace)
	}
	componentResults, stepError := (provider.(tgt.ITargetProvider)).Apply(ctx, stepEnvelope.PlanState.Deployment, stepEnvelope.Step, stepEnvelope.PlanState.Deployment.IsDryRun)
	return s.publishStepResult(ctx, stepEnvelope.Step.Target, stepEnvelope.PlanState.PlanId, stepEnvelope.StepId, stepError, []model.ComponentSpec{}, componentResults, stepEnvelope.PlanState.Namespace)
}
func (s *SolutionManager) publishStepResult(ctx context.Context, target string, planId string, stepId int, Error error, getResult []model.ComponentSpec, applyResult map[string]model.ComponentResultSpec, namespace string) error {
	errorString := ""
	if Error != nil {
		errorString = Error.Error()
	}
	return s.VendorContext.Publish(CollectStepResultTopic, v1alpha2.Event{
		Body: StepResult{
			Target:      target,
			PlanId:      planId,
			StepId:      stepId,
			GetResult:   getResult,
			ApplyResult: applyResult,
			Timestamp:   time.Now(),
			Error:       errorString,
			NameSpace:   namespace,
		},
		Metadata: map[string]string{
			"namespace": namespace,
		},
		Context: ctx,
	})
}

// handleStepResult processes the event and updates the plan state accordingly.
func (s *SolutionManager) HandleStepResult(ctx context.Context, event v1alpha2.Event) error {
	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "HandleStepResult",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	var stepResult StepResult
	// Marshal the event body to JSON
	jData, _ := json.Marshal(event.Body)
	log.InfofCtx(ctx, "Received event body: %s", string(jData))

	// Unmarshal the JSON data into stepResult
	if err := utils.UnmarshalJson(jData, &stepResult); err != nil {
		log.ErrorfCtx(ctx, "Failed to unmarshal step result: %v", err)
		return err
	}

	planId := stepResult.PlanId
	// Load the plan state object from the PlanManager
	planState, err := s.getPlanState(ctx, planId, stepResult.NameSpace)
	if err != nil {
		return fmt.Errorf("Plan not found: %s", planId)
	}
	// planState := planStateObj.(PlanState)
	lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	tryLockresult := s.KeyLockProvider.TryLock(lockName)
	log.InfofCtx(ctx, "M (Solution): Try lock result %s", tryLockresult)
	err = s.CheckJobId(ctx, planState.Deployment.JobID, planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	if err != nil {
		s.KeyLockProvider.UnLock(lockName)
		return err
	}
	// Update the plan state in the map and save the summary
	log.InfofCtx(ctx, "M(Solution): Handle step result for PlanId %s, StepId %d, Phase %s", planId, stepResult.StepId, planState.Phase)
	if err := s.saveStepResult(ctx, planState, stepResult); err != nil {
		log.ErrorCtx(ctx, "Failed to handle step result: %v", err)
		return err
	}
	return nil
}

// saveStepResult updates the plan state with the step result and saves the summary.
func (s *SolutionManager) saveStepResult(ctx context.Context, planState PlanState, stepResult StepResult) error {
	// Log the update of plan state with the step result
	if planState.Summary.TargetResults == nil {
		planState.Summary.TargetResults = make(map[string]model.TargetResultSpec)
	}
	lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
	tryLockresult := s.KeyLockProvider.TryLock(lockName)
	log.InfofCtx(ctx, "M (Solution): Try lock result %s", tryLockresult)
	log.InfofCtx(ctx, "M(Solution): Save step result for PlanId %s, StepId %d, Phase %s", planState.PlanId, stepResult.StepId, planState.Phase)
	switch planState.Phase {
	case PhaseGet:
		// Update the GetResult for the specific step
		planState.StepStates[stepResult.StepId].GetResult = stepResult.GetResult
		if stepResult.StepId != len(planState.Steps)-1 {
			if err := s.publishDeploymentStep(ctx, stepResult.StepId+1, planState, planState.Remove, planState.Steps[stepResult.StepId+1]); err != nil {
				log.ErrorfCtx(ctx, "M(Solution): publish deployment step failed PlanId %s, stepId %s", planState.PlanId, 0)
			}
		} else {
			err := s.handlePlanComplete(ctx, planState)
			if err != nil {
				log.ErrorfCtx(ctx, "M(Solution): Failed to handle plan completion: %v", err)
				lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
				s.KeyLockProvider.UnLock(lockName)
			}
		}
	case PhaseApply:
		planState.CompletedSteps++
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
			s.upsertPlanState(ctx, planState.PlanId, planState)
			// Save the summary information
			if err := s.saveSummaryInfo(ctx, planState, model.SummaryStateRunning); err != nil {
				log.ErrorfCtx(ctx, "Failed to save summary progress: %v", err)
			}
			return s.handleAllPlanCompletetion(ctx, planState)
		} else {
			// Handle success case and update the target result status and message
			targetResultSpec := model.TargetResultSpec{Status: "OK", Message: "", ComponentResults: stepResult.ApplyResult}
			planState.Summary.UpdateTargetResult(stepResult.Target, targetResultSpec)
			planState.Summary.CurrentDeployed += len(stepResult.ApplyResult)
			if planState.TargetResult[stepResult.Target] == 0 {
				planState.TargetResult[stepResult.Target] = 1
				planState.Summary.SuccessCount++
			}
			// publish next step execute event
			if stepResult.StepId != len(planState.Steps)-1 {
				if err := s.publishDeploymentStep(ctx, stepResult.StepId+1, planState, planState.Remove, planState.Steps[stepResult.StepId+1]); err != nil {
					log.ErrorfCtx(ctx, "M(Solution): publish deployment step failed PlanId %s, stepId %s", planState.PlanId, 0)
				}
				// Save the summary information
				if err := s.saveSummaryInfo(ctx, planState, model.SummaryStateRunning); err != nil {
					log.ErrorfCtx(ctx, "Failed to save summary progress: %v", err)
				}
			} else {
				// If no components are deployed, set success count to target count
				if planState.Summary.CurrentDeployed == 0 && planState.Summary.AllAssignedDeployed {
					planState.Summary.SuccessCount = planState.Summary.TargetCount
				}
				log.InfofCtx(ctx, "M(Solution): Plan state %s is completed %s", planState.Phase, planState.PlanId)
				err := s.handlePlanComplete(ctx, planState)
				if err != nil {
					log.ErrorfCtx(ctx, "M(Solution): Failed to handle plan completion: %v", err)
					lockName := api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name)
					s.KeyLockProvider.UnLock(lockName)
				}
				return nil
			}
		}
	}

	// Store the updated plan state
	s.upsertPlanState(ctx, planState.PlanId, planState)
	// Check if all steps are completed and handle plan completion
	return nil
}

// getTaskFromQueue retrieves a task from the queue for the specified target and namespace.
func (s *SolutionManager) GetTaskFromQueueByPaging(ctx context.Context, target string, namespace string, start string, size int) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doGetFromQueue",
	})
	queueName := fmt.Sprintf("%s-%s", target, namespace)
	log.InfofCtx(ctx, "M(SolutionVendor): getFromQueue %s queue length %s", queueName)
	defer span.End()
	var err error
	queueElement, lastMessageID, err := s.QueueProvider.QueryByPaging(queueName, start, size)
	var requestList []AgentRequest
	for _, element := range queueElement {
		var agentRequest AgentRequest
		err = utils.UnmarshalJson(element, &agentRequest)
		if err != nil {
			log.ErrorfCtx(ctx, "M(SolutionVendor): failed to unmarshal element - %s", err.Error())
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		requestList = append(requestList, agentRequest)
	}
	response := &ProviderPagingRequest{
		RequestList:   requestList,
		LastMessageID: lastMessageID,
	}
	if err != nil {
		log.ErrorfCtx(ctx, "M(SolutionVendor): getQueue failed - %s", err.Error())
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	data, _ := json.Marshal(response)
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}

func (s *SolutionManager) DeleteSummary(ctx context.Context, key string, namespace string) error {
	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "DeleteSummary",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, " M (Solution): delete summary, key: %s, namespace: %s", key, namespace)

	err = s.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: fmt.Sprintf("%s-%s", "summary", key),
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  Summary,
		},
	})

	if err != nil {
		if api_utils.IsNotFound(err) {
			log.DebugfCtx(ctx, " M (Solution): DeleteSummary NoutFound, id: %s, namespace: %s", key, namespace)
			return nil
		}
		log.ErrorfCtx(ctx, " M (Solution): failed to get summary[%s]: %+v", key, err)
		return err
	}

	return nil
}

func (s *SolutionManager) sendHeartbeat(ctx context.Context, id string, namespace string, remove bool, stopCh chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	action := v1alpha2.HeartBeatUpdate
	if remove {
		action = v1alpha2.HeartBeatDelete
	}

	for {
		select {
		case <-ticker.C:
			log.DebugfCtx(ctx, " M (Solution): sendHeartbeat, id: %s, namespace: %s, remove:%v", id, namespace, remove)
			s.VendorContext.Publish("heartbeat", v1alpha2.Event{
				Body: v1alpha2.HeartBeatData{
					JobId:     id,
					Scope:     namespace,
					Action:    action,
					Time:      time.Now().UTC(),
					JobAction: v1alpha2.JobUpdate,
				},
				Metadata: map[string]string{
					"namespace": namespace,
				},
				Context: ctx,
			})
		case <-stopCh:
			return // Exit the goroutine when the stop signal is received
		}
	}
}

// getTaskFromQueue retrieves a task from the queue for the specified target and namespace.
func (c *SolutionManager) GetTaskFromQueue(ctx context.Context, target string, namespace string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doGetFromQueue",
	})
	queueName := fmt.Sprintf("%s-%s", target, namespace)
	log.InfofCtx(ctx, "M(SolutionVendor): getFromQueue %s queue length %s", queueName)
	defer span.End()
	var queueElement interface{}
	var err error
	queueElement, err = c.QueueProvider.Peek(queueName)
	if err != nil {
		log.ErrorfCtx(ctx, "M(SolutionVendor): getQueue failed - %s", err.Error())
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

func (s *SolutionManager) cleanupHeartbeat(ctx context.Context, id string, namespace string, remove bool) {
	if !remove {
		return
	}

	log.DebugfCtx(ctx, " M (Solution): cleanupHeartbeat, id: %s, namespace: %s", id, namespace)
	s.VendorContext.Publish("heartbeat", v1alpha2.Event{
		Body: v1alpha2.HeartBeatData{
			JobId:     id,
			JobAction: v1alpha2.JobDelete,
		},
		Metadata: map[string]string{
			"namespace": namespace,
		},
		Context: ctx,
	})
}

func (s *SolutionManager) Reconcile(ctx context.Context, deployment model.DeploymentSpec, remove bool, namespace string, targetName string) (model.SummarySpec, error) {
	s.KeyLockProvider.Lock(api_utils.GenerateKeyLockName(namespace, deployment.Instance.ObjectMeta.Name)) // && used as split character
	defer s.KeyLockProvider.UnLock(api_utils.GenerateKeyLockName(namespace, deployment.Instance.ObjectMeta.Name))

	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "Reconcile",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, " M (Solution): reconciling deployment.InstanceName: %s, deployment.SolutionName: %s, remove: %t, namespace: %s, targetName: %s, generation: %s, jobID: %s",
		deployment.Instance.ObjectMeta.Name,
		deployment.SolutionName,
		remove,
		namespace,
		targetName,
		deployment.Generation,
		deployment.JobID)
	summary := model.SummarySpec{
		TargetResults:       make(map[string]model.TargetResultSpec),
		TargetCount:         len(deployment.Targets),
		SuccessCount:        0,
		AllAssignedDeployed: false,
		JobID:               deployment.JobID,
	}

	deploymentType := DeploymentType_Update
	if remove {
		deploymentType = DeploymentType_Delete
	}
	summary.IsRemoval = remove

	err = s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to save summary progress: %+v", err)
		return summary, err
	}
	defer func() {
		if r := recover(); r == nil {
			log.DebugfCtx(ctx, " M (Solution): Reconcile conclude Summary. Namespace: %v, deployment instance: %v, summary message: %v", namespace, deployment.Instance, summary.SummaryMessage)
			if deployment.IsDryRun {
				summary.SuccessCount = 0
			}
			s.concludeSummary(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
		} else {
			log.ErrorfCtx(ctx, " M (Solution): panic happens: %v", debug.Stack())
			panic(r)
		}
	}()

	defer func() {
		s.cleanupHeartbeat(ctx, deployment.Instance.ObjectMeta.Name, namespace, remove)
	}()

	stopCh := make(chan struct{})
	defer close(stopCh)
	go s.sendHeartbeat(ctx, deployment.Instance.ObjectMeta.Name, namespace, remove, stopCh)

	// get the components count for the deployment
	componentCount := len(deployment.Solution.Spec.Components)
	apiOperationMetrics.ApiComponentCount(
		componentCount,
		metrics.ReconcileOperation,
		metrics.UpdateOperationType,
	)

	if s.VendorContext != nil && s.VendorContext.EvaluationContext != nil {
		context := s.VendorContext.EvaluationContext.Clone()
		context.DeploymentSpec = deployment
		context.Value = deployment
		context.Component = ""
		context.Namespace = namespace
		context.Context = ctx
		deployment, err = api_utils.EvaluateDeployment(*context)
	}

	if err != nil {
		if remove {
			log.InfofCtx(ctx, " M (Solution): skipped failure to evaluate deployment spec: %+v", err)
		} else {
			summary.SummaryMessage = "failed to evaluate deployment spec: " + err.Error()
			log.ErrorfCtx(ctx, " M (Solution): failed to evaluate deployment spec: %+v", err)
			return summary, err
		}
	}

	previousDesiredState := s.getPreviousState(ctx, deployment.Instance.ObjectMeta.Name, namespace)

	var currentDesiredState, currentState model.DeploymentState
	currentDesiredState, err = NewDeploymentState(deployment)
	if err != nil {
		summary.SummaryMessage = "failed to create target manager state from deployment spec: " + err.Error()
		log.ErrorfCtx(ctx, " M (Solution): failed to create target manager state from deployment spec: %+v", err)
		return summary, err
	}
	currentState, _, err = s.Get(ctx, deployment, targetName)
	if err != nil {
		summary.SummaryMessage = "failed to get current state: " + err.Error()
		log.ErrorfCtx(ctx, " M (Solution): failed to get current state: %+v", err)
		return summary, err
	}
	desiredState := currentDesiredState
	if previousDesiredState != nil {
		desiredState = MergeDeploymentStates(previousDesiredState.State, currentDesiredState)
	}

	if remove {
		desiredState.MarkRemoveAll()
	}

	mergedState := MergeDeploymentStates(currentState, desiredState)
	var plan model.DeploymentPlan
	plan, err = PlanForDeployment(deployment, mergedState)
	if err != nil {
		summary.SummaryMessage = "failed to plan for deployment: " + err.Error()
		log.ErrorfCtx(ctx, " M (Solution): failed to plan for deployment: %+v", err)
		return summary, err
	}

	col := api_utils.MergeCollection(deployment.Solution.Spec.Metadata, deployment.Instance.Spec.Metadata)
	dep := deployment
	dep.Instance.Spec.Metadata = col
	someStepsRan := false

	targetResult := make(map[string]int)

	summary.PlannedDeployment = 0
	for _, step := range plan.Steps {
		summary.PlannedDeployment += len(step.Components)
	}
	summary.CurrentDeployed = 0
	err = s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to save summary progress: %+v", err)
		return summary, err
	}
	log.DebugfCtx(ctx, " M (Solution): reconcile save summary progress: start deploy, total %v deployments", summary.PlannedDeployment)
	// DO NOT REMOVE THIS COMMENT
	// gofail: var beforeProviders string

	plannedCount := 0
	planSuccessCount := 0
	for _, step := range plan.Steps {
		log.DebugfCtx(ctx, " M (Solution): processing step with Role %s on target %s", step.Role, step.Target)
		for _, component := range step.Components {
			log.DebugfCtx(ctx, " M (Solution): processing component %s with action %s", component.Component.Name, component.Action)
		}
		if s.IsTarget && !api_utils.ContainsString(s.TargetNames, step.Target) {
			continue
		}

		if targetName != "" && targetName != step.Target {
			continue
		}

		plannedCount++

		dep.ActiveTarget = step.Target
		agent := findAgentFromDeploymentState(mergedState, step.Target)
		if agent != "" {
			col[ENV_NAME] = agent
		} else {
			delete(col, ENV_NAME)
		}
		var override tgt.ITargetProvider
		role := step.Role
		if role == "container" {
			role = "instance"
		}
		if v, ok := s.TargetProviders[role]; ok {
			override = v
		}
		var provider providers.IProvider
		if override == nil {
			targetSpec := s.getTargetStateForStep(step.Target, deployment, previousDesiredState)
			provider, err = sp.CreateProviderForTargetRole(s.Context, step.Role, targetSpec, override)
			if err != nil {
				summary.SummaryMessage = "failed to create provider:" + err.Error()
				log.ErrorfCtx(ctx, " M (Solution): failed to create provider: %+v", err)
				return summary, err
			}
		} else {
			provider = override
		}

		if previousDesiredState != nil {
			testState := MergeDeploymentStates(previousDesiredState.State, currentState)
			if s.canSkipStep(ctx, step, step.Target, provider.(tgt.ITargetProvider), previousDesiredState.State.Components, testState) {
				log.InfofCtx(ctx, " M (Solution): skipping step with role %s on target %s", step.Role, step.Target)
				targetResult[step.Target] = 1
				planSuccessCount++
				summary.CurrentDeployed += len(step.Components)
				continue
			}
		}
		log.DebugfCtx(ctx, " M (Solution): applying step with Role %s on target %s", step.Role, step.Target)
		someStepsRan = true
		retryCount := 1
		//TODO: set to 1 for now. Although retrying can help to handle transient errors, in more cases
		// an error condition can't be resolved quickly.
		var stepError error
		var componentResults map[string]model.ComponentResultSpec

		// for _, component := range step.Components {
		// 	for k, v := range component.Component.Properties {
		// 		if strV, ok := v.(string); ok {
		// 			parser := api_utils.NewParser(strV)
		// 			eCtx := s.VendorContext.EvaluationContext.Clone()
		// 			eCtx.DeploymentSpec = deployment
		// 			eCtx.Component = component.Component.Name
		// 			val, err := parser.Eval(*eCtx)
		// 			if err == nil {
		// 				component.Component.Properties[k] = val
		// 			} else {
		// 				log.ErrorfCtx(ctx, " M (Solution): failed to evaluate property: %+v", err)
		// 				summary.SummaryMessage = fmt.Sprintf("failed to evaluate property '%s' on component '%s: %s", k, component.Component.Name, err.Error())
		// 				s.saveSummary(ctx, deployment, summary)
		// 				return summary, err
		// 			}
		// 		}
		// 	}
		// }

		for i := 0; i < retryCount; i++ {
			componentResults, stepError = (provider.(tgt.ITargetProvider)).Apply(ctx, dep, step, deployment.IsDryRun)
			if stepError == nil {
				targetResult[step.Target] = 1
				summary.AllAssignedDeployed = plannedCount == planSuccessCount
				summary.UpdateTargetResult(step.Target, model.TargetResultSpec{Status: "OK", Message: "", ComponentResults: componentResults})
				err = s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
				if err != nil {
					log.ErrorfCtx(ctx, " M (Solution): failed to save summary progress: %+v", err)
					return summary, err
				}
				break
			} else {
				targetResult[step.Target] = 0
				summary.AllAssignedDeployed = false
				targetResultStatus := fmt.Sprintf("%s Failed", deploymentType)
				targetResultMessage := fmt.Sprintf("An error occurred in %s, err: %s", deploymentType, stepError.Error())
				summary.UpdateTargetResult(step.Target, model.TargetResultSpec{Status: targetResultStatus, Message: targetResultMessage, ComponentResults: componentResults}) // TODO: this keeps only the last error on the target
				time.Sleep(5 * time.Second)                                                                                                                                   //TODO: make this configurable?
			}
		}
		if stepError != nil {
			log.ErrorfCtx(ctx, " M (Solution): failed to execute deployment step: %+v", stepError)

			successCount := 0
			for _, v := range targetResult {
				successCount += v
			}
			deployedCount := 0
			for _, ret := range componentResults {
				if (!remove && ret.Status == v1alpha2.Updated) || (remove && ret.Status == v1alpha2.Deleted) {
					// TODO: need to ensure the status updated correctly on returning from target providers.
					deployedCount += 1
				}
			}
			summary.CurrentDeployed += deployedCount
			summary.SuccessCount = successCount
			summary.AllAssignedDeployed = plannedCount == planSuccessCount
			err = stepError
			return summary, err
		}
		planSuccessCount++
		summary.CurrentDeployed += len(step.Components)
		err = s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
		if err != nil {
			log.ErrorfCtx(ctx, " M (Solution): failed to save summary progress: %+v", err)
			return summary, err
		}
		log.DebugfCtx(ctx, " M (Solution): reconcile save summary progress: current deployed %v out of total %v deployments", summary.CurrentDeployed, summary.PlannedDeployment)
	}

	mergedState.ClearAllRemoved()

	// DO NOT REMOVE THIS COMMENT
	// gofail: var beforeDeploymentError string

	if !deployment.IsDryRun {
		if len(mergedState.TargetComponent) == 0 && remove {
			log.DebugfCtx(ctx, " M (Solution): no assigned components to manage, deleting state")
			s.StateProvider.Delete(ctx, states.DeleteRequest{
				ID: deployment.Instance.ObjectMeta.Name,
				Metadata: map[string]interface{}{
					"namespace": namespace,
					"group":     model.SolutionGroup,
					"version":   "v1",
					"resource":  DeploymentState,
				},
			})
		} else {
			s.StateProvider.Upsert(ctx, states.UpsertRequest{
				Value: states.StateEntry{
					ID: deployment.Instance.ObjectMeta.Name,
					Body: SolutionManagerDeploymentState{
						Spec:  deployment,
						State: mergedState,
					},
				},
				Metadata: map[string]interface{}{
					"namespace": namespace,
					"group":     model.SolutionGroup,
					"version":   "v1",
					"resource":  DeploymentState,
				},
			})
		}
	}

	// DO NOT REMOVE THIS COMMENT
	// gofail: var afterDeploymentError string

	successCount := 0
	for _, v := range targetResult {
		successCount += v
	}
	summary.SuccessCount = successCount
	summary.AllAssignedDeployed = plannedCount == planSuccessCount

	// if solutions.components are empty,
	// we need to set summary.Skipped = true
	// and summary.SuccessCount = summary.TargetCount (instance_controller and target_controller will check whether targetCount == successCount in deletion case)
	summary.Skipped = !someStepsRan
	if summary.Skipped {
		summary.SuccessCount = summary.TargetCount
	}

	return summary, nil
}

// The deployment spec may have changed, so the previous target is not in the new deployment anymore
func (s *SolutionManager) getTargetStateForStep(target string, deployment model.DeploymentSpec, previousDeploymentState *SolutionManagerDeploymentState) model.TargetState {
	//first find the target spec in the deployment
	targetSpec, ok := deployment.Targets[target]
	if !ok {
		if previousDeploymentState != nil {
			targetSpec = previousDeploymentState.Spec.Targets[target]
		}
	}
	return targetSpec
}

// The deployment spec may have changed, so the previous target is not in the new deployment anymore
func (s *SolutionManager) GetTargetProviderForStep(target string, role string, deployment model.DeploymentSpec, previousDesiredState *SolutionManagerDeploymentState) (providers.IProvider, error) {
	var override tgt.ITargetProvider
	if role == "container" {
		role = "instance"
	}
	log.Info("get target providers %+v", s.TargetProviders)
	if v, ok := s.TargetProviders[role]; ok {
		return v, nil
	}
	targetSpec := s.getTargetStateForStep(target, deployment, previousDesiredState)
	provider, err := sp.CreateProviderForTargetRole(s.Context, role, targetSpec, override)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
func (s *SolutionManager) CheckJobId(ctx context.Context, jobID string, namespace string, objectName string) error {
	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "CheckJobId",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	oldSummary, err := s.GetSummary(ctx, objectName, namespace)
	if err != nil && !v1alpha2.IsNotFound(err) {
		log.ErrorfCtx(ctx, " M (Solution): failed to get previous summary: %+v", err)
	} else if err == nil {
		if jobID != "" && oldSummary.Summary.JobID != "" {
			var newId, oldId int64
			newId, err = strconv.ParseInt(jobID, 10, 64)
			if err != nil {
				log.ErrorfCtx(ctx, " M (Solution): failed to parse new job id: %+v", err)
				return v1alpha2.NewCOAError(err, "failed to parse new job id", v1alpha2.BadRequest)
			}
			oldId, err = strconv.ParseInt(oldSummary.Summary.JobID, 10, 64)
			if err == nil && oldId > newId {
				errMsg := fmt.Sprintf("old job id %d is greater than new job id %d", oldId, newId)
				log.ErrorfCtx(ctx, " M (Solution): %s", errMsg)
				return v1alpha2.NewCOAError(err, errMsg, v1alpha2.BadRequest)
			}
		} else {
			log.WarnfCtx(ctx, " M (Solution): JobIDs are both empty, skip id check")
		}
	}
	return nil
}

func (s *SolutionManager) saveSummaryInfo(ctx context.Context, planState PlanState, state model.SummaryState) error {
	return s.saveSummary(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Deployment.Generation, planState.Deployment.Hash, planState.Summary, state, planState.Namespace)
}

func (s *SolutionManager) saveSummary(ctx context.Context, objectName string, generation string, hash string, summary model.SummarySpec, state model.SummaryState, namespace string) error {
	// TODO: delete this state when time expires. This should probably be invoked by the vendor (via GetSummary method, for instance)
	log.DebugfCtx(ctx, " M (Solution): saving summary, objectName: %s, state: %s, namespace: %s, jobid: %s, hash %s, targetCount %d, successCount %d",
		objectName, state, namespace, summary.JobID, hash, summary.TargetCount, summary.SuccessCount)
	oldSummary, err := s.GetSummary(ctx, objectName, namespace)
	if err != nil && !v1alpha2.IsNotFound(err) {
		log.ErrorfCtx(ctx, " M (Solution): failed to get previous summary: %+v", err)
		return err
	} else if err == nil {
		if summary.JobID != "" && oldSummary.Summary.JobID != "" {
			var newId, oldId int64
			newId, err = strconv.ParseInt(summary.JobID, 10, 64)
			if err != nil {
				log.ErrorfCtx(ctx, " M (Solution): failed to parse new job id: %+v", err)
				return v1alpha2.NewCOAError(err, "failed to parse new job id", v1alpha2.BadRequest)
			}
			oldId, err = strconv.ParseInt(oldSummary.Summary.JobID, 10, 64)
			if err == nil && oldId > newId {
				errMsg := fmt.Sprintf("old job id %d is greater than new job id %d", oldId, newId)
				log.ErrorfCtx(ctx, " M (Solution): %s", errMsg)
				return v1alpha2.NewCOAError(err, errMsg, v1alpha2.BadRequest)
			}
		} else {
			log.WarnfCtx(ctx, " M (Solution): JobIDs are both empty, skip id check")
		}
	}
	_, err = s.StateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID: fmt.Sprintf("%s-%s", "summary", objectName),
			Body: model.SummaryResult{
				Summary:        summary,
				Generation:     generation,
				Time:           time.Now().UTC(),
				State:          state,
				DeploymentHash: hash,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  Summary,
		},
	})
	return err
}

func (s *SolutionManager) saveSummaryProgress(ctx context.Context, objectName string, generation string, hash string, summary model.SummarySpec, namespace string) error {
	return s.saveSummary(ctx, objectName, generation, hash, summary, model.SummaryStateRunning, namespace)
}

func (s *SolutionManager) concludeSummary(ctx context.Context, objectName string, generation string, hash string, summary model.SummarySpec, namespace string) error {
	return s.saveSummary(ctx, objectName, generation, hash, summary, model.SummaryStateDone, namespace)
}

// handleRemoteAgentExecuteResult handles the execution result from the remote agent.
func (s *SolutionManager) HandleRemoteAgentExecuteResult(ctx context.Context, asyncResult AsyncResult) v1alpha2.COAResponse {
	// Get operation ID
	operationId := asyncResult.OperationID
	// Get related info from redis - todo: timeout
	log.InfoCtx(ctx, "M(SolutionVendor): handle remote agent request %+v", asyncResult)
	operationBody, err := s.getOperationState(ctx, operationId, asyncResult.Namespace)
	if err != nil {
		log.ErrorfCtx(ctx, "M(SolutionVendor): onGetResponse failed - %s", err.Error())
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
		err := utils.UnmarshalJson(asyncResult.Body, &response)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		s.publishStepResult(ctx, operationBody.Target, operationBody.PlanId, operationBody.StepId, asyncResult.Error, response, map[string]model.ComponentResultSpec{}, operationBody.NameSpace)
		s.deleteOperationState(ctx, operationId)
		if err != nil {
			return v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"405 - delete operation Id failed\"}"),
				ContentType: "application/json",
			}
		}
		// delete from queue

		s.QueueProvider.RemoveFromQueue(queueName, operationBody.MessageId)
		return v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte("{\"result\":\"200 - handle async result successfully\"}"),
			ContentType: "application/json",
		}
	case PhaseApply:
		var response map[string]model.ComponentResultSpec
		err := utils.UnmarshalJson(asyncResult.Body, &response)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		s.publishStepResult(ctx, operationBody.Target, operationBody.PlanId, operationBody.StepId, asyncResult.Error, []model.ComponentSpec{}, response, operationBody.NameSpace)
		s.deleteOperationState(ctx, operationId)
		// delete from queue
		s.QueueProvider.RemoveFromQueue(queueName, operationBody.MessageId)
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
func (s *SolutionManager) canSkipStep(ctx context.Context, step model.DeploymentStep, target string, provider tgt.ITargetProvider, previousComponents []model.ComponentSpec, currentState model.DeploymentState) bool {

	for _, newCom := range step.Components {
		key := fmt.Sprintf("%s::%s", newCom.Component.Name, target)
		if newCom.Action == model.ComponentDelete {
			for _, c := range previousComponents {
				if c.Name == newCom.Component.Name && currentState.TargetComponent[key] != "" {
					return false // current component still exists, desired is to remove it. The step can't be skipped
				}
			}

		} else {
			found := false
			for _, c := range previousComponents {
				if c.Name == newCom.Component.Name && currentState.TargetComponent[key] != "" && !strings.HasPrefix(currentState.TargetComponent[key], "-") {
					found = true
					rule := provider.GetValidationRule(ctx)
					for _, sc := range currentState.Components {
						if sc.Name == c.Name {
							if rule.IsComponentChanged(c, newCom.Component) || rule.IsComponentChanged(sc, newCom.Component) {
								return false // component has changed, can't skip the step
							}
							break
						}
					}
					break
				}
			}
			if !found {
				return false //current component doesn't exist, desired is to update it. The step can't be skipped
			}
		}
	}
	return true
}
func (s *SolutionManager) Get(ctx context.Context, deployment model.DeploymentSpec, targetName string) (model.DeploymentState, []model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, " M (Solution): getting deployment.InstanceName: %s, deployment.SolutionName: %s, targetName: %s",
		deployment.Instance.ObjectMeta.Name,
		deployment.SolutionName,
		targetName)

	ret := model.DeploymentState{}

	var state model.DeploymentState
	state, err = NewDeploymentState(deployment)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to create manager state for deployment: %+v", err)
		return ret, nil, err
	}
	var plan model.DeploymentPlan
	plan, err = PlanForDeployment(deployment, state)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to plan for deployment: %+v", err)
		return ret, nil, err
	}
	ret = state
	ret.TargetComponent = make(map[string]string)
	retComponents := make([]model.ComponentSpec, 0)

	for _, step := range plan.Steps {
		if s.IsTarget && !api_utils.ContainsString(s.TargetNames, step.Target) {
			continue
		}
		if targetName != "" && targetName != step.Target {
			continue
		}

		deployment.ActiveTarget = step.Target

		var override tgt.ITargetProvider
		role := step.Role
		if role == "container" {
			role = "instance"
		}
		if v, ok := s.TargetProviders[role]; ok {
			override = v
		}
		var provider providers.IProvider

		if override == nil {
			provider, err = sp.CreateProviderForTargetRole(s.Context, step.Role, deployment.Targets[step.Target], override)
			if err != nil {
				log.ErrorfCtx(ctx, " M (Solution): failed to create provider: %+v", err)
				return ret, nil, err
			}
		} else {
			provider = override
		}
		var components []model.ComponentSpec
		components, err = (provider.(tgt.ITargetProvider)).Get(ctx, deployment, step.Components)

		if err != nil {
			log.WarnfCtx(ctx, " M (Solution): failed to get components: %+v", err)
			return ret, nil, err
		}
		for _, c := range components {
			key := fmt.Sprintf("%s::%s", c.Name, step.Target)
			role := c.Type
			if role == "" {
				role = "container"
			}
			ret.TargetComponent[key] = role
			found := false
			for _, rc := range retComponents {
				if rc.Name == c.Name {
					found = true
					break
				}
			}
			if !found {
				retComponents = append(retComponents, c)
			}
		}
	}
	ret.Components = retComponents
	return ret, retComponents, nil
}
func (s *SolutionManager) Enabled() bool {
	return s.Config.Properties["poll.enabled"] == "true"
}
func (s *SolutionManager) Poll() []error {
	if s.Config.Properties["poll.enabled"] == "true" && s.Context.SiteInfo.ParentSite.BaseUrl != "" && s.IsTarget {
		for _, target := range s.TargetNames {
			catalogs, err := s.ApiClientHttp.GetCatalogsWithFilter(context.Background(), "", "label", "staged_target="+target,
				s.Context.SiteInfo.ParentSite.Username,
				s.Context.SiteInfo.ParentSite.Password)
			if err != nil {
				return []error{err}
			}
			for _, c := range catalogs {
				if vs, ok := c.Spec.Properties["deployment"]; ok {
					deployment := model.DeploymentSpec{}
					jData, _ := json.Marshal(vs)
					err = utils.UnmarshalJson(jData, &deployment)
					if err != nil {
						return []error{err}
					}
					isRemove := false
					if v, ok := c.Spec.Properties["staged"]; ok {
						if vd, ok := v.(map[string]interface{}); ok {
							if v, ok := vd["removed-components"]; ok && v != nil {
								if len(v.([]interface{})) > 0 {
									isRemove = true
								}
							}
						}
					}
					_, err := s.Reconcile(context.Background(), deployment, isRemove, c.ObjectMeta.Namespace, target)
					if err != nil {
						return []error{err}
					}
					_, components, err := s.Get(context.Background(), deployment, target)
					if err != nil {
						return []error{err}
					}
					err = s.ApiClientHttp.ReportCatalogs(context.Background(),
						deployment.Instance.ObjectMeta.Name+"-"+target,
						components,
						s.Context.SiteInfo.ParentSite.Username,
						s.Context.SiteInfo.ParentSite.Password)
					if err != nil {
						return []error{err}
					}
				}
			}
		}
	}
	return nil
}

func findAgentFromDeploymentState(state model.DeploymentState, targetName string) string {
	for _, targetDes := range state.Targets {
		if targetName == targetDes.Name {
			for _, c := range targetDes.Spec.Components {
				if v, ok := c.Properties[model.ContainerImage]; ok {
					if strings.Contains(fmt.Sprintf("%v", v), SYMPHONY_AGENT) {
						return c.Name
					}
				}
			}
		}
	}
	return ""
}
func sortByDepedencies(components []model.ComponentSpec) ([]model.ComponentSpec, error) {
	size := len(components)
	inDegrees := make([]int, size)
	queue := make([]int, 0)
	for i, c := range components {
		inDegrees[i] = len(c.Dependencies)
		if inDegrees[i] == 0 {
			queue = append(queue, i)
		}
	}
	ret := make([]model.ComponentSpec, 0)
	for len(queue) > 0 {
		ret = append(ret, components[queue[0]])
		queue = queue[1:]
		for i, c := range components {
			found := false
			for _, d := range c.Dependencies {
				if d == ret[len(ret)-1].Name {
					found = true
					break
				}
			}
			if found {
				inDegrees[i] -= 1
				if inDegrees[i] == 0 {
					queue = append(queue, i)
				}
			}
		}
	}
	if len(ret) != size {
		return nil, errors.New("circular dependencies or unresolved dependencies detected in components")
	}
	return ret, nil
}

func (s *SolutionManager) upsertPlanState(ctx context.Context, planId string, planState PlanState) error {
	log.InfofCtx(ctx, "M(Solution): upsert plan state for %s", planId)
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   planId,
			Body: planState,
		},
		Metadata: map[string]interface{}{
			"namespace": planState.Namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  DeploymentPlan,
		},
	}
	_, err := s.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

// upsertOperationState upserts the operation state for the specified parameters.
func (s *SolutionManager) upsertOperationState(ctx context.Context, operationId string, stepId int, planId string, target string, action JobPhase, namespace string, remove bool, messageId string) error {
	log.InfoCtx(ctx, "upsert operationid %s", operationId)
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
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  OperationState,
		},
	}
	_, err := s.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

func (s *SolutionManager) deleteOperationState(ctx context.Context, operationId string) error {
	log.InfoCtx(ctx, "delete operationid %s", operationId)
	deleteRequest := states.DeleteRequest{
		ID: operationId,
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  OperationState,
		},
	}
	err := s.StateProvider.Delete(ctx, deleteRequest)
	return err
}

// upsertOperationState upserts the operation state for the specified parameters.
func (s *SolutionManager) deletePlanState(ctx context.Context, planId string, namespace string) error {
	log.InfofCtx(ctx, "delete plan %s", planId)
	deleteRequest := states.DeleteRequest{
		ID: planId,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  DeploymentPlan,
		},
	}
	err := s.StateProvider.Delete(ctx, deleteRequest)
	return err
}

// getPlanState retrieves the operation state for the specified operation ID.
func (s *SolutionManager) getPlanState(ctx context.Context, planId string, namespace string) (PlanState, error) {
	log.InfofCtx(ctx, "M(Solution) : Get plan %s", planId)
	getRequest := states.GetRequest{
		ID: planId,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  DeploymentPlan,
		},
	}
	var entry states.StateEntry
	entry, err := s.StateProvider.Get(ctx, getRequest)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): Failed to get deployment planstate[%s]: %+v", planId, err)
		return PlanState{}, err
	}
	var ret PlanState
	ret, err = s.getPlanStateBody(entry.Body)
	if err != nil {
		log.ErrorfCtx(ctx, "M(SolutionVendor): Failed to convert to operation state for %s", planId)
		return PlanState{}, err
	}
	return ret, err
}
func (s *SolutionManager) getPlanStateBody(body interface{}) (PlanState, error) {
	var planState PlanState
	bytes, _ := json.Marshal(body)
	err := utils.UnmarshalJson(bytes, &planState)
	if err != nil {
		return PlanState{}, err
	}
	return planState, nil
}

// getOperationState retrieves the operation state for the specified operation ID.
func (s *SolutionManager) getOperationState(ctx context.Context, operationId string, namespace string) (OperationBody, error) {
	getRequest := states.GetRequest{
		ID: operationId,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  OperationState,
		},
	}
	var entry states.StateEntry
	entry, err := s.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return OperationBody{}, err
	}
	var ret OperationBody
	ret, err = s.getOperationBody(entry.Body)
	if err != nil {
		log.ErrorfCtx(ctx, "M(SolutionVendor): Failed to convert to operation state for %s", operationId)
		return OperationBody{}, err
	}
	return ret, err
}
