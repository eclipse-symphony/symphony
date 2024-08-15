/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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

type BootstrapVendor struct {
	vendors.Vendor
	ActivationsManager *activations.ActivationsManager
}

func (b *BootstrapVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  b.Vendor.Version,
		Name:     "Bootstrap",
		Producer: "Microsoft",
	}
}

func (b *BootstrapVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := b.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range b.Managers {
		if c, ok := m.(*activations.ActivationsManager); ok {
			b.ActivationsManager = c
		}
	}
	if b.ActivationsManager == nil {
		return v1alpha2.NewCOAError(nil, "activations manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (b *BootstrapVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "bootstrap"
	if b.Route != "" {
		route = b.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost},
			Route:      route,
			Version:    b.Version,
			Handler:    b.onBootstrap,
			Parameters: []string{"flow?"},
		},
	}
}

func (b *BootstrapVendor) onBootstrap(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Bootstrap Vendor", request.Context, &map[string]string{
		"method": "onBootstrap",
	})
	defer span.End()

	vLog.Infof("V (BootstrapVendor Vendor): onBootstrap, method: %s, traceId: %s", string(request.Method), span.SpanContext().TraceID().String())

	namespace, namespaceSupplied := request.Parameters["namespace"]
	if !namespaceSupplied {
		namespace = "default"
	}

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onBootstrap-GET", pCtx, nil)
		id := request.Parameters["__flow"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			if !namespaceSupplied {
				namespace = ""
			}
			state, err = b.ActivationsManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = b.ActivationsManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			vLog.Infof("V (Bootstrap Vendor): onBootstrap failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
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
		// TODO: It might be necessary to constrain which campaigns can be activated, maybe
		// as a configuration option, or a "selector" that select different flows based on
		// posted data.
		ctx, span := observability.StartSpan("onActivations-POST", pCtx, nil)
		id := request.Parameters["__flow"]
		var activation model.ActivationState

		err := json.Unmarshal(request.Body, &activation)
		if err != nil {
			vLog.Infof("V (Bootstrap Vendor): onActivations failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		activationHash, err := activation.Spec.Hash()

		if err != nil {
			vLog.Infof("V (Bootstrap Vendor): onBootstrap failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		activationId := fmt.Sprintf("%s-%s", id, activationHash)
		activation.ObjectMeta.Name = activationId

		err = b.ActivationsManager.UpsertState(ctx, activationId, activation)
		if err != nil {
			vLog.Infof("V (BootstrapVendor Vendor): onActivations failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		// TODO: this sleep is a hack and is not guaranteed to always work. When REST API is used against a K8s state provider, creating the activation object triggers
		// the activation controller to raise the activation event as well. This is a workaround to avoid duplicated events. A proper
		// implemenation probably needs to leverage a distributed lock - such leverage a Redis lock.
		time.Sleep(1 * time.Second)

		entry, err := b.ActivationsManager.GetState(ctx, activationId, activation.ObjectMeta.Namespace)
		if err != nil {
			vLog.Infof("V (Activations Vendor): onActivations failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if !entry.Status.IsActive {
			b.Context.Publish("activation", v1alpha2.Event{
				Body: v1alpha2.ActivationData{
					Campaign:             activation.Spec.Campaign,
					ActivationGeneration: entry.ObjectMeta.Generation,
					Activation:           activationId,
					Stage:                "",
					Inputs:               activation.Spec.Inputs,
					Namespace:            activation.ObjectMeta.Namespace,
				},
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	vLog.Infof("V (Activations Vendor): onBootstrap failed - 405 method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
