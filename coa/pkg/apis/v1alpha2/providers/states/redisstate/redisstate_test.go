/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package redisstate

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/stretchr/testify/assert"
)

type TestPayload struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestWithEmptyConfig(t *testing.T) {
	provider := RedisStateProvider{}
	err := provider.Init(RedisStateProviderConfig{})
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.MissingConfig, coaErr.State)
}

func TestWithMissingHost(t *testing.T) {
	provider := RedisStateProvider{}
	err := provider.Init(RedisStateProviderConfig{
		Name:     "test",
		Password: "abc",
	})
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.MissingConfig, coaErr.State)
}

func TestInit(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS enviornment variable is not set")
	}
	provider := RedisStateProvider{}
	err := provider.Init(RedisStateProviderConfig{
		Name:     "test",
		Host:     "localhost:6379",
		Password: "",
	})
	assert.Nil(t, err)
}

func initializeProvider(t *testing.T) RedisStateProvider {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS enviornment variable is not set")
	}
	provider := RedisStateProvider{}
	err := provider.Init(RedisStateProviderConfig{
		Name:     "test",
		Host:     "localhost:6379",
		Password: "",
	})
	assert.Nil(t, err)
	return provider
}

func TestUpsertGetListAndDelete(t *testing.T) {
	provider := initializeProvider(t)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
			ETag: "testETag",
		},
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", id)

	entry, err := provider.Get(context.Background(), states.GetRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", entry.ID)
	var body []byte
	body, _ = json.Marshal(entry.Body)
	var payload TestPayload
	err = json.Unmarshal(body, &payload)
	assert.Nil(t, err)
	assert.Equal(t, "Random name", payload.Name)
	assert.Equal(t, 12345, payload.Value)
	assert.Equal(t, "testETag", entry.ETag)

	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "123", entries[0].ID)
	body, _ = json.Marshal(entries[0].Body)
	err = json.Unmarshal(body, &payload)
	assert.Nil(t, err)
	assert.Equal(t, "Random name", payload.Name)
	assert.Equal(t, 12345, payload.Value)
	assert.Equal(t, "testETag", entries[0].ETag)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}

func TestUpsertGetListAndDeleteWithNamespace(t *testing.T) {
	provider := initializeProvider(t)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
			ETag: "testETag",
		},
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", id)

	entry, err := provider.Get(context.Background(), states.GetRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", entry.ID)
	var body []byte
	body, _ = json.Marshal(entry.Body)
	var payload TestPayload
	err = json.Unmarshal(body, &payload)
	assert.Nil(t, err)
	assert.Equal(t, "Random name", payload.Name)
	assert.Equal(t, 12345, payload.Value)
	assert.Equal(t, "testETag", entry.ETag)

	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "123", entries[0].ID)
	body, _ = json.Marshal(entries[0].Body)
	err = json.Unmarshal(body, &payload)
	assert.Nil(t, err)
	assert.Equal(t, "Random name", payload.Name)
	assert.Equal(t, 12345, payload.Value)
	assert.Equal(t, "testETag", entries[0].ETag)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
}
func TestListNamespace(t *testing.T) {
	provider := initializeProvider(t)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
			ETag: "testETag",
		},
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", id)

	// namespace is different
	id, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
			ETag: "testETag",
		},
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace2",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "123", id)

	// type is different
	id, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "234",
			Body: TestPayload{
				Name:  "Random name",
				Value: 12345,
			},
			ETag: "testETag",
		},
		Metadata: map[string]interface{}{
			"resource":  "testresource2",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "234", id)

	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))

	entries, _, err = provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace2",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))

	entries, _, err = provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"resource":  "testresource2",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace2",
		},
	})
	assert.Nil(t, err)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "234",
		Metadata: map[string]interface{}{
			"resource":  "testresource2",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
}

func TestGetNonExisting(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Get(context.Background(), states.GetRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.NotFound, coaErr.State)
}

func TestUpsertStateOnly(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"spec": map[string]interface{}{
					"scope": "test",
				},
			},
		},
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"status": map[string]interface{}{
					"phase": "Running",
				},
			},
		},
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
		Options: states.UpsertOption{
			UpdateStatusOnly: true,
		},
	})
	assert.Nil(t, err)
	entry, err := provider.Get(context.Background(), states.GetRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	spec := entry.Body.(map[string]interface{})["spec"].(map[string]interface{})
	assert.Equal(t, "test", spec["scope"])
	status := entry.Body.(map[string]interface{})["status"].(map[string]interface{})
	assert.Equal(t, "Running", status["phase"])
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}

func TestLabelFilter(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app=test",
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "default",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))

	entity, _, err = provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app=test",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}

func TestLabelFilterWithNamespace(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app=test",
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))

	entity, _, err = provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app=test",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource":  "testresource",
			"group":     "testgroup",
			"namespace": "testnamespace",
		},
	})
	assert.Nil(t, err)
}

func TestLabelFilterNotEqual(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "app!=test2",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}

func TestLabelFilterBadFilter(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	_, _, err = provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: "xxxxx",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.NotNil(t, err)
	e, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadRequest, e.State)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}

func TestFieldFilterMetadata(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "field",
		FilterValue: "metadata.name=c1",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}

func TestFieldFilterDeepMetadataNotEqual(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "field",
		FilterValue: "metadata.labels.app!=xxx",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}

func TestFieldFilterStatus(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "234",
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
					"name": "c2",
				},
				"spec": map[string]interface{}{},
			},
		},
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "field",
		FilterValue: "status.phase=Running",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "234",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}

func TestSpecFilter(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "spec",
		FilterValue: `[?(@.properties.foo=="bar")]`,
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}
func TestStatusFilter(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "status",
		FilterValue: `[?(@.properties.foo=="bar")]`,
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}
func TestMultipleLabelsFilter(t *testing.T) {
	provider := initializeProvider(t)
	_, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "123",
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
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	entity, _, err := provider.List(context.Background(), states.ListRequest{
		FilterType:  "label",
		FilterValue: `app==test,app2=test2`,
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entity))
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "123",
		Metadata: map[string]interface{}{
			"resource": "testresource",
			"group":    "testgroup",
		},
	})
	assert.Nil(t, err)
}
