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
	"strconv"
	"strings"
	"sync"
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
	secret "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	log                 = logger.NewLogger("coa.runtime")
	lock                sync.Mutex
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
)

type SolutionManager struct {
	managers.Manager
	TargetProviders map[string]tgt.ITargetProvider
	StateProvider   states.IStateProvider
	ConfigProvider  config.IExtConfigProvider
	SecretProvider  secret.ISecretProvider
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
		err = json.Unmarshal(jData, &managerState)
		if err == nil {
			return &managerState
		}
	}
	log.InfofCtx(ctx, " M (Solution): failed to get previous state for instance %s in namespace %s: %+v", instance, namespace, err)
	return nil
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
	err = json.Unmarshal(jData, &result)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to deserailze deployment summary[%s]: %+v", key, err)
		return model.SummaryResult{}, err
	}

	return result, nil
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
		if v1alpha2.IsNotFound(err) {
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
	lock.Lock()
	defer lock.Unlock()

	defer func() {
		s.cleanupHeartbeat(ctx, deployment.Instance.ObjectMeta.Name, namespace, remove)
	}()

	stopCh := make(chan struct{})
	defer close(stopCh)
	go s.sendHeartbeat(ctx, deployment.Instance.ObjectMeta.Name, namespace, remove, stopCh)

	ctx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "Reconcile",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, " M (Solution): reconciling deployment.InstanceName: %s, deployment.SolutionName: %s, remove: %t, namespace: %s, targetName: %s",
		deployment.Instance.ObjectMeta.Name,
		deployment.SolutionName,
		remove,
		namespace,
		targetName)

	summary := model.SummarySpec{
		TargetResults:       make(map[string]model.TargetResultSpec),
		TargetCount:         len(deployment.Targets),
		SuccessCount:        0,
		AllAssignedDeployed: false,
	}

	deploymentType := DeploymentType_Update
	if remove {
		deploymentType = DeploymentType_Delete
	}
	summary.IsRemoval = remove

	s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
	defer func() {
		log.DebugfCtx(ctx, " M (Solution): Reconcile conclude Summary. Namespace: %v, deployment instance: %v, summary message: %v", namespace, deployment.Instance, summary.SummaryMessage)
		if deployment.IsDryRun {
			summary.SuccessCount = 0
		}
		s.concludeSummary(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
	}()

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
		desiredState = MergeDeploymentStates(&previousDesiredState.State, currentDesiredState)
	}

	if remove {
		desiredState.MarkRemoveAll()
	}

	mergedState := MergeDeploymentStates(&currentState, desiredState)
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
			targetSpec := s.getTargetStateForStep(step, deployment, previousDesiredState)
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
			testState := MergeDeploymentStates(&previousDesiredState.State, currentState)
			if s.canSkipStep(ctx, step, step.Target, provider.(tgt.ITargetProvider), previousDesiredState.State.Components, testState) {
				log.InfofCtx(ctx, " M (Solution): skipping step with role %s on target %s", step.Role, step.Target)
				targetResult[step.Target] = 1
				planSuccessCount++
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
				s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, deployment.Generation, deployment.Hash, summary, namespace)
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
			summary.SuccessCount = successCount
			summary.AllAssignedDeployed = plannedCount == planSuccessCount
			err = stepError
			return summary, err
		}
		planSuccessCount++
	}

	mergedState.ClearAllRemoved()

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
func (s *SolutionManager) getTargetStateForStep(step model.DeploymentStep, deployment model.DeploymentSpec, previousDeploymentState *SolutionManagerDeploymentState) model.TargetState {
	//first find the target spec in the deployment
	targetSpec, ok := deployment.Targets[step.Target]
	if !ok {
		if previousDeploymentState != nil {
			targetSpec = previousDeploymentState.Spec.Targets[step.Target]
		}
	}
	return targetSpec
}

func (s *SolutionManager) saveSummary(ctx context.Context, objectName string, generation string, hash string, summary model.SummarySpec, state model.SummaryState, namespace string) {
	// TODO: delete this state when time expires. This should probably be invoked by the vendor (via GetSummary method, for instance)
	log.DebugfCtx(ctx, " M (Solution): saving summary, objectName: %s, state: %s, namespace: %s, targetCount %d, successCount %d",
		objectName, state, namespace, summary.TargetCount, summary.SuccessCount)
	s.StateProvider.Upsert(ctx, states.UpsertRequest{
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
}

func (s *SolutionManager) saveSummaryProgress(ctx context.Context, objectName string, generation string, hash string, summary model.SummarySpec, namespace string) {
	s.saveSummary(ctx, objectName, generation, hash, summary, model.SummaryStateRunning, namespace)
}

func (s *SolutionManager) concludeSummary(ctx context.Context, objectName string, generation string, hash string, summary model.SummarySpec, namespace string) {
	s.saveSummary(ctx, objectName, generation, hash, summary, model.SummaryStateDone, namespace)
}

func (s *SolutionManager) canSkipStep(ctx context.Context, step model.DeploymentStep, target string, provider tgt.ITargetProvider, currentComponents []model.ComponentSpec, state model.DeploymentState) bool {

	for _, newCom := range step.Components {
		key := fmt.Sprintf("%s::%s", newCom.Component.Name, target)
		if newCom.Action == model.ComponentDelete {
			for _, c := range currentComponents {
				if c.Name == newCom.Component.Name && state.TargetComponent[key] != "" {
					return false // current component still exists, desired is to remove it. The step can't be skipped
				}
			}

		} else {
			found := false
			for _, c := range currentComponents {
				if c.Name == newCom.Component.Name && state.TargetComponent[key] != "" && !strings.HasPrefix(state.TargetComponent[key], "-") {
					found = true
					rule := provider.GetValidationRule(ctx)
					if rule.IsComponentChanged(c, newCom.Component) {
						return false // component has changed, can't skip the step
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
					err = json.Unmarshal(jData, &deployment)
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
func (s *SolutionManager) Reconcil() []error {
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
