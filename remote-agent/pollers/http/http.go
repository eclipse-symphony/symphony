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
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
)

var (
	ShouldEnd      string        = "false"
	ConcurrentJobs int           = 3
	Interval       time.Duration = 10 * time.Second
)

// HttpPoller provides service endpoints using the standard net/http client
type HttpPoller struct {
	CertProvider certs.ICertProvider
	Agent        agent.Agent
	ResponseUrl  string
	RequestUrl   string
	Client       *http.Client
	Target       string
	Namespace    string
}

// Launch the polling agent
func (h *HttpPoller) Launch() error {
	// Start the agent by handling starter requests
	var wg sync.WaitGroup
	pollingUri := fmt.Sprintf("%s?target=%s&namespace=%s&getAll=%s&preindex=%s", h.RequestUrl, h.Target, h.Namespace, "true", "0")
	var requests []map[string]interface{}
	for {
		var allRequests model.ProviderPagingRequest
		resp, err := h.Client.Get(pollingUri)
		if err != nil {
			fmt.Printf("Quitting the agent since polling failed for %s", err.Error())
			os.Exit(0)
		}
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		err = json.Unmarshal(body, &allRequests)
		requests = append(requests, allRequests.RequestList...)
		if allRequests.LastMessageID == "" {
			break
		}
		pollingUri = fmt.Sprintf("%s?target=%s&namespace=%s&getAll=%s&preindex=%s", h.RequestUrl, h.Target, h.Namespace, "true", allRequests.LastMessageID)
	}
	fmt.Printf("Found starter jobs: %d. \n", len(requests))

	for _, request := range requests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			correlationId, ok := request[contexts.ConstructHttpHeaderKeyForActivityLogContext(contexts.Activity_CorrelationId)].(string)
			if !ok {
				fmt.Println("error: correlationId not found or not a string. Using a mock one.")
				correlationId = "00000000-0000-0000-0000-000000000000"
			}
			retCtx := context.TODO()
			retCtx = context.WithValue(retCtx, contexts.Activity_CorrelationId, correlationId)

			body, err := json.Marshal(request)
			if err != nil {
				fmt.Println("error marshalling request:", err)
				return
			}
			ret := h.Agent.Handle(body, retCtx)
			ret.Namespace = h.Namespace

			// Send response back - Mock
			result, err := json.Marshal(ret)
			if err != nil {
				fmt.Println("error marshalling response:", err)
			}
			fmt.Println("Agent response:", string(result))

			responseHost := fmt.Sprintf("%s?target=%s&namespace=%s", h.ResponseUrl, h.Target, h.Namespace)
			responseUri, err := url.Parse(responseHost)
			respRet, err := h.Client.Do(&http.Request{
				URL:    responseUri,
				Method: "POST",
				Body:   io.NopCloser(strings.NewReader(string(result))),
			})
			if err != nil {
				fmt.Println("error sending response:", err)
			} else {
				fmt.Println("response status:", respRet.Status)
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("All starter requests processed. Starting polling agent.")
	time.Sleep(Interval)

	go func() {
		for ShouldEnd == "false" {
			// Mock to read request from file - Mock
			//file, err := os.ReadFile("./samples/request.json")
			// - Mock
			//body := file

			// Poll requests
			requests = h.pollRequests()
			fmt.Printf("Found jobs: %d. \n", len(requests))
			for _, req := range requests {
				wg.Add(1)
				// handle request
				go func() {
					defer wg.Done()
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
					ret.Namespace = h.Namespace
					result, err := json.Marshal(ret)
					if err != nil {
						fmt.Println("error marshalling response:", err)
					}
					fmt.Println("Agent response:", string(result))

					responseHost := fmt.Sprintf("%s?target=%s&namespace=%s", h.ResponseUrl, h.Target, h.Namespace)
					responseUri, err := url.Parse(responseHost)
					respRet, err := h.Client.Do(&http.Request{
						URL:    responseUri,
						Method: "POST",
						Body:   io.NopCloser(strings.NewReader(string(result))),
					})
					if err != nil {
						fmt.Println("error sending response:", err)
					} else {
						fmt.Println("response status:", respRet.Status)
					}
				}()
			}
			// Wait for all goroutines to finish
			wg.Wait()
			// Sleep for a while before polling again
			time.Sleep(Interval)
		}
	}()

	return nil
}

func (h *HttpPoller) pollRequests() []map[string]interface{} {
	requests := []map[string]interface{}{}
	pollingUri := fmt.Sprintf("%s?target=%s&namespace=%s", h.RequestUrl, h.Target, h.Namespace)

	for i := 0; i < ConcurrentJobs; i++ {
		resp, err := h.Client.Get(pollingUri)
		if err != nil {
			return requests
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return requests
		}
		defer resp.Body.Close()
		if body == nil {
			return requests
		}

		// Parse as unified paging format
		var pagingResponse model.ProviderPagingRequest
		err = json.Unmarshal(body, &pagingResponse)
		if err != nil {
			return requests
		}

		// Extract requests from requestList
		for _, req := range pagingResponse.RequestList {
			if _, ok := req["operationID"].(string); ok {
				requests = append(requests, req)
			}
		}
	}
	return requests
}
