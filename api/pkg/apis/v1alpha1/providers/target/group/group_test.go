package group

import (
	"context"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestGroupTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := GroupTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}

func TestGroupTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := GroupTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestGroupTargetProviderInitEmptyConfig(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
}

func TestPatchTargetPropertyCopy(t *testing.T) {
	provider := GroupTargetProvider{}
	target := model.TargetState{
		Spec: &model.TargetSpec{
			Properties: map[string]string{
				"ha-set": "ha-set1",
			},
		},
	}
	patch := map[string]string{
		"ha-set": "~COPY_ha-set",
	}
	patchedTarget, err := provider.patchTargetProperty(target, patch, nil, nil, false)
	assert.Nil(t, err)
	assert.Equal(t, "ha-set1", patchedTarget.Spec.Properties["ha-set"])
}

func TestPatchTargetPropertyRemove(t *testing.T) {
	provider := GroupTargetProvider{}
	target := model.TargetState{
		Spec: &model.TargetSpec{
			Properties: map[string]string{
				"ha-set": "ha-set1",
			},
		},
	}
	patch := map[string]string{
		"ha-set": "~REMOVE",
	}
	patchedTarget, err := provider.patchTargetProperty(target, patch, nil, nil, false)
	assert.Nil(t, err)
	assert.NotContains(t, patchedTarget.Spec.Properties, "ha-set")
}

func TestGroupTargetProviderTargetSelector(t *testing.T) {
	// testGroupPatcher := os.Getenv("TEST_GROUP_PATCHER")
	// if testGroupPatcher == "" {
	// 	t.Skip("Skipping because TEST_GROUP_PATCHER enviornment variable is not set")
	// }
	os.Setenv("SYMPHONY_API_URL", "http://localhost:8080/v1alpha2/")
	os.Setenv("USE_SERVICE_ACCOUNT_TOKENS", "false")
	config := GroupTargetProviderConfig{
		User:     "admin",
		Password: "",
	}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "ha-set",
		Type: "group",
		Properties: map[string]interface{}{
			"targetPropertySelector": map[string]interface{}{
				"ha-set": "ha-set1",
				"role":   "member",
			},
			"targetStateSelector": map[string]interface{}{
				"status": "Succeeded",
				"foo":    "barbar",
			},
			"sparePropertySelector": map[string]interface{}{
				"ha-set": "ha-set1",
				"role":   "spare",
			},
			"spareStateSelector": map[string]interface{}{
				"status": "Succeeded",
			},
			"minMatchCount": 2,
			"lowMatchAction": GroupPatchAction{
				SparePatch: map[string]string{
					"ha-sets": "~REMOVE",
					"ha-set":  "ha-set1",
					"role":    "member",
				},
				TargetPatch: map[string]string{
					"ha-sets": "~COPY_ha-set",
					"ha-set":  "~REMOVE",
					"role":    "spare",
				},
			},
			"spareComponents": []model.ComponentSpec{
				{
					Name: "spare-component",
					Properties: map[string]interface{}{
						"foo": "bar3",
					},
				},
			},
			"memberComponents": []model.ComponentSpec{
				{
					Name: "target-component",
					Properties: map[string]interface{}{
						"foo": "bar2",
					},
				},
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "nginx",
			},
			Spec: &model.InstanceSpec{
				Scope: "default",
			},
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
	assert.Nil(t, err)
}
