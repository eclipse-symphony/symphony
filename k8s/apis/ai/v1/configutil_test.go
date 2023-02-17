package v1

import (
	"testing"

	configv1 "gopls-workspace/apis/config/v1"

	configutils "gopls-workspace/configutils"

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
				Spec: ModelSpec{
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
				Spec: ModelSpec{
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
				Spec: ModelSpec{
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
		Spec: ModelSpec{
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
				Spec: ModelSpec{
					Properties: map[string]string{
						"model.project": "123",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake2",
				},
				Spec: ModelSpec{
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
		Spec: ModelSpec{
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
				Spec: ModelSpec{
					Properties: map[string]string{
						"model.project": "123",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake2",
				},
				Spec: ModelSpec{
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
		Spec: ModelSpec{
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
