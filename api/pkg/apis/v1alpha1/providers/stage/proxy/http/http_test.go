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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

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
			response = utils.AuthResponse{
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

	mgrCtx := contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			SecurityPolicy: &contexts.SecurityPolicy{
				AllowedIPRanges: []string{"127.0.0.1/8"},
			},
		},
	}
	result, paused, err := provider.Process(context.TODO(), mgrCtx, v1alpha2.ActivationData{
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
			response = utils.AuthResponse{
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

	mgrCtx := contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			SecurityPolicy: &contexts.SecurityPolicy{
				AllowedIPRanges: []string{"127.0.0.1/8"},
			},
		},
	}
	_, _, err = provider.Process(context.TODO(), mgrCtx, v1alpha2.ActivationData{
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

func TestValidateProxyBaseUrl(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		policy    *contexts.SecurityPolicy
		wantError bool
	}{
		{
			name:      "nil policy permits public IP",
			rawURL:    "http://1.2.3.4/",
			policy:    nil,
			wantError: false,
		},
		{
			name:      "nil policy rejects loopback",
			rawURL:    "http://127.0.0.1:8080/",
			policy:    nil,
			wantError: true,
		},
		{
			name:      "nil policy rejects link-local",
			rawURL:    "http://169.254.169.254/",
			policy:    nil,
			wantError: true,
		},
		{
			name:      "non-http scheme is rejected",
			rawURL:    "file:///etc/passwd",
			policy:    nil,
			wantError: true,
		},
		{
			name:   "allowedIPRanges whitelist overrides deny list",
			rawURL: "http://10.0.0.5/",
			policy: &contexts.SecurityPolicy{
				AllowedIPRanges: []string{"10.0.0.0/8"},
			},
			wantError: false,
		},
		{
			name:   "exclusive mode rejects non-whitelisted public IP",
			rawURL: "http://1.2.3.4/",
			policy: &contexts.SecurityPolicy{
				AllowedIPRanges:    []string{"10.0.0.0/8"},
				AllowListExclusive: true,
			},
			wantError: true,
		},
		{
			name:   "exclusive mode permits whitelisted IP",
			rawURL: "http://10.0.0.5/",
			policy: &contexts.SecurityPolicy{
				AllowedIPRanges:    []string{"10.0.0.0/8"},
				AllowListExclusive: true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProxyBaseUrl(tt.rawURL, tt.policy)
			if tt.wantError {
				assert.NotNil(t, err)
				assert.Equal(t, v1alpha2.BadConfig, err.(v1alpha2.COAError).State)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
