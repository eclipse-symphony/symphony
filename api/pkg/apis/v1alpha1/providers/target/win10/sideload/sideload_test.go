package sideload

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

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
	components, err := provider.Get(context.Background(), model.DeploymentSpec{})
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
	})
	assert.Equal(t, 1, len(components))
	assert.Nil(t, err)
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
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "HomeHub_1.0.4.0_x64",
				},
			},
		},
	}, []model.ComponentSpec{
		{
			Name: "HomeHub_1.0.4.0_x64__gjd5ncee18d88",
		},
	})
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
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "HomeHub_1.0.4.0_x64",
					Properties: map[string]string{
						"app.package.path": "E:\\projects\\go\\github.com\\azure\\symphony-docs\\samples\\scenarios\\homehub\\HomeHub\\HomeHub.Package\\AppPackages\\HomeHub.Package_1.0.4.0_Debug_Test\\HomeHub.Package_1.0.4.0_x64_Debug.appxbundle",
					},
				},
			},
		},
	}, false)
	assert.Nil(t, err)
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
