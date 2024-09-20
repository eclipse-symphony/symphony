/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package stage

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	symproviders "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/remote"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
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

func (t *TaskResult) GetError() error {
	if t.Error != nil {
		return t.Error
	}
	if v, ok := t.Outputs["__status"]; ok {
		switch sv := v.(type) {
		case v1alpha2.State:
			break
		case float64:
			state := v1alpha2.State(int(sv))
			stateValue := reflect.ValueOf(state)
			if stateValue.Type() != reflect.TypeOf(v1alpha2.State(0)) {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %v", sv), v1alpha2.InternalError)
			}
			t.Outputs["__status"] = state
		case int:
			state := v1alpha2.State(sv)
			stateValue := reflect.ValueOf(state)
			if stateValue.Type() != reflect.TypeOf(v1alpha2.State(0)) {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %d", sv), v1alpha2.InternalError)
			}
			t.Outputs["__status"] = state
		case string:
			vInt, err := strconv.ParseInt(sv, 10, 32)
			if err != nil {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %s", sv), v1alpha2.InternalError)
			}
			state := v1alpha2.State(vInt)
			stateValue := reflect.ValueOf(state)
			if stateValue.Type() != reflect.TypeOf(v1alpha2.State(0)) {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %d", vInt), v1alpha2.InternalError)
			}
			t.Outputs["__status"] = state
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid state %v", v), v1alpha2.InternalError)
		}

		if t.Outputs["__status"] != v1alpha2.OK {
			if v, ok := t.Outputs["__error"]; ok {
				return v1alpha2.NewCOAError(nil, utils.FormatAsString(v), t.Outputs["__status"].(v1alpha2.State))
			} else {
				return v1alpha2.NewCOAError(nil, "stage returned unsuccessful status without an error", v1alpha2.InternalError)
			}
		}
	}
	return nil
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
	// TODO: change volatileStateProvider to persistentStateProvider
	stateprovider, err := managers.GetVolatileStateProvider(config, providers)
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
func (s *StageManager) ResumeStage(ctx context.Context, status model.StageStatus, cam model.CampaignSpec) (*v1alpha2.ActivationData, error) {
	log.InfofCtx(ctx, " M (Stage): ResumeStage: %v\n", status)
	campaign, ok := status.Outputs["__campaign"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: campaign (%v) is not valid from output", status.Outputs["__campaign"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: campaign is not valid", v1alpha2.BadRequest)
	}
	activation, ok := status.Outputs["__activation"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: activation (%v) is not valid from output", status.Outputs["__activation"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: activation is not valid", v1alpha2.BadRequest)
	}
	activationGeneration, ok := status.Outputs["__activationGeneration"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: activationGeneration (%v) is not valid from output", status.Outputs["__activationGeneration"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: activationGeneration is not valid", v1alpha2.BadRequest)
	}
	site, ok := status.Outputs["__site"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: site (%v) is not valid from output", status.Outputs["__site"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: site is not valid", v1alpha2.BadRequest)
	}
	stage, ok := status.Outputs["__stage"].(string)
	if !ok {
		log.ErrorfCtx(ctx, " M (Stage): ResumeStage: stage (%v) is not valid from output", status.Outputs["__stage"])
		return nil, v1alpha2.NewCOAError(nil, "ResumeStage: stage is not valid", v1alpha2.BadRequest)
	}
	namespace, ok := status.Outputs["__namespace"].(string)
	if !ok {
		namespace = "default"
	}

	entry, err := s.StateProvider.Get(ctx, states.GetRequest{
		ID: fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
		Metadata: map[string]interface{}{
			"namespace": namespace,
		},
	})
	if err != nil {
		return nil, err
	}
	jData, _ := json.Marshal(entry.Body)
	var p PendingTask
	err = json.Unmarshal(jData, &p)
	if err == nil {
		// //find site in p.Sites
		// found := false
		// for _, s := range p.Sites {
		// 	if s == site {
		// 		found = true
		// 		break
		// 	}
		// }
		// if !found {
		// 	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("site %s is not found in pending task", site)
		// }
		//remove site from p.Sites
		newSites := make([]string, 0)
		for _, s := range p.Sites {
			if s != site {
				newSites = append(newSites, s)
			}
		}
		if len(newSites) == 0 {
			log.InfofCtx(ctx, " M (Stage): ResumeStage: all sites are done for activation %s stage %s. Check if we need to move to next stage", activation, stage)
			err := s.StateProvider.Delete(ctx, states.DeleteRequest{
				ID: fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
				Metadata: map[string]interface{}{
					"namespace": namespace,
				},
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
					eCtx.Context = ctx
					eCtx.Inputs = status.Inputs
					log.DebugfCtx(ctx, " M (Stage): ResumeStage evaluation inputs: %v", eCtx.Inputs)
					if eCtx.Inputs != nil {
						if v, ok := eCtx.Inputs["context"]; ok {
							eCtx.Value = v
						}
					}
					eCtx.Outputs = outputs
					val, err := parser.Eval(*eCtx)
					if err != nil {
						return nil, err
					}

					sVal := ""
					if val != nil {
						sVal = utils.FormatAsString(val)
					}
					if sVal != "" {
						if _, ok := cam.Stages[sVal]; ok {
							nextStage = sVal
						} else {
							return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", sVal), v1alpha2.InternalError)
						}
					}

				}
				if nextStage != "" {
					activationData := &v1alpha2.ActivationData{
						Campaign:             campaign,
						Activation:           activation,
						ActivationGeneration: activationGeneration,
						Stage:                nextStage,
						Inputs:               status.Inputs,
						Provider:             cam.Stages[nextStage].Provider,
						Config:               cam.Stages[nextStage].Config,
						Outputs:              outputs,
						TriggeringStage:      stage,
						Schedule:             cam.Stages[nextStage].Schedule,
						Namespace:            namespace,
					}
					log.InfofCtx(ctx, " M (Stage): Activating next stage: %s\n", activationData.Stage)
					return activationData, nil
				} else {
					log.InfoCtx(ctx, " M (Stage): No next stage found\n")
					return nil, nil
				}
			}
			return nil, nil
		} else {
			log.InfoCtx(ctx, " M (Stage): ResumeStage: updating pending sites %v for activation %s stage %s", newSites, activation, stage)
			p.Sites = newSites
			// TODO: clean up the remote job status entry for multi-site
			_, err := s.StateProvider.Upsert(ctx, states.UpsertRequest{
				Value: states.StateEntry{
					ID:   fmt.Sprintf("%s-%s-%s", campaign, activation, activationGeneration),
					Body: p,
				},
				Metadata: map[string]interface{}{
					"namespace": namespace,
				},
			})
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, v1alpha2.NewCOAError(err, "invalid pending task", v1alpha2.InternalError)
	}

	return nil, nil
}
func (s *StageManager) HandleDirectTriggerEvent(ctx context.Context, triggerData v1alpha2.ActivationData) model.StageStatus {
	ctx, span := observability.StartSpan("Stage Manager", ctx, &map[string]string{
		"method": "HandleDirectTriggerEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfoCtx(ctx, " M (Stage): HandleDirectTriggerEvent for campaign %s, activation %s, stage %s", triggerData.Campaign, triggerData.Activation, triggerData.Stage)

	status := model.StageStatus{
		Stage:     "",
		NextStage: "",
		Outputs: map[string]interface{}{
			"__campaign":             triggerData.Campaign,
			"__namespace":            triggerData.Namespace,
			"__activation":           triggerData.Activation,
			"__activationGeneration": triggerData.ActivationGeneration,
			"__stage":                triggerData.Stage,
			"__site":                 s.VendorContext.SiteInfo.SiteId,
		},
		Status:        v1alpha2.Untouched,
		StatusMessage: v1alpha2.Untouched.String(),
		ErrorMessage:  "",
		IsActive:      true,
	}
	var provider providers.IProvider
	factory := symproviders.SymphonyProviderFactory{}
	provider, err = factory.CreateProvider(triggerData.Provider, triggerData.Config)
	if err != nil {
		status.Status = v1alpha2.InternalError
		status.StatusMessage = v1alpha2.InternalError.String()
		status.ErrorMessage = err.Error()
		status.IsActive = false
		log.ErrorfCtx(ctx, " M (Stage): failed to create provider: %v", err)
		return status
	}
	if provider == nil {
		status.Status = v1alpha2.BadRequest
		status.StatusMessage = v1alpha2.BadRequest.String()
		status.ErrorMessage = fmt.Sprintf("provider %s is not found", triggerData.Provider)
		status.IsActive = false
		log.ErrorfCtx(ctx, " M (Stage): failed to create provider: %v", err)
		return status
	}

	if _, ok := provider.(contexts.IWithManagerContext); ok {
		provider.(contexts.IWithManagerContext).SetContext(s.Manager.Context)
	} else {
		log.ErrorfCtx(ctx, " M (Stage): provider %s does not implement IWithManagerContext", triggerData.Provider)
	}

	isRemote := false
	if _, ok := provider.(*remote.RemoteStageProvider); ok {
		isRemote = true
		provider.(*remote.RemoteStageProvider).SetOutputsContext(triggerData.Outputs)
	}

	if triggerData.Schedule != "" && !isRemote {
		log.InfofCtx(ctx, " M (Stage): send schedule event and pause stage %s for site %s", triggerData.Stage, s.VendorContext.SiteInfo.SiteId)
		s.Context.Publish("schedule", v1alpha2.Event{
			Body:    triggerData,
			Context: ctx,
		})
		status.Outputs["__status"] = v1alpha2.Paused
		status.Status = v1alpha2.Paused
		status.StatusMessage = v1alpha2.Paused.String()
		status.IsActive = false
		return status
	}

	var outputs map[string]interface{}
	outputs, _, err = provider.(stage.IStageProvider).Process(ctx, *s.Manager.Context, triggerData.Inputs)

	result := TaskResult{
		Outputs: outputs,
		Error:   err,
		Site:    s.VendorContext.SiteInfo.SiteId,
	}

	err = result.GetError()
	if err != nil {
		status.Status = v1alpha2.InternalError
		status.StatusMessage = v1alpha2.InternalError.String()
		status.ErrorMessage = err.Error()
		status.IsActive = false
		status.Outputs = carryOutPutsToErrorStatus(outputs, err, "")
		result.Outputs = carryOutPutsToErrorStatus(outputs, err, "")
		return status
	}

	// Merge outputs, provider output overwrite status.Outputs for the same key
	for k, v := range outputs {
		status.Outputs[k] = v
	}

	status.Outputs["__status"] = v1alpha2.OK
	status.Status = v1alpha2.Done
	status.StatusMessage = v1alpha2.Done.String()
	status.IsActive = false
	return status
}
func carryOutPutsToErrorStatus(outputs map[string]interface{}, err error, site string) map[string]interface{} {
	ret := make(map[string]interface{})
	statusKey := "__status"
	if site != "" {
		statusKey = fmt.Sprintf("%s.%s", statusKey, site)
	}
	errorKey := "__error"
	if site != "" {
		errorKey = fmt.Sprintf("%s.%s", errorKey, site)
	}
	for k, v := range outputs {
		ret[k] = v
	}
	if _, ok := ret[statusKey]; !ok {
		if cErr, ok := err.(v1alpha2.COAError); ok {
			ret[statusKey] = cErr.State
		} else {
			ret[statusKey] = v1alpha2.InternalError
		}
	}
	if _, ok := ret[errorKey]; !ok {
		ret[errorKey] = err.Error()
	}
	return ret
}
func (s *StageManager) HandleTriggerEvent(ctx context.Context, campaign model.CampaignSpec, triggerData v1alpha2.ActivationData) (model.StageStatus, *v1alpha2.ActivationData) {
	ctx, span := observability.StartSpan("Stage Manager", ctx, &map[string]string{
		"method": "HandleTriggerEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfoCtx(ctx, " M (Stage): HandleTriggerEvent for campaign %s, activation %s, stage %s", triggerData.Campaign, triggerData.Activation, triggerData.Stage)
	status := model.StageStatus{
		Stage:         triggerData.Stage,
		NextStage:     "",
		Outputs:       map[string]interface{}{},
		Status:        v1alpha2.Untouched,
		StatusMessage: v1alpha2.Untouched.String(),
		ErrorMessage:  "",
		IsActive:      true,
	}
	var activationData *v1alpha2.ActivationData
	if currentStage, ok := campaign.Stages[triggerData.Stage]; ok {
		sites := make([]string, 0)
		if currentStage.Contexts != "" {
			log.InfoCtx(ctx, " M (Stage): evaluating context %s", currentStage.Contexts)
			parser := utils.NewParser(currentStage.Contexts)

			eCtx := s.VendorContext.EvaluationContext.Clone()
			eCtx.Context = ctx
			eCtx.Inputs = triggerData.Inputs
			if eCtx.Inputs != nil {
				if v, ok := eCtx.Inputs["context"]; ok {
					eCtx.Value = v
				}
			}
			eCtx.Outputs = triggerData.Outputs
			var val interface{}
			val, err = parser.Eval(*eCtx)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to evaluate context: %v", err)
				return status, activationData
			}
			if valStringList, ok := val.([]string); ok {
				sites = valStringList
			} else if valList, ok := val.([]interface{}); ok {
				for _, v := range valList {
					sites = append(sites, utils.FormatAsString(v))
				}
			} else if valString, ok := val.(string); ok {
				sites = append(sites, valString)
			} else {
				status.Status = v1alpha2.BadConfig
				status.StatusMessage = v1alpha2.BadConfig.String()
				status.ErrorMessage = fmt.Sprintf("invalid context %s", currentStage.Contexts)
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): invalid context: %v", currentStage.Contexts)
				return status, activationData
			}
			log.InfofCtx(ctx, " M (Stage): evaluated context %s to %v", currentStage.Contexts, sites)
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

		// inject default inputs
		inputs["__campaign"] = triggerData.Campaign
		inputs["__namespace"] = triggerData.Namespace
		inputs["__activation"] = triggerData.Activation
		inputs["__stage"] = triggerData.Stage
		inputs["__activationGeneration"] = triggerData.ActivationGeneration
		inputs["__previousStage"] = triggerData.TriggeringStage
		inputs["__site"] = s.VendorContext.SiteInfo.SiteId
		if triggerData.Schedule != "" {
			inputs["__schedule"] = triggerData.Schedule
		}

		for k, v := range inputs {
			var val interface{}
			val, err = s.traceValue(ctx, v, inputs, triggerData.Outputs)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to evaluate input: %v", err)
				return status, activationData
			}
			inputs[k] = val
		}
		status.Inputs = map[string]interface{}{}
		for k, v := range inputs {
			if !strings.HasPrefix(k, "__") {
				status.Inputs[k] = v
			}
		}

		if triggerData.Outputs != nil {
			if v, ok := triggerData.Outputs[triggerData.Stage]; ok {
				if vs, ok := v["__state"]; ok {
					inputs["__state"] = vs
				}
			}
		}

		factory := symproviders.SymphonyProviderFactory{}
		var provider providers.IProvider
		provider, err = factory.CreateProvider(triggerData.Provider, triggerData.Config)
		if err != nil {
			status.Status = v1alpha2.InternalError
			status.StatusMessage = v1alpha2.InternalError.String()
			status.ErrorMessage = err.Error()
			status.IsActive = false
			log.ErrorfCtx(ctx, " M (Stage): failed to create provider: %v", err)
			return status, activationData
		}
		if provider == nil {
			status.Status = v1alpha2.BadRequest
			status.StatusMessage = v1alpha2.BadRequest.String()
			status.ErrorMessage = fmt.Sprintf("provider %s is not found", triggerData.Provider)
			status.IsActive = false
			log.ErrorfCtx(ctx, " M (Stage): failed to create provider: %v", err)
			return status, activationData
		}

		if _, ok := provider.(contexts.IWithManagerContext); ok {
			provider.(contexts.IWithManagerContext).SetContext(s.Manager.Context)
		} else {
			log.ErrorfCtx(ctx, " M (Stage): provider %s does not implement IWithManagerContext", triggerData.Provider)
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
					inputCopy[k] = v
				}
				inputCopy["__site"] = site

				for k, v := range inputCopy {
					var val interface{}
					val, err = s.traceValue(ctx, v, inputCopy, triggerData.Outputs)
					if err != nil {
						status.Status = v1alpha2.InternalError
						status.StatusMessage = v1alpha2.InternalError.String()
						status.ErrorMessage = err.Error()
						status.IsActive = false
						log.ErrorfCtx(ctx, " M (Stage): failed to evaluate input: %v", err)
						results <- TaskResult{
							Outputs: nil,
							Error:   err,
							Site:    site,
						}
						return
					}
					inputCopy[k] = val
				}

				if _, ok := provider.(*remote.RemoteStageProvider); ok {
					provider.(*remote.RemoteStageProvider).SetOutputsContext(triggerData.Outputs)
				}

				if triggerData.Schedule != "" {
					log.InfofCtx(ctx, " M (Stage): send schedule event and pause stage %s for site %s", triggerData.Stage, site)
					s.Context.Publish("schedule", v1alpha2.Event{
						Body:    triggerData,
						Context: ctx,
					})
					pauseRequested = true
					results <- TaskResult{
						Outputs: nil,
						Error:   nil,
						Site:    site,
					}
				} else {
					var outputs map[string]interface{}
					var pause bool
					outputs, pause, err = provider.(stage.IStageProvider).Process(ctx, *s.Manager.Context, inputCopy)

					if pause {
						log.InfofCtx(ctx, " M (Stage): stage %s in activation %s for site %s get paused result from stage provider", triggerData.Stage, triggerData.Activation, site)
						pauseRequested = true
					}
					results <- TaskResult{
						Outputs: outputs,
						Error:   err,
						Site:    site,
					}
				}
			}(&waitGroup, site, results)
		}

		waitGroup.Wait()
		close(results)

		outputs := make(map[string]interface{})
		delayedExit := false
		for result := range results {
			err = result.GetError()
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = fmt.Sprintf("%s: %s", result.Site, err.Error())
				status.IsActive = false
				site := result.Site
				if result.Site == s.Context.SiteInfo.SiteId {
					site = ""
				}
				status.Outputs = carryOutPutsToErrorStatus(nil, err, site)
				result.Outputs = carryOutPutsToErrorStatus(nil, err, site)
				log.ErrorfCtx(ctx, " M (Stage): failed to process stage %s for site %s outputs: %v", triggerData.Stage, site, err)
				delayedExit = true
			}
			for k, v := range result.Outputs {
				if result.Site == s.Context.SiteInfo.SiteId {
					outputs[k] = v
				} else {
					outputs[fmt.Sprintf("%s.%s", result.Site, k)] = v
				}
			}
			if result.Site == s.Context.SiteInfo.SiteId {
				if _, ok := result.Outputs["__status"]; !ok {
					outputs["__status"] = v1alpha2.OK
				}
			} else {
				key := fmt.Sprintf("%s.__status", result.Site)
				if _, ok := result.Outputs[key]; !ok {
					outputs[key] = v1alpha2.Untouched
				}
			}
		}

		for k, v := range outputs {
			if !strings.HasPrefix(k, "__") {
				status.Outputs[k] = v
			}
		}
		if triggerData.Outputs == nil {
			triggerData.Outputs = make(map[string]map[string]interface{})
		}
		triggerData.Outputs[triggerData.Stage] = outputs
		// If stage is paused, save the pending task and return paused status
		if pauseRequested {
			pendingTask := PendingTask{
				Sites:         sites,
				OutputContext: triggerData.Outputs,
			}
			_, err = s.StateProvider.Upsert(ctx, states.UpsertRequest{
				Value: states.StateEntry{
					ID:   fmt.Sprintf("%s-%s-%s", triggerData.Campaign, triggerData.Activation, triggerData.ActivationGeneration),
					Body: pendingTask,
				},
				Metadata: map[string]interface{}{
					"namespace": triggerData.Namespace,
				},
			})
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to save pending task: %v", err)
				return status, activationData
			}
			status.Status = v1alpha2.Paused
			status.StatusMessage = v1alpha2.Paused.String()
			status.IsActive = false
			return status, activationData
		}
		if campaign.SelfDriving {
			parser := utils.NewParser(currentStage.StageSelector)
			eCtx := s.VendorContext.EvaluationContext.Clone()
			eCtx.Context = ctx
			eCtx.Inputs = triggerData.Inputs
			if eCtx.Inputs != nil {
				if v, ok := eCtx.Inputs["context"]; ok {
					eCtx.Value = v
				}
			}
			eCtx.Outputs = triggerData.Outputs
			var val interface{}
			val, err = parser.Eval(*eCtx)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.StatusMessage = v1alpha2.InternalError.String()
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.ErrorfCtx(ctx, " M (Stage): failed to evaluate stage selector: %v", err)
				return status, activationData
			}

			sVal := ""
			if val != nil {
				sVal = utils.FormatAsString(val)
			}

			if sVal != "" {
				if nextStage, ok := campaign.Stages[sVal]; ok {
					if !delayedExit || nextStage.HandleErrors {
						status.NextStage = sVal
						activationData = &v1alpha2.ActivationData{
							Campaign:             triggerData.Campaign,
							Activation:           triggerData.Activation,
							ActivationGeneration: triggerData.ActivationGeneration,
							Stage:                sVal,
							Inputs:               triggerData.Inputs,
							Outputs:              triggerData.Outputs,
							Provider:             nextStage.Provider,
							Config:               nextStage.Config,
							TriggeringStage:      triggerData.Stage,
							Schedule:             nextStage.Schedule,
							Namespace:            triggerData.Namespace,
						}
					} else {
						status.Status = v1alpha2.InternalError
						status.StatusMessage = v1alpha2.InternalError.String()
						status.ErrorMessage = fmt.Sprintf("stage %s failed", triggerData.Stage)
						status.IsActive = false
						log.ErrorfCtx(ctx, " M (Stage): failed to process stage outputs: %v", status.ErrorMessage)
						return status, activationData
					}
				} else {
					err = v1alpha2.NewCOAError(nil, status.ErrorMessage, v1alpha2.BadRequest)
					status.Status = v1alpha2.BadRequest
					status.StatusMessage = v1alpha2.BadRequest.String()
					status.ErrorMessage = fmt.Sprintf("stage %s is not found", sVal)
					status.IsActive = false
					log.ErrorfCtx(ctx, " M (Stage): failed to find next stage: %v", err)
					return status, activationData
				}
			}
			// sVal is empty, no next stage
			status.NextStage = sVal
			status.IsActive = false
			status.Status = v1alpha2.Done
			status.StatusMessage = v1alpha2.Done.String()
			log.InfofCtx(ctx, " M (Stage): stage %s is done", triggerData.Stage)
			return status, activationData
		} else {
			// Not self-driving, no next stage
			status.Status = v1alpha2.Done
			status.StatusMessage = v1alpha2.Done.String()
			status.NextStage = ""
			status.IsActive = false
			log.InfofCtx(ctx, " M (Stage): stage %s is done (no next stage)", triggerData.Stage)
			return status, activationData
		}
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", triggerData.Stage), v1alpha2.BadRequest)
	status.Status = v1alpha2.InternalError
	status.StatusMessage = v1alpha2.InternalError.String()
	status.ErrorMessage = err.Error()
	status.IsActive = false
	log.ErrorfCtx(ctx, " M (Stage): failed to find stage: %v", err)
	return status, activationData
}

