/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var mrLog = logger.NewLogger("coa.runtime")

// ModelEndpoint describes a single OpenAI-compatible backend endpoint the
// router can forward requests to.
type ModelEndpoint struct {
	// Name is a unique identifier used to select this endpoint.
	Name string `json:"name"`
	// URL is the base URL of the OpenAI-compatible endpoint (e.g. https://api.openai.com).
	URL string `json:"url"`
	// Key is the API key used to authenticate against the endpoint.
	Key string `json:"key,omitempty"`
}

// ModelRouterVendor routes OpenAI-compatible requests to one of the configured
// backend endpoints. Configuration is supplied through vendor properties:
//   - "endpoints": a JSON array of ModelEndpoint objects.
//   - "defaultEndpoint": (optional) the name of the endpoint to use when the
//     request does not explicitly select one.
//
// Dynamic routing conditions (e.g. model-based or content-based selection) will
// be layered on top of this basic routing later.
type ModelRouterVendor struct {
	vendors.Vendor
	Endpoints       map[string]ModelEndpoint
	DefaultEndpoint string
}

func (o *ModelRouterVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "ModelRouter",
		Producer: "Microsoft",
	}
}

func (e *ModelRouterVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}

	e.Endpoints = make(map[string]ModelEndpoint)
	if config.Properties != nil {
		if raw, ok := config.Properties["endpoints"]; ok && strings.TrimSpace(raw) != "" {
			var endpoints []ModelEndpoint
			if err := json.Unmarshal([]byte(raw), &endpoints); err != nil {
				return v1alpha2.NewCOAError(err, "failed to parse model router endpoints", v1alpha2.BadConfig)
			}
			for _, ep := range endpoints {
				// Resolve expressions such as $env:OPENAI_API_KEY against the
				// process environment.
				ep.Name = coa_utils.ParseProperty(ep.Name)
				ep.URL = coa_utils.ParseProperty(ep.URL)
				ep.Key = coa_utils.ParseProperty(ep.Key)
				if ep.Name == "" {
					return v1alpha2.NewCOAError(nil, "model router endpoint is missing a name", v1alpha2.BadConfig)
				}
				if ep.URL == "" {
					return v1alpha2.NewCOAError(nil, fmt.Sprintf("model router endpoint '%s' is missing a url", ep.Name), v1alpha2.BadConfig)
				}
				e.Endpoints[ep.Name] = ep
			}
		}
		e.DefaultEndpoint = config.Properties["defaultEndpoint"]
	}

	if e.DefaultEndpoint == "" && len(e.Endpoints) == 1 {
		for name := range e.Endpoints {
			e.DefaultEndpoint = name
		}
	}

	return nil
}

func (o *ModelRouterVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "modelrouter"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/chat/completions",
			Version: o.Version,
			Handler: o.onProxy("/v1/chat/completions"),
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/completions",
			Version: o.Version,
			Handler: o.onProxy("/v1/completions"),
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/embeddings",
			Version: o.Version,
			Handler: o.onProxy("/v1/embeddings"),
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/models",
			Version: o.Version,
			Handler: o.onProxy("/v1/models"),
		},
	}
}

// onProxy returns a handler that forwards the incoming request to the selected
// backend endpoint, appending the given OpenAI-compatible path.
func (c *ModelRouterVendor) onProxy(openAIPath string) v1alpha2.COAHandler {
	return func(request v1alpha2.COARequest) v1alpha2.COAResponse {
		pCtx, span := observability.StartSpan("ModelRouter Vendor", request.Context, &map[string]string{
			"method": "onProxy",
			"path":   openAIPath,
		})
		defer span.End()
		mrLog.InfofCtx(pCtx, "V (ModelRouter): onProxy, method: %s, path: %s", request.Method, openAIPath)

		endpoint, err := c.resolveEndpoint(request)
		if err != nil {
			mrLog.ErrorfCtx(pCtx, "V (ModelRouter): failed to resolve endpoint, err: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte(err.Error()),
				ContentType: "text/plain",
			})
		}

		targetURL := strings.TrimRight(endpoint.URL, "/") + openAIPath

		contentType := request.ContentType
		if contentType == "" {
			contentType = "application/json"
		}

		req, err := http.NewRequestWithContext(pCtx, request.Method, targetURL, bytes.NewReader(request.Body))
		if err != nil {
			mrLog.ErrorfCtx(pCtx, "V (ModelRouter): failed to build request, err: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.InternalError,
				Body:        []byte(err.Error()),
				ContentType: "text/plain",
			})
		}
		req.Header.Set("Content-Type", contentType)
		if endpoint.Key != "" {
			req.Header.Set("Authorization", "Bearer "+endpoint.Key)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			mrLog.ErrorfCtx(pCtx, "V (ModelRouter): request to endpoint '%s' failed, err: %v", endpoint.Name, err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.InternalError,
				Body:        []byte(err.Error()),
				ContentType: "text/plain",
			})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			mrLog.ErrorfCtx(pCtx, "V (ModelRouter): failed to read response from endpoint '%s', err: %v", endpoint.Name, err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.InternalError,
				Body:        []byte(err.Error()),
				ContentType: "text/plain",
			})
		}

		respContentType := resp.Header.Get("Content-Type")
		if respContentType == "" {
			respContentType = "application/json"
		}

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.State(resp.StatusCode),
			Body:        body,
			ContentType: respContentType,
		})
	}
}

// resolveEndpoint selects the backend endpoint for a request. An explicit
// endpoint can be requested through the "endpoint" query parameter; otherwise
// the configured default endpoint is used.
func (c *ModelRouterVendor) resolveEndpoint(request v1alpha2.COARequest) (ModelEndpoint, error) {
	if len(c.Endpoints) == 0 {
		return ModelEndpoint{}, v1alpha2.NewCOAError(nil, "no model router endpoints are configured", v1alpha2.BadConfig)
	}

	name := request.Parameters["endpoint"]
	if name == "" {
		name = c.DefaultEndpoint
	}
	if name == "" {
		return ModelEndpoint{}, v1alpha2.NewCOAError(nil, "no endpoint specified and no default endpoint configured", v1alpha2.BadConfig)
	}

	endpoint, ok := c.Endpoints[name]
	if !ok {
		return ModelEndpoint{}, v1alpha2.NewCOAError(nil, fmt.Sprintf("model router endpoint '%s' is not configured", name), v1alpha2.BadConfig)
	}
	return endpoint, nil
}
