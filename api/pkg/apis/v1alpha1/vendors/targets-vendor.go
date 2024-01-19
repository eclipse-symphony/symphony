/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/targets"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/valyala/fasthttp"
)

var tLog = logger.NewLogger("coa.runtime")

type TargetsVendor struct {
	vendors.Vendor
	TargetsManager *targets.TargetsManager
}

func (o *TargetsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Targets",
		Producer: "Microsoft",
	}
}

func (e *TargetsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*targets.TargetsManager); ok {
			e.TargetsManager = c
		}
	}
	if e.TargetsManager == nil {
		return v1alpha2.NewCOAError(nil, "targets manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *TargetsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "targets"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route + "/registry",
			Version:    o.Version,
			Handler:    o.onRegistry,
			Parameters: []string{"name?"},
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/bootstrap",
			Version: o.Version,
			Handler: o.onBootstrap,
		},
		{
			Methods:    []string{fasthttp.MethodGet},
			Route:      route + "/ping",
			Version:    o.Version,
			Handler:    o.onHeartBeat,
			Parameters: []string{"name"},
		},
		{
			Methods:    []string{fasthttp.MethodPut},
			Route:      route + "/status",
			Version:    o.Version,
			Handler:    o.onStatus,
			Parameters: []string{"name", "component?"},
		},
		{
			Methods:    []string{fasthttp.MethodGet},
			Route:      route + "/download",
			Version:    o.Version,
			Handler:    o.onDownload,
			Parameters: []string{"doc-type", "name"},
		},
	}
}

