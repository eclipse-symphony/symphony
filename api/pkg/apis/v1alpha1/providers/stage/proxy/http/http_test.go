/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func TestInitWithMap(t *testing.T) {
	provider := HTTPProxyStageProvider{}
	input := map[string]string{}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
}
func TestSuccessfulProcess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/processor":
			response = model.StageStatus{
				Status: v1alpha2.Done,
				Outputs: map[string]interface{}{
					"foo": "bar",
				},
			}
		default:
			response = AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")

	provider := HTTPProxyStageProvider{}
	err := provider.Init(HTTPProxyStageProviderConfig{})
	assert.Nil(t, err)

	result, paused, err := provider.Process(context.TODO(), contexts.ManagerContext{}, v1alpha2.ActivationData{
		Inputs: map[string]interface{}{
			"foo": "bar",
		},
		Proxy: &v1alpha2.ProxySpec{
			Config: map[string]interface{}{
				"baseUrl":  ts.URL + "/",
				"user":     "admin",
				"password": "",
			},
		},
	})
	assert.Nil(t, err)
	assert.False(t, paused)
	assert.NotNil(t, result)
	assert.Equal(t, "bar", result["foo"])
}
func TestFailedProcess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/processor":
			response = model.StageStatus{
				Status: v1alpha2.InternalError,
				Outputs: map[string]interface{}{
					"foo": "bar",
				},
			}
		default:
			response = AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")

	provider := HTTPProxyStageProvider{}
	err := provider.Init(HTTPProxyStageProviderConfig{})
	assert.Nil(t, err)

	_, _, err = provider.Process(context.TODO(), contexts.ManagerContext{}, v1alpha2.ActivationData{
		Inputs: map[string]interface{}{
			"foo": "bar",
		},
		Proxy: &v1alpha2.ProxySpec{
			Config: map[string]interface{}{
				"baseUrl":  ts.URL + "/",
				"user":     "admin",
				"password": "",
			},
		},
	})
	assert.Equal(t, err.(v1alpha2.COAError).State, v1alpha2.InternalError)
}
func TestNoServer(t *testing.T) {

	provider := HTTPProxyStageProvider{}
	err := provider.Init(HTTPProxyStageProviderConfig{})
	assert.Nil(t, err)

	_, _, err = provider.Process(context.TODO(), contexts.ManagerContext{}, v1alpha2.ActivationData{
		Inputs: map[string]interface{}{
			"foo": "bar",
		},
		Proxy: &v1alpha2.ProxySpec{
			Config: map[string]interface{}{
				"baseUrl":  "http://bad/",
				"user":     "admin",
				"password": "",
			},
		},
	})
	assert.NotNil(t, err)
}
