package adb

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

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
	components, err := provider.Get(context.Background(), model.DeploymentSpec{})
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
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "MyApp",
					Properties: map[string]string{
						model.AppPackage: "com.sec.hiddenmenu",
					},
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
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "MyApp",
					Properties: map[string]string{
						model.AppPackage: "doesnt.exist",
					},
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
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "MyApp",
					Properties: map[string]string{
						model.AppPackage: "com.companyname.beacon",
						model.AppImage:   "E:\\projects\\go\\github.com\\torrent-org\\mobile\\Beacon\\Beacon\\bin\\Debug\\net7.0-android\\com.companyname.beacon-Signed.apk",
					},
				},
			},
		},
	}, false)
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
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "MyApp",
					Properties: map[string]string{
						model.AppPackage: "com.companyname.beacon",
						model.AppImage:   "E:\\projects\\go\\github.com\\torrent-org\\mobile\\Beacon\\Beacon\\bin\\Debug\\net7.0-android\\com.companyname.beacon-Signed.apk",
					},
				},
			},
		},
	}, nil)
	assert.Nil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &AdbProvider{}
	err := provider.Init(AdbProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
