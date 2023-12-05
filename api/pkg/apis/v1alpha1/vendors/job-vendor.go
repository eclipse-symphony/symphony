/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/jobs"
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

var jLog = logger.NewLogger("coa.runtime")

type JobVendor struct {
	vendors.Vendor
	myMessages  []string
	JobsManager *jobs.JobsManager
}

func (o *JobVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Job",
		Producer: "Microsoft",
	}
}

func (e *JobVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*jobs.JobsManager); ok {
			e.JobsManager = c
		}
	}
	if e.JobsManager == nil {
		return v1alpha2.NewCOAError(nil, "jobs manager is not supplied", v1alpha2.MissingConfig)
	}
	e.myMessages = make([]string, 0)
	e.Vendor.Context.Subscribe("trace", func(topic string, event v1alpha2.Event) error {
		msg := event.Body.(string)
		e.myMessages = append(e.myMessages, msg)
		if len(e.myMessages) > 20 {
			e.myMessages = e.myMessages[1:]
		}
		return nil
	})
	e.Vendor.Context.Subscribe("job", func(topic string, event v1alpha2.Event) error {
		err := e.JobsManager.HandleJobEvent(context.Background(), event)
		if err != nil && v1alpha2.IsDelayed(err) {
			go e.Vendor.Context.Publish(topic, event)
		}
		return err
	})
	e.Vendor.Context.Subscribe("heartbeat", func(topic string, event v1alpha2.Event) error {
		return e.JobsManager.HandleHeartBeatEvent(context.Background(), event)
	})
	e.Vendor.Context.Subscribe("schedule", func(topic string, event v1alpha2.Event) error {
		return e.JobsManager.HandleScheduleEvent(context.Background(), event)
	})

	if err != nil {
		return err
	}

	return nil
}

func (o *JobVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "jobs"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route,
			Version: o.Version,
			Handler: o.onHello,
		},
	}
}

func (c *JobVendor) onHello(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Job Vendor", request.Context, &map[string]string{
		"method": "onHello",
	})
	defer span.End()

	switch request.Method {
	case fasthttp.MethodPost:
		var activationData v1alpha2.ActivationData
		err := json.Unmarshal(request.Body, &activationData)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - bad request\"}"),
				ContentType: "application/json",
			})
		}
		c.Vendor.Context.Publish("activation", v1alpha2.Event{
			Body: activationData,
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
