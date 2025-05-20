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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/campaigns"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/materialize"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/mock"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/wait"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
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
	s.Vendor.Context.Subscribe("activation", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}

			var actData v1alpha2.ActivationData
			jData, _ := json.Marshal(event.Body)
			err := utils2.UnmarshalJson(jData, &actData)
			if err != nil {
				log.ErrorCtx(ctx, "V (Stage): event body of activation event is not ActivationData ")
				return v1alpha2.NewCOAError(nil, "event body is not an activation job", v1alpha2.BadRequest)
			}
			log.InfofCtx(ctx, "V (Stage): handling activation event for activation %s in namespace %s", actData.Activation, actData.Namespace)
			campaignName := api_utils.ConvertReferenceToObjectName(actData.Campaign)

			campaign, err := s.CampaignsManager.GetState(ctx, campaignName, actData.Namespace)
			if err != nil {
				log.ErrorfCtx(ctx, "V (Stage): unable to find campaign %s with error: %+v", campaignName, err)
				err = s.reportActivationStatusWithBadRequest(actData.Activation, actData.Namespace, err)
				// If report status succeeded, return an empty err so the subscribe function will not be retried
				// The actual error will be stored in Activation cr
				return err
			}
			activation, err := s.ActivationsManager.GetState(ctx, actData.Activation, actData.Namespace)
			if err != nil {
				log.ErrorfCtx(ctx, "V (Stage): unable to find activation: %+v", err)
				return nil
			}

			evt, err := s.StageManager.HandleActivationEvent(ctx, actData, *campaign.Spec, activation)
			if err != nil {
				err = s.reportActivationStatusWithBadRequest(actData.Activation, actData.Namespace, err)
				// If report status succeeded, return an empty err so the subscribe function will not be retried
				// The actual error will be stored in Activation cr
				return err
			}

			if evt != nil {
				s.Vendor.Context.Publish("trigger", v1alpha2.Event{
					Body:    *evt,
					Context: ctx,
				})
			}
			// DO NOT REMOVE THIS COMMENT
			// gofail: var afterPublishTrigger string
			return nil
		},
		Group: "0",
	})
	s.Vendor.Context.Subscribe("trigger", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}

			status := model.StageStatus{
				Stage:         "",
				NextStage:     "",
				Outputs:       map[string]interface{}{},
				Status:        v1alpha2.Untouched,
				StatusMessage: v1alpha2.Untouched.String(),
				ErrorMessage:  "",
				IsActive:      true,
			}
			triggerData := v1alpha2.ActivationData{}
			jData, _ := json.Marshal(event.Body)
			err := utils2.UnmarshalJson(jData, &triggerData)
			if err != nil {
				err = v1alpha2.NewCOAError(nil, "event body is not an activation job", v1alpha2.BadRequest)
				sLog.ErrorfCtx(ctx, "V (Stage): failed to deserialize activation data: %v", err)
				err = s.reportActivationStatusWithBadRequest(triggerData.Activation, triggerData.Namespace, err)
				// If report status succeeded, return an empty err so the subscribe function will not be retried
				// The actual error will be stored in Activation cr
				return err
			}
			log.InfoCtx(ctx, "V (Stage): handling trigger event for activation %s stage %s in namespace %s",
				triggerData.Activation, triggerData.Stage, triggerData.Namespace)

			status.Outputs["__namespace"] = triggerData.Namespace
			_, err = s.ActivationsManager.GetState(ctx, triggerData.Activation, triggerData.Namespace)
			if err != nil {
				sLog.ErrorfCtx(ctx, "V (Stage): unable to find activation: %+v", err)
				return nil
			}
			campaignName := api_utils.ConvertReferenceToObjectName(triggerData.Campaign)
			campaign, err := s.CampaignsManager.GetState(ctx, campaignName, triggerData.Namespace)
			if err != nil {
				sLog.ErrorfCtx(ctx, "V (Stage): failed to get campaign spec: %v", err)
				err = s.reportActivationStatusWithBadRequest(triggerData.Activation, triggerData.Namespace, err)
				// If report status succeeded, return an empty err so the subscribe function will not be retried
				// The actual error will be stored in Activation cr
				return err
			}
			status.Stage = triggerData.Stage
			status.ErrorMessage = ""
			status.Status = v1alpha2.Running
			status.StatusMessage = v1alpha2.Running.String()
			if triggerData.NeedsReport {
				sLog.DebugfCtx(ctx, "V (Stage): activation %s, stage %s in namespace %s reporting status: %v", triggerData.Activation, triggerData.Stage, triggerData.Namespace, status)
				s.Vendor.Context.Publish("report", v1alpha2.Event{
					Body:    status,
					Context: ctx,
				})
			} else {
				err = s.ActivationsManager.ReportStageStatus(ctx, triggerData.Activation, triggerData.Namespace, status)
				if err != nil {
					sLog.Errorf("V (Stage): failed to report accepted status: %v (%v)", status.ErrorMessage, err)
					return err
				}
			}

			status, activation := s.StageManager.HandleTriggerEvent(ctx, *campaign.Spec, triggerData)

			if triggerData.NeedsReport {
				sLog.DebugfCtx(ctx, "V (Stage): reporting status: %v", status)
				s.Vendor.Context.Publish("report", v1alpha2.Event{
					Body:    status,
					Context: ctx,
				})

			} else {
				err = s.ActivationsManager.ReportStageStatus(ctx, triggerData.Activation, triggerData.Namespace, status)
				if err != nil {
					sLog.ErrorfCtx(ctx, "V (Stage): failed to report status: %v (%v)", status.ErrorMessage, err)
					return err
				}
				if activation != nil && status.NextStage != "" && status.Status != v1alpha2.Paused {
					s.Vendor.Context.Publish("trigger", v1alpha2.Event{
						Body:    *activation,
						Context: ctx,
					})
				}
			}
			log.InfoCtx(ctx, "V (Stage): Finished handling trigger event")
			return nil
		},
	})
	s.Vendor.Context.Subscribe("job-report", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			sLog.DebugfCtx(ctx, "V (Stage): handling job report event: %v", event)
			jData, _ := json.Marshal(event.Body)
			var status model.StageStatus
			utils2.UnmarshalJson(jData, &status)
			campaign, ok := status.Outputs["__campaign"].(string)
			if !ok {
				sLog.ErrorfCtx(ctx, "V (Stage): failed to get campaign name from job report")
				return v1alpha2.NewCOAError(nil, "job-report: campaign is not valid", v1alpha2.BadRequest)
			}
			namespace, ok := status.Outputs["__namespace"].(string)
			if !ok {
				sLog.ErrorfCtx(ctx, "V (Stage): failed to get namespace from job report, use default instead")
				namespace = "default"
			}
			activation, ok := status.Outputs["__activation"].(string)
			if !ok {
				sLog.ErrorfCtx(ctx, "V (Stage): failed to get activation name from job report")
				return v1alpha2.NewCOAError(nil, "job-report: activation is not valid", v1alpha2.BadRequest)
			}

			err = s.ActivationsManager.ReportStageStatus(ctx, activation, namespace, status)
			if err != nil {
				sLog.ErrorfCtx(ctx, "V (Stage): failed to report status: %v (%v)", status.ErrorMessage, err)
				return err
			}

			if status.Status == v1alpha2.Done || status.Status == v1alpha2.OK {
				campaignName := api_utils.ConvertReferenceToObjectName(campaign)
				campaign, err := s.CampaignsManager.GetState(ctx, campaignName, namespace)
				if err != nil {
					sLog.ErrorfCtx(ctx, "V (Stage): failed to get campaign spec '%s': %v", campaign, err)
					return err
				}
				if campaign.Spec.SelfDriving {
					activation, err := s.StageManager.ResumeStage(ctx, status, *campaign.Spec)
					if err != nil {
						status.Status = v1alpha2.InternalError
						status.StatusMessage = v1alpha2.InternalError.String()
						status.IsActive = false
						status.ErrorMessage = fmt.Sprintf("failed to resume stage: %v", err)
						sLog.ErrorfCtx(ctx, "V (Stage): failed to resume stage: %v", err)
					}
					if activation != nil {
						s.Vendor.Context.Publish("trigger", v1alpha2.Event{
							Body:    *activation,
							Context: ctx,
						})
					}
				}
			}

			return nil
		},
	})
	s.Vendor.Context.Subscribe("remote-job", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			// Unwrap data package from event body
			jData, _ := json.Marshal(event.Body)
			var job v1alpha2.JobData
			utils2.UnmarshalJson(jData, &job)
			jData, _ = json.Marshal(job.Body)
			var dataPackage v1alpha2.InputOutputData
			err := utils2.UnmarshalJson(jData, &dataPackage)
			if err != nil {
				return err
			}

			// restore schedule
			var schedule = ""
			if v, ok := dataPackage.Inputs["__schedule"]; ok {
				schedule = utils.FormatAsString(v)
			}

			triggerData := v1alpha2.ActivationData{
				Activation:           utils.FormatAsString(dataPackage.Inputs["__activation"]),
				ActivationGeneration: utils.FormatAsString(dataPackage.Inputs["__activationGeneration"]),
				Campaign:             utils.FormatAsString(dataPackage.Inputs["__campaign"]),
				Stage:                utils.FormatAsString(dataPackage.Inputs["__stage"]),
				Inputs:               dataPackage.Inputs,
				Outputs:              dataPackage.Outputs,
				Schedule:             schedule,
				NeedsReport:          true,
				Namespace:            utils.FormatAsString(dataPackage.Inputs["__namespace"]),
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
			case "mock":
				triggerData.Provider = "providers.stage.mock"
				config, err := mock.MockStageProviderConfigFromMap(s.Vendor.Config.Properties)
				if err != nil {
					return err
				}
				triggerData.Config = config
			default:
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("operation %v is not supported", dataPackage.Inputs["operation"]), v1alpha2.BadRequest)
			}
			status := s.StageManager.HandleDirectTriggerEvent(ctx, triggerData)
			sLog.DebugfCtx(ctx, "V (Stage): reporting status: %v", status)
			s.Vendor.Context.Publish("report", v1alpha2.Event{
				Body:    status,
				Context: ctx,
			})
			return nil
		},
	})
	return nil
}

func (s *StageVendor) reportActivationStatusWithBadRequest(activation string, namespace string, err error) error {
	status := model.StageStatus{
		Stage:         "",
		NextStage:     "",
		Outputs:       map[string]interface{}{},
		Status:        v1alpha2.BadRequest,
		StatusMessage: v1alpha2.BadRequest.String(),
		ErrorMessage:  err.Error(),
		IsActive:      false,
	}
	err = s.ActivationsManager.ReportStageStatus(context.TODO(), activation, namespace, status)
	if err != nil {
		sLog.Errorf("V (Stage): failed to report error status on activtion %s/%s: %v (%v)", namespace, activation, status.ErrorMessage, err)
	}
	return err
}
