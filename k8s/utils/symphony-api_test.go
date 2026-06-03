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

	solutionversion_v1 "gopls-workspace/apis/solution/v1"
)

func TestK8SSolutionVersionToAPISolutionVersionState(t *testing.T) {
	solutionversionYaml := `apiVersion: solution.symphony/v1
kind: SolutionVersion
metadata: 
  name: sample-staged-solutionversion
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
	solutionversion := &solutionversion_v1.SolutionVersion{}
	err := yaml.Unmarshal([]byte(solutionversionYaml), solutionversion)
	assert.NoError(t, err)

	expectedProperties := map[string]interface{}{
		"foo": "bar",
		"bar": map[string]interface{}{
			"baz": "qux",
		},
	}
	actualProperties := map[string]interface{}{}
	err = json.Unmarshal(solutionversion.Spec.Components[0].Properties.Raw, &actualProperties)
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
	err = json.Unmarshal(solutionversion.Spec.Components[0].Sidecars[0].Properties.Raw, &actualSidecarProperties)
	assert.NoError(t, err)

	assert.Equal(t, expectedSidecarProperties, actualSidecarProperties)

	apiSolutionVersionState, err := K8SSolutionVersionToAPISolutionVersionState(*solutionversion)

	assert.NoError(t, err)
	assert.Equal(t, solutionversion.Name, apiSolutionVersionState.ObjectMeta.Name)
	assert.Equal(t, solutionversion.Spec.Components[0].Name, apiSolutionVersionState.Spec.Components[0].Name)
	assert.Equal(t, solutionversion.Spec.Components[0].Type, apiSolutionVersionState.Spec.Components[0].Type)
	assert.Equal(t, expectedProperties, apiSolutionVersionState.Spec.Components[0].Properties)

	assert.Equal(t, solutionversion.Spec.Components[0].Sidecars[0].Name, apiSolutionVersionState.Spec.Components[0].Sidecars[0].Name)
	assert.Equal(t, solutionversion.Spec.Components[0].Sidecars[0].Type, apiSolutionVersionState.Spec.Components[0].Sidecars[0].Type)
	assert.Equal(t, expectedSidecarProperties, apiSolutionVersionState.Spec.Components[0].Sidecars[0].Properties)
}

func TestK8SSolutionVersionToAPISolutionVersionStateNullProperty(t *testing.T) {
	solutionversionYaml := `apiVersion: solution.symphony/v1
kind: SolutionVersion
metadata: 
  name: sample-staged-solutionversion
spec:  
  components:
  - name: staged-component
    sidecars:
    - name: sidecar1
      type: container
      properties:
        container.image: "symphony/sidecar"
        env.foo: "bar"
        nestedEnv:
          baz: "qux"
`
	solutionversion := &solutionversion_v1.SolutionVersion{}
	err := yaml.Unmarshal([]byte(solutionversionYaml), solutionversion)
	assert.NoError(t, err)

	apiSolutionVersionState, err := K8SSolutionVersionToAPISolutionVersionState(*solutionversion)

	assert.NoError(t, err)
	assert.Equal(t, solutionversion.Name, apiSolutionVersionState.ObjectMeta.Name)
	assert.Equal(t, solutionversion.Spec.Components[0].Name, apiSolutionVersionState.Spec.Components[0].Name)
	assert.Equal(t, solutionversion.Spec.Components[0].Type, apiSolutionVersionState.Spec.Components[0].Type)
}

func TestK8SSidecarSpecToAPISidecarSpecNullProperty(t *testing.T) {
	solutionversionYaml := `apiVersion: solution.symphony/v1
kind: SolutionVersion
metadata: 
  name: sample-staged-solutionversion
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
`
	solutionversion := &solutionversion_v1.SolutionVersion{}
	err := yaml.Unmarshal([]byte(solutionversionYaml), solutionversion)
	assert.NoError(t, err)

	_, err = K8SSidecarSpecToAPISidecarSpec(solutionversion.Spec.Components[0].Sidecars[0])
	assert.NoError(t, err)
}
