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
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
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
	TargetProvider  tgt.ITargetProvider
	StateProvider   states.IStateProvider
	ConfigProvider  config.IConfigProvider
	SecretProvoider secret.ISecretProvider
}

func (s *SolutionManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {

	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}

	for _, v := range providers {
		if p, ok := v.(tgt.ITargetProvider); ok {
			s.TargetProvider = p
			break
		}
	}

	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}

	configProvider, err := managers.GetConfigProvider(config, providers)
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

func (s *SolutionManager) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	if s.TargetProvider != nil {
		return s.TargetProvider.NeedsUpdate(ctx, desired, current)
	}
	return !model.SlicesCover(desired, current)
}
func (s *SolutionManager) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	if s.TargetProvider != nil {
		return s.TargetProvider.NeedsRemove(ctx, desired, current)
	}
	return model.SlicesAny(desired, current)
}

func (s *SolutionManager) Apply(ctx context.Context, deployment model.DeploymentSpec) (model.SummarySpec, error) {
	lock.Lock()
	defer lock.Unlock()

	_, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "Apply",
	})
	log.Info(" M (Solution): applying deployment")

	summary := model.SummarySpec{
		TargetResults: make(map[string]model.TargetResultSpec),
		TargetCount:   len(deployment.Targets),
		SuccessCount:  0,
	}

	var err error
	deployment, err = utils.EvaluateDeployment(utils.EvaluationContext{
		ConfigProvider: s.ConfigProvider,
		SecretProvider: s.SecretProvoider,
		DeploymentSpec: deployment,
		Component:      "",
	})
	if err != nil {
		return summary, err
	}

	// at manager level, if we found deployment spec hasn't been changed, skip apply
	// not to scale out the manager, a shared state provider such as Redis state provider
	// needs to be used
	notLastSeen := true
	name := deployment.Instance.DisplayName
	if name == "" {
		name = deployment.Instance.Name
	}
	state, err := s.StateProvider.Get(ctx, states.GetRequest{
		ID: name,
	})
	if err == nil {
		var seenDeployment model.DeploymentSpec
		jData, _ := json.Marshal(state.Body)

		err = json.Unmarshal(jData, &seenDeployment)
		if err == nil {
			equal, _ := seenDeployment.DeepEquals(deployment)
			if equal {
				notLastSeen = false
			}
		}
	}

	for k, v := range deployment.Assignments {
		if v != "" {
			components := make([]model.ComponentSpec, 0)
			// get components that are assigned to the current target
			for _, component := range deployment.Solution.Components {
				if strings.Contains(v, "{"+component.Name+"}") {
					components = append(components, component)
				}
			}
			//sort components by depedencies
			components, err := sortByDepedencies(components)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				log.Errorf(" M (Solution): failed to sort components: %+v", err)
				return summary, err
			}

			for key, target := range deployment.Targets {
				if key == k {
					summary.TargetResults[key] = model.TargetResultSpec{Status: "OK", Message: ""}
					groups := collectGroups(components)
					index := 0
					var err error
					var provider providers.IProvider
					for i, group := range groups {
						if strings.HasPrefix(group.Type, "staged:") {
							continue
						}
						provider, err = sp.CreateProviderForTargetRole(group.Type, target, s.TargetProvider)
						if err != nil {
							observ_utils.CloseSpanWithError(span, err)
							log.Errorf(" M (Solution): failed to create provider: %+v", err)
							return summary, err
						}
						col := utils.MergeCollection(deployment.Solution.Metadata, deployment.Instance.Metadata)
						agent := findAgent(target)
						if agent != "" {
							col[ENV_NAME] = agent
						}
						var current []model.ComponentSpec
						dep := deployment
						dep.Instance.Metadata = col
						dep.ActiveTarget = key
						dep.Solution.Components = components

						if i == len(groups)-1 {

							dep.ComponentStartIndex = index
							dep.ComponentEndIndex = len(components)

							for counter := 0; counter < 3; counter++ {
								current, err = (provider.(tgt.ITargetProvider)).Get(ctx, dep)
								if err == nil {
									if (notLastSeen && target.ForceRedeploy) || (provider.(tgt.ITargetProvider)).NeedsUpdate(ctx, components[index:], current) {
										err = (provider.(tgt.ITargetProvider)).Apply(ctx, dep, false)
										if err == nil {
											break
										} else {
											summary.TargetResults[key] = model.TargetResultSpec{Status: "Error", Message: err.Error()}
										}
									} else {
										break
									}
								}
								time.Sleep(5 * time.Second)
							}
							if err != nil {
								observ_utils.CloseSpanWithError(span, err)
								log.Errorf(" M (Solution): %+v", err)
								return summary, err
							}
						} else {
							dep.ComponentStartIndex = index
							dep.ComponentEndIndex = groups[i+1].Index

							for counter := 0; counter < 3; counter++ {
								current, err = (provider.(tgt.ITargetProvider)).Get(ctx, dep)
								if err == nil {
									if (notLastSeen && target.ForceRedeploy) || (provider.(tgt.ITargetProvider)).NeedsUpdate(ctx, components[index:groups[i+1].Index], current) {
										err = (provider.(tgt.ITargetProvider)).Apply(ctx, dep, false)
										if err == nil {
											break
										} else {
											summary.TargetResults[key] = model.TargetResultSpec{Status: "Error", Message: err.Error()}
										}
									} else {
										break
									}
								}
								time.Sleep(5 * time.Second)
							}
							if err != nil {
								observ_utils.CloseSpanWithError(span, err)
								log.Errorf(" M (Solution): %+v", err)
								return summary, err
							}
							index = groups[i+1].Index
						}
						if err != nil && !group.CanSkip {
							observ_utils.CloseSpanWithError(span, err)
							log.Errorf(" M (Solution): %+v", err)
							return summary, err
						}
					}
					if vk, ok := summary.TargetResults[key]; ok {
						if vk.Status == "OK" {
							summary.SuccessCount += 1
						}
					}
					break
				}
			}
		}
	}

	s.StateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: deployment,
		},
	})

	observ_utils.CloseSpanWithError(span, nil)
	return summary, nil
}

