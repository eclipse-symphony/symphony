/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package adb

import (
	"context"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

func TestInitWithNil(t *testing.T) {
	provider := AdbProvider{}
	err := provider.Init(AdbProviderConfig{})
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name": "name",
	}
	provider := AdbProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestGetEmptyDesired(t *testing.T) {
	testAndroid := os.Getenv("TEST_ANDROID_ADB")
	if testAndroid != "yes" {
		t.Skip("Skipping becasue TEST_ANDROID_ADB is missing or not set to 'yes'")
	}
	provider := AdbProvider{}
	err := provider.Init(AdbProviderConfig{
		Name: "adb",
	})
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{}, nil)
	assert.Equal(t, 0, len(components))
	assert.Nil(t, err)
}

func TestGetOneDesired(t *testing.T) {
	testAndroid := os.Getenv("TEST_ANDROID_ADB")
	if testAndroid != "yes" {
		t.Skip("Skipping becasue TEST_ANDROID_ADB is missing or not set to 'yes'")
	}
	provider := AdbProvider{}
	err := provider.Init(AdbProviderConfig{
		Name: "adb",
	})
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "MyApp",
						Properties: map[string]interface{}{
							model.AppPackage: "com.sec.hiddenmenu",
						},
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: model.ComponentUpdate,
			Component: model.ComponentSpec{
				Name: "MyApp",
				Properties: map[string]interface{}{
					model.AppPackage: "com.sec.hiddenmenu",
				},
			},
		},
	})
	assert.Equal(t, 1, len(components))
	assert.Nil(t, err)
}

func TestGetOneDesiredNotFound(t *testing.T) {
	testAndroid := os.Getenv("TEST_ANDROID_ADB")
	if testAndroid != "yes" {
		t.Skip("Skipping becasue TEST_ANDROID_ADB is missing or not set to 'yes'")
	}
	provider := AdbProvider{}
	err := provider.Init(AdbProviderConfig{
		Name: "adb",
	})
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "MyApp",
						Properties: map[string]interface{}{
							model.AppPackage: "doesnt.exist",
						},
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: model.ComponentUpdate,
			Component: model.ComponentSpec{
				Name: "MyApp",
				Properties: map[string]interface{}{
					model.AppPackage: "doesnt.exist",
				},
			},
		},
	})
	assert.Equal(t, 0, len(components))
	assert.Nil(t, err)
}

func TestApply(t *testing.T) {
	testAndroid := os.Getenv("TEST_ANDROID_ADB")
	if testAndroid != "yes" {
		t.Skip("Skipping becasue TEST_ANDROID_ADB is missing or not set to 'yes'")
	}
	provider := AdbProvider{}
	err := provider.Init(AdbProviderConfig{
		Name: "adb",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "MyApp",
		Properties: map[string]interface{}{
			model.AppPackage: "com.companyname.beacon",
			model.AppImage:   "E:\\projects\\go\\github.com\\torrent-org\\mobile\\Beacon\\Beacon\\bin\\Debug\\net7.0-android\\com.companyname.beacon-Signed.apk",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestRemove(t *testing.T) {
	testAndroid := os.Getenv("TEST_ANDROID_ADB")
	if testAndroid != "yes" {
		t.Skip("Skipping becasue TEST_ANDROID_ADB is missing or not set to 'yes'")
	}
	provider := AdbProvider{}
	err := provider.Init(AdbProviderConfig{
		Name: "adb",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "MyApp",
		Properties: map[string]interface{}{
			model.AppPackage: "com.companyname.beacon",
			model.AppImage:   "E:\\projects\\go\\github.com\\torrent-org\\mobile\\Beacon\\Beacon\\bin\\Debug\\net7.0-android\\com.companyname.beacon-Signed.apk",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestGetFailed(t *testing.T) {
	provider := AdbProvider{}
	err := provider.Init(AdbProviderConfig{
		Name: "adb",
	})
	assert.Nil(t, err)
	_, err = provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)
	assert.Nil(t, err)

	_, err = provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "MyApp",
						Properties: map[string]interface{}{
							model.AppPackage: "doesnt.exist",
						},
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: model.ComponentUpdate,
			Component: model.ComponentSpec{
				Name: "MyApp",
				Properties: map[string]interface{}{
					model.AppPackage: "doesnt.exist",
				},
			},
		},
	})
	assert.NotNil(t, err)
}

func TestApplyFailed(t *testing.T) {
	provider := AdbProvider{}
	err := provider.Init(AdbProviderConfig{
		Name: "adb",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "MyApp",
		Properties: map[string]interface{}{
			model.AppPackage: "com.companyname.beacon",
			model.AppImage:   "E:\\projects\\go\\github.com\\torrent-org\\mobile\\Beacon\\Beacon\\bin\\Debug\\net7.0-android\\com.companyname.beacon-Signed.apk",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)

	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &AdbProvider{}
	err := provider.Init(AdbProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
