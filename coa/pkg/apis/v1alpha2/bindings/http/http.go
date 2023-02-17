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

package http

import (
	"encoding/json"
	"fmt"
	"strings"

	v1alpha2 "github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	autogen "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/certs/autogen"
	localfile "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/certs/localfile"
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
}

// Launch fasthttp server
func (h *HttpBinding) Launch(config HttpBindingConfig, endpoints []v1alpha2.Endpoint) error {
	handler := h.useRouter(endpoints)

	pipeline, err := BuildPipeline(config)
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

	go func() {
		if config.TLS {
			cert, key, _ := h.CertProvider.GetCert("localhost") //TODO: user proper host/DNS name
			fasthttp.ListenAndServeTLSEmbed(fmt.Sprintf(":%d", config.Port), cert, key, pipeline.Apply(handler))
		} else {
			fasthttp.ListenAndServe(fmt.Sprintf(":%d", config.Port), pipeline.Apply(handler))
		}
	}()
	return nil
}

func (h *HttpBinding) useRouter(endpoints []v1alpha2.Endpoint) fasthttp.RequestHandler {
	router := h.getRouter(endpoints)
	return router.Handler
}
func (h *HttpBinding) getRouter(endpoints []v1alpha2.Endpoint) *routing.Router {
	router := routing.New()
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

func wrapAsHTTPHandler(endpoint v1alpha2.Endpoint, handler v1alpha2.COAHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {
		req := v1alpha2.COARequest{
			Body:    reqCtx.PostBody(),
			Route:   string(reqCtx.Request.URI().Path()),
			Method:  string(reqCtx.Method()),
			Context: reqCtx,
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
				req.Parameters[k] = ""
			} else {
				req.Parameters[k] = v.(string)
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
			reqCtx.SetStatusCode(int(resp.State))
		}
	}
}
