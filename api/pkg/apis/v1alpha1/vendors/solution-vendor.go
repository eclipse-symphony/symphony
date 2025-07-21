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

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/valyala/fasthttp"
)

type SolutionVendor struct {
	vendors.Vendor
	SolutionManager *solution.SolutionManager
}

func (o *SolutionVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Solution",
		Producer: "Microsoft",
	}
}

func (e *SolutionVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*solution.SolutionManager); ok {
			e.SolutionManager = c
		}
	}
	if e.SolutionManager == nil {
		return v1alpha2.NewCOAError(nil, "solution manager is not supplied", v1alpha2.MissingConfig)
	}
	e.Vendor.Context.Subscribe(model.DeploymentStepTopic, v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := event.Context
			if ctx == nil {
				ctx = context.TODO()
			}
			log.InfoCtx(ctx, "V(Solution): subscribe deployment-step and begin to apply step ")
			// get data
			err := e.SolutionManager.HandleDeploymentStep(ctx, event)
			if err != nil {
				log.ErrorfCtx(ctx, "V(Solution): Failed to handle deployment-step: %v", err)
			}
			return err
		},
		Group: "Solution-vendor",
	})
	e.Vendor.Context.Subscribe(model.DeploymentPlanTopic, v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := event.Context
			if ctx == nil {
				ctx = context.TODO()
			}
			log.InfoCtx(ctx, "V(Solution): Begin to execute deployment-plan")
			err := e.SolutionManager.HandleDeploymentPlan(ctx, event)
			if err != nil {
				log.ErrorfCtx(ctx, "V(Solution): Failed to handle deployment plan: %v", err)
			}
			return err
		},
		Group: "stage-vendor",
	})
	e.Vendor.Context.Subscribe(model.CollectStepResultTopic, v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := event.Context
			if ctx == nil {
				ctx = context.TODO()
			}
			err := e.SolutionManager.HandleStepResult(ctx, event)
			if err != nil {
				log.ErrorfCtx(ctx, "V(Solution): Failed to handle step result: %v", err)
				return err
			}
			return err
		},
		Group: "stage-vendor",
	})
	return nil
}

func (o *SolutionVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "solution"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost, fasthttp.MethodGet, fasthttp.MethodDelete},
			Route:   route + "/instances", //this route is to support ITargetProvider interface via a proxy provider
			Version: o.Version,
			Handler: o.onApplyDeployment,
		},
		{
			Methods:    []string{fasthttp.MethodPost},
			Route:      route + "/reconcile",
			Version:    o.Version,
			Parameters: []string{"delete?"},
			Handler:    o.onReconcile,
		},
		{
			Methods: []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:   route + "/queue",
			Version: o.Version,
			Handler: o.onQueue,
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/tasks",
			Version: o.Version,
			Handler: o.onGetRequest,
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/task/getResult",
			Version: o.Version,
			Handler: o.onGetResponse,
		},
	}
}

