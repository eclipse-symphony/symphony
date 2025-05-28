/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package conformance

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func RequiredPropertiesAndMetadata[P target.ITargetProvider](t *testing.T, p P) {
	desired := []model.ComponentSpec{
		{
			Name:       "test-1",
			Properties: map[string]interface{}{},
			Metadata:   map[string]string{},
		},
	}

	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Component: model.ComponentSpec{
					Name:       "test-1",
					Properties: map[string]interface{}{},
					Metadata:   map[string]string{},
				},
			},
		},
	}

	rule := p.GetValidationRule(context.Background())

	for _, property := range rule.ComponentValidationRule.RequiredProperties {
		desired[0].Properties[property] = "dummy property"
		step.Components[0].Component.Properties[property] = "dummy property"
	}

	for _, metadata := range rule.ComponentValidationRule.RequiredMetadata {
		desired[0].Metadata[metadata] = "dummy metadata"
		step.Components[0].Component.Metadata[metadata] = "dummy metadata"
	}

	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: desired,
			},
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   1,
	}
	_, err := p.Apply(context.Background(), deployment, step, true)
	assert.Nil(t, err)
}
func AnyRequiredPropertiesMissing[P target.ITargetProvider](t *testing.T, p P) {

	desired := []model.ComponentSpec{
		{
			Name:       "test-1",
			Properties: map[string]interface{}{},
			Metadata:   map[string]string{},
		},
	}

	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Component: model.ComponentSpec{
					Name:       "test-1",
					Properties: map[string]interface{}{},
					Metadata:   map[string]string{},
				},
			},
		},
	}

	rule := p.GetValidationRule(context.Background())

	for _, metadata := range rule.ComponentValidationRule.RequiredMetadata {
		desired[0].Metadata[metadata] = "dummy metadata"
	}

	for i, _ := range rule.ComponentValidationRule.RequiredProperties {
		desired[0].Properties = make(map[string]interface{}, len(rule.ComponentValidationRule.RequiredProperties)-1)
		slice := append(append([]string{}, rule.ComponentValidationRule.RequiredProperties[:i]...), rule.ComponentValidationRule.RequiredProperties[i+1:]...)
		for _, property := range slice {
			desired[0].Properties[property] = "dummy property"
		}
		deployment := model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{},
			},
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
					Components: desired,
				},
			},
			ComponentStartIndex: 0,
			ComponentEndIndex:   1,
		}
		_, err := p.Apply(context.Background(), deployment, step, true)
		assert.NotNil(t, err)
		coaErr := err.(v1alpha2.COAError)
		condition := coaErr.State == v1alpha2.BadRequest || coaErr.State == v1alpha2.ValidateFailed
		assert.True(t, condition, "Expected coaErr.State to be either BadRequest or ValidateFailed, but got %v", coaErr.State)
	}
}
func ConformanceSuite[P target.ITargetProvider](t *testing.T, p P) {
	t.Run("Level=Basic", func(t *testing.T) {
		RequiredPropertiesAndMetadata(t, p)
		AnyRequiredPropertiesMissing(t, p)
	})
}
