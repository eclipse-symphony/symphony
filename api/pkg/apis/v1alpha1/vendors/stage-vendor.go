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
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/campaigns"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution"
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
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type StageVendor struct {
	vendors.Vendor
	StageManager       *stage.StageManager
	CampaignsManager   *campaigns.CampaignsManager
	ActivationsManager *activations.ActivationsManager
	SolutionManager    *solution.SolutionManager
	PlanManager        *PlanManager
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

func NewPlanManager() *PlanManager {
	return &PlanManager{}
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
		if c, ok := m.(*solution.SolutionManager); ok {
			s.SolutionManager = c
		}
	}
	s.PlanManager = NewPlanManager()
	if s.StageManager == nil {
		return v1alpha2.NewCOAError(nil, "stage manager is not supplied", v1alpha2.MissingConfig)
	}
	if s.CampaignsManager == nil {
		return v1alpha2.NewCOAError(nil, "campaigns manager is not supplied", v1alpha2.MissingConfig)
	}
	if s.ActivationsManager == nil {
		return v1alpha2.NewCOAError(nil, "activations manager is not supplied", v1alpha2.MissingConfig)
	}
	if s.SolutionManager == nil {
		return v1alpha2.NewCOAError(nil, "solution manager is not supplied", v1alpha2.MissingConfig)
	}
	s.Vendor.Context.Subscribe("activation", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}

			var actData v1alpha2.ActivationData
			jData, _ := json.Marshal(event.Body)
			err := json.Unmarshal(jData, &actData)
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
			err := json.Unmarshal(jData, &triggerData)
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
			json.Unmarshal(jData, &status)
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
			json.Unmarshal(jData, &job)
			jData, _ = json.Marshal(job.Body)
			var dataPackage v1alpha2.InputOutputData
			err := json.Unmarshal(jData, &dataPackage)
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
	s.Vendor.Context.Subscribe(DeploymentPlanTopic, v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}

			log.InfoCtx(ctx, "V(StageVendor): Begin to execute deployment-plan")
			return s.handleDeploymentPlan(ctx, event)
		},
		Group: "stage-vendor",
	})

	s.Vendor.Context.Subscribe(CollectStepResultTopic, v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := event.Context
			if ctx == nil {
				ctx = context.TODO()
			}
			return s.handleStepResult(ctx, event)
		},
		Group: "stage-vendor",
	})
	return nil
}

