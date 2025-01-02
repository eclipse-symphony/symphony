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
	"strconv"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/sites"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/staging"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/sync"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/trails"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

var fLog = logger.NewLogger("coa.runtime")

type FederationVendor struct {
	vendors.Vendor
	SitesManager    *sites.SitesManager
	CatalogsManager *catalogs.CatalogsManager
	StagingManager  *staging.StagingManager
	SyncManager     *sync.SyncManager
	TrailsManager   *trails.TrailsManager
	SolutionManager *solution.SolutionManager
	apiClient       utils.ApiClient
}

func (f *FederationVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  f.Vendor.Version,
		Name:     "Federation",
		Producer: "Microsoft",
	}
}
func (f *FederationVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := f.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}

	for _, m := range f.Managers {
		if c, ok := m.(*sites.SitesManager); ok {
			f.SitesManager = c
		}
		if c, ok := m.(*staging.StagingManager); ok {
			f.StagingManager = c
		}
		if c, ok := m.(*catalogs.CatalogsManager); ok {
			f.CatalogsManager = c
		}
		if c, ok := m.(*sync.SyncManager); ok {
			f.SyncManager = c
		}
		if c, ok := m.(*trails.TrailsManager); ok {
			f.TrailsManager = c
		}
		if c, ok := m.(*solution.SolutionManager); ok {
			f.SolutionManager = c
		}
	}
	if f.StagingManager == nil {
		return v1alpha2.NewCOAError(nil, "staging manager is not supplied", v1alpha2.MissingConfig)
	}
	if f.SitesManager == nil {
		return v1alpha2.NewCOAError(nil, "sites manager is not supplied", v1alpha2.MissingConfig)
	}
	if f.CatalogsManager == nil {
		return v1alpha2.NewCOAError(nil, "catalogs manager is not supplied", v1alpha2.MissingConfig)
	}
	if f.SolutionManager == nil {
		return v1alpha2.NewCOAError(nil, "solution manager is not supplied", v1alpha2.MissingConfig)
	}
	f.apiClient, err = utils.GetParentApiClient(f.Vendor.Context.SiteInfo.ParentSite.BaseUrl)
	if err != nil {
		return err
	}
	f.Vendor.Context.Subscribe("catalog", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			sites, err := f.SitesManager.ListState(context.TODO())
			if err != nil {
				return err
			}
			for _, site := range sites {
				if site.Spec.Name != f.Vendor.Context.SiteInfo.SiteId {
					event.Metadata["site"] = site.Spec.Name
					ctx := context.TODO()
					if event.Context != nil {
						ctx = event.Context
					}
					f.StagingManager.HandleJobEvent(ctx, event) //TODO: how to handle errors in this case?
				}
			}
			return nil
		},
	})
	f.Vendor.Context.Subscribe("remote", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			_, ok := event.Metadata["site"]
			if !ok {
				return v1alpha2.NewCOAError(nil, "site is not supplied", v1alpha2.BadRequest)
			}
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			f.StagingManager.HandleJobEvent(ctx, event) //TODO: how to handle errors in this case?
			return nil
		},
	})
	f.Vendor.Context.Subscribe("report", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			fLog.DebugfCtx(ctx, "V (Federation): received report event: %v", event)
			jData, _ := json.Marshal(event.Body)
			var status model.StageStatus
			err := json.Unmarshal(jData, &status)
			if err == nil {
				ctx := context.TODO()
				if event.Context != nil {
					ctx = event.Context
				}
				err := f.apiClient.SyncStageStatus(ctx, status,
					f.Vendor.Context.SiteInfo.ParentSite.Username,
					f.Vendor.Context.SiteInfo.ParentSite.Password)
				if err != nil {
					fLog.ErrorfCtx(ctx, "V (Federation): error while syncing activation status: %v", err)
					return err
				}
			}
			return v1alpha2.NewCOAError(nil, "report is not an activation status", v1alpha2.BadRequest)
		},
	})
	f.Vendor.Context.Subscribe("trail", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			if f.TrailsManager != nil {
				jData, _ := json.Marshal(event.Body)
				var trails []v1alpha2.Trail
				err := json.Unmarshal(jData, &trails)
				if err == nil {
					return f.TrailsManager.Append(ctx, trails)
				}
			}
			return nil
		},
	},
	)
	f.Vendor.Context.Subscribe("deployment-step", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			log.InfoCtx(ctx, "V(Federation): subscribe deployment-step and begin to apply step ")
			// get data
			var stepEnvelope StepEnvelope
			jData, _ := json.Marshal(event.Body)
			if err := json.Unmarshal(jData, &stepEnvelope); err != nil {
				log.ErrorfCtx(ctx, "V (Federation): failed to unmarshal step envelope: %v", err)
				return err
			}
			switch stepEnvelope.Phase {
			case PhaseGet:
				if FindAgentFromDeploymentState(stepEnvelope.Step.Components, stepEnvelope.Step.Target) {
					// if true {
					operationId := uuid.New().String()
					providerGetRequest := &ProviderGetRequest{
						AgentRequest: AgentRequest{
							OperationID: operationId,
							Provider:    stepEnvelope.Step.Role,
							Action:      string(PhaseGet),
						},
						References: stepEnvelope.Step.Components,
					}
					err = f.upsertOperationState(ctx, operationId, stepEnvelope.StepId, stepEnvelope.PlanId, stepEnvelope.Step.Target, stepEnvelope.Phase, stepEnvelope.Namespace, stepEnvelope.Remove)
					if err != nil {
						log.ErrorCtx(ctx, "error in insert operation Id %s", operationId)
					}
					f.StagingManager.QueueProvider.Enqueue(fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.Namespace), providerGetRequest)
					log.InfoCtx(ctx, "V(Federation): enqueue get %s-%s %+v ", stepEnvelope.Step.Target, stepEnvelope.Namespace, providerGetRequest)
				} else {
					// get provider todo : is dry run
					provider, err := f.SolutionManager.GetTargetProviderForStep(stepEnvelope.Step.Target, stepEnvelope.Step.Role, stepEnvelope.Deployment, stepEnvelope.PlanState.PreviousDesiredState)
					if err != nil {
						log.ErrorfCtx(ctx, " M (Solution): failed to create provider & Failed to save summary progress: %v", err)
						return f.publishStepResult(ctx, stepEnvelope, false, err, make(map[string]model.ComponentResultSpec))
					}
					log.InfoCtx(ctx, "get step components %+v", stepEnvelope.Step.Components)
					log.InfoCtx(ctx, "get step components %+v", stepEnvelope.Deployment)
					dep := stepEnvelope.Deployment
					dep.ActiveTarget = stepEnvelope.Step.Target
					components, stepError := (provider.(tgt.ITargetProvider)).Get(ctx, dep, stepEnvelope.Step.Components)
					success := true
					if stepError != nil {
						success = false
					}
					log.InfoCtx(ctx, "get component %+v remove %s", components, stepEnvelope.Remove)
					stepResult := &StepResult{
						Target:         stepEnvelope.Step.Target,
						PlanId:         stepEnvelope.PlanId,
						StepId:         stepEnvelope.StepId,
						Success:        success,
						Remove:         stepEnvelope.Remove,
						Phase:          PhaseGet,
						retComoponents: components,
						Error:          stepError,
					}
					jData, _ := json.Marshal(stepResult)
					log.InfofCtx(ctx, "Publishing step-result: %s", string(jData))
					f.Vendor.Context.Publish("step-result", v1alpha2.Event{
						Metadata: map[string]string{
							"namespace": stepEnvelope.Namespace,
						},
						Body:    jData,
						Context: ctx,
					})
				}
			case PhaseApply:
				if FindAgentFromDeploymentState(stepEnvelope.Step.Components, stepEnvelope.Step.Target) {
					operationId := uuid.New().String()
					providApplyRequest := &ProviderApplyRequest{
						AgentRequest: AgentRequest{
							OperationID: operationId,
							Provider:    stepEnvelope.Step.Role,
							Action:      string(PhaseApply),
						},
						Deployment: stepEnvelope.Deployment,
						Step:       stepEnvelope.Step,
						IsDryRun:   stepEnvelope.Deployment.IsDryRun,
					}
					log.InfoCtx(ctx, "V(Federation): enqueue %s-%s %+v ", stepEnvelope.Step.Target, stepEnvelope.Namespace, providApplyRequest)
					f.StagingManager.QueueProvider.Enqueue(fmt.Sprintf("%s-%s", stepEnvelope.Step.Target, stepEnvelope.Namespace), providApplyRequest)
					err = f.upsertOperationState(ctx, operationId, stepEnvelope.StepId, stepEnvelope.PlanId, stepEnvelope.Step.Target, stepEnvelope.Phase, stepEnvelope.Namespace, stepEnvelope.Remove)
					if err != nil {
						log.ErrorCtx(ctx, "error in insert operation Id %s", operationId)
					}
				} else {
					// get provider todo : is dry run
					provider, err := f.SolutionManager.GetTargetProviderForStep(stepEnvelope.Step.Target, stepEnvelope.Step.Role, stepEnvelope.Deployment, stepEnvelope.PlanState.PreviousDesiredState)
					if err != nil {
						log.ErrorfCtx(ctx, " M (Solution): failed to create provider & Failed to save summary progress: %v", err)
						return f.publishStepResult(ctx, stepEnvelope, false, err, make(map[string]model.ComponentResultSpec))
					}
					previousDesiredState := stepEnvelope.PlanState.PreviousDesiredState
					currentState := stepEnvelope.PlanState.CurrentState
					step := stepEnvelope.Step
					if previousDesiredState != nil {
						testState := solution.MergeDeploymentStates(&previousDesiredState.State, currentState)
						if f.SolutionManager.CanSkipStep(ctx, step, step.Target, provider.(tgt.ITargetProvider), previousDesiredState.State.Components, testState) {
							log.InfofCtx(ctx, " M (Solution): skipping step with role %s on target %s", step.Role, step.Target)
							f.Vendor.Context.Publish("step-result", v1alpha2.Event{
								Metadata: map[string]string{
									"namespace": stepEnvelope.Namespace,
								},
								Body: StepResult{
									Target:     stepEnvelope.Step.Target,
									PlanId:     stepEnvelope.PlanId,
									StepId:     stepEnvelope.StepId,
									Success:    true,
									Remove:     stepEnvelope.Remove,
									Phase:      PhaseApply,
									Components: map[string]model.ComponentResultSpec{},
									Timestamp:  time.Now(),
								},
							})
							return nil
						}
					}
					componentResults, stepError := (provider.(tgt.ITargetProvider)).Apply(ctx, stepEnvelope.Deployment, stepEnvelope.Step, stepEnvelope.Deployment.IsDryRun)
					if stepError != nil {
						return f.publishStepResult(ctx, stepEnvelope, false, stepError, componentResults)
					}
					stepResult := &StepResult{
						Target:     stepEnvelope.Step.Target,
						PlanId:     stepEnvelope.PlanId,
						StepId:     stepEnvelope.StepId,
						Success:    true,
						Remove:     stepEnvelope.Remove,
						Phase:      PhaseApply,
						Components: componentResults,
						Timestamp:  time.Now(),
					}
					f.publishStepResultSucceed(ctx, stepResult, stepEnvelope.Namespace)
				}
			}
			return nil
		},
		Group: "federation-vendor",
	})
	//now register the current site
	return f.SitesManager.UpsertSpec(context.Background(), f.Context.SiteInfo.SiteId, model.SiteSpec{
		Name:       f.Context.SiteInfo.SiteId,
		Properties: f.Context.SiteInfo.Properties,
		IsSelf:     true,
	})
}
func (f *FederationVendor) publishStepResultSucceed(ctx context.Context, stepResult *StepResult, namespace string) error {
	jData, _ := json.Marshal(stepResult)
	log.InfofCtx(ctx, "Publishing step-result: %s", string(jData))
	f.Vendor.Context.Publish("step-result", v1alpha2.Event{
		Metadata: map[string]string{
			"namespace": namespace,
		},
		Body: jData,
	})
	return nil
}
func FindAgentFromDeploymentState(stepComponents []model.ComponentStep, targetName string) bool {
	log.Info("compare between state and target name %+v, %s", stepComponents, targetName)
	for _, component := range stepComponents {
		log.Info("compare between state and target name %+v, %s", component, component.Component.Name)
		if component.Component.Name == targetName {
			if component.Component.Type == "remote-agent" {
				log.Info("It is remote call ")
				return true
			}

		}
	}
	return false
}
func (f *FederationVendor) publishStepResult(ctx context.Context, stepEnvelope StepEnvelope, success bool, Error error, components map[string]model.ComponentResultSpec) error {
	return f.Vendor.Context.Publish("step-result", v1alpha2.Event{
		Metadata: map[string]string{
			"namespace": stepEnvelope.Namespace,
		},
		Body: StepResult{
			Target:     stepEnvelope.Step.Target,
			PlanId:     stepEnvelope.PlanId,
			StepId:     stepEnvelope.StepId,
			Success:    success,
			Components: components,
			Timestamp:  time.Now(),
			Remove:     stepEnvelope.Remove,
			Error:      Error,
		},
	})
}

