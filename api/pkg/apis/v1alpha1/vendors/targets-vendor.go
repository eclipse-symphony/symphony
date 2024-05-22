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

	"github.com/eclipse-symphony/symphony/api/constants"
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
			Parameters: []string{"name", "version?"},
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/registry",
			Version: o.Version,
			Handler: o.onRegistryList,
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/bootstrap",
			Version: o.Version,
			Handler: o.onBootstrap,
		},
		{
			Methods:    []string{fasthttp.MethodPost},
			Route:      route + "/ping",
			Version:    o.Version,
			Handler:    o.onHeartBeat,
			Parameters: []string{"name", "version"},
		},
		{
			Methods:    []string{fasthttp.MethodPut},
			Route:      route + "/status",
			Version:    o.Version,
			Handler:    o.onStatus,
			Parameters: []string{"name", "version", "component?"},
		},
		{
			Methods:    []string{fasthttp.MethodGet},
			Route:      route + "/download",
			Version:    o.Version,
			Handler:    o.onDownload,
			Parameters: []string{"doc-type", "name", "version"},
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
	tLog.Infof("V (Targets) : onRegistry, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	version := request.Parameters["__version"]
	rootResource := request.Parameters["__name"]
	var id string
	var resourceId string
	if version != "" {
		id = rootResource + "-" + version
		resourceId = rootResource + ":" + version
	} else {
		id = rootResource
		resourceId = rootResource
	}
	tLog.Infof("V (Targets): onCampaigns, id: %s, version: %s", id, version)

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onRegistry-GET", pCtx, nil)
		var err error
		var state interface{}

		if version == "latest" {
			state, err = c.TargetsManager.GetLatestState(ctx, rootResource, namespace)
		} else {
			state, err = c.TargetsManager.GetState(ctx, id, namespace)
		}

		if err != nil {
			tLog.Infof("V (Targets) : onRegistry Get failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		jData, _ := utils.FormatObject(state, false, request.Parameters["path"], request.Parameters["doc-type"])
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

		if version == "" || version == "latest" {
			tLog.Infof("V (Targets): onRegistry Post failed - version is required for POST, traceId: %s", span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte("version is required for POST"),
			})
		}

		binding := request.Parameters["with-binding"]
		var target model.TargetState
		err := json.Unmarshal(request.Body, &target)
		if err != nil {
			tLog.Infof("V (Targets) : onRegistry Post failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if target.ObjectMeta.Name == "" {
			target.ObjectMeta.Name = id
		}
		if target.Spec.Version == "" && version != "" {
			target.Spec.Version = version
		}
		if target.Spec.RootResource == "" && rootResource != "" {
			target.Spec.RootResource = rootResource
		}

		if binding != "" {
			if binding == "staging" {
				target.Spec.ForceRedeploy = true
				if target.Spec.Topologies == nil {
					target.Spec.Topologies = make([]model.TopologySpec, 0)
				}
				found := false
				for _, t := range target.Spec.Topologies {
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
					if len(target.Spec.Topologies) == 0 {
						target.Spec.Topologies = append(target.Spec.Topologies, model.TopologySpec{})
					}
					if target.Spec.Topologies[len(target.Spec.Topologies)-1].Bindings == nil {
						target.Spec.Topologies[len(target.Spec.Topologies)-1].Bindings = make([]model.BindingSpec, 0)
					}
					target.Spec.Topologies[len(target.Spec.Topologies)-1].Bindings = append(target.Spec.Topologies[len(target.Spec.Topologies)-1].Bindings, newb)
				}
			} else {
				tLog.Infof("V (Targets) : onRegistry failed - invalid binding, traceId: %s", span.SpanContext().TraceID().String())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.BadRequest,
					Body:  []byte("invalid binding, supported is: 'staging'"),
				})
			}
		}
		err = c.TargetsManager.UpsertState(ctx, id, target)
		if err != nil {
			tLog.Infof("V (Targets) : onRegistry failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if c.Config.Properties["useJobManager"] == "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "target",
					"namespace":  namespace,
				},
				Body: v1alpha2.JobData{
					Id:     resourceId,
					Action: v1alpha2.JobUpdate,
					Scope:  namespace,
				},
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onRegistry-DELETE", pCtx, nil)
		direct := request.Parameters["direct"]

		if c.Config.Properties["useJobManager"] == "true" && direct != "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "target",
					"namespace":  namespace,
				},
				Body: v1alpha2.JobData{
					Id:     resourceId,
					Action: v1alpha2.JobDelete,
					Scope:  namespace,
				},
			})
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.OK,
			})
		} else {
			err := c.TargetsManager.DeleteSpec(ctx, resourceId, namespace)
			if err != nil {
				tLog.Infof("V (Targets) : onRegistry Delete failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
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
	tLog.Infof("V (Targets) : onRegistry failed - method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *TargetsVendor) onRegistryList(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Solutions Vendor", request.Context, &map[string]string{
		"method": "onRegistryList",
	})
	defer span.End()
	tLog.Infof("V (Targets): onRegistryList, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = "default"
	}
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onRegistryList-GET", pCtx, nil)

		var err error
		var state interface{}
		// Change namespace back to empty to indicate ListSpec need to query all namespaces
		if !exist {
			namespace = ""
		}
		state, err = c.TargetsManager.ListState(ctx, namespace)

		if err != nil {
			tLog.Infof("V (Targets): onRegistryList failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		jData, _ := utils.FormatObject(state, true, request.Parameters["path"], request.Parameters["doc-type"])
		resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        jData,
			ContentType: "application/json",
		})
		if request.Parameters["doc-type"] == "yaml" {
			resp.ContentType = "application/text"
		}
		return resp
	}

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
	tLog.Infof("V (Targets) : onBootstrap, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())
	switch request.Method {
	case fasthttp.MethodPost:
		var authRequest AuthRequest
		err := json.Unmarshal(request.Body, &authRequest)
		if err != nil || authRequest.UserName != "symphony-test" {
			tLog.Infof("V (Targets) : onBootstrap failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
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

		resp := v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte(`{"accessToken":"` + ss + `", "tokenType": "Bearer"}`),
			ContentType: "application/json",
		}

		observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
		return resp
	}
	tLog.Infof("V (Targets) : onRegistry failed - method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
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
	tLog.Infof("V (Targets) : onStatus, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	version := request.Parameters["__version"]
	rootResource := request.Parameters["__name"]
	id := rootResource + "-" + version
	tLog.Infof("V (Targets): onCampaigns, id: %s, version: %s", id, version)

	switch request.Method {
	case fasthttp.MethodPut:
		namespace, exist := request.Parameters["namespace"]
		if !exist {
			namespace = constants.DefaultScope
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
			ObjectMeta: model.ObjectMeta{
				Name:      id,
				Namespace: namespace,
			},
			Status: model.TargetStatus{
				Properties:   properties,
				LastModified: time.Now().UTC(),
			},
		})

		if err != nil {
			tLog.Infof("V (Targets) : onStatus failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		jData, _ := json.Marshal(state)
		resp := v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        jData,
			ContentType: "application/json",
		}
		observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
		return resp
	}
	tLog.Infof("V (Targets) : onStatus failed - method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
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
	tLog.Infof("V (Targets) : onDownload, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	version := request.Parameters["__version"]
	rootResource := request.Parameters["__name"]
	id := rootResource + "-" + version
	tLog.Infof("V (Targets): onDownload, id: %s, version: %s", id, version)

	switch request.Method {
	case fasthttp.MethodGet:
		namespace, exist := request.Parameters["namespace"]
		if !exist {
			namespace = constants.DefaultScope
		}
		state, err := c.TargetsManager.GetState(pCtx, id, namespace)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		jData, err := utils.FormatObject(state, false, request.Parameters["path"], request.Parameters["__doc-type"])
		if err != nil {
			tLog.Infof("V (Targets) : onDownload failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
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

		observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
		return resp
	}
	tLog.Infof("V (Targets) : onDownload failed - method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *TargetsVendor) onHeartBeat(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Targets Vendor", request.Context, &map[string]string{
		"method": "onHeartBeat",
	})
	defer span.End()
	tLog.Infof("V (Targets) : onHeartBeat, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	version := request.Parameters["__version"]
	rootResource := request.Parameters["__name"]
	id := rootResource + "-" + version

	switch request.Method {
	case fasthttp.MethodPost:
		namespace, exist := request.Parameters["namespace"]
		if !exist {
			namespace = constants.DefaultScope
		}
		_, err := c.TargetsManager.ReportState(pCtx, model.TargetState{
			ObjectMeta: model.ObjectMeta{
				Name:      id,
				Namespace: namespace,
			},
			Status: model.TargetStatus{
				LastModified: time.Now().UTC(),
			},
		})

		if err != nil {
			tLog.Infof("V (Targets) : onHeartBeat failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		resp := v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte(`{}`),
			ContentType: "application/json",
		}
		observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
		return resp
	}
	tLog.Infof("V (Targets) : onHeartBeat failed - method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
