package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponentDeepEqual(t *testing.T) {
	c1 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]string{
			"container.createOptions": "",
			"container.image":         "possprod.azurecr.io/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_CLIENT_SECRET": "\\u003cSP Client Secret\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
		},
	}
	c2 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]string{
			"container.createOptions": "",
			"container.image":         "possprod.azurecr.io/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_CLIENT_SECRET": "\\u003cSP Client Secret\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
		},
	}
	equal, err := c1.DeepEquals(c2)
	assert.Nil(t, err)
	assert.True(t, equal)
}
func TestComponentDeepNotEqual(t *testing.T) {
	c1 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]string{
			"container.createOptions": "",
			"container.image":         "possprod.azurecr.io/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.1", //difference is here!
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_CLIENT_SECRET": "\\u003cSP Client Secret\\u003e",
			"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
			"env.STORAGE_ACCOUNT":     "voestore",
			"env.STORAGE_CONTAINER":   "snapshots",
			"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
			"env.TARGET_NAME":         "symphony-k8s-target",
		},
	}
	c2 := ComponentSpec{
		Name: "symphony-agent",
		Properties: map[string]string{
			"container.createOptions": "",
			"container.image":         "possprod.azurecr.io/symphony-agent:0.39.9",
			"container.restartPolicy": "always",
			"container.type":          "docker",
			"container.version":       "1.0",
			"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
			"env.AZURE_CLIENT_SECRET": "\\u003cSP Client Secret\\u003e",
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
