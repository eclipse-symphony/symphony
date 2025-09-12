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
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"

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
	RLog         logger.Logger
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
			h.RLog.Errorf("Quitting the agent since polling failed for %s", err.Error())
			return err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			h.RLog.Errorf("error reading response body: %v", err)
		}
		defer resp.Body.Close()
		err = json.Unmarshal(body, &allRequests)
		if err != nil {
			h.RLog.Errorf("error unmarshalling response body: %v", err)
		}
		requests = append(requests, allRequests.RequestList...)
		if allRequests.LastMessageID == "" {
			break
		}
		pollingUri = fmt.Sprintf("%s?target=%s&namespace=%s&getAll=%s&preindex=%s", h.RequestUrl, h.Target, h.Namespace, "true", allRequests.LastMessageID)
	}
	h.RLog.Infof("Found starter jobs: %d. \n", len(requests))

	for _, request := range requests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			correlationId, ok := request[contexts.ConstructHttpHeaderKeyForActivityLogContext(contexts.Activity_CorrelationId)].(string)
			if !ok {
				h.RLog.Warnf("Warning: correlationId not found or not a string. Using a mock one.")
				correlationId = "00000000-0000-0000-0000-000000000000"
			}
			retCtx := context.TODO()
			retCtx = context.WithValue(retCtx, contexts.Activity_CorrelationId, correlationId)

			body, err := json.Marshal(request)
			if err != nil {
				h.RLog.ErrorfCtx(retCtx, "error marshalling request: %v", err)
				return
			}
			ret := h.Agent.Handle(body, retCtx)
			ret.Namespace = h.Namespace
			h.RLog.InfofCtx(retCtx, "Agent response: %v", ret)
			logs, _, err := h.RLog.(*logger.FileLogger).GetLogsFromOffset(h.RLog.(*logger.FileLogger).GetLogOffset())
			if err != nil {
				h.RLog.ErrorfCtx(retCtx, "error getting logs: %v", err)
			}
			ret.Logs = logs

			// Send response back
			result, err := json.Marshal(ret)
			if err != nil {
				h.RLog.ErrorfCtx(retCtx, "error marshalling response: %v", err)
			}
			responseHost := fmt.Sprintf("%s?target=%s&namespace=%s", h.ResponseUrl, h.Target, h.Namespace)
			responseUri, err := url.Parse(responseHost)
			if err != nil {
				h.RLog.ErrorfCtx(retCtx, "error parsing response URL: %v", err)
				return
			}
			respRet, err := h.Client.Do(&http.Request{
				URL:    responseUri,
				Method: "POST",
				Body:   io.NopCloser(strings.NewReader(string(result))),
			})
			if err != nil {
				h.RLog.ErrorfCtx(retCtx, "error sending response: %v", err)
			} else {
				h.RLog.InfofCtx(retCtx, "response status: %s", respRet.Status)
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	h.RLog.Infof("All starter requests processed. Starting polling agent.")
	time.Sleep(Interval)

	go func() {
		for ShouldEnd == "false" {

			// Poll requests
			requests = h.pollRequests()
			h.RLog.Infof("Found jobs: %d. \n", len(requests))
			for _, req := range requests {
				wg.Add(1)
				// handle request
				go func() {
					defer wg.Done()
					correlationId, ok := req[contexts.ConstructHttpHeaderKeyForActivityLogContext(contexts.Activity_CorrelationId)].(string)
					if !ok {
						h.RLog.Warnf("Warning: correlationId not found or not a string. Using a mock one.")
						correlationId = "00000000-0000-0000-0000-000000000000"
					}
					retCtx := context.TODO()
					retCtx = context.WithValue(retCtx, contexts.Activity_CorrelationId, correlationId)

					body, err := json.Marshal(req)
					if err != nil {
						h.RLog.ErrorfCtx(retCtx, "error marshalling request: %v", err)
						return
					}
					ret := h.Agent.Handle(body, retCtx)
					ret.Namespace = h.Namespace
					h.RLog.InfofCtx(retCtx, "Agent response: %v", ret)

					logs, _, err := h.RLog.(*logger.FileLogger).GetLogsFromOffset(h.RLog.(*logger.FileLogger).GetLogOffset())
					if err != nil {
						h.RLog.ErrorfCtx(retCtx, "error getting logs: %v", err)
					}
					ret.Logs = logs
					result, err := json.Marshal(ret)

					if err != nil {
						h.RLog.ErrorfCtx(retCtx, "error marshalling response: %v", err)
					}

					responseHost := fmt.Sprintf("%s?target=%s&namespace=%s", h.ResponseUrl, h.Target, h.Namespace)
					responseUri, err := url.Parse(responseHost)
					if err != nil {
						h.RLog.ErrorfCtx(retCtx, "error parsing response URL: %v", err)
						return
					}
					respRet, err := h.Client.Do(&http.Request{
						URL:    responseUri,
						Method: "POST",
						Body:   io.NopCloser(strings.NewReader(string(result))),
					})
					if err != nil {
						h.RLog.ErrorfCtx(retCtx, "error sending response: %v", err)
					} else {
						h.RLog.InfofCtx(retCtx, "response status: %s", respRet.Status)
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
