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
