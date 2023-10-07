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

package vendors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/campaigns"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/stage"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage/materialize"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage/wait"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type StageVendor struct {
	vendors.Vendor
	StageManager       *stage.StageManager
	CampaignsManager   *campaigns.CampaignsManager
	ActivationsManager *activations.ActivationsManager
}

func (s *StageVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  s.Vendor.Version,
		Name:     "Stage",
		Producer: "Microsoft",
	}
}

func (o *StageVendor) GetEndpoints() []v1alpha2.Endpoint {
	return []v1alpha2.Endpoint{}
}

func (s *StageVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := s.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range s.Managers {
		if c, ok := m.(*stage.StageManager); ok {
			s.StageManager = c
		}
		if c, ok := m.(*campaigns.CampaignsManager); ok {
			s.CampaignsManager = c
		}
		if c, ok := m.(*activations.ActivationsManager); ok {
			s.ActivationsManager = c
		}
	}
	if s.StageManager == nil {
		return v1alpha2.NewCOAError(nil, "stage manager is not supplied", v1alpha2.MissingConfig)
	}
	if s.CampaignsManager == nil {
		return v1alpha2.NewCOAError(nil, "campaigns manager is not supplied", v1alpha2.MissingConfig)
	}
	if s.ActivationsManager == nil {
		return v1alpha2.NewCOAError(nil, "activations manager is not supplied", v1alpha2.MissingConfig)
	}
	s.Vendor.Context.Subscribe("activation", func(topic string, event v1alpha2.Event) error {
		var actData v1alpha2.ActivationData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &actData)
		if err != nil {
			return v1alpha2.NewCOAError(nil, "event body is not an activation job", v1alpha2.BadRequest)
		}
		campaign, err := s.CampaignsManager.GetSpec(context.Background(), actData.Campaign)
		if err != nil {
			return err
		}
		activation, err := s.ActivationsManager.GetSpec(context.Background(), actData.Activation)
		if err != nil {
			return err
		}

		evt, err := s.StageManager.HandleActivationEvent(context.Background(), actData, *campaign.Spec, activation)
		if err != nil {
			return err
		}
		if evt != nil {
			s.Vendor.Context.Publish("trigger", v1alpha2.Event{
				Body: *evt,
			})
		}
		return nil
	})
	s.Vendor.Context.Subscribe("trigger", func(topic string, event v1alpha2.Event) error {
		log.Info("V (Stage): handling trigger event")
		status := model.ActivationStatus{
			Stage:        "",
			NextStage:    "",
			Outputs:      nil,
			Status:       v1alpha2.Untouched,
			ErrorMessage: "",
			IsActive:     true,
		}
		triggerData := v1alpha2.ActivationData{}
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &triggerData)
		if err != nil {
			err = v1alpha2.NewCOAError(nil, "event body is not an activation job", v1alpha2.BadRequest)
			status.Status = v1alpha2.BadRequest
			status.ErrorMessage = err.Error()
			status.IsActive = false
			sLog.Errorf("V (Stage): failed to deserialize activation data: %v", err)
			err = s.ActivationsManager.ReportStatus(context.Background(), triggerData.Activation, status)
			if err != nil {
				sLog.Errorf("V (Stage): failed to report error status: %v (%v)", status.ErrorMessage, err)
			}
		}
		campaign, err := s.CampaignsManager.GetSpec(context.Background(), triggerData.Campaign)
		if err != nil {
			status.Status = v1alpha2.BadRequest
			status.ErrorMessage = err.Error()
			status.IsActive = false
			sLog.Errorf("V (Stage): failed to get campaign spec: %v", err)
			err = s.ActivationsManager.ReportStatus(context.Background(), triggerData.Activation, status)
			if err != nil {
				sLog.Errorf("V (Stage): failed to report error status: %v (%v)", status.ErrorMessage, err)
			}
		}
		status.Stage = triggerData.Stage
		status.ActivationGeneration = triggerData.ActivationGeneration
		status.ErrorMessage = ""
		status.Status = v1alpha2.Running
		err = s.ActivationsManager.ReportStatus(context.Background(), triggerData.Activation, status)
		if err != nil {
			sLog.Errorf("V (Stage): failed to report accepted status: %v (%v)", status.ErrorMessage, err)
			return err
		}
		status, activation := s.StageManager.HandleTriggerEvent(context.Background(), *campaign.Spec, triggerData)
		err = s.ActivationsManager.ReportStatus(context.Background(), triggerData.Activation, status)
		if err != nil {
			sLog.Errorf("V (Stage): failed to report status: %v (%v)", status.ErrorMessage, err)
			return err
		}
		if activation != nil && status.Status != v1alpha2.Done && status.Status != v1alpha2.Paused {
			s.Vendor.Context.Publish("trigger", v1alpha2.Event{
				Body: *activation,
			})
		}
		return nil
	})
	s.Vendor.Context.Subscribe("job-report", func(topic string, event v1alpha2.Event) error {
		sLog.Debugf("V (Stage): handling job report event: %v", event)
		jData, _ := json.Marshal(event.Body)
		var status model.ActivationStatus
		json.Unmarshal(jData, &status)
		if status.Status == v1alpha2.Done || status.Status == v1alpha2.OK {
			campaign, err := s.CampaignsManager.GetSpec(context.Background(), status.Outputs["__campaign"].(string))
			if err != nil {
				sLog.Errorf("V (Stage): failed to get campaign spec '%s': %v", status.Outputs["__campaign"].(string), err)
				return err
			}
			if campaign.Spec.SelfDriving {
				activation, err := s.StageManager.ResumeStage(status, *campaign.Spec)
				if err != nil {
					status.Status = v1alpha2.InternalError
					status.IsActive = false
					status.ErrorMessage = fmt.Sprintf("failed to resume stage: %v", err)
					sLog.Errorf("V (Stage): failed to resume stage: %v", err)
				}
				if activation != nil {
					s.Vendor.Context.Publish("trigger", v1alpha2.Event{
						Body: *activation,
					})
				}
			}
		}

		//TODO: later site overrides reports from earlier sites
		err = s.ActivationsManager.ReportStatus(context.Background(), status.Outputs["__activation"].(string), status)
		if err != nil {
			sLog.Errorf("V (Stage): failed to report status: %v (%v)", status.ErrorMessage, err)
			return err
		}
		return nil
	})
	s.Vendor.Context.Subscribe("remote-job", func(topic string, event v1alpha2.Event) error {
		jData, _ := json.Marshal(event.Body)
		var job v1alpha2.JobData
		json.Unmarshal(jData, &job)
		jData, _ = json.Marshal(job.Body)
		var dataPackage v1alpha2.InputOutputData
		err := json.Unmarshal(jData, &dataPackage)
		if err != nil {
			return err
		}

		triggerData := v1alpha2.ActivationData{
			Activation:           dataPackage.Inputs["__activation"].(string),
			ActivationGeneration: dataPackage.Inputs["__activationGeneration"].(string),
			Campaign:             dataPackage.Inputs["__campaign"].(string),
			Stage:                dataPackage.Inputs["__stage"].(string),
			Inputs:               dataPackage.Inputs,
			Outputs:              dataPackage.Outputs,
		}

		triggerData.Inputs["__origin"] = event.Metadata["origin"]

		switch dataPackage.Inputs["operation"] {
		case "wait":
			triggerData.Provider = "providers.stage.wait"
			config, err := wait.WaitStageProviderConfigFromVendorMap(s.Vendor.Config.Properties)
			if err != nil {
				return err
			}
			triggerData.Config = config
		case "materialize":
			triggerData.Provider = "providers.stage.materialize"
			config, err := materialize.MaterializeStageProviderConfigFromVendorMap(s.Vendor.Config.Properties)
			if err != nil {
				return err
			}
			triggerData.Config = config
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("operation %v is not supported", dataPackage.Inputs["operation"]), v1alpha2.BadRequest)
		}
		status := s.StageManager.HandleDirectTriggerEvent(context.Background(), triggerData)
		sLog.Debugf("V (Stage): reporting status: %v", status)
		s.Vendor.Context.Publish("report", v1alpha2.Event{
			Body: status,
		})
		return nil
	})
	return nil
}
