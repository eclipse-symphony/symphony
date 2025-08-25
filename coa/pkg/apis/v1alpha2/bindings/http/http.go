/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
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

var (
	ClientCAFile = os.Getenv("CLIENT_CA_FILE")
)

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
	var err error
	h.pipeline, err = BuildPipeline(config, pubsubProvider)

	if err != nil {
		return err
	}

	caCertPool := x509.NewCertPool()

	if config.TLS {
		switch config.CertProvider.Type {
		case "certs.autogen":
			h.CertProvider = &autogen.AutoGenCertProvider{}
		case "certs.localfile":
			h.CertProvider = &localfile.LocalCertFileProvider{}
			localConfig := &localfile.LocalCertFileProviderConfig{}
			data, err := json.Marshal(config.CertProvider.Config)
			if err != nil {
				log.Errorf("B (HTTP): failed to marshall config %+v", err)
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("B (HTTP): failed to marshall config"), v1alpha2.BadConfig)
			}
			err = json.Unmarshal(data, &localConfig)
			if err != nil {
				log.Errorf("B (HTTP): failed to unmarshall config %+v", err)
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("B (HTTP): failed to unmarshall config"), v1alpha2.BadConfig)
			}
			certFile, err := os.Open(localConfig.CertFile)
			if err != nil {
				log.Errorf("B (HTTP): failed to open certificate file %+v", err)
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("B (HTTP): failed to open certificate file %+v", err), v1alpha2.BadConfig)
			}
			certData, err := io.ReadAll(certFile)
			if err != nil {
				log.Errorf("B (HTTP): failed to read certificate file %+v", err)
				return v1alpha2.NewCOAError(err, "B (HTTP): failed to read certificate file", v1alpha2.InternalError)
			}
			certs, err := h.parseCertificates(certData)
			if err != nil {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("Failed to parse the symphony CA file, %s", localConfig.CertFile), v1alpha2.BadConfig)
			}
			for _, cert := range certs {
				caCertPool.AddCert(cert)
			}
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("cert provider type '%s' is not recognized", config.CertProvider.Type), v1alpha2.BadConfig)
		}
		err = h.CertProvider.Init(config.CertProvider.Config)
		if err != nil {
			return err
		}
	}

	// Load the PEM file
	if ClientCAFile != "" {
		pemData, err := os.ReadFile(ClientCAFile)
		if err != nil {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Client cert file '%s' is not read successfully", ClientCAFile), v1alpha2.BadConfig)
		}

		// Parse the certificates
		certs, err := h.parseCertificates(pemData)
		if err != nil {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Failed to parse the client cert file, %s", ClientCAFile), v1alpha2.BadConfig)
		}
		for _, cert := range certs {
			caCertPool.AddCert(cert)
		}
	}
	fs := &fasthttp.FS{
		Root:               "/", // Directory to serve files from
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: true, // Default file names to look for
		Compress:           true, // Generate directory listing if index file is not found
	}
	fileServerHandler := fs.NewRequestHandler()

	handler := h.useRouter(endpoints, fileServerHandler)

	h.server = &fasthttp.Server{
		Handler: h.pipeline.Apply(handler),
		TLSConfig: &tls.Config{
			ClientAuth: tls.VerifyClientCertIfGiven,
			ClientCAs:  caCertPool,
		},
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

func (h *HttpBinding) parseCertificates(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	var block *pem.Block
	var rest = pemData

	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// Shutdown fasthttp server
func (h *HttpBinding) Shutdown(ctx context.Context) error {
	if err := h.pipeline.Shutdown(ctx); err != nil {
		return err
	}
	return h.server.ShutdownWithContext(ctx)
}

func (h *HttpBinding) useRouter(endpoints []v1alpha2.Endpoint, fileHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
	router := h.getRouter(endpoints, fileHandler)
	return router.Handler
}
func (h *HttpBinding) getRouter(endpoints []v1alpha2.Endpoint, fileHandler fasthttp.RequestHandler) *routing.Router {
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
	router.Handle("GET", "/v1alpha2/files/{filepath:*}", fileHandler)
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
