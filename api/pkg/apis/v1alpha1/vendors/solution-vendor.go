/*
Copyright 2022 The COA Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vendors

import (
	"context"
	"encoding/json"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
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

func (e *SolutionVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider) error {
	err := e.Vendor.Init(config, factories, providers)
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

func (o *SolutionVendor) HasLoop() bool {
	return true
}

func (o *SolutionVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "solution"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost, fasthttp.MethodGet, fasthttp.MethodDelete},
			Route:   route + "/instances",
			Version: o.Version,
			Handler: o.onApplyDeployment,
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/needsupdate",
			Version: o.Version,
			Handler: o.onNeedsUpdate,
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/needsremove",
			Version: o.Version,
			Handler: o.onNeedsRemove,
		},
	}
}

type TwoComponentSlices struct {
	Current []model.ComponentSpec `json:"current"`
	Desired []model.ComponentSpec `json:"desired"`
}

func (c *SolutionVendor) onNeedsUpdate(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Solution Vendor", context.Background(), &map[string]string{
		"method": "onNeedsUpdate",
	})
	log.Info("V (Solution): onNeedsUpdate")

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onNeedsUpdate", request.Context, nil)
		slices := new(TwoComponentSlices)
		err := json.Unmarshal(request.Body, &slices)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		b := c.SolutionManager.NeedsUpdate(ctx, slices.Desired, slices.Current)
		if b {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.OK,
				Body:  []byte("{\"result\":\"200\"}"),
			})
		} else {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte("{\"result\":\"5001\"}"),
			})
		}
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
func (c *SolutionVendor) onNeedsRemove(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Solution Vendor", context.Background(), &map[string]string{
		"method": "onNeedsRemove",
	})
	log.Info("V (Solution): onNeedsRemove")

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onNeedsRemove", request.Context, nil)
		slices := new(TwoComponentSlices)
		err := json.Unmarshal(request.Body, &slices)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		b := c.SolutionManager.NeedsRemove(ctx, slices.Desired, slices.Current)

		if b {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.OK,
				Body:  []byte("{\"result\":\"200\"}"),
			})
		} else {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte("{\"result\":\"5001\"}"),
			})
		}
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

func (c *SolutionVendor) onApplyDeployment(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Solution Vendor", request.Context, nil)

	log.Infof("V (Solution): received request %s", request.Method)

	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("Apply Deployment", request.Context, nil)
		deployment := new(model.DeploymentSpec)
		err := json.Unmarshal(request.Body, &deployment)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := c.doDeploy(ctx, *deployment)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("Get Components", request.Context, nil)
		deployment := new(model.DeploymentSpec)
		err := json.Unmarshal(request.Body, &deployment)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := c.doGet(ctx, *deployment)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("Delete Components", request.Context, nil)
		var deployment model.DeploymentSpec
		err := json.Unmarshal(request.Body, &deployment)
		if err != nil {
			return v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			}
		}
		response := c.doRemove(ctx, deployment)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
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

func (c *SolutionVendor) doGet(ctx context.Context, deployment model.DeploymentSpec) v1alpha2.COAResponse {
	components, err := c.SolutionManager.Get(ctx, deployment)
	if err != nil {
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	data, _ := json.Marshal(components)
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}
func (c *SolutionVendor) doDeploy(ctx context.Context, deployment model.DeploymentSpec) v1alpha2.COAResponse {
	summary, err := c.SolutionManager.Apply(ctx, deployment)
	data, _ := json.Marshal(summary)
	if err != nil {
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  data,
		}
	}
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}
func (c *SolutionVendor) doRemove(ctx context.Context, deployment model.DeploymentSpec) v1alpha2.COAResponse {
	summary, err := c.SolutionManager.Remove(ctx, deployment)
	data, _ := json.Marshal(summary)
	if err != nil {
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  data,
		}
	}
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}
