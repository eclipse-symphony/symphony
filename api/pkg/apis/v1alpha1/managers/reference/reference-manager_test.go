/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package reference

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	refmock "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/http"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	reporterProvider := http.HTTPReporter{}
	reporterProvider.Init(http.HTTPReporterConfig{})
	manager := ReferenceManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.reference":     "mock",
			"providers.volatilestate": "memory",
			"providers.reporter":      "report",
		},
	}, map[string]providers.IProvider{
		"mock":   &referenceProvider,
		"memory": &stateProvider,
		"report": &reporterProvider,
	})
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{
		Values: map[string]interface{}{
			"abc": "def",
		},
	})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	reporterProvider := http.HTTPReporter{}
	reporterProvider.Init(http.HTTPReporterConfig{})
	manager := ReferenceManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.reference":     "mock",
			"providers.volatilestate": "memory",
			"providers.reporter":      "report",
		},
	}, map[string]providers.IProvider{
		"mock":   &referenceProvider,
		"memory": &stateProvider,
		"report": &reporterProvider,
	})
	assert.Nil(t, err)
	target, err := manager.Get("mock", "abc", "", "", "", "", "", "")
	assert.Nil(t, err)
	assert.Equal(t, "\"def\"", string(target))
}

func TestGetExt(t *testing.T) {
	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{
		Values: map[string]interface{}{
			"abc": "def",
		},
	})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	reporterProvider := http.HTTPReporter{}
	reporterProvider.Init(http.HTTPReporterConfig{})
	manager := ReferenceManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.reference":     "mock",
			"providers.volatilestate": "memory",
			"providers.reporter":      "report",
		},
	}, map[string]providers.IProvider{
		"mock":   &referenceProvider,
		"memory": &stateProvider,
		"report": &reporterProvider,
	})
	assert.Nil(t, err)
	_, err = manager.GetExt("mock", "", "abc", "", "", "", "abc", "", "", "", "", "")
	assert.Nil(t, err)
}

func TestGetExtDownload(t *testing.T) {
	modelSpec := &model.ModelSpec{}
	modelSpec.DisplayName = "test"
	modelSpec.Properties = map[string]string{
		"model.type":      "customvision",
		"model.endpoint":  "endpoint",
		"model.project":   "project",
		"model.version.1": "00000000-0000-0000-0000-000000000000",
	}

	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{
		Values: map[string]interface{}{
			"abc": map[string]interface{}{
				"spec": modelSpec,
			},
			"project": "project",
		},
	})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	reporterProvider := http.HTTPReporter{}
	reporterProvider.Init(http.HTTPReporterConfig{})
	manager := ReferenceManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.reference":     "mock",
			"providers.volatilestate": "memory",
			"providers.reporter":      "report",
		},
	}, map[string]providers.IProvider{
		"mock":   &referenceProvider,
		"memory": &stateProvider,
		"report": &reporterProvider,
	})
	assert.Nil(t, err)
	_, err = manager.GetExt("mock", "", "abc", "", "", "", "abc", "download", "", "", "latest", "")
	assert.Nil(t, err)
}

type AnyType struct {
	Name  string
	Value uint64
}

func TestGetAnyType(t *testing.T) {
	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{
		Values: map[string]interface{}{
			"abc": AnyType{
				Name:  "def",
				Value: 12345,
			},
		},
	})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	reporterProvider := http.HTTPReporter{}
	reporterProvider.Init(http.HTTPReporterConfig{})
	manager := ReferenceManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.reference":     "mock",
			"providers.volatilestate": "memory",
			"providers.reporter":      "report",
		},
	}, map[string]providers.IProvider{
		"mock":   &referenceProvider,
		"memory": &stateProvider,
		"report": &reporterProvider,
	})
	assert.Nil(t, err)
	target := AnyType{}
	data, err := manager.Get("mock", "abc", "", "", "", "", "", "")
	assert.Nil(t, err)
	err = json.Unmarshal(data, &target)
	assert.Nil(t, err)
	assert.Equal(t, "def", target.Name)
	assert.Equal(t, uint64(12345), target.Value)
}

func TestPoll(t *testing.T) {
	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	reporterProvider := http.HTTPReporter{}
	reporterProvider.Init(http.HTTPReporterConfig{})
	manager := ReferenceManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.reference":     "mock",
			"providers.volatilestate": "memory",
			"providers.reporter":      "report",
			"poll.enabled":            "true",
		},
	}, map[string]providers.IProvider{
		"mock":   &referenceProvider,
		"memory": &stateProvider,
		"report": &reporterProvider,
	})
	assert.Nil(t, err)
	res := manager.Enabled()
	assert.True(t, res)
	errPoll := manager.Poll()
	assert.Nil(t, errPoll)
	errRec := manager.Poll()
	assert.Nil(t, errRec)
}