func (f *FederationVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "federation"
	if f.Route != "" {
		route = f.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodPost, fasthttp.MethodGet},
			Route:      route + "/sync",
			Version:    f.Version,
			Handler:    f.onSync,
			Parameters: []string{"site?"},
		},
		{
			Methods:    []string{fasthttp.MethodPost, fasthttp.MethodGet},
			Route:      route + "/registry",
			Version:    f.Version,
			Handler:    f.onRegistry,
			Parameters: []string{"name?"},
		},
		{
			Methods:    []string{fasthttp.MethodPost},
			Route:      route + "/status",
			Version:    f.Version,
			Handler:    f.onStatus,
			Parameters: []string{"name"},
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/trail",
			Version: f.Version,
			Handler: f.onTrail,
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/k8shook",
			Version: f.Version,
			Handler: f.onK8sHook,
		}, {
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/tasks",
			Version: f.Version,
			Handler: f.onGetRequest,
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/task/getResult",
			Version: f.Version,
			Handler: f.onGetResponse,
		},
	}
}
func (f *FederationVendor) onGetRequest(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onGetRequest",
	})
	defer span.End()
	var agentRequest AgentRequest
	target := request.Parameters["target"]
	namespace := request.Parameters["namespace"]
	sLog.InfoCtx(ctx, "V(Federation): get request from remote agent %+v", agentRequest)
	return f.getTaskFromQueue(ctx, target, namespace)
}

