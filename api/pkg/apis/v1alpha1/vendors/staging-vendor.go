/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var gLog = logger.NewLogger("coa.runtime")

type StagingVendor struct {
	vendors.Vendor
}

func (f *StagingVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  f.Vendor.Version,
		Name:     "Staging",
		Producer: "Microsoft",
	}
}
func (f *StagingVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := f.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	return nil
}
func (f *StagingVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "federation"
	if f.Route != "" {
		route = f.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:   route + "/download",
			Version: f.Version,
			Handler: f.onDownload,
		},
	}
}
func (f *StagingVendor) onDownload(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Staging Vendor", request.Context, &map[string]string{
		"method": "onDownload",
	})
	defer span.End()

	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	return resp
}
