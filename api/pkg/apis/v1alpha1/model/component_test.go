/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponentDeepEqual(t *testing.T) {
	c1 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]interface{}{
			"container.createOptions": "",
			ContainerImage:            "ghcr.io/eclipse-symphony/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	c2 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]interface{}{
			"container.createOptions": "",
			ContainerImage:            "ghcr.io/eclipse-symphony/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
			"nested": map[string]interface{}{
				"key2": "value2", // should match even if order is different
				"key1": "value1",
			},
		},
	}
	equal, err := c1.DeepEquals(c2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestComponentDeepNotEqual(t *testing.T) {
	c1 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]interface{}{
			"container.createOptions": "",
			ContainerImage:            "ghcr.io/eclipse-symphony/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.1", //difference is here!
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
		},
	}
	c2 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]interface{}{
			"container.createOptions": "",
			ContainerImage:            "ghcr.io/eclipse-symphony/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
		},
	}
	equal, err := c1.DeepEquals(c2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestComponentNestedDeepNotEqual(t *testing.T) {
	c1 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]interface{}{
			"container.createOptions": "",
			ContainerImage:            "ghcr.io/eclipse-symphony/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_CLIENT_SECRET": "${{$secret(key,value)}}",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	c2 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]interface{}{
			"container.createOptions": "",
			ContainerImage:            "ghcr.io/eclipse-symphonyony/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_CLIENT_SECRET": "${{$secret(key,value)}}",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
			"nested": map[string]interface{}{
				"key1": "value1", // key2 is missing
			},
		},
	}
	equal, err := c1.DeepEquals(c2)
	assert.Nil(t, err)
	assert.False(t, equal)

	equal, err = c2.DeepEquals(c1)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestComponentDeepEqualEmpty(t *testing.T) {
	component1 := ComponentSpec{
		Name: "symphony-agent",
	}
	res, err := component1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a ComponentSpec type")
	assert.False(t, res)
}

func TestComponentDeepNotEqualMetadata(t *testing.T) {
	c1 := ComponentSpec{
		Name: "symphony-agent",
		Metadata: map[string]string{
			"key1": "value1",
		},
		Properties: map[string]interface{}{
			"container.createOptions": "",
			ContainerImage:            "ghcr.io/eclipse-symphony/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
		},
	}
	c2 := ComponentSpec{
		Name: "symphony-agent",
		Metadata: map[string]string{
			"key1": "value2", // difference is here!
		},
		Properties: map[string]interface{}{
			"container.createOptions": "",
			ContainerImage:            "ghcr.io/eclipse-symphony/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
		},
	}

	equal, err := c1.DeepEquals(c2)
	assert.Nil(t, err)
	assert.False(t, equal)
}
