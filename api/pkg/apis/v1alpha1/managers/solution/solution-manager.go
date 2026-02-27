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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/keylock"
	secret "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	log                 = logger.NewLogger("coa.runtime")
	jobListLock         sync.RWMutex
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

	defaultTimeout = 60 * time.Minute
)

type JobIdentifier struct {
	Namespace string
	Name      string
}

type SolutionManager struct {
	SummaryManager
	TargetProviders map[string]tgt.ITargetProvider
	ConfigProvider  config.IExtConfigProvider
	SecretProvider  secret.ISecretProvider
	KeyLockProvider keylock.IKeyLockProvider
	IsTarget        bool
	TargetNames     []string
	ApiClientHttp   api_utils.ApiClient
	jobList         map[JobIdentifier]map[string]context.CancelFunc
}

type SolutionManagerDeploymentState struct {
	Spec  model.DeploymentSpec  `json:"spec,omitempty"`
	State model.DeploymentState `json:"state,omitempty"`
}

func (s *SolutionManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.SummaryManager.Init(context, config, providers)
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

	s.initCancelMap()
	return nil
}

func (s *SolutionManager) initCancelMap() {
	s.jobList = make(map[JobIdentifier]map[string]context.CancelFunc)
}

func (s *SolutionManager) GetSummary(ctx context.Context, summaryId string, name string, namespace string) (model.SummaryResult, error) {
	return s.SummaryManager.GetSummary(ctx, fmt.Sprintf("%s-%s", "summary", summaryId), name, namespace)
}

