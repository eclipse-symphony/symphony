/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	autogen "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/autogen"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type TestCustomClaims struct {
	User string `json:"user"`
	jwt.RegisteredClaims
}

type TestAuthRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func testHttpRequestHelper(context context.Context, t *testing.T, method string, url string, body []byte, expectedStatusCode int, expectedBody string) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context, method, url, bytes.NewBuffer(body))
	observ_utils.PropagateSpanContextToHttpRequestHeader(req)
	assert.Nil(t, err)
	resp, err := client.Do(req)
	assert.Nil(t, err)

	defer resp.Body.Close()
	assert.Equal(t, expectedStatusCode, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	if expectedBody != "" {
		assert.Equal(t, expectedBody, string(bodyBytes))
	}
}

func testHttpRequestHelperWithHeaders(context context.Context, t *testing.T, method string, url string, body []byte, headers map[string]string, expectedStatusCode int, expectedBody string, expectedResponseHeaders map[string]string) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context, method, url, bytes.NewBuffer(body))
	observ_utils.PropagateSpanContextToHttpRequestHeader(req)
	assert.Nil(t, err)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	assert.Nil(t, err)

	defer resp.Body.Close()
	assert.Equal(t, expectedStatusCode, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	if expectedBody != "" {
		assert.Equal(t, expectedBody, string(bodyBytes))
	}

	for k, v := range expectedResponseHeaders {
		assert.Equal(t, v, resp.Header.Get(k))
	}
}

func TestHTTPEcho(t *testing.T) {
	config := HttpBindingConfig{
		Port: 8080,
		TLS:  false,
	}
	binding := HttpBinding{}
	endpoints := []v1alpha2.Endpoint{
		{
			Methods: []string{"GET"},
			Route:   "greetings",
			Version: "v1",
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				return v1alpha2.COAResponse{
					Body:  []byte("Hi there!!"),
					State: v1alpha2.OK,
				}
			},
		},
		{
			Methods: []string{"GET"},
			Route:   "greetings2",
			Version: "v1",
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				return v1alpha2.COAResponse{
					Body:  []byte("Hi " + c.Parameters["name"] + "!!"),
					State: v1alpha2.OK,
				}
			},
		},
		{
			Methods:    []string{"GET"},
			Route:      "greetings3",
			Version:    "v1",
			Parameters: []string{"name"},
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				return v1alpha2.COAResponse{
					Body:  []byte("Hi " + c.Parameters["__name"] + "!!!"),
					State: v1alpha2.OK,
				}
			},
		},
		{
			Methods: []string{"POST"},
			Route:   "greetingsWithMetadata",
			Version: "v1",
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				metadata := c.Metadata
				value := metadata["key"]
				return v1alpha2.COAResponse{
					Metadata: map[string]string{
						"key": value,
					},
					Body:  []byte("Hi " + value + "!!!!"),
					State: v1alpha2.OK,
				}
			},
		},
	}
	err := binding.Launch(config, endpoints, nil)
	assert.Nil(t, err)

	// wait for http server startup
	time.Sleep(5 * time.Second)

	testHttpRequestHelper(context.Background(), t, fasthttp.MethodGet, "http://localhost:8080/v1/greetings", nil, 200, "Hi there!!")

	// query args
	testHttpRequestHelper(context.Background(), t, fasthttp.MethodGet, "http://localhost:8080/v1/greetings2?name=John", nil, 200, "Hi John!!")

	// path parameters
	testHttpRequestHelper(context.Background(), t, fasthttp.MethodGet, "http://localhost:8080/v1/greetings3/John", nil, 200, "Hi John!!!")

	// req metadata and resp metadata
	req4Metadata := map[string]string{
		"key": "Alice",
	}
	b, _ := json.Marshal(req4Metadata)
	testHttpRequestHelperWithHeaders(
		context.Background(),
		t,
		fasthttp.MethodPost,
		"http://localhost:8080/v1/greetingsWithMetadata",
		nil,
		map[string]string{
			v1alpha2.COAMetaHeader: string(b),
		},
		200,
		"Hi Alice!!!!",
		map[string]string{
			v1alpha2.COAMetaHeader: string(b),
		})
}

