/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"encoding/json"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestStageSpecSchedule(t *testing.T) {
	jsonString := `{"schedule": "2021-01-30T08:30:10+08:00"}`

	var newStage StageSpec
	var err = json.Unmarshal([]byte(jsonString), &newStage)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	targetTime := "2021-01-30T08:30:10+08:00"
	if newStage.Schedule != targetTime {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
}

func TestInstanceSpecDeepEquals(t *testing.T) {
	interval := "1h"
	spec := createDummyInstanceSpec("spec", interval)
	spec_update := createDummyInstanceSpec("spec", interval)

	// Test DisplayName
	spec_update.DisplayName = "spec_update"
	assert.False(t, spec.DeepEquals(spec_update))
	spec_update.DisplayName = spec.DisplayName

	// Test Scope
	spec_update.Scope = "nondefault"
	assert.False(t, spec.DeepEquals(spec_update))
	spec_update.Scope = spec.Scope

	// Test Parameters
	spec_update.Parameters = nil
	assert.False(t, spec.DeepEquals(spec_update))
	spec_update.Parameters = spec.Parameters

	// Test Metadata
	spec_update.Metadata = nil
	assert.True(t, spec.DeepEquals(spec_update))

	// Test Solution
	spec_update.Solution = "solution:version2"
	assert.False(t, spec.DeepEquals(spec_update))
	spec_update.Solution = spec.Solution

	// Test Target
	spec_update.Target.Selector["group"] = "rtos-demo"
	assert.False(t, spec.DeepEquals(spec_update))
	spec_update.Target.Selector["group"] = spec.Target.Selector["group"]

	// Test Topologies
	spec_update.Topologies[0].Bindings[0].Role = "ingress"
	assert.False(t, spec.DeepEquals(spec_update))
	spec_update.Topologies[0].Bindings[0].Role = spec.Topologies[0].Bindings[0].Role

	// Test Pipelines
	spec_update.Pipelines[0].Name = "pipeline1"
	assert.False(t, spec.DeepEquals(spec_update))
	spec_update.Pipelines[0].Name = spec.Pipelines[0].Name

	// Test IsDryRun
	spec_update.IsDryRun = !spec.IsDryRun
	assert.False(t, spec.DeepEquals(spec_update))
	spec_update.IsDryRun = spec.IsDryRun

	// Test ReconciliationPolicy
	spec_update.ReconciliationPolicy = nil
	assert.False(t, spec.DeepEquals(spec_update))

}

func createDummyInstanceSpec(name string, interval string) InstanceSpec {
	return InstanceSpec{
		DisplayName: name,
		Scope:       "default",
		Parameters: map[string]string{
			"foo": "bar",
		},
		Metadata: map[string]string{},
		Solution: "solution:version1",
		Target: model.TargetSelector{
			Name: "target-1",
			Selector: map[string]string{
				"group": "demo",
			},
		},
		Topologies: []model.TopologySpec{
			{
				Bindings: []model.BindingSpec{
					{
						Config: map[string]string{
							"inCluster": "true",
						},
						Provider: "providers.target.k8s",
						Role:     "instance",
					},
				},
			},
		},
		Pipelines: []model.PipelineSpec{
			{
				Name:  "pipeline0",
				Skill: "skill-d1574858-24bb-41eb-b3cb-ec27c3199c94",
				Parameters: map[string]string{
					"device_displayname":   "fdfd",
					"device_id":            "device-1",
					"fps":                  "10",
					"instance_displayname": "dfdf",
					"rtsp":                 "rtsp://:@20.212.158.240/2.mkv",
					"skill_displayname":    "fsdfdf",
				},
			},
		},
		IsDryRun: false,
		ReconciliationPolicy: &ReconciliationPolicySpec{
			Interval: &interval,
			State:    "active",
		},
	}
}