func (f *FederationVendor) onGetResponse(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onGetResponse",
	})
	defer span.End()

	var asyncResult AsyncResult
	err := json.Unmarshal(request.Body, &asyncResult)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (FederationVendor): onGetResponse failed - %s", err.Error())
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	sLog.InfoCtx(ctx, "V(Federation): get async result from remote agent %+v", asyncResult)
	return f.handleRemoteAgentExecuteResult(ctx, asyncResult)
}

func (f *FederationVendor) handleRemoteAgentExecuteResult(ctx context.Context, asyncResult AsyncResult) v1alpha2.COAResponse {
	// get opertaion Id
	operationId := asyncResult.OperationID
	// get related info from redis- todo: timeout
	log.InfoCtx(ctx, "handle remote agent request %+v ", asyncResult)
	operationBody, err := f.getOperationState(ctx, operationId)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (FederationVendor): onGetResponse failed - %s", err.Error())
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}

	switch operationBody.Action {
	case PhaseGet:
		// send to stp result
		var response []model.ComponentSpec
		err := json.Unmarshal(asyncResult.Body, &response)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}

		log.InfoCtx(ctx, "get response %+v", response)
		if asyncResult.Error != nil {
			log.InfoCtx(ctx, "publish step result  with error ")
			f.Vendor.Context.Publish("step-result", v1alpha2.Event{
				Metadata: map[string]string{
					"namespace": operationBody.NameSpace,
				},
				Body: StepResult{
					Target:         operationBody.Target,
					PlanId:         operationBody.PlanId,
					StepId:         operationBody.StepId,
					Success:        false,
					retComoponents: response,
					Phase:          PhaseGet,
					Remove:         operationBody.Remove,
					Timestamp:      time.Now(),
					Error:          asyncResult.Error,
				},
			})
		} else {
			log.InfoCtx(ctx, "publish step result  without error ")
			f.Vendor.Context.Publish("step-result", v1alpha2.Event{
				Metadata: map[string]string{
					"namespace": operationBody.NameSpace,
				},
				Body: StepResult{
					Target:         operationBody.Target,
					PlanId:         operationBody.PlanId,
					StepId:         operationBody.StepId,
					Phase:          PhaseGet,
					Success:        true,
					Remove:         operationBody.Remove,
					retComoponents: response,
					Timestamp:      time.Now(),
				},
			})
		}
		deleteRequest := states.DeleteRequest{
			ID: operationId,
		}

		err = f.StagingManager.StateProvider.Delete(ctx, deleteRequest)
		if err != nil {
			return v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"delete operation Id failed\"}"),
				ContentType: "application/json",
			}
		}
		return v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte("{\"result\":\"200 - handle async result successfully\"}"),
			ContentType: "application/json",
		}
	case PhaseApply:
		var response map[string]model.ComponentResultSpec
		err := json.Unmarshal(asyncResult.Body, &response)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		if asyncResult.Error != nil {
			f.Vendor.Context.Publish("step-result", v1alpha2.Event{
				Metadata: map[string]string{
					"namespace": operationBody.NameSpace,
				},
				Body: StepResult{
					Target:     operationBody.Target,
					PlanId:     operationBody.PlanId,
					StepId:     operationBody.StepId,
					Success:    true,
					Phase:      PhaseApply,
					Remove:     operationBody.Remove,
					Components: response,
					Timestamp:  time.Now(),
				},
			})
		} else {
			f.Vendor.Context.Publish("step-result", v1alpha2.Event{
				Metadata: map[string]string{
					"namespace": operationBody.NameSpace,
				},
				Body: StepResult{
					Target:     operationBody.Target,
					PlanId:     operationBody.PlanId,
					StepId:     operationBody.StepId,
					Success:    true,
					Phase:      PhaseApply,
					Remove:     operationBody.Remove,
					Components: response,
					Timestamp:  time.Now(),
				},
			})
		}
		deleteRequest := states.DeleteRequest{
			ID: operationId,
		}

		err = f.StagingManager.StateProvider.Delete(ctx, deleteRequest)
		if err != nil {
			return v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"delete operation Id failed\"}"),
				ContentType: "application/json",
			}
		}
		return v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte("{\"result\":\"200 - get response successfully\"}"),
			ContentType: "application/json",
		}
	}
	return v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
}

