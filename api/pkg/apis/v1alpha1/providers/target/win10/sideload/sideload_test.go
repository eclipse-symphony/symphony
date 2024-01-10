/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package sideload

import (
	"context"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name":                "name",
		"ipAddress":           "0.0.0.0",
		"winAppDeployCmdPath": "c:\\Program Files (x86)",
		"pin":                 "pin",
		"networkUser":         "user",
		"networkPassword":     "",
		"silent":              "true",
	}
	provider := Win10SideLoadProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}
func TestGetEmptyDesired(t *testing.T) {
	testWin10 := os.Getenv("TEST_WIN10_SIDELOAD")
	if testWin10 != "yes" {
		t.Skip("Skipping becasue TEST_WIN10_SIDELOAD is missing or not set to 'yes'")
	}
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name:                "win10sideload",
		IPAddress:           "192.168.50.55",
		WinAppDeployCmdPath: "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
	})
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{}, nil)
	assert.Equal(t, 0, len(components))
	assert.Nil(t, err)
}
func TestGetOneDesired(t *testing.T) {
	testWin10 := os.Getenv("TEST_WIN10_SIDELOAD")
	if testWin10 != "yes" {
		t.Skip("Skipping becasue TEST_WIN10_SIDELOAD is missing or not set to 'yes'")
	}
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name:                "win10sideload",
		IPAddress:           "192.168.50.55",
		WinAppDeployCmdPath: "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
	})
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "HomeHub_1.0.4.0_x64",
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "HomeHub_1.0.4.0_x64",
			},
		},
	})
	assert.Equal(t, 1, len(components))
	assert.Nil(t, err)
}
func TestNeedUpdate(t *testing.T) {
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name: "win10sideload",
	})
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
	}
	componentDesired := []model.ComponentSpec{component}
	componentCurrent := []model.ComponentSpec{component}
	res := provider.NeedsUpdate(context.Background(), componentDesired, componentCurrent)
	assert.False(t, res)
}
func TestNeedRemove(t *testing.T) {
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name: "win10sideload",
	})
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
	}
	componentDesired := []model.ComponentSpec{component}
	componentCurrent := []model.ComponentSpec{component}
	res := provider.NeedsRemove(context.Background(), componentDesired, componentCurrent)
	assert.True(t, res)
}
func TestRemove(t *testing.T) {
	testWin10 := os.Getenv("TEST_WIN10_SIDELOAD")
	if testWin10 != "yes" {
		t.Skip("Skipping becasue TEST_WIN10_SIDELOAD is missing or not set to 'yes'")
	}
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name:                "win10sideload",
		IPAddress:           "192.168.50.55",
		WinAppDeployCmdPath: "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "delete",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}
func TestApply(t *testing.T) {
	testWin10 := os.Getenv("TEST_WIN10_SIDELOAD")
	if testWin10 != "yes" {
		t.Skip("Skipping becasue TEST_WIN10_SIDELOAD is missing or not set to 'yes'")
	}
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name:                "win10sideload",
		IPAddress:           "192.168.50.55",
		WinAppDeployCmdPath: "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
		Properties: map[string]interface{}{
			"app.package.path": "E:\\projects\\go\\github.com\\azure\\symphony-docs\\samples\\scenarios\\homehub\\HomeHub\\HomeHub.Package\\AppPackages\\HomeHub.Package_1.0.4.0_Debug_Test\\HomeHub.Package_1.0.4.0_x64_Debug.appxbundle",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "update",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}
func TestApplyUpdateFailed(t *testing.T) {
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name:                "win10sideload",
		IPAddress:           "0.0.0.0",
		WinAppDeployCmdPath: "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
		Pin:                 "pin",
	})
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
		Properties: map[string]interface{}{
			"app.package.path": "C:\\",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "update",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}
func TestApplySlientDelete(t *testing.T) {
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name:                "win10sideload",
		IPAddress:           "0.0.0.0",
		WinAppDeployCmdPath: "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
		Silent:              true,
		Pin:                 "pin",
	})
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
		Properties: map[string]interface{}{
			"app.package.path": "C:\\",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "delete",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}
func TestGetUnknownPath(t *testing.T) {
	provider := Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name:                "win10sideload",
		IPAddress:           "0.0.0.0",
		WinAppDeployCmdPath: "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
		Silent:              true,
		Pin:                 "pin",
	})
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
		Properties: map[string]interface{}{
			"app.package.path": "C:\\",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
	}
	step := []model.ComponentStep{{
		Action:    "delete",
		Component: component,
	}}
	_, err = provider.Get(context.Background(), deployment, step)
	assert.NotNil(t, err)
}
func TestConformanceSuite(t *testing.T) {
	provider := &Win10SideLoadProvider{}
	err := provider.Init(Win10SideLoadProviderConfig{
		Name:                "win10sideload",
		IPAddress:           "192.168.50.55",
		WinAppDeployCmdPath: "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
	})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
