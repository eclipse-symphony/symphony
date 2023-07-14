/*
MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE
*/

package httpstate

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	states "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
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