func (f *FederationVendor) getTaskFromQueue(ctx context.Context, target string, namespace string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doGetFromQueue",
	})
	queueName := fmt.Sprintf("%s-%s", target, namespace)
	sLog.InfoCtx(ctx, "V (FederationVendor): getFromqueue %s queue length %s", queueName)
	defer span.End()

	queueElement, err := f.StagingManager.QueueProvider.Dequeue(queueName)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (FederationVendor): getqueue failed - %s", err.Error())
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	data, _ := json.Marshal(queueElement)
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}

// for operation state storage
func (f *FederationVendor) upsertOperationState(ctx context.Context, operationId string, stepId int, planId string, target string, action JobPhase, namespace string, remove bool) error {
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: operationId,
			Body: map[string]interface{}{
				"StepId":    stepId,
				"PlanId":    planId,
				"Target":    target,
				"Action":    action,
				"namespace": namespace,
				"Remove":    remove,
			}},
	}
	_, err := f.StagingManager.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

// for get operation state
func (f *FederationVendor) getOperationState(ctx context.Context, operationId string) (OperationBody, error) {
	getRequest := states.GetRequest{
		ID: operationId,
	}
	var entry states.StateEntry
	entry, err := f.StagingManager.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return OperationBody{}, err
	}
	var ret OperationBody
	ret, err = f.getOperationBody(entry.Body)
	if err != nil {
		log.ErrorfCtx(ctx, "Failed to convert to operation state for %s ", operationId)
		return OperationBody{}, err
	}
	return ret, err
}
func (f *FederationVendor) getOperationBody(body interface{}) (OperationBody, error) {
	var operationBody OperationBody
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &operationBody)
	if err != nil {
		return OperationBody{}, err
	}
	return operationBody, nil
}