type MyCustomClaims struct {
	User string `json:"user"`
	jwt.RegisteredClaims
}
type AuthRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func (c *TargetsVendor) onRegistry(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Targets Vendor", request.Context, &map[string]string{
		"method": "onRegistry",
	})
	defer span.End()
	log.Infof("V (Targets): onRegistry %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	scope, exist := request.Parameters["scope"]
	if !exist {
		scope = "default"
	}
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onRegistry-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change scope back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				scope = ""
			}
			state, err = c.TargetsManager.ListSpec(ctx, scope)
			isArray = true
		} else {
			state, err = c.TargetsManager.GetSpec(ctx, id, scope)
		}
		if err != nil {
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
		ctx, span := observability.StartSpan("onRegistry-POST", pCtx, nil)
		id := request.Parameters["__name"]
		binding := request.Parameters["with-binding"]
		var target model.TargetSpec
		err := json.Unmarshal(request.Body, &target)
		if err != nil {
			log.Errorf("V (Targets): onRegistry failed to unmarshall request body, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if binding != "" {
			if binding == "staging" {
				target.ForceRedeploy = true
				if target.Topologies == nil {
					target.Topologies = make([]model.TopologySpec, 0)
				}
				found := false
				for _, t := range target.Topologies {
					if t.Bindings != nil {
						for _, b := range t.Bindings {
							if b.Role == "instance" && b.Provider == "providers.target.staging" {
								found = true
								break
							}
						}
					}
				}
				if !found {
					newb := model.BindingSpec{
						Role:     "instance",
						Provider: "providers.target.staging",
						Config: map[string]string{
							"inCluster":  "true",
							"targetName": id,
						},
					}
					if len(target.Topologies) == 0 {
						target.Topologies = append(target.Topologies, model.TopologySpec{})
					}
					if target.Topologies[len(target.Topologies)-1].Bindings == nil {
						target.Topologies[len(target.Topologies)-1].Bindings = make([]model.BindingSpec, 0)
					}
					target.Topologies[len(target.Topologies)-1].Bindings = append(target.Topologies[len(target.Topologies)-1].Bindings, newb)
				}
			} else {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.BadRequest,
					Body:  []byte("invalid binding, supported is: 'staging'"),
				})
			}
		}
		err = c.TargetsManager.UpsertSpec(ctx, id, scope, target)
		if err != nil {
			log.Errorf("V (Targets): onRegistry failed to upsert spec, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if c.Config.Properties["useJobManager"] == "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "target",
					"scope":      scope,
				},
				Body: v1alpha2.JobData{
					Id:     id,
					Action: "UPDATE",
				},
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onRegistry-DELETE", pCtx, nil)
		id := request.Parameters["__name"]
		direct := request.Parameters["direct"]

		if c.Config.Properties["useJobManager"] == "true" && direct != "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "target",
					"scope":      scope,
				},
				Body: v1alpha2.JobData{
					Id:     id,
					Action: "DELETE",
				},
			})
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.OK,
			})
		} else {
			err := c.TargetsManager.DeleteSpec(ctx, id, scope)
			if err != nil {
				log.Errorf("V (Targets): onRegistry failed to delete spec, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}

	log.Infof("V (Targets): onRegistry returned MethodNotAllowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *TargetsVendor) onBootstrap(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Targets Vendor", request.Context, &map[string]string{
		"method": "onBootstrap",
	})
	defer span.End()
	log.Infof("V (Targets): onBootstrap %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	var authRequest AuthRequest
	err := json.Unmarshal(request.Body, &authRequest)
	if err != nil || authRequest.UserName != "symphony-test" {
		log.Error("V (Targets): onBootstrap returned Unauthorized, traceId: %s", request.Method, span.SpanContext().TraceID().String())
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.Unauthorized,
			Body:  []byte(err.Error()),
		})
	}
	mySigningKey := []byte("SymphonyKey")
	claims := MyCustomClaims{
		authRequest.UserName,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "symphony",
			Subject:   "symphony",
			ID:        "1",
			Audience:  []string{"*"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString(mySigningKey)

	log.Info("V (Targets): onBootstrap succeeded, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        []byte(`{"accessToken":"` + ss + `", "tokenType": "Bearer"}`),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *TargetsVendor) onStatus(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Targets Vendor", request.Context, &map[string]string{
		"method": "onStatus",
	})
	defer span.End()
	log.Infof("V (Targets): onStatus %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	scope, exist := request.Parameters["scope"]
	if !exist {
		scope = "default"
	}
	var dict map[string]interface{}
	json.Unmarshal(request.Body, &dict)

	properties := make(map[string]string)
	if k, ok := dict["status"]; ok {
		var insideKey map[string]interface{}
		j, _ := json.Marshal(k)
		json.Unmarshal(j, &insideKey)
		if p, ok := insideKey["properties"]; ok {
			jk, _ := json.Marshal(p)
			json.Unmarshal(jk, &properties)
		}
	}

	for k, v := range request.Parameters {
		if !strings.HasPrefix(k, "__") {
			properties[k] = v
		}
	}

	state, err := c.TargetsManager.ReportState(pCtx, model.TargetState{
		Id: request.Parameters["__name"],
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FabricGroup,
			"resource": "targets",
			"scope":    scope,
		},
		Status: properties,
	})

	if err != nil {
		log.Errorf("V (Targets): onStatus failed to report state, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		})
	}
	jData, _ := json.Marshal(state)

	log.Info("V (Targets): onStatus succeeded, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        jData,
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *TargetsVendor) onDownload(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Targets Vendor", request.Context, &map[string]string{
		"method": "onDownload",
	})
	defer span.End()
	log.Infof("V (Targets): onDownload %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	scope, exist := request.Parameters["scope"]
	if !exist {
		scope = "default"
	}
	state, err := c.TargetsManager.GetSpec(pCtx, request.Parameters["__name"], scope)
	if err != nil {
		log.Errorf("V (Targets): onDownload failed to get target spec, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		})
	}
	jData, err := utils.FormatObject(state, false, request.Parameters["path"], request.Parameters["__doc-type"])
	if err != nil {
		log.Errorf("V (Targets): onDownload failed to format target object, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		})
	}
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        jData,
		ContentType: "application/json",
	}

	if request.Parameters["__doc-type"] == "yaml" {
		resp.ContentType = "application/text"
	}

	log.Info("V (Targets): onDownload succeeded, traceId: %s", span.SpanContext().TraceID().String())
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *TargetsVendor) onHeartBeat(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Targets Vendor", request.Context, &map[string]string{
		"method": "onHeartBeat",
	})
	defer span.End()
	log.Infof("V (Targets): onHeartBeat %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	scope, exist := request.Parameters["scope"]
	if !exist {
		scope = "default"
	}
	_, err := c.TargetsManager.ReportState(pCtx, model.TargetState{
		Id: request.Parameters["__name"],
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FabricGroup,
			"resource": "targets",
			"scope":    scope,
		},
		Status: map[string]string{
			"ping": time.Now().UTC().String(),
		},
	})

	if err != nil {
		log.Errorf("V (Targets): failed to report state, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		})
	}

	log.Info("V (Targets): onHeartBeat succeeded, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        []byte(`{}`),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
