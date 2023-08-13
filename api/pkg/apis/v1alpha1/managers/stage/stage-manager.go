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
	"os"
	"sync"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	symproviders "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type StageManager struct {
	managers.Manager
}

type TaskResult struct {
	Outputs map[string]interface{}
	Site    string
	Error   error
}

func (s *StageManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
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
func (s *StageManager) HandleTriggerEvent(ctx context.Context, campaign model.CampaignSpec, triggerData v1alpha2.ActivationData) (model.ActivationStatus, *v1alpha2.ActivationData) {
	status := model.ActivationStatus{
		Stage:        "",
		NextStage:    "",
		Outputs:      nil,
		Status:       v1alpha2.Untouched,
		ErrorMessage: "",
		IsActive:     true,
	}
	var activationData *v1alpha2.ActivationData
	if currentStage, ok := campaign.Stages[triggerData.Stage]; ok {
		// stage definition inputs override activation inputs
		inputs := triggerData.Inputs
		if inputs == nil {
			inputs = make(map[string]interface{})
		}
		if currentStage.Inputs != nil {
			for k, v := range currentStage.Inputs {
				if _, ok := v.(string); !ok {
					inputs[k] = v
				} else {
					parser := utils.NewParser(v.(string))
					val, err := parser.Eval(utils.EvaluationContext{Inputs: triggerData.Inputs, Outputs: triggerData.Outputs})
					if err != nil {
						status.Status = v1alpha2.InternalError
						status.ErrorMessage = err.Error()
						status.IsActive = false
						return status, activationData
					}
					inputs[k] = val
				}
			}
		}
		sites := make([]string, 0)
		if currentStage.Contexts != "" {
			parser := utils.NewParser(currentStage.Contexts)
			val, err := parser.Eval(utils.EvaluationContext{Inputs: triggerData.Inputs, Outputs: triggerData.Outputs})
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = err.Error()
				status.IsActive = false
				return status, activationData
			}
			if _, ok := val.([]string); ok {
				sites = val.([]string)
			} else if _, ok := val.(string); ok {
				sites = append(sites, val.(string))
			} else {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = fmt.Sprintf("invalid context %s", currentStage.Contexts)
				status.IsActive = false
				return status, activationData
			}
		} else {
			sites = append(sites, os.Getenv("SYMPHONY_SITE_ID"))
		}

		// inject default inputs
		inputs["__campaign"] = triggerData.Campaign
		inputs["__activation"] = triggerData.Activation

		factory := symproviders.SymphonyProviderFactory{}
		provider, err := factory.CreateProvider(triggerData.Provider, triggerData.Config)
		if err != nil {
			status.Status = v1alpha2.InternalError
			status.ErrorMessage = err.Error()
			status.IsActive = false
			return status, activationData
		}

		numTasks := len(sites)
		waitGroup := sync.WaitGroup{}
		results := make(chan TaskResult, numTasks)

		for _, site := range sites {
			waitGroup.Add(1)
			go func(wg *sync.WaitGroup, site string, results chan<- TaskResult) {
				defer wg.Done()
				inputCopy := make(map[string]interface{})
				for k, v := range inputs {
					inputCopy[k] = v
				}
				inputCopy["__site"] = site
				outputs, err := provider.(stage.IStageProvider).Process(ctx, inputCopy)
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
				return status, activationData
			}
			for k, v := range result.Outputs {
				if result.Site == os.Getenv("SYMPHONY_SITE_ID") {
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
			parser := utils.NewParser(currentStage.StageSelector)
			val, err := parser.Eval(utils.EvaluationContext{Inputs: triggerData.Inputs, Outputs: triggerData.Outputs})
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = err.Error()
				status.IsActive = false
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
					return status, activationData
				}
			}
			status.NextStage = sVal
			if val == "" {
				status.IsActive = false
				status.Status = v1alpha2.Done
			}
			return status, activationData
		} else {
			status.NextStage = ""
			status.IsActive = false
			return status, activationData
		}
	}
	err := v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", triggerData.Stage), v1alpha2.BadRequest)
	status.Status = v1alpha2.InternalError
	status.ErrorMessage = err.Error()
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
