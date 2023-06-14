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
	"strconv"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/reference"
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

var log = logger.NewLogger("coa.runtime")

type AgentVendor struct {
	vendors.Vendor
	ReferenceManager *reference.ReferenceManager
}

func (o *AgentVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Agent",
		Producer: "Microsoft",
	}
}

func (e *AgentVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*reference.ReferenceManager); ok {
			e.ReferenceManager = c
		}
	}
	if e.ReferenceManager == nil {
		return v1alpha2.NewCOAError(nil, "reference manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *AgentVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "agent"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodGet, fasthttp.MethodPost},
			Route:   route + "/references",
			Version: o.Version,
			Handler: o.onReference,
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/config",
			Version: o.Version,
			Handler: o.onConfig,
		},
	}
}
func (c *AgentVendor) onConfig(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Agent Vendor", request.Context, nil)

	log.Infof("V (Agent): onConfig %s", request.Method)

	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("Apply Config", request.Context, nil)
		response := c.doApplyConfig(ctx, request.Parameters, request.Body)
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
func (c *AgentVendor) onReference(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Agent Vendor", request.Context, nil)

	log.Infof("V (Agent): onReference %s", request.Method)

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("Get References", request.Context, nil)
		response := c.doGet(ctx, request.Parameters)
		return observ_utils.CloseSpanWithCOAResponse(span, response)
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("Report Status", request.Context, nil)
		response := c.doPost(ctx, request.Parameters, request.Body)
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

func (c *AgentVendor) doGet(ctx context.Context, parameters map[string]string) v1alpha2.COAResponse {
	var scope = "default"
	var kind = ""
	var ref = ""
	var group = ""
	var id = ""
	var version = ""
	var fieldSelector = ""
	var labelSelector = ""
	var instance = ""
	var lookup = ""
	var platform = ""
	var flavor = ""
	var iteration = ""
	var alias = ""
	if v, ok := parameters["scope"]; ok {
		scope = v
	}
	if v, ok := parameters["ref"]; ok {
		ref = v
	}
	if v, ok := parameters["kind"]; ok {
		kind = v
	}
	if v, ok := parameters["version"]; ok {
		version = v
	}
	if v, ok := parameters["group"]; ok {
		group = v
	}
	if v, ok := parameters["id"]; ok {
		id = v
	}
	if v, ok := parameters["field-selector"]; ok {
		fieldSelector = v
	}
	if v, ok := parameters["label-selector"]; ok {
		labelSelector = v
	}
	if v, ok := parameters["instance"]; ok {
		instance = v
	}
	if v, ok := parameters["platform"]; ok {
		platform = v
	}
	if v, ok := parameters["flavor"]; ok {
		flavor = v
	}
	if v, ok := parameters["lookup"]; ok {
		lookup = v
	}
	if v, ok := parameters["iteration"]; ok {
		iteration = v
	}
	if v, ok := parameters["alias"]; ok {
		alias = v
	}

	var data []byte
	var err error
	if instance != "" {
		data, err = c.ReferenceManager.GetExt(ref, scope, id, group, kind, version, instance, "symphony.microsoft.com", "instances", "v1", "", alias)
	} else if lookup != "" {
		data, err = c.ReferenceManager.GetExt(ref, scope, id, group, kind, version, instance, lookup, platform, flavor, iteration, "")
	} else {
		data, err = c.ReferenceManager.Get(ref, id, scope, group, kind, version, labelSelector, fieldSelector)
	}
	if err != nil {
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}

func (c *AgentVendor) doApplyConfig(ctx context.Context, parameters map[string]string, data []byte) v1alpha2.COAResponse {
	var config managers.ProviderConfig
	err := json.Unmarshal(data, &config)
	if err != nil {
		return v1alpha2.COAResponse{
			State: v1alpha2.BadRequest,
			Body:  []byte(err.Error()),
		}
	}
	// TODO: The following is temporary implementation. A proper mechanism to reconfigure providers/managers/vendors is needed. This doesn't handle scaling out
	// (when multiple vendor instances are behind load balancer), either
	switch config.Type {
	case "providers.reference.customvision":
		for _, p := range c.ReferenceManager.ReferenceProviders {
			err = p.Reconfigure(config.Config)
			if err != nil {
				return v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				}
			}
		}
	}
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        []byte("{}"),
		ContentType: "application/json",
	}
}

func (c *AgentVendor) doPost(ctx context.Context, parameters map[string]string, data []byte) v1alpha2.COAResponse {
	var scope = "default"
	var kind = ""
	var group = ""
	var id = ""
	var version = ""
	var overwrite = false
	if v, ok := parameters["scope"]; ok {
		scope = v
	}
	if v, ok := parameters["kind"]; ok {
		kind = v
	}
	if v, ok := parameters["version"]; ok {
		version = v
	}
	if v, ok := parameters["group"]; ok {
		group = v
	}
	if v, ok := parameters["id"]; ok {
		id = v
	}
	if v, ok := parameters["overwrite"]; ok {
		overwrite, _ = strconv.ParseBool(v)
	}
	properties := make(map[string]string)
	err := json.Unmarshal(data, &properties)
	if err != nil {
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	err = c.ReferenceManager.Report(id, scope, group, kind, version, properties, overwrite)
	if err != nil {
		return v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		}
	}
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}
