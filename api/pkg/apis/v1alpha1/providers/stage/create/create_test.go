/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package create

import (
	"context"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestDeployInstance(t *testing.T) {
	testDeploy := os.Getenv("TEST_DEPLOY_INSTANCE")
	if testDeploy != "yes" {
		t.Skip("Skipping becasue TEST_DEPLOY_INSTANCE is missing or not set to 'yes'")
	}
	provider := CreateStageProvider{}
	err := provider.Init(CreateStageProviderConfig{
		BaseUrl:      "http://localhost:8082/v1alpha2/",
		User:         "admin",
		Password:     "",
		WaitCount:    3,
		WaitInterval: 5,
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "instance",
		"objectName": "redis-server",
		"object": map[string]interface{}{
			"displayName": "redis-server",
			"solution":    "sample-redis",
			"target": map[string]interface{}{
				"name": "sample-docker-target",
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "OK", outputs["status"])
}
