/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package httpstate

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/stretchr/testify/assert"
)

type TestPayload struct {
	Name  string
	Value int
}

func TestInitWithEmptyConfig(t *testing.T) {
	provider := HttpStateProvider{}
	err := provider.Init(HttpStateProviderConfig{})
	assert.NotNil(t, err)
}
func TestInitWithConfig(t *testing.T) {
	provider := HttpStateProvider{}
	err := provider.Init(HttpStateProviderConfig{
		Url: "http://localhost:3500/v1.0/state/statestore",
	})
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	provider := HttpStateProvider{}
	err := provider.InitWithMap(
		map[string]string{
			"url":               "http://localhost:3500/v1.0/state/statestore",
			"postNameInPath":    "false",
			"postBodyKeyName":   "key",
			"postBodyValueName": "value",
			"postAsArray":       "true",
		},
	)
	assert.Nil(t, err)
}

func TestInitWithMapWithError(t *testing.T) {
	provider := HttpStateProvider{}
	err := provider.InitWithMap(
		map[string]string{
			"url":               "http://localhost:3500/v1.0/state/statestore",
			"postNameInPath":    "false",
			"postBodyKeyName":   "key",
			"postBodyValueName": "value",
			"postAsArray":       "This is causing an Error :)",
		},
	)
	assert.Error(t, err, "invalid bool value in the 'postAsArray' setting of Http state provider")
}

func TestID(t *testing.T) {
	provider := HttpStateProvider{}
	provider.Init(HttpStateProviderConfig{
		Name: "name",
	})

	assert.Equal(t, "name", provider.ID())
}

func TestSetContext(t *testing.T) {
	provider := HttpStateProvider{}
	provider.Init(HttpStateProviderConfig{
		Name: "name",
	})
	provider.SetContext(&contexts.ManagerContext{})
	assert.NotNil(t, provider.Context)
}

func TestUpSert(t *testing.T) {
	testDapr := os.Getenv("TEST_DAPR")
	if testDapr == "" {
		t.Skip("Skipping because TEST_DAPR enviornment variable is not set")
	}
	provider := HttpStateProvider{}
	err := provider.Init(HttpStateProviderConfig{
		Url:               "http://localhost:3500/v1.0/state/statestore",
		PostNameInPath:    false,
		PostBodyKeyName:   "key",
		PostBodyValueName: "value",
		PostAsArray:       true,
	})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name 3",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", id)
}

func TestList(t *testing.T) {
	testDapr := os.Getenv("TEST_DAPR")
	if testDapr == "" {
		t.Skip("Skipping because TEST_DAPR enviornment variable is not set")
	}
	provider := HttpStateProvider{}
	err := provider.Init(HttpStateProviderConfig{
		Url:               "http://localhost:3500/v1.0/state/statestore",
		PostNameInPath:    false,
		PostBodyKeyName:   "key",
		PostBodyValueName: "value",
		PostAsArray:       true,
		NotFoundAs204:     true,
	})
	assert.Nil(t, err)
	_, _, err = provider.List(context.Background(), states.ListRequest{})
	assert.NotNil(t, err)
}

func TestDelete(t *testing.T) {
	testDapr := os.Getenv("TEST_DAPR")
	if testDapr == "" {
		t.Skip("Skipping because TEST_DAPR enviornment variable is not set")
	}
	provider := HttpStateProvider{}
	err := provider.Init(HttpStateProviderConfig{
		Url:               "http://localhost:3500/v1.0/state/statestore",
		PostNameInPath:    false,
		PostBodyKeyName:   "key",
		PostBodyValueName: "value",
		PostAsArray:       true,
		NotFoundAs204:     true,
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
	})
	assert.Nil(t, err)
}

func TestMemoryStateProviderConfigFromMapNil(t *testing.T) {
	_, err := HttpStateProviderConfigFromMap(nil)
	assert.NotNil(t, err)
}