func (s *SolutionManager) DeleteSummary(ctx context.Context, summaryId string, namespace string) error {
	// Slient side delete summary is soft delete: will only add a deleted flag.
	return s.SummaryManager.DeleteSummary(ctx, summaryId, namespace, true)
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

	if deployment.IsInActive {
		log.InfofCtx(ctx, " M (Solution): deployment is not active, remove the deployment")
		remove = true
	}
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
	summaryId := deployment.Instance.ObjectMeta.GetSummaryId()
	if summaryId == "" {
		log.ErrorfCtx(ctx, " M (Solution): object GUID is null: %+v", err)
		return summary, err
	}

	err = s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, summaryId, deployment.Generation, deployment.Hash, summary, namespace)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to save summary progress: %+v", err)
		return summary, err
	}
	defer func() {
		if r := recover(); r == nil {
			log.DebugfCtx(ctx, " M (Solution): Reconcile conclude Summary. Namespace: %v, deployment instance: %v, summary message: %v", namespace, deployment.Instance, summary.SummaryMessage)
			s.concludeSummary(ctx, deployment.Instance.ObjectMeta.Name, summaryId, deployment.Generation, deployment.Hash, summary, namespace)
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

	previousDesiredState := s.GetDeploymentState(ctx, deployment.Instance.ObjectMeta.Name, namespace)

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

	summary.PlannedDeployment = 0
	for _, step := range plan.Steps {
		summary.PlannedDeployment += len(step.Components)
	}
	summary.CurrentDeployed = 0
	err = s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, summaryId, deployment.Generation, deployment.Hash, summary, namespace)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Solution): failed to save summary progress: %+v", err)
		return summary, err
	}
	log.DebugfCtx(ctx, " M (Solution): reconcile save summary progress: start deploy, total %v deployments", summary.PlannedDeployment)
	// DO NOT REMOVE THIS COMMENT
	// gofail: var beforeProviders string

	plannedCount := 0
	planSuccessCount := 0
	log.DebugfCtx(ctx, " M (Solution): count of plan.Steps: %v", len(plan.Steps))
	for _, step := range plan.Steps {
		select {
		case <-ctx.Done():
			// Context canceled or timed out
			log.DebugCtx(ctx, " M (Solution): reconcile canceled")
			err = v1alpha2.NewCOAError(nil, "Reconciliation was canceled.", v1alpha2.Canceled)
			return summary, err
		default:
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
			var stepError error
			var componentResults = make(map[string]model.ComponentResultSpec)
			if previousDesiredState != nil {
				testState := MergeDeploymentStates(&previousDesiredState.State, currentState)
				if s.canSkipStep(ctx, step, step.Target, provider.(tgt.ITargetProvider), previousDesiredState.State.Components, testState) {
					summary.UpdateTargetResult(step.Target, model.TargetResultSpec{Status: "OK", Message: "", ComponentResults: componentResults})
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

			defaultScope := deployment.Instance.Spec.Scope
			// ensure to restore the original scope defined in instance in case the scope is changed during the deployment
			defer func() {
				deployment.Instance.Spec.Scope = defaultScope
			}()
			for i := 0; i < retryCount; i++ {
				deployment.Instance.Spec.Scope = getCurrentApplicationScope(ctx, deployment.Instance, deployment.Targets[step.Target])
				componentResults, stepError = (provider.(tgt.ITargetProvider)).Apply(ctx, dep, step, deployment.IsDryRun)
				if stepError == nil {
					targetResult[step.Target] = 1
					summary.AllAssignedDeployed = plannedCount == planSuccessCount
					summary.UpdateTargetResult(step.Target, model.TargetResultSpec{Status: "OK", Message: "", ComponentResults: componentResults})
					err = s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, summaryId, deployment.Generation, deployment.Hash, summary, namespace)
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

					if v1alpha2.IsCanceled(stepError) {
						log.ErrorfCtx(ctx, " M (Solution): reconcile canceled: %+v", stepError)
						break
					}

					time.Sleep(5 * time.Second) //TODO: make this configurable?
				}
				deployment.Instance.Spec.Scope = defaultScope
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
				if deployment.IsDryRun || deployment.IsInActive {
					summary.SuccessCount = 0
				} else {
					summary.SuccessCount = successCount
				}
				summary.AllAssignedDeployed = plannedCount == planSuccessCount
				err = stepError
				return summary, err
			}
			planSuccessCount++
			summary.CurrentDeployed += len(step.Components)
			err = s.saveSummaryProgress(ctx, deployment.Instance.ObjectMeta.Name, summaryId, deployment.Generation, deployment.Hash, summary, namespace)
			if err != nil {
				log.ErrorfCtx(ctx, " M (Solution): failed to save summary progress: %+v", err)
				return summary, err
			}
			log.DebugfCtx(ctx, " M (Solution): reconcile save summary progress: current deployed %v out of total %v deployments", summary.CurrentDeployed, summary.PlannedDeployment)
		}
	}

	mergedState.ClearAllRemoved()

	// DO NOT REMOVE THIS COMMENT
	// gofail: var beforeDeploymentError string

	if !deployment.IsDryRun {
		if len(mergedState.TargetComponent) == 0 && remove {
			log.DebugfCtx(ctx, " M (Solution): no assigned components to manage, deleting state")
			s.DeleteDeploymentState(ctx, deployment.Instance.ObjectMeta.Name, namespace)
		} else {
			s.UpsertDeploymentState(ctx, deployment.Instance.ObjectMeta.Name, namespace, deployment, mergedState)
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

	summary.Skipped = !someStepsRan

	if deployment.IsDryRun || deployment.IsInActive {
		summary.SuccessCount = 0
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

func (s *SolutionManager) saveSummary(ctx context.Context, objectName string, summaryId string, generation string, hash string, summary model.SummarySpec, state model.SummaryState, namespace string) error {
	// TODO: delete this state when time expires. This should probably be invoked by the vendor (via GetSummary method, for instance)
	log.DebugfCtx(ctx, " M (Solution): saving summary, objectName: %s, summaryId: %s, state: %v, namespace: %s, jobid: %s, hash %s, targetCount %d, successCount %d",
		objectName, summaryId, state, namespace, summary.JobID, hash, summary.TargetCount, summary.SuccessCount)
	return s.SummaryManager.UpsertSummary(ctx, fmt.Sprintf("%s-%s", "summary", summaryId), generation, hash, summary, state, namespace)
}

func (s *SolutionManager) saveSummaryProgress(ctx context.Context, objectName string, summaryId string, generation string, hash string, summary model.SummarySpec, namespace string) error {
	return s.saveSummary(ctx, objectName, summaryId, generation, hash, summary, model.SummaryStateRunning, namespace)
}

func (s *SolutionManager) concludeSummary(ctx context.Context, objectName string, summaryId string, generation string, hash string, summary model.SummarySpec, namespace string) error {
	return s.saveSummary(ctx, objectName, summaryId, generation, hash, summary, model.SummaryStateDone, namespace)
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
	defaultScope := deployment.Instance.Spec.Scope

	for _, step := range plan.Steps {
		if s.IsTarget && !api_utils.ContainsString(s.TargetNames, step.Target) {
			continue
		}
		if targetName != "" && targetName != step.Target {
			continue
		}

		deployment.ActiveTarget = step.Target
		deployment.Instance.Spec.Scope = getCurrentApplicationScope(ctx, deployment.Instance, deployment.Targets[step.Target])

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
		deployment.Instance.Spec.Scope = defaultScope
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

func getCurrentApplicationScope(ctx context.Context, instance model.InstanceState, target model.TargetState) string {
	log.InfofCtx(ctx, " M (Solution): getting current application scope, instance scope: %s, target application scope: %s", instance.Spec.Scope, target.Spec.SolutionScope)
	if instance.Spec.Scope == "" {
		if target.Spec.SolutionScope == "" {
			return "default"
		}
		return target.Spec.SolutionScope
	}
	if target.Spec.SolutionScope != "" && target.Spec.SolutionScope != instance.Spec.Scope {
		message := fmt.Sprintf(" M (Solution): target application scope: %s is inconsistent with instance scope: %s", target.Spec.SolutionScope, instance.Spec.Scope)
		log.WarnfCtx(ctx, message)
		observ_utils.EmitUserAuditsLogs(ctx, message)
	}
	return instance.Spec.Scope
}

func (s *SolutionManager) ReconcileWithCancelWrapper(ctx context.Context, deployment model.DeploymentSpec, remove bool, namespace string, targetName string) (model.SummarySpec, error) {
	instance := deployment.Instance.ObjectMeta.Name
	log.InfofCtx(ctx, " M (Solution): onReconcile create context with timeout, instance: %s, job id: %s, isRemove: %v", instance, deployment.JobID, remove)

	// Two conditions when the context will be canceled:
	//   1. default timeout reached;
	//   2. cancel() is called
	cancelCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	if !remove {
		// Track the CancelFunc for ongoing reconcile job
		s.addCancelFunc(ctx, namespace, instance, deployment.JobID, cancel)
	}

	defer func() {
		// call cancel to release resource and clean up the cancelFunc in jobList
		log.InfofCtx(ctx, " M (Solution): onReconcile complete, namespace: %s, instance: %s, job id: %s, isRemove: %v", namespace, instance, deployment.JobID, remove)
		cancel()
		if !remove {
			s.cleanUpCancelFunc(namespace, instance, deployment.JobID)
		}
	}()

	summary, err := s.Reconcile(cancelCtx, deployment, remove, namespace, targetName)
	return summary, err
}

func (s *SolutionManager) addCancelFunc(ctx context.Context, namespace string, instance string, jobID string, cancel context.CancelFunc) {
	jobListLock.Lock()
	defer jobListLock.Unlock()
	log.InfofCtx(ctx, " M (Solution): AddCancelFunc, namespace: %s, instance: %s, job id: %s", namespace, instance, jobID)

	jobKey := JobIdentifier{Namespace: namespace, Name: instance}
	if _, exists := s.jobList[jobKey]; !exists {
		s.jobList[jobKey] = make(map[string]context.CancelFunc)
	}

	s.jobList[jobKey][jobID] = cancel
}

func (s *SolutionManager) cleanUpCancelFunc(namespace string, instance string, jobID string) {
	jobListLock.Lock()
	defer jobListLock.Unlock()

	jobKey := JobIdentifier{Namespace: namespace, Name: instance}
	if _, exists := s.jobList[jobKey]; exists {
		delete(s.jobList[jobKey], jobID)
	}

	if len(s.jobList[jobKey]) == 0 {
		delete(s.jobList, jobKey)
	}
}

func (s *SolutionManager) CancelPreviousJobs(ctx context.Context, namespace string, instance string, jobID string) error {
	jobListLock.RLock()
	defer jobListLock.RUnlock()
	log.InfofCtx(ctx, " M (Solution): CancelPreviousJobs, namespace: %s, instance: %s, job id: %s", namespace, instance, jobID)

	jobKey := JobIdentifier{Namespace: namespace, Name: instance}
	for jid, cancelJob := range s.jobList[jobKey] {
		if convertJobIdToInt(jid) < convertJobIdToInt(jobID) {
			// only cancel jobs prior to the delete job
			log.InfofCtx(ctx, " M (Solution): CancelPreviousJobs, found previous job id: %s", jid)
			if cancelJob != nil {
				cancelJob()
				log.InfofCtx(ctx, " M (Solution): CancelPreviousJobs, cancelled job id: %s", jid)
			}
		}
	}
	return nil
}

func convertJobIdToInt(jobID string) int {
	num, err := strconv.Atoi(jobID)
	if err != nil {
		return 0
	}
	return num
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