// handleStepResult processes the event and updates the plan state accordingly.
func (s *StageVendor) handleStepResult(ctx context.Context, event v1alpha2.Event) error {
	var stepResult StepResult

	// Marshal the event body to JSON
	jData, _ := json.Marshal(event.Body)
	log.InfofCtx(ctx, "Received event body: %s", string(jData))

	// Unmarshal the JSON data into stepResult
	if err := json.Unmarshal(jData, &stepResult); err != nil {
		log.ErrorfCtx(ctx, "Failed to unmarshal step result: %v", err)
		return err
	}

	planId := stepResult.PlanId

	// Load the plan state object from the PlanManager
	planStateObj, exists := s.PlanManager.Plans.Load(planId)
	if !exists {
		log.ErrorCtx(ctx, "Plan not found: %s", planId)
		return fmt.Errorf("Plan not found: %s", planId)
	}
	planState := planStateObj.(*PlanState)

	// Update the plan state in the map and save the summary
	if err := s.saveStepResult(ctx, planState, stepResult); err != nil {
		log.ErrorCtx(ctx, "Failed to update plan state: %v", err)
		return err
	}

	return nil
}
func (s *StageVendor) handleDeploymentPlan(ctx context.Context, event v1alpha2.Event) error {
	var planEnvelope PlanEnvelope
	jData, _ := json.Marshal(event.Body)
	err := json.Unmarshal(jData, &planEnvelope)
	if err != nil {
		log.ErrorCtx(ctx, "failed to unmarshal plan envelope :%v", err)
		return err
	}
	planState := s.createPlanState(ctx, planEnvelope)
	s.SaveSummaryInfo(ctx, planState, model.SummaryStateRunning)
	if planState.isCompleted() {
		return s.handlePlanComplete(ctx, planState)

	}
	for i, step := range planEnvelope.Plan.Steps {
		switch planEnvelope.Phase {
		case PhaseGet:
			log.InfoCtx(ctx, "phase get begin deployment %+v", planEnvelope.Deployment)
			if err := s.publishStepResult(ctx, i, step, planState, planEnvelope); err != nil {
				log.InfoCtx(ctx, "V(Federation): publish deployment step failed PlanId %s, stepId %s", planEnvelope.PlanId, i)
				return err
			}
		case PhaseApply:
			planState.Summary.PlannedDeployment += len(step.Components)
			log.InfoCtx(ctx, "V(Federation): publish deployment step id %s step %+v", i, step.Role)
			if err := s.publishStepResult(ctx, i, step, planState, planEnvelope); err != nil {
				log.InfoCtx(ctx, "V(Federation): publish deployment step failed PlanId %s, stepId %s", planEnvelope.PlanId, i)
				return err
			}
		}
	}
	log.InfoCtx(ctx, "V(Federation): store plan id %s in map %+v", planEnvelope.PlanId)
	s.PlanManager.Plans.Store(planEnvelope.PlanId, planState)
	return nil
}
func (s *StageVendor) publishStepResult(ctx context.Context, stepId int, step model.DeploymentStep, planState *PlanState, planEnvelope PlanEnvelope) error {
	log.InfoCtx(ctx, "V(StageVendor): publish deployment step for PlanId %s StepId %s", planState.PlanId, stepId)
	if err := s.Vendor.Context.Publish("deployment-step", v1alpha2.Event{
		Body: StepEnvelope{
			Step:      step,
			Remove:    planEnvelope.Remove,
			StepId:    stepId,
			PlanState: planState,
		},
		Context: ctx,
	}); err != nil {
		log.InfoCtx(ctx, "V(StageVendor): publish deployment step failed PlanId %s, stepId %s", planEnvelope.PlanId, stepId)
		return err
	}
	return nil
}

// create inital plan state
func (s *StageVendor) createPlanState(ctx context.Context, planEnvelope PlanEnvelope) *PlanState {
	return &PlanState{
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
	}
}

// saveStepResult updates the plan state with the step result and saves the summary.
func (s *StageVendor) saveStepResult(ctx context.Context, planState *PlanState, stepResult StepResult) error {
	// Log the update of plan state with the step result
	log.InfoCtx(ctx, "V(Stage): Update plan state %v with step result %v phase %s", planState, stepResult, planState.Phase)
	planState.CompletedSteps++

	switch planState.Phase {
	case PhaseGet:
		// Update the GetResult for the specific step
		planState.StepStates[stepResult.StepId].GetResult = stepResult.GetResult
	case PhaseApply:
		if stepResult.Error != nil {
			// Handle error case and update the target result status and message
			targetResultStatus := fmt.Sprintf("%s Failed", deploymentTypeMap[planState.Remove])
			targetResultMessage := fmt.Sprintf("Failed to create provider %s, err: %s", deploymentTypeMap[planState.Remove], stepResult.Error)
			targetResultSpec := model.TargetResultSpec{Status: targetResultStatus, Message: targetResultMessage}
			planState.Summary.UpdateTargetResult(stepResult.Target, targetResultSpec)
			planState.Summary.AllAssignedDeployed = false
			for _, ret := range stepResult.ApplyResult {
				if (!planState.Remove && ret.Status == v1alpha2.Updated) || (planState.Remove && ret.Status == v1alpha2.Deleted) {
					planState.Summary.CurrentDeployed++
				}
			}
		} else {
			// Handle success case and update the target result status and message
			if planState.TargetResult[stepResult.Target] == 0 {
				planState.TargetResult[stepResult.Target] = 1
				planState.Summary.SuccessCount++
			}
			targetResultSpec := model.TargetResultSpec{Status: "OK", Message: "", ComponentResults: stepResult.ApplyResult}
			planState.Summary.UpdateTargetResult(stepResult.Target, targetResultSpec)
			log.InfoCtx(ctx, "Update plan state target spec %v", targetResultSpec)
			planState.Summary.CurrentDeployed += len(stepResult.ApplyResult)
		}

		// If no components are deployed, set success count to target count
		if planState.Summary.CurrentDeployed == 0 && planState.Summary.AllAssignedDeployed {
			planState.Summary.SuccessCount = planState.Summary.TargetCount
		}

		// Save the summary information
		if err := s.SaveSummaryInfo(ctx, planState, model.SummaryStateRunning); err != nil {
			log.ErrorfCtx(ctx, "Failed to save summary progress: %v", err)
		}
	}

	// Store the updated plan state
	s.PlanManager.Plans.Store(planState.PlanId, planState)

	// Check if all steps are completed and handle plan completion
	if planState.isCompleted() {
		return s.handlePlanComplete(ctx, planState)
	}

	return nil
}

