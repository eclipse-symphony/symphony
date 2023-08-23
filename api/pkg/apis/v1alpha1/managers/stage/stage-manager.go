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

package stage

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	symproviders "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage/remote"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type StageManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

type TaskResult struct {
	Outputs map[string]interface{}
	Site    string
	Error   error
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
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
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
func (s *StageManager) ResumeStage(status model.ActivationStatus, cam model.CampaignSpec) (*v1alpha2.ActivationData, error) {
	campaign := status.Outputs["__campaign"].(string)
	activation := status.Outputs["__activation"].(string)
	activationGeneration := status.Outputs["__activationGeneration"].(string)
	site := status.Outputs["__site"].(string)
	stage := status.Outputs["__stage"].(string)

	entry, err := s.StateProvider.Get(context.Background(), states.GetRequest{
		ID: fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
	})
	if err != nil {
		return nil, err
	}
	if p, ok := entry.Body.(PendingTask); ok {
		// //find site in p.Sites
		// found := false
		// for _, s := range p.Sites {
		// 	if s == site {
		// 		found = true
		// 		break
		// 	}
		// }
		// if !found {
		// 	return nil, fmt.Errorf("site %s is not found in pending task", site)
		// }
		//remove site from p.Sites
		newSites := make([]string, 0)
		for _, s := range p.Sites {
			if s != site {
				newSites = append(newSites, s)
			}
		}
		if len(newSites) == 0 {
			err := s.StateProvider.Delete(context.Background(), states.DeleteRequest{
				ID: fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
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
					eCtx.Inputs = status.Inputs
					eCtx.Outputs = outputs
					val, err := parser.Eval(*eCtx)
					if err != nil {
						return nil, err
					}
					sVal := ""
					if val != nil {
						sVal = val.(string)
					}
					if sVal != "" {
						if _, ok := cam.Stages[sVal]; ok {
							nextStage = sVal
						} else {
							return nil, fmt.Errorf("stage %s is not found", sVal)
						}
					}
				}
				if nextStage != "" {
					activationData := &v1alpha2.ActivationData{
						Campaign:             campaign,
						Activation:           activation,
						ActivationGeneration: activationGeneration,
						Stage:                nextStage,
						Inputs:               status.Outputs,
						Provider:             cam.Stages[nextStage].Provider,
						Config:               cam.Stages[nextStage].Config,
						Outputs:              outputs,
					}
					return activationData, nil
				} else {
					return nil, nil
				}
			}
			return nil, nil
		} else {
			p.Sites = newSites
			_, err := s.StateProvider.Upsert(context.Background(), states.UpsertRequest{
				Value: states.StateEntry{
					ID:   fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
					Body: p,
				},
			})
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, fmt.Errorf("invalid pending task")
	}

	return nil, nil
}
func (s *StageManager) HandleDirectTriggerEvent(ctx context.Context, triggerData v1alpha2.ActivationData) model.ActivationStatus {
	status := model.ActivationStatus{
		Stage:        "",
		NextStage:    "",
		Outputs:      nil,
		Status:       v1alpha2.Untouched,
		ErrorMessage: "",
		IsActive:     true,
	}
	factory := symproviders.SymphonyProviderFactory{}
	provider, err := factory.CreateProvider(triggerData.Provider, triggerData.Config)
	if err != nil {
		status.Status = v1alpha2.InternalError
		status.ErrorMessage = err.Error()
		status.IsActive = false
		return status
	}
	if _, ok := provider.(*remote.RemoteStageProvider); ok {
		provider.(*remote.RemoteStageProvider).SetOutputsContext(triggerData.Outputs)
	}
	outputs, _, err := provider.(stage.IStageProvider).Process(ctx, *s.Manager.Context, triggerData.Inputs)
	if err != nil {
		status.Status = v1alpha2.InternalError
		status.ErrorMessage = err.Error()
		status.IsActive = false
		status.Outputs = outputs //TODO: this is not good. It assumes a process carries over inputs to outputs, and outputs are always returned even if there are errors. Downstream, some code relies on properties like __activation and __campaign in output collection.
		return status
	}
	status.Outputs = outputs
	status.Status = v1alpha2.Done
	status.IsActive = false
	return status
}
func (s *StageManager) HandleTriggerEvent(ctx context.Context, campaign model.CampaignSpec, triggerData v1alpha2.ActivationData) (model.ActivationStatus, *v1alpha2.ActivationData) {
	_, span := observability.StartSpan("Stage Manager", ctx, &map[string]string{
		"method": "HandleTriggerEvent",
	})
	log.Info(" M (Stage): HandleTriggerEvent")

	status := model.ActivationStatus{
		Stage:        triggerData.Stage,
		NextStage:    "",
		Outputs:      nil,
		Status:       v1alpha2.Untouched,
		ErrorMessage: "",
		IsActive:     true,
	}
	var activationData *v1alpha2.ActivationData
	if currentStage, ok := campaign.Stages[triggerData.Stage]; ok {
		sites := make([]string, 0)
		if currentStage.Contexts != "" {
			parser := utils.NewParser(currentStage.Contexts)

			eCtx := s.VendorContext.EvaluationContext.Clone()
			eCtx.Inputs = triggerData.Inputs
			eCtx.Outputs = triggerData.Outputs
			val, err := parser.Eval(*eCtx)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.Errorf(" M (Stage): failed to evaluate context: %v", err)
				observ_utils.CloseSpanWithError(span, err)
				return status, activationData
			}
			if _, ok := val.([]string); ok {
				sites = val.([]string)
			} else if _, ok := val.([]interface{}); ok {
				for _, v := range val.([]interface{}) {
					sites = append(sites, v.(string))
				}
			} else if _, ok := val.(string); ok {
				sites = append(sites, val.(string))
			} else {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = fmt.Sprintf("invalid context %s", currentStage.Contexts)
				status.IsActive = false
				log.Errorf(" M (Stage): invalid context: %v", currentStage.Contexts)
				observ_utils.CloseSpanWithError(span, err)
				return status, activationData
			}
		} else {
			sites = append(sites, s.VendorContext.SiteInfo.SiteId)
		}

		inputs := triggerData.Inputs
		if inputs == nil {
			inputs = make(map[string]interface{})
		}

		if currentStage.Inputs != nil {
			for k, v := range currentStage.Inputs {
				inputs[k] = v
			}
		}
		for k, v := range inputs {
			if _, ok := v.(string); ok {
				parser := utils.NewParser(v.(string))
				eCtx := s.VendorContext.EvaluationContext.Clone()
				eCtx.Inputs = triggerData.Inputs
				eCtx.Outputs = triggerData.Outputs
				val, err := parser.Eval(*eCtx)
				if err != nil {
					status.Status = v1alpha2.InternalError
					status.ErrorMessage = err.Error()
					status.IsActive = false
					log.Errorf(" M (Stage): failed to evaluate input: %v", err)
					observ_utils.CloseSpanWithError(span, err)
					return status, activationData
				}
				inputs[k] = val
			}
		}

		// inject default inputs
		inputs["__campaign"] = triggerData.Campaign
		inputs["__activation"] = triggerData.Activation
		inputs["__stage"] = triggerData.Stage
		inputs["__activationGeneration"] = triggerData.ActivationGeneration

		factory := symproviders.SymphonyProviderFactory{}
		provider, err := factory.CreateProvider(triggerData.Provider, triggerData.Config)
		if err != nil {
			status.Status = v1alpha2.InternalError
			status.ErrorMessage = err.Error()
			status.IsActive = false
			log.Errorf(" M (Stage): failed to create provider: %v", err)
			observ_utils.CloseSpanWithError(span, err)
			return status, activationData
		}

		numTasks := len(sites)
		waitGroup := sync.WaitGroup{}
		results := make(chan TaskResult, numTasks)
		pauseRequested := false

		for _, site := range sites {
			waitGroup.Add(1)
			go func(wg *sync.WaitGroup, site string, results chan<- TaskResult) {
				defer wg.Done()
				inputCopy := make(map[string]interface{})
				for k, v := range inputs {
					if _, ok := v.(string); ok {
						sv := v.(string)
						sv = strings.ReplaceAll(sv, "__site", site)
						sv = strings.ReplaceAll(sv, "__campaign", triggerData.Campaign)
						sv = strings.ReplaceAll(sv, "__activation", triggerData.Activation)
						sv = strings.ReplaceAll(sv, "__activationGeneration", triggerData.ActivationGeneration)
						inputCopy[k] = sv
					} else {
						inputCopy[k] = v
					}
				}
				inputCopy["__site"] = site
				if _, ok := provider.(*remote.RemoteStageProvider); ok {
					provider.(*remote.RemoteStageProvider).SetOutputsContext(triggerData.Outputs)
				}
				outputs, pause, err := provider.(stage.IStageProvider).Process(ctx, *s.Manager.Context, inputCopy)
				if pause {
					pauseRequested = true
				}
				results <- TaskResult{
					Outputs: outputs,
					Error:   err,
					Site:    site,
				}
			}(&waitGroup, site, results)
		}

		waitGroup.Wait()
		close(results)

		outputs := make(map[string]interface{})
		for result := range results {
			if result.Error != nil {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = fmt.Sprintf("%s: %s", result.Site, result.Error.Error())
				status.IsActive = false
				log.Errorf(" M (Stage): failed to process stage outputs: %v", result.Error)
				observ_utils.CloseSpanWithError(span, result.Error)
				return status, activationData
			}
			for k, v := range result.Outputs {
				if result.Site == s.Context.SiteInfo.SiteId {
					outputs[k] = v
				} else {
					outputs[fmt.Sprintf("%s.%s", result.Site, k)] = v
				}
			}
		}

		if triggerData.Outputs == nil {
			triggerData.Outputs = make(map[string]map[string]interface{})
		}
		triggerData.Outputs[triggerData.Stage] = outputs
		if campaign.SelfDriving {
			if pauseRequested {
				pendingTask := PendingTask{
					Sites:         sites,
					OutputContext: triggerData.Outputs,
				}
				_, err := s.StateProvider.Upsert(ctx, states.UpsertRequest{
					Value: states.StateEntry{
						ID:   fmt.Sprintf("%s-%s-%s", triggerData.Campaign, triggerData.Activation, triggerData.ActivationGeneration),
						Body: pendingTask,
					},
				})
				if err != nil {
					status.Status = v1alpha2.InternalError
					status.ErrorMessage = err.Error()
					status.IsActive = false
					log.Errorf(" M (Stage): failed to save pending task: %v", err)
					observ_utils.CloseSpanWithError(span, err)
					return status, activationData
				}
				status.Status = v1alpha2.Paused
				status.IsActive = false
				return status, activationData
			}

			parser := utils.NewParser(currentStage.StageSelector)
			eCtx := s.VendorContext.EvaluationContext.Clone()
			eCtx.Inputs = triggerData.Inputs
			eCtx.Outputs = triggerData.Outputs
			val, err := parser.Eval(*eCtx)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.Errorf(" M (Stage): failed to evaluate stage selector: %v", err)
				observ_utils.CloseSpanWithError(span, err)
				return status, activationData
			}
			sVal := ""
			if val != nil {
				sVal = val.(string)
			}
			if sVal != "" {
				if nextStage, ok := campaign.Stages[sVal]; ok {
					status.NextStage = sVal
					activationData = &v1alpha2.ActivationData{
						Campaign:             triggerData.Campaign,
						Activation:           triggerData.Activation,
						ActivationGeneration: triggerData.ActivationGeneration,
						Stage:                sVal,
						Inputs:               outputs,
						Outputs:              triggerData.Outputs,
						Provider:             nextStage.Provider,
						Config:               nextStage.Config,
					}
				} else {
					err = v1alpha2.NewCOAError(nil, status.ErrorMessage, v1alpha2.BadRequest)
					status.Status = v1alpha2.BadRequest
					status.ErrorMessage = fmt.Sprintf("stage %s is not found", sVal)
					status.IsActive = false
					log.Errorf(" M (Stage): failed to find next stage: %v", err)
					observ_utils.CloseSpanWithError(span, err)
					return status, activationData
				}
			}
			status.NextStage = sVal
			if sVal == "" {
				status.IsActive = false
				status.Status = v1alpha2.Done
			} else {
				if pauseRequested {
					status.IsActive = false
					status.Status = v1alpha2.Paused
				} else {
					status.IsActive = true
					status.Status = v1alpha2.Running
				}
			}
			log.Infof(" M (Stage): stage %s is done", triggerData.Stage)
			observ_utils.CloseSpanWithError(span, nil)
			return status, activationData
		} else {
			status.Status = v1alpha2.Done
			status.NextStage = ""
			status.IsActive = false
			log.Infof(" M (Stage): stage %s is done (no next stage)", triggerData.Stage)
			observ_utils.CloseSpanWithError(span, nil)
			return status, activationData
		}
	}
	err := v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", triggerData.Stage), v1alpha2.BadRequest)
	status.Status = v1alpha2.InternalError
	status.ErrorMessage = err.Error()
	status.IsActive = false
	log.Errorf(" M (Stage): failed to find stage: %v", err)
	observ_utils.CloseSpanWithError(span, err)
	return status, activationData
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
		if activation.Status != nil && activation.Status.Stage != "" && activation.Status.NextStage != stage {
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not the next stage", stage), v1alpha2.BadRequest)
		}
		return &v1alpha2.ActivationData{
			Campaign:             actData.Campaign,
			Activation:           actData.Activation,
			ActivationGeneration: actData.ActivationGeneration,
			Stage:                stage,
			Inputs:               activation.Spec.Inputs,
			Provider:             stageSpec.Provider,
			Config:               stageSpec.Config,
		}, nil
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", stage), v1alpha2.BadRequest)
}