func (c *SolutionVendor) onQueue(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onQueue",
	})
	defer span.End()
	instance := request.Parameters["instance"]
	sLog.InfofCtx(rContext, "V (Solution): onQueue, method: %s, %s", request.Method, instance)

	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onQueue-GET", rContext, nil)
		defer span.End()
		instance := request.Parameters["instance"]
		instanceName := request.Parameters["name"]

		if instance == "" {
			sLog.ErrorCtx(ctx, "V (Solution): onQueue failed - 400 instance parameter is not found")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - instance parameter is not found\"}"),
				ContentType: "application/json",
			})
		}
		summary, err := c.SolutionManager.GetSummary(ctx, instance, instanceName, namespace)
		data, _ := json.Marshal(summary)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onQueue failed - %s", err.Error())
			if api_utils.IsNotFound(err) {
				errorMsg := fmt.Sprintf("instance '%s' is not found in namespace %s", instance, namespace)
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.NotFound,
					Body:  []byte(errorMsg),
				})
			} else {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
					Body:  data,
				})
			}
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        data,
			ContentType: "application/json",
		})
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onQueue-POST", rContext, nil)
		defer span.End()

		// DO NOT REMOVE THIS COMMENT
		// gofail: var onQueueError string

		instance := request.Parameters["instance"]
		delete := request.Parameters["delete"]
		objectType := request.Parameters["objectType"]
		target := request.Parameters["target"]

		if objectType == "" { // For backward compatibility
			objectType = "instance"
		}

		if target == "true" {
			objectType = "target"
		}

		if objectType == "deployment" {
			deployment, err := model.ToDeployment(request.Body)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State:       v1alpha2.DeserializeError,
					ContentType: "application/json",
					Body:        []byte(fmt.Sprintf(`{"result":"%s"}`, err.Error())),
				})
			}
			instance = deployment.Instance.ObjectMeta.Name
		}

		if instance == "" {
			sLog.ErrorCtx(ctx, "V (Solution): onQueue failed - 400 instance parameter is not found")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - instance parameter is not found\"}"),
				ContentType: "application/json",
			})
		}
		action := v1alpha2.JobUpdate
		if delete == "true" {
			action = v1alpha2.JobDelete
		}
		c.Vendor.Context.Publish("job", v1alpha2.Event{
			Metadata: map[string]string{
				"objectType": objectType,
				"namespace":  namespace,
			},
			Body: v1alpha2.JobData{
				Id:     instance,
				Scope:  namespace,
				Action: action,
				Data:   request.Body,
			},
			Context: ctx,
		})
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte("{\"result\":\"200 - instance reconcilation job accepted\"}"),
			ContentType: "application/json",
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onQueue-DELETE", rContext, nil)
		defer span.End()
		instance := request.Parameters["instance"]

		if instance == "" {
			sLog.ErrorCtx(ctx, "V (Solution): onQueue failed - 400 instance parameter is not found")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - instance parameter is not found\"}"),
				ContentType: "application/json",
			})
		}

		err := c.SolutionManager.DeleteSummary(ctx, instance, namespace)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onQueue DeleteSummary failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			ContentType: "application/json",
		})
	}
	sLog.ErrorCtx(rContext, "V (Solution): onQueue failed - 405 method not allowed")
	return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	})
}

func (c *SolutionVendor) onReconcile(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onReconcile",
	})
	defer span.End()

	sLog.InfofCtx(rContext, "V (Solution): onReconcile, method: %s", request.Method)
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}
	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onReconcile-POST", rContext, nil)
		defer span.End()
		var deployment model.DeploymentSpec
		err := utils2.UnmarshalJson(request.Body, &deployment)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onReconcile failed POST - unmarshal request %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		delete := request.Parameters["delete"]
		remove := delete == "true"
		targetName := ""
		if request.Metadata != nil {
			if v, ok := request.Metadata["active-target"]; ok {
				targetName = v
			}
		}
		summary, err := c.SolutionManager.AsyncReconcile(ctx, deployment, remove, namespace, targetName)
		data, _ := json.Marshal(summary)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onReconcile failed POST - reconcile %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  data,
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        data,
			ContentType: "application/json",
		})
	}
	sLog.ErrorCtx(rContext, "V (Solution): onReconcile failed - 405 method not allowed")
	return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	})
}

