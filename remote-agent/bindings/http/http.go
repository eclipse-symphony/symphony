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
	"io/ioutil"
	"net/url"
	"time"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	autogen "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/autogen"
	localfile "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/localfile"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
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
	Agent        agent.Agent
	ResponseUrl  *url.URL
	RequestUrl   *url.URL
}

// Launch the polling agent
func (h *HttpBinding) Launch(config HttpBindingConfig) error {
	//handler := h.useRouter(endpoints)
	var err error

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
		// poll http response from a url - Mock
		//httpclient := &http.Client{}
		for {
			//This is the correct logic - Mock
			// resp, err := httpclient.Get(h.RequestUrl.Host)
			// if err != nil {
			// 	fmt.Println("error:", err)
			// 	time.Sleep(5 * time.Second) // Retry after a delay
			// 	continue
			// }

			// Mock to read request from file - Mock
			file, err := ioutil.ReadFile("./samples/request.json")

			// read response body - Mock
			// body, err := io.ReadAll(resp.Body)

			// - Mock
			body := file

			if err != nil {
				fmt.Println("error reading body:", err)
			} else {
				fmt.Println("response body:", string(body))
			}

			// close response body - Mock
			// resp.Body.Close()

			requests := make([]map[string]interface{}, 0)
			err = json.Unmarshal(body, &requests)

			for _, req := range requests {
				// handle request
				go func() {
					// TODO Ack the requests

					correlationId, ok := req[contexts.ConstructHttpHeaderKeyForActivityLogContext(contexts.Activity_CorrelationId)].(string)
					if !ok {
						fmt.Println("error: correlationId not found or not a string")
						correlationId = "00000000-0000-0000-0000-000000000000"
					}
					retCtx := context.TODO()
					retCtx = context.WithValue(retCtx, contexts.Activity_CorrelationId, correlationId)

					body, err := json.Marshal(req)
					if err != nil {
						fmt.Println("error marshalling request:", err)
						return
					}
					ret := h.Agent.Handle(body, retCtx)
					fmt.Println("Agent response:", string(ret.Body))

					// Send response back - Mock
					// respBody, err := json.Marshal(ret)
					// if err != nil {
					// 	fmt.Println("error marshalling response:", err)
					// }
					// respRet, err := httpclient.Do(&http.Request{
					// 	URL:    h.ResponseUrl,
					// 	Method: "POST",
					// 	Body:   io.NopCloser(strings.NewReader(string(respBody))),
					// })
					// if err != nil {
					// 	fmt.Println("error sending response:", err)
					// } else {
					// 	fmt.Println("response status:", respRet.Status)
					// }
				}()
			}

			// Sleep for a while before polling again
			time.Sleep(15 * time.Second)
		}

	}()
	return nil
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