func (s *StageManager) traceValue(ctx context.Context, v interface{}, inputs map[string]interface{}, outputs map[string]map[string]interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		parser := utils.NewParser(val)
		context := s.Context.VencorContext.EvaluationContext.Clone()
		context.Context = ctx
		context.DeploymentSpec = s.Context.VencorContext.EvaluationContext.DeploymentSpec
		context.Inputs = inputs
		context.Outputs = outputs
		if context.Inputs != nil {
			if v, ok := context.Inputs["context"]; ok {
				context.Value = v
			}
		}
		v, err := parser.Eval(*context)
		if err != nil {
			return "", err
		}
		switch vt := v.(type) {
		case string:
			return vt, nil
		default:
			return s.traceValue(ctx, v, inputs, outputs)
		}
	case []interface{}:
		ret := []interface{}{}
		for _, v := range val {
			tv, err := s.traceValue(ctx, v, inputs, outputs)
			if err != nil {
				return "", err
			}
			ret = append(ret, tv)
		}
		return ret, nil
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range val {
			tv, err := s.traceValue(ctx, v, inputs, outputs)
			if err != nil {
				return "", err
			}
			ret[k] = tv
		}
		return ret, nil
	default:
		return val, nil
	}
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
		if activation.Status != nil && activation.Status.StageHistory != nil && len(activation.Status.StageHistory) != 0 &&
			activation.Status.StageHistory[len(activation.Status.StageHistory)-1].Stage != "" &&
			activation.Status.StageHistory[len(activation.Status.StageHistory)-1].NextStage != stage {
			log.ErrorfCtx(ctx, " M (Stage): current stage is %s, expected next stage is %s, actual next stage is %s",
				activation.Status.StageHistory[len(activation.Status.StageHistory)-1].Stage,
				activation.Status.StageHistory[len(activation.Status.StageHistory)-1].NextStage,
				stage)
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
			TriggeringStage:      stage,
			Schedule:             stageSpec.Schedule,
			Namespace:            actData.Namespace,
		}, nil
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", stage), v1alpha2.BadRequest)
}
