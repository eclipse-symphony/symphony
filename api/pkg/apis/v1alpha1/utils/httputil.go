/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

// DoHTTPRequest executes an HTTP request with status-code checking and
// optional retry on transient failures (network errors and 5xx).
// It always closes the response body and returns the body bytes on success.
//
// client may be nil, in which case http.DefaultClient is used.
// maxRetries is the number of additional attempts after the first failure;
// 0 means the request is only tried once. Retries use exponential backoff.
// The operation string is included in returned error messages to help
// locate the failing call site.
func DoHTTPRequest(client *http.Client, req *http.Request, maxRetries int, operation string) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 2 * time.Second // Initial retry interval.
	b.MaxInterval = 30 * time.Second    // Maximum retry interval.

	var bodyBytes []byte
	attempt := 0
	err := backoff.Retry(func() error {
		// Replay the request body on retries. Requests built with
		// http.NewRequest and a bytes/strings reader get GetBody for free.
		if attempt > 0 && req.Body != nil {
			if req.GetBody == nil {
				// The body was consumed by the first attempt and cannot be
				// replayed, so retrying is pointless.
				return backoff.Permanent(v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: request body cannot be replayed for retry", operation), v1alpha2.InternalError))
			}
			body, bErr := req.GetBody()
			if bErr != nil {
				return backoff.Permanent(v1alpha2.NewCOAError(bErr, fmt.Sprintf("%s: failed to replay request body", operation), v1alpha2.InternalError))
			}
			req.Body = body
		}
		attempt++

		resp, dErr := client.Do(req)
		if dErr != nil {
			return maybePermanent(dErr, attempt, maxRetries)
		}
		defer resp.Body.Close()

		bodyBytes, dErr = io.ReadAll(resp.Body)
		if dErr != nil {
			return maybePermanent(dErr, attempt, maxRetries)
		}

		if resp.StatusCode >= 500 {
			// 5xx is transient, retry if attempts remain
			return maybePermanent(httpError(resp.StatusCode, bodyBytes, operation), attempt, maxRetries)
		}
		if resp.StatusCode >= 400 {
			// 4xx is not retriable
			return backoff.Permanent(httpError(resp.StatusCode, bodyBytes, operation))
		}
		return nil
	}, b)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

// maybePermanent wraps err as permanent once all retries are exhausted, so
// backoff.Retry stops and returns the error to the caller.
func maybePermanent(err error, attempt int, maxRetries int) error {
	if attempt > maxRetries {
		return backoff.Permanent(err)
	}
	return err
}

// httpError builds a COAError from an HTTP status code, carrying the
// operation name and response body in the message.
func httpError(statusCode int, body []byte, operation string) error {
	coaErr := v1alpha2.FromHTTPResponseCode(statusCode, body)
	return v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: http status %d: %s", operation, statusCode, coaErr.Message), coaErr.State)
}
