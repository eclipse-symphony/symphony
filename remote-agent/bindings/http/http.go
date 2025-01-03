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
	"net/http"
	"os"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
)

var ShouldEnd string = "false"

// HttpBinding provides service endpoints as a fasthttp web server
type HttpBinding struct {
	CertProvider certs.ICertProvider
	Agent        agent.Agent
	ResponseUrl  string
	RequestUrl   string
	Client       *http.Client
	Target       string
	Namespace    string
}

// Launch the polling agent
func (h *HttpBinding) Launch() error {
	//handler := h.useRouter(endpoints)
	var err error

	if err != nil {
		return err
	}

	go func() {
		for {
			//This is the correct logic - Mock
			// pollingUri := fmt.Sprintf("%s?target=%s&namespace=%s", h.RequestUrl, h.Target, h.Namespace)
			// resp, err := h.Client.Get(pollingUri)
			// if err != nil {
			// 	fmt.Println("error:", err)
			// 	time.Sleep(5 * time.Second) // Retry after a delay
			// 	continue
			// }

			// Mock to read request from file - Mock
			file, err := os.ReadFile("./samples/request.json")

			// read response body - Mock
			//body, err := io.ReadAll(resp.Body)

			// - Mock
			body := file

			if err != nil {
				fmt.Println("error reading body:", err)
			} else {
				fmt.Println("response body:", string(body))
			}

			// close response body - Mock
			//defer resp.Body.Close()

			requests := make([]map[string]interface{}, 0)
			err = json.Unmarshal(body, &requests)

			for _, req := range requests {
				// handle request
				go func() {
					// TODO Ack the requests

					correlationId, ok := req[contexts.ConstructHttpHeaderKeyForActivityLogContext(contexts.Activity_CorrelationId)].(string)
					if !ok {
						fmt.Println("error: correlationId not found or not a string. Using a mock one.")
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
					result, err := json.Marshal(ret)
					if err != nil {
						fmt.Println("error marshalling response:", err)
					}
					fmt.Println("Agent response:", string(result))

					// Send response back - Mock
					// respBody, err := json.Marshal(ret)
					// if err != nil {
					// 	fmt.Println("error marshalling response:", err)
					// }
					// responseHost := fmt.Sprintf("%s?target=%s&namespace=%s", h.ResponseUrl, h.Target, h.Namespace)
					// responseUri, err := url.Parse(responseHost)
					// respRet, err := h.Client.Do(&http.Request{
					// 	URL:    responseUri,
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

	go func() {
		for {
			if ShouldEnd == "true" {
				os.Exit(0)
			}
			time.Sleep(30 * time.Second)
		}
	}()
	return nil
}
