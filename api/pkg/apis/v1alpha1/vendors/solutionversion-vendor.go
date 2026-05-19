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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutionversion"
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

type SolutionVersionVendor struct {
	vendors.Vendor
	SolutionVersionManager *solutionversion.SolutionVersionManager
}

func (o *SolutionVersionVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "SolutionVersion",
		Producer: "Microsoft",
	}
}

func (e *SolutionVersionVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*solutionversion.SolutionVersionManager); ok {
			e.SolutionVersionManager = c
		}
	}
	if e.SolutionVersionManager == nil {
		return v1alpha2.NewCOAError(nil, "solutionversion manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *SolutionVersionVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "solutionversion"
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
	}
}
func (c *SolutionVersionVendor) onQueue(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("SolutionVersion Vendor", request.Context, &map[string]string{
		"method": "onQueue",
	})
	defer span.End()
	instance := request.Parameters["instance"]
	sLog.InfofCtx(rContext, "V (SolutionVersion): onQueue, method: %s, %s", request.Method, instance)

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
			sLog.ErrorCtx(ctx, "V (SolutionVersion): onQueue failed - 400 instance parameter is not found")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - instance parameter is not found\"}"),
				ContentType: "application/json",
			})
		}
		summary, err := c.SolutionVersionManager.GetSummary(ctx, instance, instanceName, namespace)
		data, _ := json.Marshal(summary)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (SolutionVersion): onQueue failed - %s", err.Error())
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
		}

		if instance == "" {
			sLog.ErrorCtx(ctx, "V (SolutionVersion): onQueue failed - 400 instance parameter is not found")
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
			sLog.ErrorCtx(ctx, "V (SolutionVersion): onQueue failed - 400 instance parameter is not found")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - instance parameter is not found\"}"),
				ContentType: "application/json",
			})
		}

		err := c.SolutionVersionManager.DeleteSummary(ctx, instance, namespace)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (SolutionVersion): onQueue DeleteSummary failed - %s", err.Error())
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
	sLog.ErrorCtx(rContext, "V (SolutionVersion): onQueue failed - 405 method not allowed")
	return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	})
}
func (c *SolutionVersionVendor) onReconcile(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("SolutionVersion Vendor", request.Context, &map[string]string{
		"method": "onReconcile",
	})
	defer span.End()

	sLog.InfofCtx(rContext, "V (SolutionVersion): onReconcile, method: %s", request.Method)
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
			sLog.ErrorfCtx(ctx, "V (SolutionVersion): onReconcile failed POST - unmarshal request %s", err.Error())
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
		summary, err := c.SolutionVersionManager.Reconcile(ctx, deployment, delete == "true", namespace, targetName)
		data, _ := json.Marshal(summary)
		if err != nil {
			sLog.ErrorfCtx(ctx, "V (SolutionVersion): onReconcile failed POST - reconcile %s", err.Error())
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
	sLog.ErrorCtx(rContext, "V (SolutionVersion): onReconcile failed - 405 method not allowed")
	return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	})
}

func (c *SolutionVersionVendor) onApplyDeployment(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rContext, span := observability.StartSpan("SolutionVersion Vendor", request.Context, &map[string]string{
		"method": "onApplyDeployment",
	})
	defer span.End()

	sLog.InfofCtx(rContext, "V (SolutionVersion): onApplyDeployment %s", request.Method)
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
			sLog.ErrorfCtx(ctx, "V (SolutionVersion): onApplyDeployment failed - %s", err.Error())
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
			sLog.ErrorfCtx(ctx, "V (SolutionVersion): onApplyDeployment failed - %s", err.Error())
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
			sLog.ErrorfCtx(ctx, "V (SolutionVersion): onApplyDeployment failed - %s", err.Error())
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := c.doRemove(ctx, deployment, namespace, targetName)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	}
	sLog.ErrorCtx(rContext, "V (SolutionVersion): onApplyDeployment failed - 405 method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *SolutionVersionVendor) doGet(ctx context.Context, deployment model.DeploymentSpec, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("SolutionVersion Vendor", ctx, &map[string]string{
		"method": "doGet",
	})
	defer span.End()
	sLog.InfoCtx(ctx, "V (SolutionVersion): doGet")

	_, components, err := c.SolutionVersionManager.Get(ctx, deployment, targetName)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (SolutionVersion): doGet failed - %s", err.Error())
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
func (c *SolutionVersionVendor) doDeploy(ctx context.Context, deployment model.DeploymentSpec, namespace string, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("SolutionVersion Vendor", ctx, &map[string]string{
		"method": "doDeploy",
	})
	defer span.End()
	sLog.InfoCtx(ctx, "V (SolutionVersion): doDeploy")
	summary, err := c.SolutionVersionManager.Reconcile(ctx, deployment, false, namespace, targetName)
	data, _ := json.Marshal(summary)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (SolutionVersion): doDeploy failed - %s", err.Error())
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
func (c *SolutionVersionVendor) doRemove(ctx context.Context, deployment model.DeploymentSpec, namespace string, targetName string) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("SolutionVersion Vendor", ctx, &map[string]string{
		"method": "doRemove",
	})
	defer span.End()

	sLog.InfoCtx(ctx, "V (SolutionVersion): doRemove")
	summary, err := c.SolutionVersionManager.Reconcile(ctx, deployment, true, namespace, targetName)
	data, _ := json.Marshal(summary)
	if err != nil {
		sLog.ErrorfCtx(ctx, "V (SolutionVersion): doRemove failed - %s", err.Error())
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
