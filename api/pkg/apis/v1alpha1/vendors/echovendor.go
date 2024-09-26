/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/valyala/fasthttp"
)

type EchoVendor struct {
	vendors.Vendor
	myMessages []string
	lock       sync.Mutex
}

func (o *EchoVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Echo",
		Producer: "Microsoft",
	}
}

func (e *EchoVendor) GetMessages() []string {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.myMessages
}

func (e *EchoVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	e.myMessages = make([]string, 0)
	e.Vendor.Context.Subscribe("trace", func(topic string, event v1alpha2.Event) error {
		e.lock.Lock()
		defer e.lock.Unlock()
		msg := utils.FormatAsString(event.Body)
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
	ctx, span := observability.StartSpan("Echo Vendor", request.Context, &map[string]string{
		"method": "onHello",
	})
	defer span.End()

	switch request.Method {
	case fasthttp.MethodGet:
		c.lock.Lock()
		defer c.lock.Unlock()
		message := "Hello from Symphony K8s control plane (S8C)"
		if len(c.myMessages) > 0 {
			for _, m := range c.myMessages {
				message = message + "\r\n" + m
			}
		}
		resp := v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte(message),
			ContentType: "text/plain",
		}
		return observ_utils.CloseSpanWithCOAResponse(span, resp)
	case fasthttp.MethodPost:
		c.Vendor.Context.Publish("trace", v1alpha2.Event{
			Body:    string(request.Body),
			Context: ctx,
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
	return resp
}