func TestCacheLifespan(t *testing.T) {
	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	reporterProvider := http.HTTPReporter{}
	reporterProvider.Init(http.HTTPReporterConfig{})
	manager := ReferenceManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.reference":     "mock",
			"providers.volatilestate": "memory",
			"cacheLifespan":           "5",
			"providers.reporter":      "report",
		},
	}, map[string]providers.IProvider{
		"mock":   &referenceProvider,
		"memory": &stateProvider,
		"report": &reporterProvider,
	})
	assert.Nil(t, err)
	stamp := time.Now()
	stamp2 := time.Now()
	data, err := manager.Get("mock", "timestamp", "", "", "", "", "", "")
	assert.Nil(t, err)
	err = json.Unmarshal(data, &stamp)
	assert.Nil(t, err)
	assert.LessOrEqual(t, time.Since(stamp).Seconds(), float64(5))
	data, err = manager.Get("mock", "timestamp", "", "", "", "", "", "")
	assert.Nil(t, err)
	err = json.Unmarshal(data, &stamp2)
	assert.Nil(t, err)
	assert.Equal(t, stamp, stamp2)
	time.Sleep(10 * time.Second)
	data, err = manager.Get("mock", "timestamp", "", "", "", "", "", "")
	assert.Nil(t, err)
	err = json.Unmarshal(data, &stamp)
	assert.Nil(t, err)
	assert.LessOrEqual(t, time.Since(stamp).Seconds(), float64(5))
}
func TestFillParametersFromInstance(t *testing.T) {
	instance := model.InstanceSpec{
		Pipelines: []model.PipelineSpec{
			{
				Name:  "pipeline1",
				Skill: "skill1",
				Parameters: map[string]string{
					"c": "value-c",
					"a": "value-a",
				},
			},
		},
	}
	skill := model.SkillSpec{
		Parameters: map[string]string{
			"a": "default-a",
			"c": "default-c",
		},
		Nodes: []model.NodeSpec{
			{
				Id: "1",
				Configurations: map[string]string{
					"v-a": "$param(a)",
					"v-c": "$param(c)",
				},
			},
		},
	}
	data1, _ := json.Marshal(skill)
	data2, _ := json.Marshal(instance)
	data, err := fillParameters(data1, data2, "skill1", "pipeline1")
	assert.Nil(t, err)
	var updatedSkill model.SkillSpec
	err = json.Unmarshal(data, &updatedSkill)
	assert.Nil(t, err)
	assert.Equal(t, "value-a", updatedSkill.Nodes[0].Configurations["v-a"])
	assert.Equal(t, "value-c", updatedSkill.Nodes[0].Configurations["v-c"])
}
func TestFillParametersFromInstanceMixWithTopParameters(t *testing.T) {
	instance := model.InstanceSpec{
		Parameters: map[string]string{
			"a": "value-a",
		},
		Pipelines: []model.PipelineSpec{
			{
				Name:  "pipeline1",
				Skill: "skill1",
				Parameters: map[string]string{
					"c": "value-c",
				},
			},
		},
	}
	skill := model.SkillSpec{
		Parameters: map[string]string{
			"a": "default-a",
			"c": "default-c",
		},
		Nodes: []model.NodeSpec{
			{
				Id: "1",
				Configurations: map[string]string{
					"v-a": "$param(a)",
					"v-c": "$param(c)",
				},
			},
		},
	}
	data1, _ := json.Marshal(skill)
	data2, _ := json.Marshal(instance)
	data, err := fillParameters(data1, data2, "skill1", "pipeline1")
	assert.Nil(t, err)
	var updatedSkill model.SkillSpec
	err = json.Unmarshal(data, &updatedSkill)
	assert.Nil(t, err)
	assert.Equal(t, "value-a", updatedSkill.Nodes[0].Configurations["v-a"])
	assert.Equal(t, "value-c", updatedSkill.Nodes[0].Configurations["v-c"])
}
func TestFillParametersFromInstanceMissingA(t *testing.T) {
	instance := model.InstanceSpec{
		Pipelines: []model.PipelineSpec{
			{
				Name:  "pipeline1",
				Skill: "skill1",
				Parameters: map[string]string{
					"c": "value-c",
				},
			},
		},
	}
	skill := model.SkillSpec{
		Parameters: map[string]string{
			"a": "default-a",
			"c": "default-c",
		},
		Nodes: []model.NodeSpec{
			{
				Id: "1",
				Configurations: map[string]string{
					"v-a": "$param(a)",
					"v-c": "$param(c)",
				},
			},
		},
	}
	data1, _ := json.Marshal(skill)
	data2, _ := json.Marshal(instance)
	data, err := fillParameters(data1, data2, "skill1", "pipeline1")
	assert.Nil(t, err)
	var updatedSkill model.SkillSpec
	err = json.Unmarshal(data, &updatedSkill)
	assert.Nil(t, err)
	assert.Equal(t, "default-a", updatedSkill.Nodes[0].Configurations["v-a"])
	assert.Equal(t, "value-c", updatedSkill.Nodes[0].Configurations["v-c"])
}
