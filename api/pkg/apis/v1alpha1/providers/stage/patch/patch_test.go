/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package patch

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

func TestPatchSolution(t *testing.T) {
	testPatchSolution := os.Getenv("TEST_PATCH_SOLUTION")
	if testPatchSolution != "yes" {
		t.Skip("Skipping becasue TEST_PATCH_SOLUTION is missing or not set to 'yes'")
	}
	provider := PatchStageProvider{}
	err := provider.Init(PatchStageProviderConfig{
		BaseUrl:  "http://localhost:8082/v1alpha2/",
		User:     "admin",
		Password: "",
	})

	provider.SetContext(&contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":   "solution",
		"objectName":   "test-app",
		"patchSource":  "catalog",
		"patchContent": "ai-config",
		"component":    "frontend",
		"property":     "deployment.replicas",
		"subKey":       "",
		"dedupKey":     "flavor",
		"patchAction":  "add",
	})
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, "OK", outputs["status"])
}
func TestPatchSolutionWholeComponent(t *testing.T) {
	testPatchSolution := os.Getenv("TEST_PATCH_SOLUTION")
	if testPatchSolution != "yes" {
		t.Skip("Skipping becasue TEST_PATCH_SOLUTION is missing or not set to 'yes'")
	}
	provider := PatchStageProvider{}
	err := provider.Init(PatchStageProviderConfig{
		BaseUrl:  "http://localhost:8082/v1alpha2/",
		User:     "admin",
		Password: "",
	})

	provider.SetContext(&contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":  "solution",
		"objectName":  "test-app",
		"patchSource": "inline",
		"patchContent": model.ComponentSpec{
			Name: "test-ingress2",
			Type: "ingress",
			Properties: map[string]interface{}{
				"ingressClassName": "nginx",
				"rules": []map[string]interface{}{
					{
						"http": map[string]interface{}{
							"paths": []interface{}{
								map[string]interface{}{
									"path":     "/testpath",
									"backend":  map[string]interface{}{"serviceName": "test-app", "servicePort": 100 + 200},
									"pathType": "Prefix",
								},
							},
						},
					},
				},
			},
		},
		"patchAction": "add",
	})
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, "OK", outputs["status"])
}
