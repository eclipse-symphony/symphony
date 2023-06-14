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
	"encoding/json"
	"strings"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/targets"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
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
	tLog.Info(" M (Targets) : onRegistry")

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onRegistry-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			state, err = c.TargetsManager.ListSpec(ctx)
			isArray = true
		} else {
			state, err = c.TargetsManager.GetSpec(ctx, id)
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
		err = c.TargetsManager.UpsertSpec(ctx, id, target)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if c.Config.Properties["useJobManager"] == "true" {
			c.Context.Publish("job", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": "target",
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
			err := c.TargetsManager.DeleteSpec(ctx, id)
			if err != nil {
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
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	span.End()
	return resp
}

func (c *TargetsVendor) onBootstrap(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Targets Vendor", request.Context, nil)
	var authRequest AuthRequest
	err := json.Unmarshal(request.Body, &authRequest)
	if err != nil || authRequest.UserName != "symphony-test" {
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
	span.End()
	return resp
}

func (c *TargetsVendor) onStatus(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Targets Vendor", request.Context, nil)

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

	state, err := c.TargetsManager.ReportState(request.Context, model.TargetState{
		Id: request.Parameters["__name"],
		Metadata: map[string]string{
			"version":  "v1",
			"group":    "symphony.microsoft.com",
			"resource": "targets",
		},
		Status: properties,
	})

	if err != nil {
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
	span.End()
	return resp
}

func (c *TargetsVendor) onDownload(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Targets Vendor", request.Context, nil)

	state, err := c.TargetsManager.GetSpec(request.Context, request.Parameters["__name"])
	if err != nil {
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.InternalError,
			Body:  []byte(err.Error()),
		})
	}
	jData, err := utils.FormatObject(state, false, request.Parameters["path"], request.Parameters["__doc-type"])
	if err != nil {
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
	span.End()
	return resp
}

func (c *TargetsVendor) onHeartBeat(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Targets Vendor", request.Context, nil)

	_, err := c.TargetsManager.ReportState(request.Context, model.TargetState{
		Id: request.Parameters["__name"],
		Metadata: map[string]string{
			"version":  "v1",
			"group":    "symphony.microsoft.com",
			"resource": "targets",
		},
		Status: map[string]string{
			"ping": time.Now().UTC().String(),
		},
	})

	if err != nil {
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
	span.End()
	return resp
}
