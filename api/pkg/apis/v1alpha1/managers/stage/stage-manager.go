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
func (s *StageManager) HandleTriggerEvent(ctx context.Context, event v1alpha2.Event) (*v1alpha2.ActivationData, error) {
	baseUrl, err := utils.GetString(s.Manager.Config.Properties, "baseUrl")
	if err != nil {
		return nil, err
	}
	user, err := utils.GetString(s.Manager.Config.Properties, "user")
	if err != nil {
		return nil, err
	}
	password, err := utils.GetString(s.Manager.Config.Properties, "password")
	if err != nil {
		return nil, err
	}
	triggerData := v1alpha2.ActivationData{}

	status := model.ActivationStatus{
		Stage:        "",
		NextStage:    "",
		Outputs:      nil,
		Status:       v1alpha2.Untouched,
		ErrorMessage: "",
		IsActive:     true,
	}

	var aok bool
	if triggerData, aok = event.Body.(v1alpha2.ActivationData); !aok {
		err = v1alpha2.NewCOAError(nil, "event body is not an activation job", v1alpha2.BadRequest)
		status.Status = v1alpha2.BadRequest
		status.ErrorMessage = err.Error()
		status.IsActive = false
		err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
		return nil, err
	}

	campaign, err := utils.GetCampaign(baseUrl, triggerData.Campaign, user, password)
	if err != nil {
		status.Status = v1alpha2.InternalError
		status.ErrorMessage = err.Error()
		status.IsActive = false
		err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
		return nil, err
	}
	status.Stage = triggerData.Stage
	status.ActivationGeneration = triggerData.ActivationGeneration
	if currentStage, ok := campaign.Spec.Stages[triggerData.Stage]; ok {
		factory := symproviders.SymphonyProviderFactory{}
		provider, err := factory.CreateProvider(triggerData.Provider, triggerData.Config)
		if err != nil {
			status.Status = v1alpha2.InternalError
			status.ErrorMessage = err.Error()
			status.IsActive = false
			err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
			return nil, err
		}

		// stage definition inputs override activation inputs
		inputs := triggerData.Inputs
		if currentStage.Inputs != nil {
			for k, v := range currentStage.Inputs {
				inputs[k] = v
			}
		}

		outputs, err := provider.(stage.IStageProvider).Process(ctx, inputs)
		if err != nil {
			status.Status = v1alpha2.InternalError
			status.ErrorMessage = err.Error()
			status.IsActive = false
			err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
			return nil, err
		}
		status.Status = v1alpha2.OK
		status.Outputs = outputs
		var activationData *v1alpha2.ActivationData
		if campaign.Spec.SelfDriving {
			parser := utils.NewParser(currentStage.StageSelector)
			val, err := parser.Eval(utils.EvaluationContext{Inputs: triggerData.Inputs, Outputs: outputs})
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = err.Error()
				status.IsActive = false
				err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
				return nil, err
			}
			if val != "" {
				if nextStage, ok := campaign.Spec.Stages[val]; ok {
					status.NextStage = val
					activationData = &v1alpha2.ActivationData{
						Campaign:             triggerData.Campaign,
						Activation:           triggerData.Activation,
						ActivationGeneration: triggerData.ActivationGeneration,
						Stage:                val,
						Inputs:               outputs,
						Provider:             nextStage.Provider,
						Config:               nextStage.Config,
					}
				} else {
					err = v1alpha2.NewCOAError(nil, status.ErrorMessage, v1alpha2.BadRequest)
					status.Status = v1alpha2.BadRequest
					status.ErrorMessage = err.Error()
					status.IsActive = false
					err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
					return nil, err
				}
			}
			status.NextStage = val
			if val == "" {
				status.IsActive = false
			}
			err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
		} else {
			status.NextStage = ""
			status.IsActive = false
			err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
		}
		return activationData, err
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", triggerData.Stage), v1alpha2.BadRequest)
	status.Status = v1alpha2.InternalError
	status.ErrorMessage = err.Error()
	err = utils.ReportActivationStatus(baseUrl, triggerData.Activation, user, password, status)
	return nil, err
}

func (s *StageManager) HandleActivationEvent(ctx context.Context, event v1alpha2.Event) (*v1alpha2.ActivationData, error) {
	baseUrl, err := utils.GetString(s.Manager.Config.Properties, "baseUrl")
	if err != nil {
		return nil, err
	}
	user, err := utils.GetString(s.Manager.Config.Properties, "user")
	if err != nil {
		return nil, err
	}
	password, err := utils.GetString(s.Manager.Config.Properties, "password")
	if err != nil {
		return nil, err
	}
	var actData v1alpha2.ActivationData
	var aok bool
	if actData, aok = event.Body.(v1alpha2.ActivationData); !aok {
		return nil, v1alpha2.NewCOAError(nil, "event body is not an activation job", v1alpha2.BadRequest)
	}

	campaign, err := utils.GetCampaign(baseUrl, actData.Campaign, user, password)
	if err != nil {
		return nil, err
	}
	activation, err := utils.GetActivation(baseUrl, actData.Activation, user, password)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}
	stage := actData.Stage
	if _, ok := campaign.Spec.Stages[stage]; !ok {
		stage = campaign.Spec.FirstStage
	}
	if stage == "" {
		return nil, v1alpha2.NewCOAError(nil, "no stage found", v1alpha2.BadRequest)
	}
	if stageSpec, ok := campaign.Spec.Stages[stage]; ok {
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
