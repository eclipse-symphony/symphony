/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/jobs"
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
	e.Vendor.Context.Subscribe("trace", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			msg := utils.FormatAsString(event.Body)
			e.myMessages = append(e.myMessages, msg)
			if len(e.myMessages) > 20 {
				e.myMessages = e.myMessages[1:]
			}
			return nil
		},
		Group: "job",
	})
	e.Vendor.Context.Subscribe("job", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			err := e.JobsManager.HandleJobEvent(ctx, event)
			if err != nil && v1alpha2.IsDelayed(err) {
				return err
			}
			// job reconciler already has a retry mechanism, return nil to avoid retrying
			return nil
		},
	})
	e.Vendor.Context.Subscribe("heartbeat", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			return e.JobsManager.HandleHeartBeatEvent(ctx, event)
		},
	})
	e.Vendor.Context.Subscribe("schedule", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			return e.JobsManager.HandleScheduleEvent(ctx, event)
		},
	})

	return err
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
	ctx, span := observability.StartSpan("Job Vendor", request.Context, &map[string]string{
		"method": "onHello",
	})
	defer span.End()

	jLog.InfofCtx(ctx, "V (Job): onHello, method: %s", string(request.Method))
	switch request.Method {
	case fasthttp.MethodPost:
		var activationData v1alpha2.ActivationData
		err := json.Unmarshal(request.Body, &activationData)
		if err != nil {
			jLog.ErrorfCtx(ctx, "V (Job): onHello failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\":\"400 - bad request\"}"),
				ContentType: "application/json",
			})
		}
		jLog.InfofCtx(ctx, "V (Job): onHello, activationData: %v", activationData)
		err = c.Vendor.Context.Publish("activation", v1alpha2.Event{
			Body:    activationData,
			Context: ctx,
		})
		if err != nil {
			jLog.ErrorfCtx(ctx, "V (Job): onHello failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.InternalError,
				Body:        []byte(err.Error()),
				ContentType: "application/json",
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	jLog.ErrorCtx(ctx, "V (Job): onHello failed - 405 method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)

	return resp
}