func (c *SolutionVendor) onApplyDeployment(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onApplyDeployment",
	})
	defer span.End()

	sLog.InfofCtx(rContext, "V (Solution): onApplyDeployment %s", request.Method)
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}
	targetName := ""
	if request.Metadata != nil {
		if v, ok := request.Metadata["active-target"]; ok {
			targetName = v
		}
	}
	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("Apply Deployment", rContext, nil)
		defer span.End()
		deployment := new(model.DeploymentSpec)
		err := utils2.UnmarshalJson(request.Body, &deployment)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onApplyDeployment failed - %s", err.Error())
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := c.doDeploy(ctx, *deployment, namespace, targetName)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("Get Components", rContext, nil)
		defer span.End()
		deployment := new(model.DeploymentSpec)
		err := utils2.UnmarshalJson(request.Body, &deployment)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onApplyDeployment failed - %s", err.Error())
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := c.doGet(ctx, *deployment, targetName)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("Delete Components", rContext, nil)
		defer span.End()
		var deployment model.DeploymentSpec
		err := utils2.UnmarshalJson(request.Body, &deployment)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (Solution): onApplyDeployment failed - %s", err.Error())
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := c.doRemove(ctx, deployment, namespace, targetName)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	}
	sLog.ErrorCtx(rContext, "V (Solution): onApplyDeployment failed - 405 method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *SolutionVendor) doGet(ctx context.Context, deployment model.DeploymentSpec, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doGet",
	})
	defer span.End()
	sLog.InfoCtx(ctx, "V (Solution): doGet")

	_, components, err := c.SolutionManager.Get(ctx, deployment, targetName)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (Solution): doGet failed - %s", err.Error())
		response := v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
		observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
		return response
	}
	data, _ := json.Marshal(components)
	response := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
	return response
}
func (c *SolutionVendor) doDeploy(ctx context.Context, deployment model.DeploymentSpec, namespace string, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doDeploy",
	})
	defer span.End()
	sLog.InfoCtx(ctx, "V (Solution): doDeploy")
	summary, err := c.SolutionManager.Reconcile(ctx, deployment, false, namespace, targetName)
	data, _ := json.Marshal(summary)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (Solution): doDeploy failed - %s", err.Error())
		response := v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  data,
		}
		observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
		return response
	}
	response := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
	return response
}
func (c *SolutionVendor) doRemove(ctx context.Context, deployment model.DeploymentSpec, namespace string, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", ctx, &map[string]string{
		"method": "doRemove",
	})
	defer span.End()

	sLog.InfoCtx(ctx, "V (Solution): doRemove")
	summary, err := c.SolutionManager.Reconcile(ctx, deployment, true, namespace, targetName)
	data, _ := json.Marshal(summary)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (Solution): doRemove failed - %s", err.Error())
		response := v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  data,
		}
		observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
		return response
	}
	response := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, response)
	return response
}

// onGetRequest handles the get request from the remote agent.
func (c *SolutionVendor) onGetRequest(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onGetRequest",
	})
	defer span.End()
	var agentRequest model.AgentRequest
	sLog.InfoCtx(ctx, "V(Solution): onGetRequest")
	target := request.Parameters["target"]
	namespace := request.Parameters["namespace"]
	getAll, exists := request.Parameters["getAll"]

	// Extract correlationId from request parameters
	correlationId := request.Parameters["correlationId"]

	if exists && getAll == "true" {
		// Logic to handle getALL parameter
		sLog.InfoCtx(ctx, "V(Solution): getALL request from remote agent %+v", agentRequest)

		start, startExist := request.Parameters["preindex"]
		if !startExist {
			start = "0"
		}
		sizeStr, sizeExist := request.Parameters["size"]
		var size int
		var err error
		if !sizeExist {
			size = 10
		} else {
			size, err = strconv.Atoi(sizeStr)
			if err != nil {
				// Handle the error, for example, set a default value or return an error
				size = 10
			}
		}

		return c.SolutionManager.GetTaskFromQueueByPaging(ctx, target, namespace, start, size, correlationId)
	}
	return c.SolutionManager.GetTaskFromQueue(ctx, target, namespace, correlationId)
}

// onGetResponse handles the get response from the remote agent.
func (c *SolutionVendor) onGetResponse(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onGetResponse",
	})
	defer span.End()
	sLog.InfoCtx(ctx, "V (Solution): onGetResponse")
	var asyncResult model.AsyncResult
	err := utils.UnmarshalJson(request.Body, &asyncResult)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V(Solution): onGetResponse failed - %s", err.Error())
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	sLog.InfoCtx(ctx, "V(Solution): get async result from remote agent %+v", asyncResult)
	return c.SolutionManager.HandleRemoteAgentExecuteResult(ctx, asyncResult)
}
