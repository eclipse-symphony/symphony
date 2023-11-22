/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package solution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	sp "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers"
	tgt "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target"
	api_utils "github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	config "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config"
	secret "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	states "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")
var lock sync.Mutex

const (
	SYMPHONY_AGENT string = "/symphony-agent:"
	ENV_NAME       string = "SYMPHONY_AGENT_ADDRESS"
)

type SolutionManager struct {
	managers.Manager
	TargetProviders map[string]tgt.ITargetProvider
	StateProvider   states.IStateProvider
	ConfigProvider  config.IExtConfigProvider
	SecretProvoider secret.ISecretProvider
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

	stateprovider, err := managers.GetStateProvider(config, providers)
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
		s.SecretProvoider = secretProvider
	} else {
		return err
	}

	return nil
}

func (s *SolutionManager) getPreviousState(ctx context.Context, instance string) *SolutionManagerDeploymentState {
	state, err := s.StateProvider.Get(ctx, states.GetRequest{
		ID: instance,
	})
	if err == nil {
		var managerState SolutionManagerDeploymentState
		jData, _ := json.Marshal(state.Body)
		err = json.Unmarshal(jData, &managerState)
		if err == nil {
			return &managerState
		}
		return nil
	}
	return nil
}
func (s *SolutionManager) GetSummary(ctx context.Context, key string) (model.SummaryResult, error) {
	// lock.Lock()
	// defer lock.Unlock()

	iCtx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "GetSummary",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Info(" M (Solution): get summary")

	state, err := s.StateProvider.Get(iCtx, states.GetRequest{
		ID: fmt.Sprintf("%s-%s", "summary", key),
	})
	if err != nil {
		log.Errorf(" M (Solution): failed to get deployment summary[%s]: %+v", key, err)
		return model.SummaryResult{}, err
	}

	var result model.SummaryResult
	jData, _ := json.Marshal(state.Body)
	err = json.Unmarshal(jData, &result)
	if err != nil {
		log.Errorf(" M (Solution): failed to deserailze deployment summary[%s]: %+v", key, err)
		return model.SummaryResult{}, err
	}

	return result, nil
}

func (s *SolutionManager) sendHeartbeat(id string, remove bool, stopCh chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	action := "update"
	if remove {
		action = "delete"
	}

	for {
		select {
		case <-ticker.C:
			s.VendorContext.Publish("heartbeat", v1alpha2.Event{
				Body: v1alpha2.HeartBeatData{
					JobId:  id,
					Action: action,
					Time:   time.Now().UTC(),
				},
			})
		case <-stopCh:
			return // Exit the goroutine when the stop signal is received
		}
	}
}

