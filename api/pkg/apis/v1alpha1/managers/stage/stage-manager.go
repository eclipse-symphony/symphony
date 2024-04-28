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
		case int:
			state := v1alpha2.State(sv)
			stateValue := reflect.ValueOf(state)
			if stateValue.Type() != reflect.TypeOf(v1alpha2.State(0)) {
				return fmt.Errorf("invalid state %d", sv)
			}
			t.Outputs["__status"] = state
		case string:
			vInt, err := strconv.ParseInt(sv, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid state %s", sv)
			}
			state := v1alpha2.State(vInt)
			stateValue := reflect.ValueOf(state)
			if stateValue.Type() != reflect.TypeOf(v1alpha2.State(0)) {
				return fmt.Errorf("invalid state %d", vInt)
			}
			t.Outputs["__status"] = state
		default:
			return fmt.Errorf("invalid state %v", v)
		}

		if t.Outputs["__status"] != v1alpha2.OK {
			if v, ok := t.Outputs["__error"]; ok {
				return v1alpha2.NewCOAError(nil, v.(string), t.Outputs["__status"].(v1alpha2.State))
			} else {
				return fmt.Errorf("stage returned unsuccessful status without an error")
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
	log.Debugf(" M (Stage): ResumeStage: %v\n", status)
	campaign, ok := status.Outputs["__campaign"].(string)
	if !ok {
		log.Errorf(" M (Stage): ResumeStage: campaign (%v) is not valid from output", status.Outputs["__campaign"])
		return nil, fmt.Errorf("ResumeStage: campaign is not valid")
	}
	activation, ok := status.Outputs["__activation"].(string)
	if !ok {
		log.Errorf(" M (Stage): ResumeStage: activation (%v) is not valid from output", status.Outputs["__activation"])
		return nil, fmt.Errorf("ResumeStage: activation is not valid")
	}
	activationGeneration, ok := status.Outputs["__activationGeneration"].(string)
	if !ok {
		log.Errorf(" M (Stage): ResumeStage: activationGeneration (%v) is not valid from output", status.Outputs["__activationGeneration"])
		return nil, fmt.Errorf("ResumeStage: activationGeneration is not valid")
	}
	site, ok := status.Outputs["__site"].(string)
	if !ok {
		log.Errorf(" M (Stage): ResumeStage: site (%v) is not valid from output", status.Outputs["__site"])
		return nil, fmt.Errorf("ResumeStage: site is not valid")
	}
	stage, ok := status.Outputs["__stage"].(string)
	if !ok {
		log.Errorf(" M (Stage): ResumeStage: stage (%v) is not valid from output", status.Outputs["__stage"])
		return nil, fmt.Errorf("ResumeStage: stage is not valid")
	}
	namespace, ok := status.Outputs["__namespace"].(string)
	if !ok {
		namespace = "default"
	}

	entry, err := s.StateProvider.Get(context.TODO(), states.GetRequest{
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
			err := s.StateProvider.Delete(context.TODO(), states.DeleteRequest{
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
					eCtx.Inputs = status.Inputs
					log.Debugf(" M (Stage): ResumeStage evaluation inputs: %v", eCtx.Inputs)
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
						Inputs:               status.Inputs,
						Provider:             cam.Stages[nextStage].Provider,
						Config:               cam.Stages[nextStage].Config,
						Outputs:              outputs,
						TriggeringStage:      stage,
						Schedule:             cam.Stages[nextStage].Schedule,
						Namespace:            namespace,
					}
					log.Debugf(" M (Stage): Activating next stage: %s\n", activationData.Stage)
					return activationData, nil
				} else {
					log.Debugf(" M (Stage): No next stage found\n")
					return nil, nil
				}
			}
			return nil, nil
		} else {
			p.Sites = newSites
			_, err := s.StateProvider.Upsert(context.TODO(), states.UpsertRequest{
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
		return nil, fmt.Errorf("invalid pending task")
	}

	return nil, nil
}
func (s *StageManager) HandleDirectTriggerEvent(ctx context.Context, triggerData v1alpha2.ActivationData) model.ActivationStatus {
	ctx, span := observability.StartSpan("Stage Manager", ctx, &map[string]string{
		"method": "HandleDirectTriggerEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	status := model.ActivationStatus{
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
		Status:       v1alpha2.Untouched,
		ErrorMessage: "",
		IsActive:     true,
	}
	var provider providers.IProvider
	factory := symproviders.SymphonyProviderFactory{}
	provider, err = factory.CreateProvider(triggerData.Provider, triggerData.Config)
	if err != nil {
		status.Status = v1alpha2.InternalError
		status.ErrorMessage = err.Error()
		status.IsActive = false
		return status
	}
	if provider == nil {
		status.Status = v1alpha2.BadRequest
		status.ErrorMessage = fmt.Sprintf("provider %s is not found", triggerData.Provider)
		status.IsActive = false
		return status
	}

	if _, ok := provider.(contexts.IWithManagerContext); ok {
		provider.(contexts.IWithManagerContext).SetContext(s.Manager.Context)
	} else {
		log.Errorf(" M (Stage): provider %s does not implement IWithManagerContext", triggerData.Provider)
	}

	isRemote := false
	if _, ok := provider.(*remote.RemoteStageProvider); ok {
		isRemote = true
		provider.(*remote.RemoteStageProvider).SetOutputsContext(triggerData.Outputs)
	}

	if triggerData.Schedule != nil && !isRemote {
		s.Context.Publish("schedule", v1alpha2.Event{
			Body: triggerData,
		})
		status.Outputs["__status"] = v1alpha2.Delayed
		status.Status = v1alpha2.Paused
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
func (s *StageManager) HandleTriggerEvent(ctx context.Context, campaign model.CampaignSpec, triggerData v1alpha2.ActivationData) (model.ActivationStatus, *v1alpha2.ActivationData) {
	ctx, span := observability.StartSpan("Stage Manager", ctx, &map[string]string{
		"method": "HandleTriggerEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Info(" M (Stage): HandleTriggerEvent")
	status := model.ActivationStatus{
		Stage:     triggerData.Stage,
		NextStage: "",
		Outputs: map[string]interface{}{
			"__campaign":             triggerData.Campaign,
			"__namespace":            triggerData.Namespace,
			"__activation":           triggerData.Activation,
			"__activationGeneration": triggerData.ActivationGeneration,
			"__stage":                triggerData.Stage,
			"__site":                 s.VendorContext.SiteInfo.SiteId,
		},
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
			log.Debugf(" M (Stage): HandleTriggerEvent evaluation inputs 1: %v", eCtx.Inputs)
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
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.Errorf(" M (Stage): failed to evaluate context: %v", err)
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

		log.Debugf(" M (Stage): HandleTriggerEvent before evaluation inputs 2: %v", inputs)

		// inject default inputs
		inputs["__campaign"] = triggerData.Campaign
		inputs["__namespace"] = triggerData.Namespace
		inputs["__activation"] = triggerData.Activation
		inputs["__stage"] = triggerData.Stage
		inputs["__activationGeneration"] = triggerData.ActivationGeneration
		inputs["__previousStage"] = triggerData.TriggeringStage
		inputs["__site"] = s.VendorContext.SiteInfo.SiteId
		if triggerData.Schedule != nil {
			jSchedule, _ := json.Marshal(triggerData.Schedule)
			inputs["__schedule"] = string(jSchedule)
		}
		for k, v := range inputs {
			var val interface{}
			val, err = s.traceValue(v, inputs, triggerData.Outputs)
			if err != nil {
				status.Status = v1alpha2.InternalError
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.Errorf(" M (Stage): failed to evaluate input: %v", err)
				return status, activationData
			}
			inputs[k] = val
		}

		if triggerData.Outputs != nil {
			if v, ok := triggerData.Outputs[triggerData.Stage]; ok {
				if vs, ok := v["__state"]; ok {
					inputs["__state"] = vs
				}
			}
		}

		log.Debugf(" M (Stage): HandleTriggerEvent after evaluation inputs 2: %v", inputs)

		factory := symproviders.SymphonyProviderFactory{}
		var provider providers.IProvider
		provider, err = factory.CreateProvider(triggerData.Provider, triggerData.Config)
		if err != nil {
			status.Status = v1alpha2.InternalError
			status.ErrorMessage = err.Error()
			status.IsActive = false
			log.Errorf(" M (Stage): failed to create provider: %v", err)
			return status, activationData
		}
		if provider == nil {
			status.Status = v1alpha2.BadRequest
			status.ErrorMessage = fmt.Sprintf("provider %s is not found", triggerData.Provider)
			status.IsActive = false
			log.Errorf(" M (Stage): failed to create provider: %v", err)
			return status, activationData
		}

		if _, ok := provider.(contexts.IWithManagerContext); ok {
			provider.(contexts.IWithManagerContext).SetContext(s.Manager.Context)
		} else {
			log.Errorf(" M (Stage): provider %s does not implement IWithManagerContext", triggerData.Provider)
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
					val, err = s.traceValue(v, inputCopy, triggerData.Outputs)
					if err != nil {
						status.Status = v1alpha2.InternalError
						status.ErrorMessage = err.Error()
						status.IsActive = false
						log.Errorf(" M (Stage): failed to evaluate input: %v", err)
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

				if triggerData.Schedule != nil {
					s.Context.Publish("schedule", v1alpha2.Event{
						Body: triggerData,
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
				status.ErrorMessage = fmt.Sprintf("%s: %s", result.Site, err.Error())
				status.IsActive = false
				site := result.Site
				if result.Site == s.Context.SiteInfo.SiteId {
					site = ""
				}
				status.Outputs = carryOutPutsToErrorStatus(nil, err, site)
				result.Outputs = carryOutPutsToErrorStatus(nil, err, site)
				log.Errorf(" M (Stage): failed to process stage outputs: %v", err)
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
					outputs[fmt.Sprintf("%s.__status", result.Site)] = v1alpha2.OK
				}
			}
		}

		for k, v := range outputs {
			status.Outputs[k] = v
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
					status.ErrorMessage = err.Error()
					status.IsActive = false
					log.Errorf(" M (Stage): failed to save pending task: %v", err)
					return status, activationData
				}
				status.Status = v1alpha2.Paused
				status.IsActive = false
				return status, activationData
			}

			parser := utils.NewParser(currentStage.StageSelector)
			eCtx := s.VendorContext.EvaluationContext.Clone()
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
				status.ErrorMessage = err.Error()
				status.IsActive = false
				log.Errorf(" M (Stage): failed to evaluate stage selector: %v", err)
				return status, activationData
			}
			sVal := ""
			if val != nil {
				sVal = val.(string)
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
						status.ErrorMessage = fmt.Sprintf("stage %s failed", triggerData.Stage)
						status.IsActive = false
						log.Errorf(" M (Stage): failed to process stage outputs: %v", status.ErrorMessage)
						return status, activationData
					}
				} else {
					err = v1alpha2.NewCOAError(nil, status.ErrorMessage, v1alpha2.BadRequest)
					status.Status = v1alpha2.BadRequest
					status.ErrorMessage = fmt.Sprintf("stage %s is not found", sVal)
					status.IsActive = false
					log.Errorf(" M (Stage): failed to find next stage: %v", err)
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
			return status, activationData
		} else {
			status.Status = v1alpha2.Done
			status.NextStage = ""
			status.IsActive = false
			log.Infof(" M (Stage): stage %s is done (no next stage)", triggerData.Stage)
			return status, activationData
		}
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", triggerData.Stage), v1alpha2.BadRequest)
	status.Status = v1alpha2.InternalError
	status.ErrorMessage = err.Error()
	status.IsActive = false
	log.Errorf(" M (Stage): failed to find stage: %v", err)
	return status, activationData
}

func (s *StageManager) traceValue(v interface{}, inputs map[string]interface{}, outputs map[string]map[string]interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		parser := utils.NewParser(val)
		context := s.Context.VencorContext.EvaluationContext.Clone()
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
			return s.traceValue(v, inputs, outputs)
		}
	case []interface{}:
		ret := []interface{}{}
		for _, v := range val {
			tv, err := s.traceValue(v, inputs, outputs)
			if err != nil {
				return "", err
			}
			ret = append(ret, tv)
		}
		return ret, nil
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range val {
			tv, err := s.traceValue(v, inputs, outputs)
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
			TriggeringStage:      stage,
			Schedule:             stageSpec.Schedule,
			Namespace:            actData.Namespace,
		}, nil
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("stage %s is not found", stage), v1alpha2.BadRequest)
}
