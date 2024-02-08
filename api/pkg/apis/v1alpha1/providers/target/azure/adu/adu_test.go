/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package adu

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

func TestInitWithNil(t *testing.T) {
	provider := ADUTargetProvider{}
	err := provider.Init(nil)
	assert.NotNil(t, err)
}

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name": "name",
	}
	provider := ADUTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":     "name",
		"tenantId": "tid",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":     "name",
		"tenantId": "tid",
		"clientId": "cid",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":         "name",
		"tenantId":     "tid",
		"clientId":     "cid",
		"clientSecret": "cscre",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":               "name",
		"tenantId":           "tid",
		"clientId":           "cid",
		"clientSecret":       "cscre",
		"aduAccountEndpoint": "ac",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":               "name",
		"tenantId":           "tid",
		"clientId":           "cid",
		"clientSecret":       "cscre",
		"aduAccountEndpoint": "acend",
		"aduAccountInstance": "acinst",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":               "name",
		"tenantId":           "tid",
		"clientId":           "cid",
		"clientSecret":       "cscre",
		"aduAccountEndpoint": "acend",
		"aduAccountInstance": "acinst",
		"aduGroup":           "agroup",
	}
	err = provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestGetFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	assert.Nil(t, err)

	provider := &ADUTargetProvider{}
	err = provider.Init(ADUTargetProviderConfig{
		Name:               "test",
		TenantId:           "00000000-0000-0000-0000-000000000000",
		ClientId:           "00000000-0000-0000-0000-000000000000",
		ClientSecret:       "testsecret",
		ADUAccountEndpoint: u.Host,
		ADUAccountInstance: "testinstance",
		ADUGroup:           "testgroup",
	})
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "test",
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	steps := []model.ComponentStep{
		{
			Action:    "update",
			Component: component,
		},
	}

	_, err = provider.Get(context.Background(), deployment, steps)
	assert.NotNil(t, err)
}

func TestApplyFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	assert.Nil(t, err)

	provider := &ADUTargetProvider{}
	err = provider.Init(ADUTargetProviderConfig{
		Name:               "test",
		TenantId:           "00000000-0000-0000-0000-000000000000",
		ClientId:           "00000000-0000-0000-0000-000000000000",
		ClientSecret:       "testsecret",
		ADUAccountEndpoint: u.Host,
		ADUAccountInstance: "testinstance",
		ADUGroup:           "testgroup",
	})
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "test",
		Properties: map[string]interface{}{
			"update.name":     "update",
			"update.provider": "provider",
			"update.version":  "1.0.0",
		},
	}

	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	components := []model.ComponentStep{
		{
			Action:    "update",
			Component: component,
		},
	}
	step := model.DeploymentStep{
		Components: components,
	}

	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)

	components = []model.ComponentStep{
		{
			Action:    "delete",
			Component: component,
		},
	}
	step = model.DeploymentStep{
		Components: components,
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestConformanceSuite(t *testing.T) {
	provider := &ADUTargetProvider{}
	err := provider.Init(ADUTargetProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