func (s *SolutionManager) Reconcile(ctx context.Context, deployment model.DeploymentSpec, remove bool) (model.SummarySpec, error) {
	lock.Lock()
	defer lock.Unlock()

	stopCh := make(chan struct{})
	defer close(stopCh)
	go s.sendHeartbeat(deployment.Instance.Name, remove, stopCh)

	iCtx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "Reconcile",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Info(" M (Solution): reconciling")

	summary := model.SummarySpec{
		TargetResults: make(map[string]model.TargetResultSpec),
		TargetCount:   len(deployment.Targets),
		SuccessCount:  0,
	}

	if s.VendorContext != nil && s.VendorContext.EvaluationContext != nil {
		context := s.VendorContext.EvaluationContext.Clone()
		context.DeploymentSpec = deployment
		context.Component = ""
		deployment, err = api_utils.EvaluateDeployment(*context)
	}

	if err != nil {
		if remove {
			log.Infof(" M (Solution): skipped failure to evaluate deployment spec: %+v", err)
		} else {
			summary.SummaryMessage = "failed to evaluate deployment spec: " + err.Error()
			log.Errorf(" M (Solution): failed to evaluate deployment spec: %+v", err)
			s.saveSummary(iCtx, deployment, summary)
			return summary, err
		}
	}

	previousDesiredState := s.getPreviousState(iCtx, deployment.Instance.Name)
	currentDesiredState, err := NewDeploymentState(deployment)
	if err != nil {
		summary.SummaryMessage = "failed to create target manager state from deployment spec: " + err.Error()
		log.Errorf(" M (Solution): failed to create target manager state from deployment spec: %+v", err)
		s.saveSummary(iCtx, deployment, summary)
		return summary, err
	}
	currentState, _, err := s.Get(iCtx, deployment)
	if err != nil {
		summary.SummaryMessage = "failed to get current state: " + err.Error()
		log.Errorf(" M (Solution): failed to get current state: %+v", err)
		s.saveSummary(iCtx, deployment, summary)
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

	plan, err := PlanForDeployment(deployment, mergedState)
	if err != nil {
		summary.SummaryMessage = "failed to plan for deployment: " + err.Error()
		log.Errorf(" M (Solution): failed to plan for deployment: %+v", err)
		s.saveSummary(iCtx, deployment, summary)
		return summary, err
	}

	col := api_utils.MergeCollection(deployment.Solution.Metadata, deployment.Instance.Metadata)
	dep := deployment
	dep.Instance.Metadata = col
	someStepsRan := false

	for _, step := range plan.Steps {
		dep.ActiveTarget = step.Target
		agent := findAgent(deployment.Targets[step.Target])
		if agent != "" {
			col[ENV_NAME] = agent
		} else {
			delete(col, ENV_NAME)
		}
		var override tgt.ITargetProvider
		if v, ok := s.TargetProviders[step.Target]; ok {
			override = v
		}
		var provider providers.IProvider
		provider, err = sp.CreateProviderForTargetRole(s.Context, step.Role, deployment.Targets[step.Target], override)
		if err != nil {
			summary.SummaryMessage = "failed to create provider:" + err.Error()
			log.Errorf(" M (Solution): failed to create provider: %+v", err)
			s.saveSummary(ctx, deployment, summary)
			return summary, err
		}

		if previousDesiredState != nil {
			testState := MergeDeploymentStates(&previousDesiredState.State, currentState)
			if s.canSkipStep(iCtx, step, step.Target, provider.(tgt.ITargetProvider), previousDesiredState.State.Components, testState) {
				continue
			}
		}
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
		// 				log.Errorf(" M (Solution): failed to evaluate property: %+v", err)
		// 				summary.SummaryMessage = fmt.Sprintf("failed to evaluate property '%s' on component '%s: %s", k, component.Component.Name, err.Error())
		// 				s.saveSummary(ctx, deployment, summary)
		// 				observ_utils.CloseSpanWithError(span, &err)
		// 				return summary, err
		// 			}
		// 		}
		// 	}
		// }

		for i := 0; i < retryCount; i++ {
			componentResults, stepError = (provider.(tgt.ITargetProvider)).Apply(iCtx, dep, step, false)
			if stepError == nil {
				summary.UpdateTargetResult(step.Target, model.TargetResultSpec{Status: "OK", Message: "", ComponentResults: componentResults})
				break
			} else {
				summary.UpdateTargetResult(step.Target, model.TargetResultSpec{Status: "Error", Message: stepError.Error(), ComponentResults: componentResults}) // TODO: this keeps only the last error on the target
				time.Sleep(5 * time.Second)                                                                                                                      //TODO: make this configurable?
			}
		}
		if stepError != nil {
			log.Errorf(" M (Solution): failed to execute deployment step: %+v", stepError)
			s.saveSummary(iCtx, deployment, summary)
			err = stepError
			return summary, err
		}
	}

	mergedState.ClearAllRemoved()

	// TODO: delete the state if the mergedState is empty (doesn't have any ComponentTarget assignements)
	s.StateProvider.Upsert(iCtx, states.UpsertRequest{
		Value: states.StateEntry{
			ID: deployment.Instance.Name,
			Body: SolutionManagerDeploymentState{
				Spec:  deployment,
				State: mergedState,
			},
		},
	})

	summary.Skipped = !someStepsRan
	if summary.Skipped {
		summary.SuccessCount = summary.TargetCount
	}
	summary.IsRemoval = remove
	s.saveSummary(iCtx, deployment, summary)
	return summary, nil
}
func (s *SolutionManager) saveSummary(ctx context.Context, deployment model.DeploymentSpec, summary model.SummarySpec) {
	// TODO: delete this state when time expires. This should probably be invoked by the vendor (via GetSummary method, for instance)
	s.StateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID: fmt.Sprintf("%s-%s", "summary", deployment.Instance.Name),
			Body: model.SummaryResult{
				Summary:    summary,
				Generation: deployment.Generation,
				Time:       time.Now().UTC(),
			},
		},
	})
}
func (s *SolutionManager) canSkipStep(ctx context.Context, step model.DeploymentStep, target string, provider tgt.ITargetProvider, currentComponents []model.ComponentSpec, state model.DeploymentState) bool {

	for _, newCom := range step.Components {
		key := fmt.Sprintf("%s::%s", newCom.Component.Name, target)
		if newCom.Action == "delete" {
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
func (s *SolutionManager) Get(ctx context.Context, deployment model.DeploymentSpec) (model.DeploymentState, []model.ComponentSpec, error) {
	iCtx, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Info(" M (Solution): getting deployment")

	ret := model.DeploymentState{}

	state, err := NewDeploymentState(deployment)
	if err != nil {
		log.Errorf(" M (Solution): failed to create manager state for deployment: %+v", err)
		return ret, nil, err
	}
	plan, err := PlanForDeployment(deployment, state)
	if err != nil {
		log.Errorf(" M (Solution): failed to plan for deployment: %+v", err)
		return ret, nil, err
	}
	ret = state
	ret.TargetComponent = make(map[string]string)
	retComponents := make([]model.ComponentSpec, 0)
	for _, step := range plan.Steps {
		var override tgt.ITargetProvider
		if v, ok := s.TargetProviders[step.Target]; ok {
			override = v
		}
		var provider providers.IProvider
		provider, err = sp.CreateProviderForTargetRole(s.Context, step.Role, deployment.Targets[step.Target], override)
		if err != nil {
			log.Errorf(" M (Solution): failed to create provider: %+v", err)
			return ret, nil, err
		}
		var components []model.ComponentSpec
		components, err = (provider.(tgt.ITargetProvider)).Get(iCtx, deployment, step.Components)
		if err != nil {
			log.Errorf(" M (Solution): failed to get: %+v", err)
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
	return false
}
func (s *SolutionManager) Poll() []error {
	return nil
}
func (s *SolutionManager) Reconcil() []error {
	return nil
}
func findAgent(target model.TargetSpec) string {
	for _, c := range target.Components {
		if v, ok := c.Properties[model.ContainerImage]; ok {
			if strings.Contains(fmt.Sprintf("%v", v), SYMPHONY_AGENT) {
				return c.Name
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
