/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/instances"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var iLog = logger.NewLogger("coa.runtime")

type InstancesVendor struct {
	vendors.Vendor
	InstancesManager *instances.InstancesManager
}

func (o *InstancesVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Instances",
		Producer: "Microsoft",
	}
}

func (e *InstancesVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*instances.InstancesManager); ok {
			e.InstancesManager = c
		}
	}
	if e.InstancesManager == nil {
		return v1alpha2.NewCOAError(nil, "instances manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *InstancesVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "instances"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onInstances,
			Parameters: []string{"name?"},
		},
	}
}

func (c *InstancesVendor) onInstances(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Instances Vendor", request.Context, &map[string]string{
		"method": "onInstances",
	})
	defer span.End()
	iLog.InfofCtx(pCtx, "V (Instances): onInstances, method: %s", request.Method)

	id := request.Parameters["__name"]
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	iLog.InfofCtx(pCtx, "V (Instances): onInstances, method: %s", string(request.Method))
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onInstances-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change partition back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				namespace = ""
			}
			state, err = c.InstancesManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.InstancesManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			iLog.ErrorfCtx(ctx, "V (Instances): onInstances failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
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
		ctx, span := observability.StartSpan("onInstances-POST", pCtx, nil)
		solution := request.Parameters["solution"]
		target := request.Parameters["target"]
		target_selector := request.Parameters["target-selector"]

		var instance model.InstanceState

		if solution != "" && (target != "" || target_selector != "") {
			instance = model.InstanceState{
				ObjectMeta: model.ObjectMeta{
					Name:      id,
					Namespace: namespace,
				},
				Spec: &model.InstanceSpec{
					DisplayName: id,
					Solution:    solution,
				},
			}
			if target != "" {
				instance.Spec.Target = model.TargetSelector{
					Name: target,
				}
			} else {
				parts := strings.Split(target_selector, "=")
				if len(parts) != 2 {
					iLog.ErrorCtx(ctx, "V (Instances): onInstances failed - invalid target selector format")
					return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
						State: v1alpha2.InternalError,
						Body:  []byte("invalid target selector format. Expected: <property>=<value>"),
					})
				}
				instance.Spec.Target = model.TargetSelector{
					Selector: map[string]string{
						parts[0]: parts[1],
					},
				}
			}
		} else {
			err := json.Unmarshal(request.Body, &instance)
			if err != nil {
				iLog.ErrorfCtx(ctx, "V (Instances): onInstances failed - %s", err.Error())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
			if instance.ObjectMeta.Name == "" {
				instance.ObjectMeta.Name = id
			}
		}
		err := c.InstancesManager.UpsertState(ctx, id, instance)
		if err != nil {
			iLog.ErrorfCtx(ctx, "V (Instances): onInstances failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if c.Config.Properties["useJobManager"] == "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "instance",
					"namespace":  instance.ObjectMeta.Namespace,
				},
				Body: v1alpha2.JobData{
					Id:     id,
					Action: v1alpha2.JobUpdate,
					Scope:  instance.ObjectMeta.Namespace,
				},
				Context: ctx,
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onInstances-DELETE", pCtx, nil)
		direct := request.Parameters["direct"]
		if c.Config.Properties["useJobManager"] == "true" && direct != "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "instance",
					"namespace":  namespace,
				},
				Body: v1alpha2.JobData{
					Id:     id,
					Action: v1alpha2.JobDelete,
					Scope:  namespace,
				},
				Context: ctx,
			})
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.OK,
			})
		} else {
			err := c.InstancesManager.DeleteState(ctx, id, namespace)
			if err != nil {
				iLog.ErrorfCtx(ctx, "V (Instances): onInstances failed - %s", err.Error())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	iLog.InfoCtx(pCtx, "V (Instances): onInstances failed - 405 method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
