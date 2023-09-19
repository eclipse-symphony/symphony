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
			ContainerImage:            "possprod.azurecr.io/symphony-agent:0.39.9",
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
			ContainerImage:            "possprod.azurecr.io/symphony-agent:0.39.9",
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
			ContainerImage:            "possprod.azurecr.io/symphony-agent:0.39.9",
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
			ContainerImage:            "possprod.azurecr.io/symphony-agent:0.39.9",
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
			ContainerImage:            "possprod.azurecr.io/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_CLIENT_SECRET": "$secret(key,value)",
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
			ContainerImage:            "possprod.azurecr.io/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_CLIENT_SECRET": "$secret(key,value)",
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
