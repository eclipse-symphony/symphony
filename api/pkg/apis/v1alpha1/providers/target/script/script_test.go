/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package script

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitMissingGet tests that we can init with a map fails if get is missing
// func TestInitMissingGet(t *testing.T) {
// 	provider := ScriptProvider{}
// 	err := provider.InitWithMap(map[string]string{
// 		"scriptFolder":  ".",
// 		"stagingFolder": ".",
// 		"applyScript":   "a",
// 		"removeScript":  "b",
// 	})
// 	require.NotNil(t, err)
// }

// // TestInitMissingApply tests that we can init with a map fails if apply is missing
// func TestInitMissingApply(t *testing.T) {
// 	provider := ScriptProvider{}
// 	err := provider.InitWithMap(map[string]string{
// 		"scriptFolder":  ".",
// 		"stagingFolder": ".",
// 		"getScript":     "a",
// 		"removeScript":  "b",
// 	})
// 	require.NotNil(t, err)
// }

// // TestInitMissingRemove tests that we can init with a map fails if remove is missing
// func TestInitMissingRemove(t *testing.T) {
// 	provider := ScriptProvider{}
// 	err := provider.InitWithMap(map[string]string{
// 		"scriptFolder":  ".",
// 		"stagingFolder": ".",
// 		"getScript":     "a",
// 		"applyScript":   "b",
// 	})
// 	require.NotNil(t, err)
// 	assert.Equal(t, err.Error(), "Bad Config: invalid script provider config, exptected 'removeScript'")
// }

// // TestInitWithMap tests that we can init with a map
// func TestInitWithMap(t *testing.T) {
// 	provider := ScriptProvider{}
// 	err := provider.InitWithMap(map[string]string{
// 		"name":          "test",
// 		"needsUpdate":   "mock-needsupdate.sh",
// 		"needsRemove":   "mock-needsremove.sh",
// 		"stagingFolder": "./staging",
// 		"scriptFolder":  "https://raw.githubusercontent.com/eclipse-symphony/symphony/main/docs/samples/script-provider",
// 		"applyScript":   "mock-apply.sh",
// 		"removeScript":  "mock-remove.sh",
// 		"getScript":     "mock-get.sh",
// 		"scriptEngine":  "bash",
// 	})
// 	require.Nil(t, err)
// }

// // TestGet tests that we can get a script
// func TestGet(t *testing.T) {
// 	provider := ScriptProvider{}
// 	currentFolder, _ := filepath.Abs(".")
// 	err := provider.Init(ScriptProviderConfig{
// 		ScriptFolder: "",
// 		GetScript:    filepath.Join(currentFolder, "mock-get.sh"),
// 	})
// 	require.Nil(t, err)
// 	components, err := provider.Get(context.Background(), model.DeploymentSpec{
// 		Solution: model.SolutionState{
// 			Spec: &model.SolutionSpec{
// 				Components: []model.ComponentSpec{
// 					{
// 						Name: "com1",
// 					},
// 				},
// 			},
// 		},
// 		Instance: model.InstanceState{
// 			Spec: &model.InstanceSpec{
// 				Scope: "test-scope",
// 			},
// 		},
// 	}, []model.ComponentStep{
// 		{
// 			Action: model.ComponentUpdate,
// 			Component: model.ComponentSpec{
// 				Name: "com1",
// 			},
// 		},
// 	})

// 	assert.Nil(t, err)
// 	assert.Equal(t, 1, len(components))
// }

// // TestRemoveScript tests that we can remove a script
// func TestRemoveScript(t *testing.T) {
// 	provider := ScriptProvider{}
// 	err := provider.Init(ScriptProviderConfig{
// 		RemoveScript: "mock-remove.sh",
// 	})
// 	assert.Nil(t, err)
// 	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
// 		Solution: model.SolutionState{
// 			Spec: &model.SolutionSpec{
// 				Components: []model.ComponentSpec{
// 					{
// 						Name: "com1",
// 					},
// 				},
// 			},
// 		},
// 		Instance: model.InstanceState{
// 			Spec: &model.InstanceSpec{
// 				Scope: "test-scope",
// 			},
// 		},
// 	}, model.DeploymentStep{
// 		Components: []model.ComponentStep{
// 			{
// 				Action: model.ComponentDelete,
// 				Component: model.ComponentSpec{
// 					Name: "com1",
// 				},
// 			},
// 		},
// 	}, false)
// 	assert.Nil(t, err)
// }

// // TestApplyScript tests that we can apply a script
// func TestApplyScript(t *testing.T) {
// 	provider := ScriptProvider{}
// 	err := provider.Init(ScriptProviderConfig{
// 		ApplyScript: "mock-apply.sh",
// 	})
// 	assert.Nil(t, err)
// 	results, err := provider.Apply(context.Background(), model.DeploymentSpec{
// 		Solution: model.SolutionState{
// 			Spec: &model.SolutionSpec{
// 				Components: []model.ComponentSpec{
// 					{
// 						Name: "com1",
// 					},
// 				},
// 			},
// 		},
// 		Instance: model.InstanceState{
// 			Spec: &model.InstanceSpec{
// 				Scope: "test-scope",
// 			},
// 		},
// 	}, model.DeploymentStep{
// 		Components: []model.ComponentStep{
// 			{
// 				Action: model.ComponentUpdate,
// 				Component: model.ComponentSpec{
// 					Name: "com1",
// 					Parameters: map[string]string{
// 						"path": "echo",
// 						"args": "hello",
// 					},
// 				},
// 			},
// 		},
// 	}, false)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 1, len(results))
// }

func TestApplyScriptWithoutPath(t *testing.T) {
	provider := ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{})
	assert.Nil(t, err)
	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "com1",
					},
				},
			},
		},
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				Scope: "test-scope",
			},
		},
	}, model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action: model.ComponentUpdate,
				Component: model.ComponentSpec{
					Name: "com1",
				},
			},
		},
	}, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "expected 'path'")
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &ScriptProvider{}
	err := provider.Init(ScriptProviderConfig{})
	require.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