// handlePlanComplete handles the completion of a plan and updates its status.
func (s *StageVendor) handlePlanComplete(ctx context.Context, planState *PlanState) error {
	log.InfoCtx(ctx, "V(Stage): Plan state %s is completed %v", planState.Phase, planState)
	if !planState.Summary.AllAssignedDeployed {
		planState.Status = "failed"
	}
	log.InfoCtx(ctx, "V(Stage): Plan state is completed %v", planState.Summary.AllAssignedDeployed)
	switch planState.Phase {
	case PhaseGet:
		if err := s.handleGetPlanCompletetion(ctx, planState); err != nil {
			log.ErrorfCtx(ctx, "V(Stage): Failed to handle get plan completion: %v", err)
			return err
		}
	case PhaseApply:
		if err := s.handleApplyPlanCompletetion(ctx, planState); err != nil {
			log.ErrorfCtx(ctx, "V(Stage): Failed to handle apply plan completion: %v", err)
			return err
		}
	}
	s.PlanManager.DeletePlan(planState.PlanId)
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

// handleGetPlanCompletetion handles the completion of the get plan phase.
func (s *StageVendor) handleGetPlanCompletetion(ctx context.Context, planState *PlanState) error {
	// Collect result
	log.InfoCtx(ctx, "V(Stage): Begin to get current state %v", planState)
	Plan, err := s.threeStateMerge(ctx, planState)
	if err != nil {
		log.ErrorfCtx(ctx, "V(Stage): Failed to merge states: %v", err)
		return err
	}
	s.Vendor.Context.Publish(DeploymentPlanTopic, v1alpha2.Event{
		Metadata: map[string]string{
			"Id": planState.Deployment.JobID,
		},
		Body: PlanEnvelope{
			Plan:                 Plan,
			Deployment:           planState.Deployment,
			MergedState:          planState.MergedState,
			CurrentState:         planState.CurrentState,
			PreviousDesiredState: planState.PreviousDesiredState,
			PlanId:               planState.Deployment.Instance.ObjectMeta.Name,
			Remove:               planState.Remove,
			Namespace:            planState.Namespace,
			Phase:                PhaseApply,
		},
		Context: ctx,
	})
	return nil
}

// threeStateMerge merges the current, previous, and desired states to create a deployment plan.
func (s *StageVendor) threeStateMerge(ctx context.Context, planState *PlanState) (model.DeploymentPlan, error) {
	currentState := model.DeploymentState{}
	currentState.TargetComponent = make(map[string]string)

	for _, StepState := range planState.StepStates {
		for _, c := range StepState.GetResult {
			key := fmt.Sprintf("%s::%s", c.Name, StepState.Target)
			role := c.Type
			if role == "" {
				role = "container"
			}
			log.InfoCtx(ctx, "V(Stage): Store key value in current key: %s value: %s", key, role)
			currentState.TargetComponent[key] = role
		}
	}
	log.InfoCtx(ctx, "V(Stage): Compute current state %v for Plan ID: %s", currentState, planState.PlanId)
	planState.CurrentState = currentState
	previousDesiredState := s.SolutionManager.GetPreviousState(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Namespace)
	planState.PreviousDesiredState = previousDesiredState
	var currentDesiredState model.DeploymentState
	currentDesiredState, err := solution.NewDeploymentState(planState.Deployment)
	if err != nil {
		log.ErrorfCtx(ctx, "V(Stage): Failed to get current desired state: %+v", err)
		return model.DeploymentPlan{}, err
	}
	log.InfoCtx(ctx, "V(Stage): Get current desired state %+v", currentDesiredState)
	desiredState := currentDesiredState
	if previousDesiredState != nil {
		desiredState = solution.MergeDeploymentStates(&previousDesiredState.State, currentDesiredState)
	}
	log.InfoCtx(ctx, "V(Stage): Get desired state %+v", desiredState)
	if planState.Remove {
		desiredState.MarkRemoveAll()
		log.InfoCtx(ctx, "V(Stage): After remove desired state %+v", desiredState)
	}

	mergedState := solution.MergeDeploymentStates(&currentState, desiredState)
	planState.MergedState = mergedState
	Plan, err := solution.PlanForDeployment(planState.Deployment, mergedState)
	if err != nil {
		log.ErrorfCtx(ctx, "V(Stage): Plan generate error")
		return model.DeploymentPlan{}, err
	}
	log.InfoCtx(ctx, "V(Stage): Begin to publish topic to deployment plan %v merged state %v get plan %v", planState, mergedState, Plan)
	return Plan, nil
}

func (s *StageVendor) SaveSummaryInfo(ctx context.Context, planState *PlanState, state model.SummaryState) error {
	return s.SolutionManager.SaveSummary(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Deployment.Generation, planState.Deployment.Hash, planState.Summary, model.SummaryStateRunning, planState.Namespace)
}
func (s *StageVendor) handleApplyPlanCompletetion(ctx context.Context, planState *PlanState) error {
	log.InfofCtx(ctx, "handle plan completetion:begin to handle plan completetion %v", planState)
	if err := s.SaveSummaryInfo(ctx, planState, model.SummaryStateDone); err != nil {
		log.ErrorfCtx(ctx, "Failed to save summary progress done: %v", err)
		return err
	}
	// update summary
	if err := s.SolutionManager.ConcludeSummary(ctx, planState.Deployment.Instance.ObjectMeta.Name, planState.Deployment.Generation, planState.Deployment.Hash, planState.Summary, planState.Namespace); err != nil {
		log.ErrorfCtx(ctx, "handle plan completetion: failed to conclude summary: %v", err)
		return err
	}
	planState.MergedState.ClearAllRemoved()

	if !planState.Deployment.IsDryRun {
		if len(planState.MergedState.TargetComponent) == 0 && planState.Remove {
			log.DebugfCtx(ctx, " M (Solution): no assigned components to manage, deleting state")
			s.SolutionManager.StateProvider.Delete(ctx, states.DeleteRequest{
				ID: planState.Deployment.Instance.ObjectMeta.Name,
				Metadata: map[string]interface{}{
					"namespace": planState.Namespace,
					"group":     model.SolutionGroup,
					"version":   "v1",
					"resource":  DeploymentState,
				},
			})
		} else {
			s.SolutionManager.StateProvider.Upsert(ctx, states.UpsertRequest{
				Value: states.StateEntry{
					ID: planState.Deployment.Instance.ObjectMeta.Name,
					Body: solution.SolutionManagerDeploymentState{
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
	s.SolutionManager.KeyLockProvider.UnLock(api_utils.GenerateKeyLockName(planState.Namespace, planState.Deployment.Instance.ObjectMeta.Name))
	return nil
}
func (p *PlanState) IsExpired() bool {
	log.Info("time now")
	log.Info("time expired")
	return time.Now().After(p.ExpireTime)
}

func (p *PlanState) isCompleted() bool {
	return p.CompletedSteps == p.TotalSteps
}
func (pm *PlanManager) GetPlan(planId string) (*PlanState, bool) {
	if value, ok := pm.Plans.Load(planId); ok {
		return value.(*PlanState), true
	}
	return nil, false
}
func (pm *PlanManager) DeletePlan(planId string) {
	pm.Plans.Delete(planId)
}
