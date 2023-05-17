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

package conformance

import (
	"context"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func RequiredPropertiesAndMetadata[P target.ITargetProvider](t *testing.T, p P) {
	desired := []model.ComponentSpec{
		{
			Name:       "test-1",
			Properties: map[string]string{},
			Metadata:   map[string]string{},
		},
	}

	rule := p.GetValidationRule(context.Background())

	for _, property := range rule.RequiredProperties {
		desired[0].Properties[property] = "dummy property"
	}

	for _, metadata := range rule.RequiredMetadata {
		desired[0].Metadata[metadata] = "dummy metadata"
	}

	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: desired,
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   1,
	}
	assert.Nil(t, p.Apply(context.Background(), deployment, true))
}
func AnyRequiredPropertiesMissing[P target.ITargetProvider](t *testing.T, p P) {

	desired := []model.ComponentSpec{
		{
			Name:       "test-1",
			Properties: map[string]string{},
			Metadata:   map[string]string{},
		},
	}

	rule := p.GetValidationRule(context.Background())

	for _, metadata := range rule.RequiredMetadata {
		desired[0].Metadata[metadata] = "dummy metadata"
	}

	for i, _ := range rule.RequiredProperties {
		desired[0].Properties = make(map[string]string, len(rule.RequiredProperties)-1)
		slice := append(append([]string{}, rule.RequiredProperties[:i]...), rule.RequiredProperties[i+1:]...)
		for _, property := range slice {
			desired[0].Properties[property] = "dummy property"
		}
		deployment := model.DeploymentSpec{
			Solution: model.SolutionSpec{
				Components: desired,
			},
			ComponentStartIndex: 0,
			ComponentEndIndex:   1,
		}
		err := p.Apply(context.Background(), deployment, true)
		assert.NotNil(t, err)
		coaErr := err.(v1alpha2.COAError)
		assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
	}
}
func ConformanceSuite[P target.ITargetProvider](t *testing.T, p P) {
	t.Run("Level=Basic", func(t *testing.T) {
		RequiredPropertiesAndMetadata(t, p)
		AnyRequiredPropertiesMissing(t, p)
	})
}