func (c *FederationVendor) onStatus(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onStatus",
	})
	defer span.End()

	tLog.InfoCtx(pCtx, "V (Federation): OnStatus")
	switch request.Method {
	case fasthttp.MethodPost:
		var state model.SiteState
		json.Unmarshal(request.Body, &state)

		err := c.SitesManager.ReportState(pCtx, state)

		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (f *FederationVendor) onRegistry(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onRegistry",
	})
	defer span.End()

	tLog.Info("V (Federation): onRegistry")
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onRegistry-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			state, err = f.SitesManager.ListState(ctx)
			isArray = true
		} else {
			state, err = f.SitesManager.GetState(ctx, id)
		}
		if err != nil {
			if utils.IsNotFound(err) {
				errorMsg := fmt.Sprintf("site '%s' is not found", id)
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.NotFound,
					Body:  []byte(errorMsg),
				})
			} else {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}
		jData, _ := utils.FormatObject(state, isArray, request.Parameters["path"], request.Parameters["doc-type"])
		resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        jData,
			ContentType: "application/json",
		})
		if request.Parameters["doc-type"] == "yaml" {
			resp.ContentType = "text/plain"
		}
		return resp
	case fasthttp.MethodPost:
		// TODO: POST federation/registry need to pass SiteState as request body
		ctx, span := observability.StartSpan("onRegistry-POST", pCtx, nil)
		id := request.Parameters["__name"]

		var site model.SiteSpec
		err := json.Unmarshal(request.Body, &site)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		//TODO: generate site key pair as needed
		err = f.SitesManager.UpsertSpec(ctx, id, site)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onRegistry-DELETE", pCtx, nil)
		id := request.Parameters["__name"]
		err := f.SitesManager.DeleteSpec(ctx, id)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
