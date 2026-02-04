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

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
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
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/cancel",
			Version: o.Version,
			Handler: o.onCancel,
		},
		{
			Methods: []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:   route + "/queue",
			Version: o.Version,
			Handler: o.onQueue,
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
			if utils.IsNotFound(err) {
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

			if delete == "true" {
				c.SolutionManager.CancelPreviousJobs(ctx, namespace, instance, deployment.JobID)
			}
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
		err := c.Vendor.Context.Publish("job", v1alpha2.Event{
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
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.InternalError,
				Body:        []byte("{\"result\":\"500 - failed to publish job pubsub event\")),"),
				ContentType: "application/json",
			})
		}
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
		targetName := ""
		if request.Metadata != nil {
			if v, ok := request.Metadata["active-target"]; ok {
				targetName = v
			}
		}

		summary, err := c.SolutionManager.ReconcileWithCancelWrapper(ctx, deployment, delete == "true", namespace, targetName)
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

func (c *SolutionVendor) onCancel(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("Solution Vendor", request.Context, &map[string]string{
		"method": "onCancel",
	})
	defer span.End()

	namespace := request.Parameters["namespace"]
	instance := request.Parameters["instance"]
	jobId := request.Parameters["jobId"]

	sLog.InfofCtx(rContext, "V (Solution): onCancel instance: %s job ID: %s", instance, jobId)
	c.SolutionManager.CancelPreviousJobs(rContext, namespace, instance, jobId)

	return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        []byte("{\"result\":\"200 - OK\"}"),
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