func (s *SolutionManager) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	lock.Lock()
	defer lock.Unlock()

	_, span := observability.StartSpan("Solution Manager", ctx, &map[string]string{
		"method": "Get",
	})
	log.Info(" M (Solution): getting deployment")

	ret := make([]model.ComponentSpec, 0)

	for k, v := range deployment.Assignments {
		if v != "" {
			components := make([]model.ComponentSpec, 0)
			// get components that are assigned to the current target
			for _, component := range deployment.Solution.Components {
				if strings.Contains(v, "{"+component.Name+"}") {
					components = append(components, component)
				}
			}
			//sort components by depedencies
			components, err := sortByDepedencies(components)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				log.Errorf(" M (Solution): failed to sort components: %+v", err)
				return ret, err
			}

			for key, target := range deployment.Targets {
				if key == k {
					groups := collectGroups(components)
					index := 0
					var err error
					var provider providers.IProvider
					for i, group := range groups {
						if strings.HasPrefix(group.Type, "staged:") {
							continue
						}
						provider, err = sp.CreateProviderForTargetRole(group.Type, target, s.TargetProvider)
						if err != nil {
							observ_utils.CloseSpanWithError(span, err)
							log.Errorf(" M (Solution): failed to create provider: %+v", err)
							return ret, err
						}
						col := utils.MergeCollection(deployment.Solution.Metadata, deployment.Instance.Metadata)
						agent := findAgent(target)
						if agent != "" {
							col[ENV_NAME] = agent
						}
						var current []model.ComponentSpec

						dep := deployment
						dep.Instance.Metadata = col
						dep.ActiveTarget = key
						dep.Solution.Components = components

						if i == len(groups)-1 {

							dep.ComponentStartIndex = index
							dep.ComponentEndIndex = len(components)

							current, err = (provider.(tgt.ITargetProvider)).Get(ctx, dep)
							if err != nil {
								observ_utils.CloseSpanWithError(span, err)
								log.Errorf(" M (Solution):  %+v", err)
								return ret, err
							}
							ret = append(ret, current...)
						} else {
							dep.ComponentStartIndex = index
							dep.ComponentEndIndex = groups[i+1].Index

							current, err = (provider.(tgt.ITargetProvider)).Get(ctx, dep)
							if err != nil {
								observ_utils.CloseSpanWithError(span, err)
								log.Errorf(" M (Solution): %+v", err)
								return ret, err
							}
							ret = append(ret, current...)
							index = groups[i+1].Index
						}
					}
					break
				}
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}
func updateSummary(summary model.SummarySpec, target string, err error) (model.SummarySpec, error) {
	sczErr, ok := err.(v1alpha2.COAError)
	if ok {
		summary.TargetResults[target] = model.TargetResultSpec{
			Status:  sczErr.State.String(),
			Message: sczErr.Message,
		}
	} else {
		summary.TargetResults[target] = model.TargetResultSpec{
			Status:  "Internal Error",
			Message: err.Error(),
		}
	}
	return summary, err
}
func (s *SolutionManager) Remove(ctx context.Context, deployment model.DeploymentSpec) (model.SummarySpec, error) {
	lock.Lock()
	defer lock.Unlock()

	summary := model.SummarySpec{
		TargetResults: make(map[string]model.TargetResultSpec),
		TargetCount:   len(deployment.Targets),
		SuccessCount:  len(deployment.Targets),
	}
	for k, v := range deployment.Assignments {
		if v != "" {
			components := make([]model.ComponentSpec, 0)
			// get components that are assigned to the current target
			for _, component := range deployment.Solution.Components {
				if strings.Contains(v, "{"+component.Name+"}") {
					components = append(components, component)
				}
			}
			//sort components by depedencies
			components, err := sortByDepedencies(components)
			if err != nil {
				return updateSummary(summary, v, err)
			}
			for key, target := range deployment.Targets {
				if key == k {
					cd, _ := json.Marshal(components)
					log.Debug(string(cd))
					groups := collectGroups(components)
					index := 0
					for i, group := range groups {
						if strings.HasPrefix(group.Type, "staged:") {
							continue
						}
						provider, err := sp.CreateProviderForTargetRole(group.Type, target, s.TargetProvider)
						if err != nil {
							return updateSummary(summary, k, err)
						}
						col := utils.MergeCollection(deployment.Solution.Metadata, deployment.Instance.Metadata)
						agent := findAgent(target)
						if agent != "" {
							col[ENV_NAME] = agent
						}
						var current []model.ComponentSpec
						dep := deployment
						dep.Instance.Metadata = col
						dep.ActiveTarget = key
						dep.Solution.Components = components
						if i == len(groups)-1 {
							dep.ComponentStartIndex = index
							dep.ComponentEndIndex = len(components)

							for counter := 0; counter < 3; counter++ {
								current, err = (provider.(tgt.ITargetProvider)).Get(ctx, dep)
								if err == nil {
									if (provider.(tgt.ITargetProvider)).NeedsRemove(ctx, components[index:], current) {
										err = (provider.(tgt.ITargetProvider)).Remove(ctx, dep, current)
										if err == nil {
											break
										} else {
											summary.TargetResults[key] = model.TargetResultSpec{Status: "Error", Message: err.Error()}
											log.Errorf(" M (Solution): failed to remove: %+v", err)
										}
									} else {
										break
									}
								} else {
									log.Errorf(" M (Solution): failed to get components during remove: %+v", err)
								}
								time.Sleep(5 * time.Second)
							}
						} else {
							dep.ComponentStartIndex = index
							dep.ComponentEndIndex = groups[i+1].Index

							for counter := 0; counter < 3; counter++ {
								current, err = (provider.(tgt.ITargetProvider)).Get(ctx, dep)
								if err == nil {
									if (provider.(tgt.ITargetProvider)).NeedsRemove(ctx, components[index:groups[i+1].Index], current) {
										err = (provider.(tgt.ITargetProvider)).Remove(ctx, dep, current)
										if err == nil {
											break
										} else {
											summary.TargetResults[key] = model.TargetResultSpec{Status: "Error", Message: err.Error()}
										}
									} else {
										break
									}
								}
								time.Sleep(5 * time.Second)
							}
							index = groups[i+1].Index
						}
						if err != nil && !group.CanSkip {
							return summary, err
						}
					}
					if vk, ok := summary.TargetResults[key]; ok {
						if vk.Status == "OK" {
							summary.SuccessCount -= 1
						}
					}
					break
				}
			}
		}
	}
	name := deployment.Instance.DisplayName
	if name == "" {
		name = deployment.Instance.Name
	}
	s.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
	})
	return summary, nil
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

type marks struct {
	Type    string
	Index   int
	CanSkip bool
}

func collectGroups(components []model.ComponentSpec) []marks {
	ret := make([]marks, 0)
	currentType := "INVALID"
	for i, c := range components {
		if c.Type != currentType {
			if _, ok := c.Properties["errors.ignore"]; ok {
				ret = append(ret, marks{Type: c.Type, Index: i, CanSkip: true})
			} else {
				ret = append(ret, marks{Type: c.Type, Index: i, CanSkip: false})
			}
			currentType = c.Type
		}
	}
	return ret
}