func (f *FederationVendor) onSync(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onSync",
	})
	defer span.End()

	tLog.Info("V (Federation): onSync")
	switch request.Method {
	case fasthttp.MethodPost:
		var status model.StageStatus
		err := json.Unmarshal(request.Body, &status)
		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Federation): failed to unmarshal stage status: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}
		err = f.Vendor.Context.Publish("job-report", v1alpha2.Event{
			Body:    status,
			Context: pCtx,
		})
		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Federation): failed to publish job report: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		tLog.Debugf("V (Federation): published job report: %v", status)
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onSync-GET", pCtx, nil)
		id := request.Parameters["__site"]
		count := request.Parameters["count"]
		namespace, exist := request.Parameters["namespace"]
		if !exist {
			namespace = "default"
		}
		if count == "" {
			count = "1"
		}
		intCount, err := strconv.Atoi(count)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}
		batch, err := f.StagingManager.GetABatchForSite(id, intCount)

		pack := model.SyncPackage{
			Origin: f.Context.SiteInfo.SiteId,
		}

		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		catalogs := make([]model.CatalogState, 0)
		jobs := make([]v1alpha2.JobData, 0)
		for _, c := range batch {
			if c.Action == v1alpha2.JobRun { //TODO: I don't really like this
				jobs = append(jobs, c)
			} else {
				catalog, err := f.CatalogsManager.GetState(ctx, c.Id, namespace)
				if err != nil {
					return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
						State: v1alpha2.InternalError,
						Body:  []byte(err.Error()),
					})
				}
				catalogs = append(catalogs, catalog)
			}
		}
		pack.Catalogs = catalogs
		pack.Jobs = jobs
		jData, _ := utils.FormatObject(pack, true, request.Parameters["path"], request.Parameters["doc-type"])
		resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        jData,
			ContentType: "application/json",
		})
		if request.Parameters["doc-type"] == "yaml" {
			resp.ContentType = "text/plain"
		}
		return resp
	}
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
func (f *FederationVendor) onTrail(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onTrail",
	})
	defer span.End()

	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	return resp
}
func (f *FederationVendor) onK8sHook(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onK8sHook",
	})
	defer span.End()

	tLog.Info("V (Federation): onK8sHook")
	switch request.Method {
	case fasthttp.MethodPost:
		objectType := request.Parameters["objectType"]
		if objectType == "catalog" {
			var catalog model.CatalogState
			err := json.Unmarshal(request.Body, &catalog)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.BadRequest,
					Body:  []byte(err.Error()),
				})
			}
			err = f.Vendor.Context.Publish("catalog", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": catalog.Spec.CatalogType,
				},
				Body: v1alpha2.JobData{
					Id:     catalog.ObjectMeta.Name,
					Action: v1alpha2.JobUpdate, //TODO: handle deletion, this probably requires BetBachForSites return flags
					Body:   catalog,
				},
				Context: ctx,
			})
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.OK,
			})
		}
	}

	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
