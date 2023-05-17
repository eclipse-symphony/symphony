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
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/valyala/fasthttp"
)

type EchoVendor struct {
	vendors.Vendor
	myMessages []string
}

func (o *EchoVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Echo",
		Producer: "Microsoft",
	}
}

func (e *EchoVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	e.myMessages = make([]string, 0)
	e.Vendor.Context.Subscribe("trace", func(topic string, event v1alpha2.Event) error {
		msg := event.Body.(string)
		e.myMessages = append(e.myMessages, msg)
		if len(e.myMessages) > 20 {
			e.myMessages = e.myMessages[1:]
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *EchoVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "greetings"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodGet, fasthttp.MethodPost},
			Route:   route,
			Version: o.Version,
			Handler: o.onHello,
		},
	}
}

func (c *EchoVendor) onHello(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Echo Vendor", request.Context, nil)
	switch request.Method {
	case fasthttp.MethodGet:
		message := "Hello from Symphony K8s control plane (S8C)"
		if len(c.myMessages) > 0 {
			for _, m := range c.myMessages {
				message = message + "\r\n" + m
			}
		}
		resp := v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte(message),
			ContentType: "application/text",
		}
		return observ_utils.CloseSpanWithCOAResponse(span, resp)
	case fasthttp.MethodPost:
		c.Vendor.Context.Publish("trace", v1alpha2.Event{
			Body: string(request.Body),
		})
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
