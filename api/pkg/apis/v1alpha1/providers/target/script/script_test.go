package script

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

func TestInitMissingGet(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.InitWithMap(map[string]string{
		"scriptFolder":  ".",
		"stagingFolder": ".",
		"applyScript":   "a",
		"removeScript":  "b",
	})
	assert.NotNil(t, err)
}

func TestInitMissingApply(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.InitWithMap(map[string]string{
		"scriptFolder":  ".",
		"stagingFolder": ".",
		"getScript":     "a",
		"removeScript":  "b",
	})
	assert.NotNil(t, err)
}
func TestInitMissingRemove(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.InitWithMap(map[string]string{
		"scriptFolder":  ".",
		"stagingFolder": ".",
		"getScript":     "a",
		"applyScript":   "b",
	})
	assert.NotNil(t, err)
}
func TestGet(t *testing.T) {
	provider := ScriptProvider{}
	currentFolder, _ := filepath.Abs(".")
	err := provider.Init(ScriptProviderConfig{
		ScriptFolder: "",
		GetScript:    filepath.Join(currentFolder, "mock-get.sh"),
	})
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "com1",
				},
			},
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
}
func TestNeedsUpdateEmptyScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{})
	assert.Nil(t, err)
	b := provider.NeedsUpdate(context.Background(), []model.ComponentSpec{
		{
			Name: "com1",
		},
	},
		[]model.ComponentSpec{
			{
				Name: "com1",
			},
		})

	assert.Nil(t, err)
	assert.False(t, b)
}
func TestNeedsUpdateScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		NeedsUpdate: "mock-needsupdate.sh",
	})
	assert.Nil(t, err)
	b := provider.NeedsUpdate(context.Background(), []model.ComponentSpec{
		{
			Name: "com1",
		},
	},
		[]model.ComponentSpec{
			{
				Name: "com1",
			},
		})

	assert.Nil(t, err)
	assert.True(t, b)
}
func TestNeedsRemoveEmptyScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{})
	assert.Nil(t, err)
	b := provider.NeedsRemove(context.Background(), []model.ComponentSpec{
		{
			Name: "com1",
		},
	},
		[]model.ComponentSpec{
			{
				Name: "com1",
			},
		})

	assert.Nil(t, err)
	assert.True(t, b)
}
func TestNeedsRemoveScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		NeedsRemove: "mock-needsremove.sh",
	})
	assert.Nil(t, err)
	b := provider.NeedsRemove(context.Background(), []model.ComponentSpec{
		{
			Name: "com1",
		},
	},
		[]model.ComponentSpec{
			{
				Name: "com1",
			},
		})

	assert.Nil(t, err)
	assert.False(t, b)
}
func TestRemoveScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		RemoveScript: "mock-remove.sh",
	})
	assert.Nil(t, err)
	err = provider.Remove(context.Background(), model.DeploymentSpec{},
		[]model.ComponentSpec{
			{
				Name: "com1",
			},
		})

	assert.Nil(t, err)
}

func TestApplyScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		ApplyScript: "mock-apply.sh",
	})
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{}, false)
	assert.Nil(t, err)
}

func TestGetScriptFromUrl(t *testing.T) {
	testScriptProvider := os.Getenv("TEST_SCRIPT_PROVIDER")
	if testScriptProvider == "" {
		t.Skip("Skipping because TEST_SCRIPT_PROVIDER environment variable is not set")
	}
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		GetScript:    "mock-get.sh",
		ApplyScript:  "mock-apply.sh",
		RemoveScript: "mock-remove.sh",
		ScriptFolder: "https://demopolicies.blob.core.windows.net/gatekeeper",
	})
	assert.Nil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
