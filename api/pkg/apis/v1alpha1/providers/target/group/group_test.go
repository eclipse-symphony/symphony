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
	assert.NotNil(t, err)
}

func TestGroupTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := GroupTargetProviderConfigFromMap(map[string]string{})
	assert.NotNil(t, err)
}

func TestGroupTargetProviderInitEmptyConfig(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
}

func TestAComponentAssignmentsSingle(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components := []model.ComponentSpec{
		{
			Name:       "component1",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
	}
	targets := []model.TargetState{
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target1",
				Namespace: "default",
			},
		},
	}
	assignments, err := provider.assignComponents(components, targets)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assignments))
	assert.Equal(t, 1, len(assignments["target1"]))
	assert.Equal(t, "component1", assignments["target1"][0])
}

func TestAComponentAssignmentsDouble(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components := []model.ComponentSpec{
		{
			Name:       "component1",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
		{
			Name:       "component2",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
	}
	targets := []model.TargetState{
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target1",
				Namespace: "default",
			},
		},
	}
	assignments, err := provider.assignComponents(components, targets)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assignments))
	assert.Equal(t, 2, len(assignments["target1"]))
	assert.Equal(t, "component1", assignments["target1"][0])
	assert.Equal(t, "component2", assignments["target1"][1])
}

func TestAComponentAssignmentsSingleTwoTargets(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components := []model.ComponentSpec{
		{
			Name:       "component1",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
	}
	targets := []model.TargetState{
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target1",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target2",
				Namespace: "default",
			},
		},
	}
	assignments, err := provider.assignComponents(components, targets)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(assignments))
	assert.Equal(t, 1, len(assignments["target1"]))
	assert.Equal(t, "component1", assignments["target1"][0])
	assert.Equal(t, 0, len(assignments["target2"]))
}

func TestAComponentAssignmentsDoubleTwoTargets(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components := []model.ComponentSpec{
		{
			Name:       "component1",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
		{
			Name:       "component2",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
	}
	targets := []model.TargetState{
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target1",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target2",
				Namespace: "default",
			},
		},
	}
	assignments, err := provider.assignComponents(components, targets)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(assignments))
	assert.Equal(t, 1, len(assignments["target1"]))
	assert.Equal(t, 1, len(assignments["target2"]))
	assert.Equal(t, "component1", assignments["target1"][0])
	assert.Equal(t, "component2", assignments["target2"][0])
}

func TestAComponentAssignmentsDoubleThreeTargets(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components := []model.ComponentSpec{
		{
			Name:       "component1",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
		{
			Name:       "component2",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
	}
	targets := []model.TargetState{
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target1",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target2",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target3",
				Namespace: "default",
			},
		},
	}
	assignments, err := provider.assignComponents(components, targets)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(assignments))
	assert.Equal(t, 1, len(assignments["target1"]))
	assert.Equal(t, 1, len(assignments["target2"]))
	assert.Equal(t, 0, len(assignments["target3"]))
	assert.Equal(t, "component1", assignments["target1"][0])
	assert.Equal(t, "component2", assignments["target2"][0])
}

func TestAComponentAssignmentsDoubleThreeTargetsExisting(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components := []model.ComponentSpec{
		{
			Name:       "component1",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
		{
			Name:       "component2",
			Type:       "group",
			Properties: map[string]interface{}{},
		},
	}
	targets := []model.TargetState{
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target1",
				Namespace: "default",
			},
			Status: model.TargetStatus{
				Properties: map[string]string{
					"component:component1": `{"Name":"component1","Type":"group","Properties":{}}`,
				},
			},
		},
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target2",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "target3",
				Namespace: "default",
			},
			Status: model.TargetStatus{
				Properties: map[string]string{
					"component:component2": `{"Name":"component2","Type":"group","Properties":{}}`,
				},
			},
		},
	}
	assignments, err := provider.assignComponents(components, targets)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(assignments))
	assert.Equal(t, 1, len(assignments["target1"]))
	assert.Equal(t, 0, len(assignments["target2"]))
	assert.Equal(t, 1, len(assignments["target3"]))
	assert.Equal(t, "component1", assignments["target1"][0])
	assert.Equal(t, "component2", assignments["target3"][0])
}

func TestGroupTargetProviderApply(t *testing.T) {
	testGroupPatcher := os.Getenv("TEST_GROUP_APPLY")
	if testGroupPatcher == "" {
		t.Skip("Skipping because TEST_GROUP_APPLY enviornment variable is not set")
	}
	os.Setenv("SYMPHONY_API_URL", "http://localhost:8082/v1alpha2/")
	os.Setenv("USE_SERVICE_ACCOUNT_TOKENS", "false")
	config := GroupTargetProviderConfig{
		User:     "admin",
		Password: "",
		TargetSelector: model.TargetSelector{
			LabelSelector: map[string]string{
				"haSet": "ha-set1",
				"kind":  "ipc",
			},
		},
	}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	component1 := model.ComponentSpec{
		Name:       "Softdpac_2",
		Type:       "container",
		Properties: map[string]interface{}{},
	}
	component2 := model.ComponentSpec{
		Name:       "Softdpac_3",
		Type:       "container",
		Properties: map[string]interface{}{},
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
				Components: []model.ComponentSpec{component1, component2},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component1,
			},
			{
				Action:    model.ComponentUpdate,
				Component: component2,
			},
		},
	}
	_, err = provider.Apply(context.Background(), model.TargetProviderApplyReference{
		Deployment: deployment,
		Step:       step,
		IsDryRun:   false,
	})
	assert.Nil(t, err)
}
