/*
Copyright 2022 The COA Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package solution

import (
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestFindAgentEmpty(t *testing.T) {
	agent := findAgent(model.TargetSpec{})
	assert.Equal(t, "", agent)
}
func TestFindAgentMatch(t *testing.T) {
	agent := findAgent(model.TargetSpec{
		Components: []model.ComponentSpec{
			{
				Name: "symphony-agent",
				Properties: map[string]string{
					"container.image": "possprod.azurecr.io/symphony-agent:0.38.0",
				},
			},
		},
	})
	assert.Equal(t, "symphony-agent", agent)
}
func TestFindAgentNotMatch(t *testing.T) {
	agent := findAgent(model.TargetSpec{
		Components: []model.ComponentSpec{
			{
				Name: "symphony-agent",
				Properties: map[string]string{
					"container.image": "possprod.azurecr.io/symphony-api:0.38.0",
				},
			},
		},
	})
	assert.Equal(t, "", agent)
}
func TestSortByDepedenciesSingleChain(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com2"},
		},
		{
			Name:         "com2",
			Dependencies: []string{"com1"},
		},
		{
			Name: "com1",
		},
	}
	ret, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, "com1", ret[0].Name)
	assert.Equal(t, "com2", ret[1].Name)
	assert.Equal(t, "com3", ret[2].Name)
}
func TestSortByDepedenciesSingleCircle(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com2"},
		},
		{
			Name:         "com2",
			Dependencies: []string{"com1"},
		},
		{
			Name:         "com1",
			Dependencies: []string{"com3"},
		},
	}
	_, err := sortByDepedencies(components)
	assert.NotNil(t, err)
}
func TestSortByDepedenciesSelfCircle(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com2"},
		},
		{
			Name:         "com2",
			Dependencies: []string{"com1"},
		},
		{
			Name:         "com1",
			Dependencies: []string{"com1"}, // note: generally self-depedencies should not be allowed
		},
	}
	_, err := sortByDepedencies(components)
	assert.NotNil(t, err)
}
func TestSortByDepedenciesNoDependencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name: "com3",
		},
		{
			Name: "com2",
		},
		{
			Name: "com1",
		},
	}
	ret, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, "com3", ret[0].Name)
	assert.Equal(t, "com2", ret[1].Name)
	assert.Equal(t, "com1", ret[2].Name)
}
func TestSortByDepedenciesParitalDependencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com1"},
		},
		{
			Name: "com2",
		},
		{
			Name: "com1",
		},
	}
	ret, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, "com2", ret[0].Name)
	assert.Equal(t, "com1", ret[1].Name)
	assert.Equal(t, "com3", ret[2].Name)
}
func TestSortByDepedenciesMultiDependencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com1", "com2"},
		},
		{
			Name: "com2",
		},
		{
			Name:         "com1",
			Dependencies: []string{"com2"},
		},
	}
	ret, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, "com2", ret[0].Name)
	assert.Equal(t, "com1", ret[1].Name)
	assert.Equal(t, "com3", ret[2].Name)
}
func TestSortByDepedenciesForeignDependencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com4"},
		},
		{
			Name: "com2",
		},
		{
			Name: "com1",
		},
	}
	_, err := sortByDepedencies(components)
	assert.NotNil(t, err)
}
func TestSortByDepedenciesAllSelfReferences(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com3"}, //note: unlike TestSortByDepedenciesSelfCircle, this self-depedency is not resolved
		},
		{
			Name: "com2",
		},
		{
			Name:         "com1",
			Dependencies: []string{"com2"},
		},
	}
	_, err := sortByDepedencies(components)
	assert.NotNil(t, err)
}
func TestGetRolesUnique(t *testing.T) {
	components := []model.ComponentSpec{
		{Type: "a"},
		{Type: "b"},
		{Type: "c"},
	}
	groups := collectGroups(components)
	assert.Equal(t, 3, len(groups))
	assert.Equal(t, "a", groups[0].Type)
	assert.Equal(t, "b", groups[1].Type)
	assert.Equal(t, "c", groups[2].Type)
	assert.Equal(t, 0, groups[0].Index)
	assert.Equal(t, 1, groups[1].Index)
	assert.Equal(t, 2, groups[2].Index)
}
func TestGetRolesAllSame(t *testing.T) {
	components := []model.ComponentSpec{
		{Type: "a"},
		{Type: "a"},
		{Type: "a"},
	}
	groups := collectGroups(components)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, "a", groups[0].Type)
	assert.Equal(t, 0, groups[0].Index)
}
func TestGetRolesAllEmpty(t *testing.T) {
	components := []model.ComponentSpec{{}, {}, {}}
	groups := collectGroups(components)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, "", groups[0].Type)
	assert.Equal(t, 0, groups[0].Index)
}
func TestGetRolesEmptyNonEmpty(t *testing.T) {
	components := []model.ComponentSpec{
		{},
		{Type: "a"},
		{Type: "a"},
	}
	groups := collectGroups(components)
	assert.Equal(t, 2, len(groups))
	assert.Equal(t, "", groups[0].Type)
	assert.Equal(t, "a", groups[1].Type)
	assert.Equal(t, 0, groups[0].Index)
	assert.Equal(t, 1, groups[1].Index)
}
func TestGetRolesMixed(t *testing.T) {
	components := []model.ComponentSpec{
		{},
		{Type: "a"},
		{Type: "b"},
		{Type: "a"},
		{},
		{Type: "b"},
	}
	groups := collectGroups(components)
	assert.Equal(t, 6, len(groups))
	assert.Equal(t, "", groups[0].Type)
	assert.Equal(t, 0, groups[0].Index)
	assert.Equal(t, "a", groups[1].Type)
	assert.Equal(t, 1, groups[1].Index)
	assert.Equal(t, "b", groups[2].Type)
	assert.Equal(t, 2, groups[2].Index)
	assert.Equal(t, "a", groups[3].Type)
	assert.Equal(t, 3, groups[3].Index)
	assert.Equal(t, "", groups[4].Type)
	assert.Equal(t, 4, groups[4].Index)
	assert.Equal(t, "b", groups[5].Type)
	assert.Equal(t, 5, groups[5].Index)
}
func TestGetRolesContinuedMixed(t *testing.T) {
	components := []model.ComponentSpec{
		{},
		{Type: "a"},
		{Type: "a"},
		{Type: "b"},
		{},
		{Type: "b"},
	}
	groups := collectGroups(components)
	assert.Equal(t, 5, len(groups))
	assert.Equal(t, "", groups[0].Type)
	assert.Equal(t, 0, groups[0].Index)
	assert.Equal(t, "a", groups[1].Type)
	assert.Equal(t, 1, groups[1].Index)
	assert.Equal(t, "b", groups[2].Type)
	assert.Equal(t, 3, groups[2].Index)
	assert.Equal(t, "", groups[3].Type)
	assert.Equal(t, 4, groups[3].Index)
	assert.Equal(t, "b", groups[4].Type)
	assert.Equal(t, 5, groups[4].Index)
}
func TestGetRolesSingle(t *testing.T) {
	components := []model.ComponentSpec{
		{Type: "a"},
	}
	groups := collectGroups(components)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, "a", groups[0].Type)
	assert.Equal(t, 0, groups[0].Index)
}
func TestSortByDepedenciesSingleNoDepedencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name: "com3",
			Type: "helm",
		},
	}
	components, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
	assert.Equal(t, "com3", components[0].Name)
	groups := collectGroups(components)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, "helm", groups[0].Type)
	assert.Equal(t, 0, groups[0].Index)
}
