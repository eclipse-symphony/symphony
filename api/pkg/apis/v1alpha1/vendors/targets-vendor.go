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
			Parameters: []string{"name?"},
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
	tLog.InfofCtx(pCtx, "V (Targets) : onRegistry, method: %s", request.Method)

	id := request.Parameters["__name"]
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onRegistry-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change namespace back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				namespace = ""
			}
			state, err = c.TargetsManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.TargetsManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			tLog.ErrorfCtx(ctx, "V (Targets) : onRegistry failed - %s", err.Error())
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
		binding := request.Parameters["with-binding"]
		var target model.TargetState
		err := json.Unmarshal(request.Body, &target)
		if err != nil {
			tLog.ErrorfCtx(ctx, "V (Targets) : onRegistry failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if target.ObjectMeta.Name == "" {
			target.ObjectMeta.Name = id
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
				tLog.ErrorCtx(ctx, "V (Targets) : onRegistry failed - invalid binding")
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.BadRequest,
					Body:  []byte("invalid binding, supported is: 'staging'"),
				})
			}
		}
		err = c.TargetsManager.UpsertState(ctx, id, target)
		if err != nil {
			tLog.ErrorfCtx(ctx, "V (Targets) : onRegistry failed - %s", err.Error())
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
					Id:     id,
					Action: v1alpha2.JobUpdate,
					Scope:  namespace,
				},
				Context: ctx,
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
					Id:     id,
					Action: v1alpha2.JobDelete,
					Scope:  namespace,
				},
				Context: ctx,
			})
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.OK,
			})
		} else {
			err := c.TargetsManager.DeleteSpec(ctx, id, namespace)
			if err != nil {
				tLog.ErrorfCtx(ctx, "V (Targets) : onRegistry failed - %s", err.Error())
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
	tLog.ErrorCtx(pCtx, "V (Targets) : onRegistry failed - method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *TargetsVendor) onBootstrap(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Targets Vendor", request.Context, &map[string]string{
		"method": "onBootstrap",
	})
	defer span.End()
	tLog.InfofCtx(ctx, "V (Targets) : onBootstrap, method: %s", request.Method)
	switch request.Method {
	case fasthttp.MethodPost:
		var authRequest AuthRequest
		err := json.Unmarshal(request.Body, &authRequest)
		if err != nil || authRequest.UserName != "symphony-test" {
			tLog.ErrorfCtx(ctx, "V (Targets) : onBootstrap failed - %s", err.Error())
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
	tLog.ErrorCtx(ctx, "V (Targets) : onRegistry failed - method not allowed")
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
	tLog.InfofCtx(pCtx, "V (Targets) : onStatus, method: %s", request.Method)

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
				Name:      request.Parameters["__name"],
				Namespace: namespace,
			},
			Status: model.TargetStatus{
				Properties:   properties,
				LastModified: time.Now().UTC(),
			},
		})

		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Targets) : onStatus failed - %s", err.Error())
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
	tLog.ErrorCtx(pCtx, "V (Targets) : onStatus failed - method not allowed")
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
	tLog.InfofCtx(pCtx, "V (Targets) : onDownload, method: %s", request.Method)

	switch request.Method {
	case fasthttp.MethodGet:
		namespace, exist := request.Parameters["namespace"]
		if !exist {
			namespace = constants.DefaultScope
		}
		state, err := c.TargetsManager.GetState(pCtx, request.Parameters["__name"], namespace)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		jData, err := utils.FormatObject(state, false, request.Parameters["path"], request.Parameters["__doc-type"])
		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Targets) : onDownload failed - %s", err.Error())
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
	tLog.ErrorCtx(pCtx, "V (Targets) : onDownload failed - method not allowed")
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
	tLog.InfofCtx(pCtx, "V (Targets) : onHeartBeat, method: %s", request.Method)

	switch request.Method {
	case fasthttp.MethodPost:
		namespace, exist := request.Parameters["namespace"]
		if !exist {
			namespace = constants.DefaultScope
		}
		_, err := c.TargetsManager.ReportState(pCtx, model.TargetState{
			ObjectMeta: model.ObjectMeta{
				Name:      request.Parameters["__name"],
				Namespace: namespace,
			},
			Status: model.TargetStatus{
				LastModified: time.Now().UTC(),
			},
		})

		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Targets) : onHeartBeat failed - %s", err.Error())
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
	tLog.ErrorCtx(pCtx, "V (Targets) : onHeartBeat failed - method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
