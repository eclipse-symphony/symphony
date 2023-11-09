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

package k8s

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/stretchr/testify/assert"
)

func TestK8sStateProviderConfigFromMapNil(t *testing.T) {
	_, err := K8sStateProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestK8sStateProviderConfigFromMapEmpty(t *testing.T) {
	_, err := K8sStateProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestInitWithBadConfigType(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "Bad",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyFile(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "path",
	}
	provider := K8sStateProvider{}
	provider.Init(config)
	// assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}
func TestInitWithBadFile(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyData(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "bytes",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithBadData(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "bytes",
		ConfigData: "bad data",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

func TestUpSert(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	provider := K8sStateProvider{}
	err := provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s123",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "s123",
				},
				"spec": model.TargetSpec{
					Properties: map[string]string{
						"foo": "bar2",
					},
				},
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "s123", id)
}

func TestList(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	provider := K8sStateProvider{}
	err := provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s123",
			Body: model.TargetSpec{
				Properties: map[string]string{
					"foo": "bar2",
				},
			},
		},
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	assert.Nil(t, err)
	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(entries))
	assert.Equal(t, "s123", entries[0].ID)
}

func TestDelete(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	provider := K8sStateProvider{}
	err := provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s123",
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "s123",
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	assert.Nil(t, err)
}
func TestGet(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	provider := K8sStateProvider{}
	err := provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s123",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "s123",
				},
				"spec": model.TargetSpec{
					Properties: map[string]string{
						"foo": "bar2",
					},
				},
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	assert.Nil(t, err)
	item, err := provider.Get(context.Background(), states.GetRequest{
		ID: "s123",
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targetds",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "s123", item.ID)
}
func TestUpSertWithState(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	provider := K8sStateProvider{}
	err := provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s234",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "s234",
				},
				"spec": model.TargetSpec{
					Properties: map[string]string{
						"foo": "bar2",
					},
				},
				"status": map[string]interface{}{
					"properties": map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "s234", id)
}
func TestUpSertWithStateOnly(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	provider := K8sStateProvider{}
	err := provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s234",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup,
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "s234",
				},
				"status": map[string]interface{}{
					"properties": map[string]string{
						"foo": "bar2",
					},
				},
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "s234", id)
}