func TestMemoryStateProviderConfigFromMapEmpty(t *testing.T) {
	_, err := HttpStateProviderConfigFromMap(map[string]string{})
	assert.NotNil(t, err)
}
func TestMemoryStateProviderConfigFromMap(t *testing.T) {
	config, err := HttpStateProviderConfigFromMap(map[string]string{
		"name": "my-name",
		"url":  "some-url",
	})
	assert.Nil(t, err)
	assert.Equal(t, "my-name", config.Name)
}
func TestMemoryStateProviderConfigFromMapEnvOverride(t *testing.T) {
	os.Setenv("my-name", "real-name")
	config, err := HttpStateProviderConfigFromMap(map[string]string{
		"name": "$env:my-name",
		"url":  "some-url",
	})
	assert.Nil(t, err)
	assert.Equal(t, "real-name", config.Name)
}
func TestGet(t *testing.T) {
	testDapr := os.Getenv("TEST_DAPR")
	if testDapr == "" {
		t.Skip("Skipping because TEST_DAPR enviornment variable is not set")
	}
	provider := HttpStateProvider{}
	err := provider.Init(HttpStateProviderConfig{
		Url:               "http://localhost:3500/v1.0/state/statestore",
		PostNameInPath:    false,
		PostBodyKeyName:   "key",
		PostBodyValueName: "value",
		PostAsArray:       true,
		NotFoundAs204:     true,
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	entity, err := provider.Get(context.Background(), states.GetRequest{
		ID: "123",
	})
	assert.Nil(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, "123", entity.ID)

	payload := TestPayload{}
	data, err := json.Marshal(entity.Body)
	assert.Nil(t, err)
	err = json.Unmarshal(data, &payload)
	assert.Nil(t, err)
	assert.Equal(t, "Random name", payload.Name)
	assert.Equal(t, 12345, payload.Value)
	entity, err = provider.Get(context.Background(), states.GetRequest{
		ID: "890",
	})
	sczErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.NotFound, sczErr.State)
}

func TestUpsertGetDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			relativePath := r.URL.Path
			if relativePath == "" || relativePath == "/" {
				http.Error(w, "invalid path", http.StatusBadRequest)
				return
			}
			response := map[string]interface{}{
				"key": "abc",
			}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
		} else if r.Method == "POST" {
			var data []map[string]interface{}
			body, _ := ioutil.ReadAll(r.Body)
			err := json.Unmarshal([]byte(body), &data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if len(data) == 0 {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			firstItemKey := data[0]["key"].(string)
			if firstItemKey == "" {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			response := map[string]interface{}{
				"key":   data[0]["key"],
				"value": data[0]["value"],
			}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
		} else if r.Method == "DELETE" {
			relativePath := r.URL.Path
			if relativePath == "" || relativePath == "/" {
				http.Error(w, "invalid path", http.StatusBadRequest)
				return
			}
			w.Write([]byte("OK"))
		} else {
			w.Write([]byte("OK"))
		}
	}))
	defer ts.Close()

	provider := HttpStateProvider{}
	err := provider.Init(HttpStateProviderConfig{
		Url:               ts.URL,
		PostNameInPath:    false,
		PostBodyKeyName:   "key",
		PostBodyValueName: "value",
		PostAsArray:       true,
		NotFoundAs204:     true,
	})
	assert.Nil(t, err)

	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	_, err = provider.Get(context.Background(), states.GetRequest{
		ID: "123",
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
	})
	assert.Nil(t, err)

	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID:   "",
			Body: TestPayload{},
		},
	})
	assert.NotNil(t, err)

	_, err = provider.Get(context.Background(), states.GetRequest{
		ID: "",
	})
	assert.NotNil(t, err)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "",
	})
	assert.NotNil(t, err)
}

func TestClone(t *testing.T) {
	provider := HttpStateProvider{}
	provider.Init(HttpStateProviderConfig{
		Url: "http://localhost:3500/v1.0/state/statestore",
	})

	p, err := provider.Clone(HttpStateProviderConfig{
		Url:               "http://localhost:3500/v1.0/state/statestore",
		PostNameInPath:    false,
		PostBodyKeyName:   "key",
		PostBodyValueName: "value",
		PostAsArray:       true,
		NotFoundAs204:     true,
	})
	assert.NotNil(t, p)
	assert.Nil(t, err)

	p, err = provider.Clone(nil)
	assert.NotNil(t, p)
	assert.Nil(t, err)
}
