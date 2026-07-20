/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"
	"strings"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/valyala/fasthttp"
)

// SecurityPolicyVendor serves the server-side SecurityPolicy to all providers.
// It is discovered by the host and its policy is propagated to every vendor's
// VendorContext, making it accessible to providers via ManagerContext.GetSecurityPolicy().
//
// Example configuration:
//
//	{
//	  "type": "vendors.securitypolicy",
//	  "route": "security-policy",
//	  "properties": {
//	    "allowedIPRanges": "[\"10.0.5.0/24\",\"192.168.1.100\"]",
//	    "allowListExclusive": "true"
//	  }
//	}
type SecurityPolicyVendor struct {
	vendors.Vendor
	policy contexts.SecurityPolicy
}

func (v *SecurityPolicyVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  v.Vendor.Version,
		Name:     "SecurityPolicy",
		Producer: "Microsoft",
	}
}

// GetSecurityPolicy implements ISecurityPolicyVendor. The host calls this after all
// vendors are initialized and broadcasts the returned policy to every vendor.
func (v *SecurityPolicyVendor) GetSecurityPolicy() *contexts.SecurityPolicy {
	return &v.policy
}

func (v *SecurityPolicyVendor) Init(cfg vendors.VendorConfig, factories []managers.IManagerFactroy, ps map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := v.Vendor.Init(cfg, factories, ps, pubsubProvider)
	if err != nil {
		return err
	}

	// Parse allowedIPRanges: a JSON array string, e.g. "[\"10.0.0.0/8\"]".
	if raw, ok := cfg.Properties["allowedIPRanges"]; ok && raw != "" {
		var ranges []string
		if err := json.Unmarshal([]byte(raw), &ranges); err != nil {
			return v1alpha2.NewCOAError(err, "invalid allowedIPRanges in security policy vendor config", v1alpha2.BadConfig)
		}
		v.policy.AllowedIPRanges = ranges
	}

	// Parse allowListExclusive: "true" / "1" / "yes" to enable.
	if raw, ok := cfg.Properties["allowListExclusive"]; ok {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "true", "1", "yes":
			v.policy.AllowListExclusive = true
		}
	}

	return nil
}

// GetEndpoints exposes a GET endpoint that returns the current security policy as JSON.
// This can be used by administrators to inspect the active policy.
func (v *SecurityPolicyVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "security-policy"
	if v.Route != "" {
		route = v.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route,
			Version: v.Version,
			Handler: v.onGet,
		},
	}
}

func (v *SecurityPolicyVendor) onGet(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("SecurityPolicy Vendor", request.Context, &map[string]string{
		"method": "onGet",
	})
	defer span.End()

	if request.Method != fasthttp.MethodGet {
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.MethodNotAllowed,
			Body:        []byte(`{"result":"405 - method not allowed"}`),
			ContentType: "application/json",
		})
	}

	data, err := json.Marshal(v.policy)
	if err != nil {
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		})
	}
	return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	})
}
