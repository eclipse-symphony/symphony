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
	"encoding/json"
	"strings"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/instances"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
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
	tLog.Info("~ Instances Manager ~ : onInstances")

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onInstances-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			state, err = c.InstancesManager.ListSpec(ctx)
			isArray = true
		} else {
			state, err = c.InstancesManager.GetSpec(ctx, id)
		}
		if err != nil {
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
			resp.ContentType = "application/text"
		}
		return resp
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onInstances-POST", pCtx, nil)
		id := request.Parameters["__name"]

		solution := request.Parameters["solution"]
		target := request.Parameters["target"]
		target_selector := request.Parameters["target-selector"]

		var instance model.InstanceSpec

		if solution != "" && (target != "" || target_selector != "") {
			instance = model.InstanceSpec{
				DisplayName: id,
				Name:        id,
				Solution:    solution,
			}
			if target != "" {
				instance.Target = model.TargetSelector{
					Name: target,
				}
			} else {
				parts := strings.Split(target_selector, "=")
				if len(parts) != 2 {
					return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
						State: v1alpha2.InternalError,
						Body:  []byte("invalid target selector format. Expected: <property>=<value>"),
					})
				}
				instance.Target = model.TargetSelector{
					Selector: map[string]string{
						parts[0]: parts[1],
					},
				}
			}
		} else {
			err := json.Unmarshal(request.Body, &instance)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}
		err := c.InstancesManager.UpsertSpec(ctx, id, instance)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if c.Config.Properties["useJobManager"] == "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "instance",
				},
				Body: v1alpha2.JobData{
					Id:     id,
					Action: "UPDATE",
				},
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onInstances-DELETE", pCtx, nil)
		id := request.Parameters["__name"]
		direct := request.Parameters["direct"]
		if c.Config.Properties["useJobManager"] == "true" && direct != "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "instance",
				},
				Body: v1alpha2.JobData{
					Id:     id,
					Action: "DELETE",
				},
			})
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.OK,
			})
		} else {
			err := c.InstancesManager.DeleteSpec(ctx, id)
			if err != nil {
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
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	span.End()
	return resp
}
