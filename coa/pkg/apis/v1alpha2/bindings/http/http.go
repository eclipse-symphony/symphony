/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	autogen "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/autogen"
	localfile "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/localfile"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	routing "github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// MiddlewareConfig configures a HTTP middleware.
type MiddlewareConfig struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

type CertProviderConfig struct {
	Type   string                    `json:"type"`
	Config providers.IProviderConfig `json:"config"`
}

// HttpBindingConfig configures a HttpBinding.
type HttpBindingConfig struct {
	Port         int                `json:"port"`
	Pipeline     []MiddlewareConfig `json:"pipeline"`
	TLS          bool               `json:"tls"`
	CertProvider CertProviderConfig `json:"certProvider"`
}

// HttpBinding provides service endpoints as a fasthttp web server
type HttpBinding struct {
	CertProvider certs.ICertProvider
	server       *fasthttp.Server
	pipeline     Pipeline
}

// Launch fasthttp server
func (h *HttpBinding) Launch(config HttpBindingConfig, endpoints []v1alpha2.Endpoint, pubsubProvider pubsub.IPubSubProvider) error {
	handler := h.useRouter(endpoints)
	var err error
	h.pipeline, err = BuildPipeline(config, pubsubProvider)

	if err != nil {
		return err
	}

	if config.TLS {
		switch config.CertProvider.Type {
		case "certs.autogen":
			h.CertProvider = &autogen.AutoGenCertProvider{}
		case "certs.localfile":
			h.CertProvider = &localfile.LocalCertFileProvider{}
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("cert provider type '%s' is not recognized", config.CertProvider.Type), v1alpha2.BadConfig)
		}
		err = h.CertProvider.Init(config.CertProvider.Config)
		if err != nil {
			return err
		}
	}

	h.server = &fasthttp.Server{
		Handler: h.pipeline.Apply(handler),
	}

	go func() {
		if config.TLS {
			cert, key, _ := h.CertProvider.GetCert("localhost") //TODO: user proper host/DNS name
			h.server.ListenAndServeTLSEmbed(fmt.Sprintf(":%d", config.Port), cert, key)
		} else {
			h.server.ListenAndServe(fmt.Sprintf(":%d", config.Port))
		}
	}()
	return nil
}

// Shutdown fasthttp server
func (h *HttpBinding) Shutdown(ctx context.Context) error {
	if err := h.pipeline.Shutdown(ctx); err != nil {
		return err
	}
	return h.server.ShutdownWithContext(ctx)
}

func (h *HttpBinding) useRouter(endpoints []v1alpha2.Endpoint) fasthttp.RequestHandler {
	router := h.getRouter(endpoints)
	return router.Handler
}
func (h *HttpBinding) getRouter(endpoints []v1alpha2.Endpoint) *routing.Router {
	router := routing.New()
	router.SaveMatchedRoutePath = true
	for _, e := range endpoints {
		path := fmt.Sprintf("/%s/%s", e.Version, e.Route)
		for _, p := range e.Parameters {
			path += "/{" + p + "}"
		}
		for _, m := range e.Methods {
			router.Handle(m, path, wrapAsHTTPHandler(e, e.Handler))
		}
	}
	return router
}

func composeCOARequestContext(reqCtx *fasthttp.RequestCtx, actCtx *contexts.ActivityLogContext, diagCtx *contexts.DiagnosticLogContext) context.Context {
	retCtx := context.TODO()
	if reqCtx != nil {
		retCtx = context.WithValue(retCtx, v1alpha2.COAFastHTTPContextKey, reqCtx)
	}
	if actCtx != nil {
		retCtx = context.WithValue(retCtx, contexts.ActivityLogContextKey, actCtx)
	}
	if diagCtx != nil {
		retCtx = context.WithValue(retCtx, contexts.DiagnosticLogContextKey, diagCtx)
	}
	return retCtx
}

func wrapAsHTTPHandler(endpoint v1alpha2.Endpoint, handler v1alpha2.COAHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {
		actCtx := contexts.ParseActivityLogContextFromHttpRequestHeader(reqCtx)
		diagCtx := contexts.ParseDiagnosticLogContextFromHttpRequestHeader(reqCtx)
		ctx := composeCOARequestContext(reqCtx, actCtx, diagCtx)
		// patch correlation id if missing
		ctx = contexts.GenerateCorrelationIdToParentContextIfMissing(ctx)
		req := v1alpha2.COARequest{
			Body:    reqCtx.PostBody(),
			Route:   string(reqCtx.Request.URI().Path()),
			Method:  string(reqCtx.Method()),
			Context: ctx,
		}
		meta := reqCtx.Request.Header.Peek(v1alpha2.COAMetaHeader)
		if meta != nil {
			metaMap := make(map[string]string)
			json.Unmarshal(meta, &metaMap)
			req.Metadata = metaMap
		}
		req.Parameters = make(map[string]string)

		for _, p := range endpoint.Parameters {
			k := p
			if strings.HasSuffix(p, "?") {
				k = k[:len(p)-1]
			}
			v := reqCtx.UserValue(k)
			k = "__" + k
			if v == nil {
				req.Parameters[k] = "" //TODO: chance to report on missing required parameters
			} else {
				req.Parameters[k] = utils.FormatAsString(v)
			}
		}

		reqCtx.QueryArgs().VisitAll(func(key, value []byte) {
			req.Parameters[string(key)] = string(value)
		})

		resp := handler(req)

		if resp.State == v1alpha2.APIRedirect {
			reqCtx.Redirect(resp.RedirectUri, 308)
		} else {
			if len(resp.Metadata) != 0 {
				data, _ := json.Marshal(resp.Metadata)
				reqCtx.Response.Header.Set(v1alpha2.COAMetaHeader, string(data))
			}
			reqCtx.SetContentType(resp.ContentType)
			reqCtx.SetBody(resp.Body)
			reqCtx.SetStatusCode(toHttpState(resp.State))
		}
	}
}

func toHttpState(state v1alpha2.State) int {
	switch state {
	case v1alpha2.OK:
		return fasthttp.StatusOK
	case v1alpha2.Accepted:
		return fasthttp.StatusAccepted
	case v1alpha2.BadRequest:
		return fasthttp.StatusBadRequest
	case v1alpha2.Unauthorized:
		return fasthttp.StatusUnauthorized
	case v1alpha2.NotFound:
		return fasthttp.StatusNotFound
	case v1alpha2.MethodNotAllowed:
		return fasthttp.StatusMethodNotAllowed
	case v1alpha2.Conflict:
		return fasthttp.StatusConflict
	case v1alpha2.InternalError:
		return fasthttp.StatusInternalServerError
	default:
		return fasthttp.StatusInternalServerError
	}
}
