/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memorystate

import (
	"context"
	"encoding/json"
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
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProviderConfig{})
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.InitWithMap(
		map[string]string{
			"name": "name1",
		},
	)
	assert.Nil(t, err)
}

func TestID(t *testing.T) {
	provider := MemoryStateProvider{}
	provider.Init(MemoryStateProviderConfig{
		Name: "name",
	})

	assert.Equal(t, "name", provider.ID())
}

func TestSetContext(t *testing.T) {
	provider := MemoryStateProvider{}
	provider.Init(MemoryStateProviderConfig{
		Name: "name",
	})
	provider.SetContext(&contexts.ManagerContext{})
	assert.NotNil(t, provider.Context)
}

func TestUpSert(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", id)
}

func TestUpSertWithNamespace(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", id)
}

func TestList(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
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
	entries, _, err := provider.List(context.Background(), states.ListRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "123", entries[0].ID)
}

func TestListWithNamespace(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
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
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "234",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "default",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "123", entries[0].ID)
	entries, _, err = provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "234", entries[0].ID)
	entries, _, err = provider.List(context.Background(), states.ListRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(entries))
	assert.Equal(t, "123", entries[0].ID)
	assert.Equal(t, "234", entries[1].ID)
}

func TestDelete(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
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
	entries, _, err := provider.List(context.Background(), states.ListRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entries))
}

func TestDeleteWithNamespace(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entries))
}

func TestMemoryStateProviderConfigFromMapNil(t *testing.T) {
	_, err := MemoryStateProviderConfigFromMap(nil)
	assert.Nil(t, err)
}

func TestMemoryStateProviderConfigFromMapEmpty(t *testing.T) {
	_, err := MemoryStateProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestMemoryStateProviderConfigFromMap(t *testing.T) {
	config, err := MemoryStateProviderConfigFromMap(map[string]string{
		"name": "my-name",
	})
	assert.Nil(t, err)
	assert.Equal(t, "my-name", config.Name)
}
func TestMemoryStateProviderConfigFromMapEnvOverride(t *testing.T) {
	os.Setenv("my-name", "real-name")
	config, err := MemoryStateProviderConfigFromMap(map[string]string{
		"name": "$env:my-name",
	})
	assert.Nil(t, err)
	assert.Equal(t, "real-name", config.Name)
}
func TestGet(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
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

func TestGetWithNamespace(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	entity, err := provider.Get(context.Background(), states.GetRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
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

func TestUpSertEmptyID(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "", id)
}

func TestListEmptyID(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	entries, _, err := provider.List(context.Background(), states.ListRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "", entries[0].ID)
}

func TestDeleteEmptyID(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "",
	})
	assert.Nil(t, err)
	entries, _, err := provider.List(context.Background(), states.ListRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entries))
}

func TestGetEmptyID(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
		},
	})
	assert.Nil(t, err)
	entity, err := provider.Get(context.Background(), states.GetRequest{
		ID: "",
	})
	assert.Nil(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, "", entity.ID)

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

func TestClone(t *testing.T) {
	provider := MemoryStateProvider{}

	p, err := provider.Clone(MemoryStateProviderConfig{
		Name: "",
	})
	assert.NotNil(t, p)
	assert.Nil(t, err)

	p, err = provider.Clone(nil)
	assert.NotNil(t, p)
	assert.Nil(t, err)
}

func TestLabelFilter(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"spec": map[string]interface{}{},
			},
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app=test",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}

func TestLabelFilterWithNamespace(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"spec": map[string]interface{}{},
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app=test",
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}

func TestLabelFilterNotEqual(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"spec": map[string]interface{}{},
			},
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app!=test2",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}
func TestLabelFilterNotEqualWithNamespace(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"spec": map[string]interface{}{},
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app!=test2",
		Metadata: map[string]interface{}{
			"namespace": "nondefault",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}
func TestLabelFilterBadFilter(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"spec": map[string]interface{}{},
			},
		},
	})
	assert.Nil(t, err)
	_, _, err = provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "xxxxx",
	})
	assert.NotNil(t, err)
	e, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadRequest, e.State)
}
func TestFieldFilterMetadata(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
					"name": "c1",
				},
				"spec": map[string]interface{}{},
			},
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "field",
		FilterValue: "metadata.name=c1",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}
func TestFieldFilterDeepMetadataNotEqual(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
					"name": "c1",
				},
				"spec": map[string]interface{}{},
			},
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "field",
		FilterValue: "metadata.labels.app!=xxx",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}
func TestFieldFilterStatus(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
					"name": "c1",
				},
				"spec": map[string]interface{}{},
				"status": map[string]interface{}{
					"phase": "Running",
				},
			},
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "field",
		FilterValue: "status.phase=Running",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}
func TestFieldFilterStatusBadFilter(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
					"name": "c1",
				},
				"spec": map[string]interface{}{},
				"status": map[string]interface{}{
					"phase": "Running",
				},
			},
		},
	})
	assert.Nil(t, err)
	_, _, err = provider.List(context.Background(), states.ListRequest{
		FilterType:  "field",
		FilterValue: "status.phase",
	})
	assert.NotNil(t, err)
	e, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadRequest, e.State)
}
func TestSpecFilter(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"spec": map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "spec",
		FilterValue: `[?(@.properties.foo=="bar")]`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}
func TestStatusFilter(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"status": map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "status",
		FilterValue: `[?(@.properties.foo=="bar")]`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}

func TestMultipleLabelsFilter(t *testing.T) {
	provider := MemoryStateProvider{}
	err := provider.Init(MemoryStateProvider{})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app":  "test",
						"app2": "test2",
					},
				},
				"status": map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: `app==test,app2=test2`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
}
