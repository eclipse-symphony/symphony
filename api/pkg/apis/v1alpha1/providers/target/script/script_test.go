package script

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/require"
)

// TestInitMissingGet tests that we can init with a map fails if get is missing
func TestInitMissingGet(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.InitWithMap(map[string]string{
		"scriptFolder":  ".",
		"stagingFolder": ".",
		"applyScript":   "a",
		"removeScript":  "b",
	})
	require.NotNil(t, err)
}

// TestInitMissingApply tests that we can init with a map fails if apply is missing
func TestInitMissingApply(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.InitWithMap(map[string]string{
		"scriptFolder":  ".",
		"stagingFolder": ".",
		"getScript":     "a",
		"removeScript":  "b",
	})
	require.NotNil(t, err)
}

// TestInitMissingRemove tests that we can init with a map fails if remove is missing
func TestInitMissingRemove(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.InitWithMap(map[string]string{
		"scriptFolder":  ".",
		"stagingFolder": ".",
		"getScript":     "a",
		"applyScript":   "b",
	})
	require.NotNil(t, err)
}

// TestInitWithMap tests that we can init with a map
func TestInitWithMap(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.InitWithMap(map[string]string{
		"name":          "test",
		"needsUpdate":   "mock-needsupdate.sh",
		"needsRemove":   "mock-needsremove.sh",
		"stagingFolder": ".",
		"scriptFolder":  "https://demopolicies.blob.core.windows.net/gatekeeper",
		"applyScript":   "mock-apply.sh",
		"removeScript":  "mock-remove.sh",
		"getScript":     "mock-get.sh",
		"scriptEngine":  "bash",
	})
	require.Nil(t, err)
}

// TestGet tests that we can get a script
func TestGet(t *testing.T) {
	provider := ScriptProvider{}
	currentFolder, _ := filepath.Abs(".")
	err := provider.Init(ScriptProviderConfig{
		ScriptFolder: "",
		GetScript:    filepath.Join(currentFolder, "mock-get.sh"),
	})
	require.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "com1",
				},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(components))
}

// TestNeedsUpdateEmptyScript tests that we can update a script
func TestNeedsUpdateEmptyScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{})
	require.Nil(t, err)
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

	require.Nil(t, err)
	require.False(t, b)
}

// TestNeedsUpdateScript tests that we can update a script
func TestNeedsUpdateScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		NeedsUpdate: "mock-needsupdate.sh",
	})
	require.Nil(t, err)
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

	require.Nil(t, err)
	require.True(t, b)
}

// TestNeedsRemoveEmptyScript tests that we can remove a script
func TestNeedsRemoveEmptyScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{})
	require.Nil(t, err)
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

	require.Nil(t, err)
	require.True(t, b)
}

// TestNeedsRemoveScript tests that we can remove a script
func TestNeedsRemoveScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		NeedsRemove: "mock-needsremove.sh",
	})
	require.Nil(t, err)
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

	require.Nil(t, err)
	require.False(t, b)
}

// TestRemoveScript tests that we can remove a script
func TestRemoveScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		RemoveScript: "mock-remove.sh",
	})
	require.Nil(t, err)
	err = provider.Remove(context.Background(), model.DeploymentSpec{},
		[]model.ComponentSpec{
			{
				Name: "com1",
			},
		})

	require.Nil(t, err)
}

// TestApplyScript tests that we can apply a script
func TestApplyScript(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{
		ApplyScript: "mock-apply.sh",
	})
	require.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{}, false)
	require.Nil(t, err)
}

// TestGetScriptFromUrl tests that we can get a script from a url
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
	require.Nil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{})
	require.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