func TestHTTPEchoWithTLS(t *testing.T) {
	config := HttpBindingConfig{
		Port: 8888,
		TLS:  true,
		CertProvider: CertProviderConfig{
			Type: "certs.autogen",
			Config: autogen.AutoGenCertProviderConfig{
				Name: "test",
			},
		},
	}
	binding := HttpBinding{}
	endpoints := []v1alpha2.Endpoint{
		{
			Methods: []string{"GET"},
			Route:   "greetings",
			Version: "v1",
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				return v1alpha2.COAResponse{
					Body:  []byte("Hi there!!"),
					State: v1alpha2.OK,
				}
			},
		},
	}
	err := binding.Launch(config, endpoints, nil)
	assert.Nil(t, err)

	// need to wait for tls cert creation (it is in another go routine)
	time.Sleep(5 * time.Second)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(fasthttp.MethodGet, "https://localhost:8888/v1/greetings", nil)
	assert.Nil(t, err)
	resp, err := client.Do(req)
	assert.Nil(t, err)

	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, 200)
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, string(bodyBytes), "Hi there!!")
}

func TestHTTPEchoWithPipeline(t *testing.T) {
	pubsub := &memory.InMemoryPubSubProvider{}
	err := pubsub.Init(memory.InMemoryPubSubConfig{})
	assert.Nil(t, err)

	signingKey := "TestKey"
	config := HttpBindingConfig{
		Port: 8081,
		TLS:  false,
		Pipeline: []MiddlewareConfig{
			{
				Type: "middleware.http.tracing",
				Properties: map[string]interface{}{
					"pipeline": []map[string]interface{}{
						{
							"exporter": map[string]interface{}{
								"type":       "tracing.exporters.console",
								"backendUrl": "",
								"sampler": map[string]string{
									"sampleRate": "always",
								},
							},
						},
					},
				},
			},
			{
				Type:       "middleware.http.trail",
				Properties: map[string]interface{}{},
			},
			{
				Type: "middleware.http.telemetry",
				Properties: map[string]interface{}{
					"enabled":                 true,
					"maxBatchSize":            8192,
					"maxBatchIntervalSeconds": 2,
					"client":                  "coabinding-test", // will be override as uuid
				},
			},
			{
				Type: "middleware.http.cors",
				Properties: map[string]interface{}{
					"Any": "value",
				},
			},
			{
				Type: "middleware.http.jwt",
				Properties: map[string]interface{}{
					"ignorePaths": []string{"/v1/auth"},
					"verifyKey":   signingKey,
					"enableRBAC":  true,
					"roles": []map[string]string{
						{
							"role":  "administrator",
							"claim": "user",
							"value": "adminuser",
						},
					},
					"policy": map[string]interface{}{
						"administrator": map[string]map[string]string{
							"items": {
								"/v1/greetings": fasthttp.MethodGet,
							},
						},
					},
				},
			},
		},
	}
	binding := HttpBinding{}
	userRoleMap := map[string][]string{
		"adminuser":      {"administrator"},
		"reader":         {"reader"},
		"developer":      {"developer"},
		"device-manager": {"device-manager"},
		"operator":       {"operator"},
	}
	endpoints := []v1alpha2.Endpoint{
		{
			Methods: []string{"POST"},
			Route:   "auth",
			Version: "v1",
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				var authRequest TestAuthRequest
				_ = json.Unmarshal(c.Body, &authRequest)

				mySigningKey := []byte(signingKey)
				claims := TestCustomClaims{
					authRequest.UserName,
					jwt.RegisteredClaims{
						// A usual scenario is to set the expiration time relative to the current time
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						NotBefore: jwt.NewNumericDate(time.Now()),
						Issuer:    "symphony",
						Subject:   "test",
						ID:        "1",
						Audience:  []string{"*"},
					},
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				ss, _ := token.SignedString(mySigningKey)

				roles := userRoleMap[authRequest.UserName]
				if roles == nil {
					roles = nil
				}
				rolesJSON, _ := json.Marshal(roles)
				resp := v1alpha2.COAResponse{
					State:       v1alpha2.OK,
					Body:        []byte(fmt.Sprintf(`{"accessToken":"%s", "tokenType": "Bearer", "username": "%s", "roles": %s}`, ss, authRequest.UserName, rolesJSON)),
					ContentType: "application/json",
				}
				return resp
			},
		},
		{
			Methods: []string{"GET", "POST"},
			Route:   "greetings",
			Version: "v1",
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				switch c.Method {
				case fasthttp.MethodGet:
					return v1alpha2.COAResponse{
						Body:  []byte("Hi there!!"),
						State: v1alpha2.OK,
					}
				case fasthttp.MethodPost:
					reqBody := string(c.Body)
					return v1alpha2.COAResponse{
						Body:  []byte(fmt.Sprintf("Hi %s!!", reqBody)),
						State: v1alpha2.OK,
					}
				}
				return v1alpha2.COAResponse{}
			},
		},
	}
	err = binding.Launch(config, endpoints, pubsub)
	assert.Nil(t, err)

	// wait for http server startup
	time.Sleep(5 * time.Second)

	client := &http.Client{}

	user := TestAuthRequest{
		UserName: "adminuser",
		Password: "",
	}
	userJSON, _ := json.Marshal(user)
	authReq, err := http.NewRequest(fasthttp.MethodPost, "http://localhost:8081/v1/auth", bytes.NewBuffer(userJSON))
	assert.Nil(t, err)
	authResp, err := client.Do(authReq)
	assert.Nil(t, err)

	defer authResp.Body.Close()
	assert.Equal(t, authResp.StatusCode, 200)
	bodyBytes2, err := io.ReadAll(authResp.Body)
	var parsedAuthResp map[string]interface{}
	json.Unmarshal(bodyBytes2, &parsedAuthResp)
	authHeader, ok := parsedAuthResp["accessToken"].(string)
	assert.True(t, ok)

	// test valid token and rbac
	testHttpRequestHelperWithHeaders(context.Background(), t, fasthttp.MethodGet, "http://localhost:8081/v1/greetings", nil,
		map[string]string{
			"Authorization": "Bearer " + authHeader,
		}, 200, "Hi there!!", map[string]string{
			"Access-Control-Allow-Origin":      corsAllowOrigin,
			"Access-Control-Allow-Methods":     corsAllowMethods,
			"Access-Control-Allow-Credentials": corsAllowCredentials,
			"Access-Control-Allow-Headers":     corsAllowHeaders,
			"Any":                              "value",
		})

	// invalid token
	testHttpRequestHelperWithHeaders(context.Background(), t, fasthttp.MethodGet, "http://localhost:8081/v1/greetings", nil, map[string]string{
		"Authorization": "Bearer fake-token",
	}, 403, "", nil)

	// empty token
	testHttpRequestHelperWithHeaders(context.Background(), t, fasthttp.MethodGet, "http://localhost:8081/v1/greetings", nil, nil, 403, "", nil)

	// test valid token and wrong rbac (method not match)
	testHttpRequestHelperWithHeaders(context.Background(), t, fasthttp.MethodPost, "http://localhost:8081/v1/greetings", []byte("John"), map[string]string{
		"Authorization": "Bearer " + authHeader,
	}, 403, "", nil)

	// test valid token and rbac with context
	ctx, span := observability.StartSpan("HTTP-Test-Client", context.Background(), &map[string]string{
		"method":      "TestHTTPEchoWithPipeline",
		"http.method": fasthttp.MethodGet,
		"http.url":    "http://localhost:8081/v1/greetings",
	})
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	testHttpRequestHelperWithHeaders(ctx, t, fasthttp.MethodGet, "http://localhost:8081/v1/greetings", nil,
		map[string]string{
			"Authorization": "Bearer " + authHeader,
		}, 200, "Hi there!!", map[string]string{
			"Access-Control-Allow-Origin":      corsAllowOrigin,
			"Access-Control-Allow-Methods":     corsAllowMethods,
			"Access-Control-Allow-Credentials": corsAllowCredentials,
			"Access-Control-Allow-Headers":     corsAllowHeaders,
			"Any":                              "value",
		})

	time.Sleep(5 * time.Second) // wait for telemetry to send data
}
