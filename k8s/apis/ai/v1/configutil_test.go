/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package v1

import (
	"testing"

	configv1 "gopls-workspace/apis/config/v1"

	configutils "gopls-workspace/configutils"

	apimodel "github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplyModelValidationPoliciesNoItem(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"model": []configv1.ValidationPolicy{
			{
				SelectorType:   "properties",
				SelectorKey:    "",
				SelectorValue:  "",
				SpecField:      "model.project",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := ModelList{
		Items: []Model{},
	}
	for _, p := range policies["model"] {
		pack := extractModelValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", "cv-model", p.ValidationType, pack)
		assert.Nil(t, err)
		assert.Equal(t, "", ret)
	}
}

func TestApplyModelValidationPoliciesSingleItem(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"model": []configv1.ValidationPolicy{
			{
				SelectorType:   "properties",
				SelectorKey:    "",
				SelectorValue:  "",
				SpecField:      "model.project",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := ModelList{
		Items: []Model{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: apimodel.ModelSpec{
					Properties: map[string]string{
						"model.project": "123",
					},
				},
			},
		},
	}
	for _, p := range policies["model"] {
		pack := extractModelValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", readModelValiationTarget(&list.Items[0], p), p.ValidationType, pack)
		assert.Nil(t, err)
		assert.Equal(t, "", ret)
	}
}

func TestApplyModelValidationPoliciesSingleItemUpdate(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"model": []configv1.ValidationPolicy{
			{
				SelectorType:   "properties",
				SelectorKey:    "",
				SelectorValue:  "",
				SpecField:      "model.project",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := ModelList{
		Items: []Model{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: apimodel.ModelSpec{
					Properties: map[string]string{
						"model.project": "123",
					},
				},
			},
		},
	}
	for _, p := range policies["model"] {
		pack := extractModelValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", "345", p.ValidationType, pack)
		assert.Nil(t, err)
		assert.Equal(t, "", ret)
	}
}

func TestApplyModelValidationPoliciesSingleItemConflict(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"model": []configv1.ValidationPolicy{
			{
				SelectorType:   "properties",
				SelectorKey:    "",
				SelectorValue:  "",
				SpecField:      "model.project",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := ModelList{
		Items: []Model{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: apimodel.ModelSpec{
					Properties: map[string]string{
						"model.project": "123",
					},
				},
			},
		},
	}
	newModel := Model{
		ObjectMeta: metav1.ObjectMeta{
			Name: "quake2",
		},
		Spec: apimodel.ModelSpec{
			Properties: map[string]string{
				"model.project": "123",
			},
		},
	}
	for _, p := range policies["model"] {
		pack := extractModelValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake2", readModelValiationTarget(&newModel, p), p.ValidationType, pack)
		assert.Nil(t, err)
		assert.NotEqual(t, "", ret)
	}
}

func TestApplyModelValidationPoliciesUpdateDuplicate(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"model": []configv1.ValidationPolicy{
			{
				SelectorType:   "properties",
				SelectorKey:    "",
				SelectorValue:  "",
				SpecField:      "model.project",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := ModelList{
		Items: []Model{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: apimodel.ModelSpec{
					Properties: map[string]string{
						"model.project": "123",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake2",
				},
				Spec: apimodel.ModelSpec{
					Properties: map[string]string{
						"model.project": "345",
					},
				},
			},
		},
	}
	newModel := Model{
		ObjectMeta: metav1.ObjectMeta{
			Name: "quake",
		},
		Spec: apimodel.ModelSpec{
			Properties: map[string]string{
				"model.project": "345",
			},
		},
	}
	for _, p := range policies["model"] {
		pack := extractModelValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", readModelValiationTarget(&newModel, p), p.ValidationType, pack)
		assert.Nil(t, err)
		assert.NotEqual(t, "", ret)
	}
}

func TestApplyModelValidationPoliciesUpdateNoConflict(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"model": []configv1.ValidationPolicy{
			{
				SelectorType:   "properties",
				SelectorKey:    "",
				SelectorValue:  "",
				SpecField:      "model.project",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := ModelList{
		Items: []Model{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: apimodel.ModelSpec{
					Properties: map[string]string{
						"model.project": "123",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake2",
				},
				Spec: apimodel.ModelSpec{
					Properties: map[string]string{
						"model.project": "345",
					},
				},
			},
		},
	}
	newModel := Model{
		ObjectMeta: metav1.ObjectMeta{
			Name: "quake",
		},
		Spec: apimodel.ModelSpec{
			Properties: map[string]string{
				"model.project": "347",
			},
		},
	}
	for _, p := range policies["model"] {
		pack := extractModelValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", readModelValiationTarget(&newModel, p), p.ValidationType, pack)
		assert.Nil(t, err)
		assert.Equal(t, "", ret)
	}
}
