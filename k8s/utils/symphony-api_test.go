/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"

	solution_v1 "gopls-workspace/apis/solution/v1"
)

func TestK8SSolutionToAPISolutionState(t *testing.T) {
	solutionYaml := `apiVersion: solution.symphony/v1
kind: Solution
metadata: 
  name: sample-staged-solution
spec:  
  components:
  - name: staged-component
    properties:
      foo: "bar"
      bar:
        baz: "qux"
    sidecars:
    - name: sidecar1
      type: container
      properties:
        container.image: "symphony/sidecar"
        env.foo: "bar"
        nestedEnv:
          baz: "qux"
`
	solution := &solution_v1.Solution{}
	err := yaml.Unmarshal([]byte(solutionYaml), solution)
	assert.NoError(t, err)

	expectedProperties := map[string]interface{}{
		"foo": "bar",
		"bar": map[string]interface{}{
			"baz": "qux",
		},
	}
	actualProperties := map[string]interface{}{}
	err = json.Unmarshal(solution.Spec.Components[0].Properties.Raw, &actualProperties)
	assert.NoError(t, err)

	assert.Equal(t, expectedProperties, actualProperties)

	expectedSidecarProperties := map[string]interface{}{
		"container.image": "symphony/sidecar",
		"env.foo":         "bar",
		"nestedEnv": map[string]interface{}{
			"baz": "qux",
		},
	}
	actualSidecarProperties := map[string]interface{}{}
	err = json.Unmarshal(solution.Spec.Components[0].Sidecars[0].Properties.Raw, &actualSidecarProperties)
	assert.NoError(t, err)

	assert.Equal(t, expectedSidecarProperties, actualSidecarProperties)

	apiSolutionState, err := K8SSolutionToAPISolutionState(*solution)

	assert.NoError(t, err)
	assert.Equal(t, solution.Name, apiSolutionState.ObjectMeta.Name)
	assert.Equal(t, solution.Spec.Components[0].Name, apiSolutionState.Spec.Components[0].Name)
	assert.Equal(t, solution.Spec.Components[0].Type, apiSolutionState.Spec.Components[0].Type)
	assert.Equal(t, expectedProperties, apiSolutionState.Spec.Components[0].Properties)

	assert.Equal(t, solution.Spec.Components[0].Sidecars[0].Name, apiSolutionState.Spec.Components[0].Sidecars[0].Name)
	assert.Equal(t, solution.Spec.Components[0].Sidecars[0].Type, apiSolutionState.Spec.Components[0].Sidecars[0].Type)
	assert.Equal(t, expectedSidecarProperties, apiSolutionState.Spec.Components[0].Sidecars[0].Properties)
}
