package utils

import (
	"encoding/json"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestSimpleMustConstraint(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "CPU",
		Value:     "x86",
	}
	properties := map[string]string{
		"CPU": "x86",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, -2, res)
}

func TestSimpleMustConstraintMiss(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "CPU",
		Value:     "x86",
	}
	properties := map[string]string{
		"CPU": "arm",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestSimpleRejectConstraint(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "x86",
	}
	properties := map[string]string{
		"CPU": "x86",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestSimpleRejectConstraintMiss(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "x86",
	}
	properties := map[string]string{
		"CPU": "X86",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, -2, res)
}
func TestSimplePreferConstraint(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "prefer",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, 1, res)
}
func TestSimplePreferConstraintMiss(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "prefer",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "X64",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, 0, res)
}
func TestSimpleEmptyConstraint(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, -3, res)
}
func TestSimpleEmptyConstraintMiss(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "X64",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, 0, res)
}
func TestSimpleInvalidConstraint(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "bad",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	_, err := evaluateConstraint(constraint, properties)
	assert.NotNil(t, err)
}
func TestConstraintWithoutKey(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	_, err := evaluateConstraint(constraint, properties)
	assert.NotNil(t, err)
}
func TestSimpleInvalidConstraintMiss(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "bad",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "X64",
	}
	_, err := evaluateConstraint(constraint, properties)
	assert.NotNil(t, err)
}
func TestAnyRejection(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "reject",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "arm",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestAnyRejectionAllMiss(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "reject",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "dragon",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, 0, res)
}
func TestAnyEmptyOperator(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "arm",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, -3, res)
}
func TestAnyEmptyQualifierDeepReject(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestAnyEmptyQualifierDeepMust(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "CPU",
		Value:     "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "CPU",
		Value:     "x64",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, 0, res)
}
func TestAnyEmptyQualifierDeepMissingKey(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "must",
		Value:     "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Qualifier: "must",
		Value:     "x64",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	_, err := evaluateConstraint(constraint, properties)
	assert.NotNil(t, err)
}
func TestAnyBadJson(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "reject",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{"BAD JSON"},
	}
	properties := map[string]string{
		"CPU": "HIJ",
	}
	_, err := evaluateConstraint(constraint, properties)
	assert.NotNil(t, err)
}
func TestAnyPreferred(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "GPU",
		Value: "t4",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "prefer",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "arm",
		"GPU": "t4",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, 2, res)
}
func TestAnyMust(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraint(constraint, properties)
	assert.Nil(t, err)
	assert.Equal(t, 0, res)
}
func TestAnyBad(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "bad",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	_, err := evaluateConstraint(constraint, properties)
	assert.NotNil(t, err)
}
func TestBadOperator(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Operator:  "bad",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	_, err := evaluateConstraint(constraint, properties)
	assert.NotNil(t, err)
}
func TestIncomplete(t *testing.T) {
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Operator:  "bad",
		Key:       "CPU",
	}
	properties := map[string]string{
		"CPU": "arm",
	}
	_, err := evaluateConstraint(constraint, properties)
	assert.NotNil(t, err)
}
func TestMultipleMust(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "GPU",
		Value:     "t4",
	}
	properties := map[string]string{
		"CPU": "x64",
		"GPU": "t4",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -2, res)
}
func TestMultipleMustMissingKey(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "must",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "must",
		Value:     "t4",
	}
	properties := map[string]string{
		"CPU": "x64",
		"GPU": "t4",
	}
	_, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.NotNil(t, err)
}
func TestMultipleMustMiss(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "GPU",
		Value:     "t4",
	}
	properties := map[string]string{
		"CPU": "X64",
		"GPU": "t4",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestMultipleReject(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "GPU",
		Value:     "t4",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestMultipleRejectMiss(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "GPU",
		Value:     "t4",
	}
	properties := map[string]string{
		"CPU": "X64",
		"GPU": "orin",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -2, res)
}
func TestMultiplePreferred(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "prefer",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "prefer",
		Key:       "GPU",
		Value:     "t4",
	}
	properties := map[string]string{
		"CPU": "x64",
		"GPU": "t4",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, 2, res)
}
func TestMultiplePreferredMiss(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "prefer",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "prefer",
		Key:       "GPU",
		Value:     "t4",
	}
	properties := map[string]string{
		"CPU": "X64",
		"GPU": "T4",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, 0, res)
}
func TestMultipleEmpty(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "",
		Key:       "GPU",
		Value:     "t4",
	}
	properties := map[string]string{
		"CPU": "x64",
		"GPU": "t4",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -3, res)
}
func TestMixedMustAndReject(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestMixedPreferredandReject(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "prefer",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestMixedEmptyandReject(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -1, res)
}
func TestMixedMustAndPreferred(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "must",
		Key:       "CPU",
		Value:     "x64",
	}
	constraintB := model.ConstraintSpec{
		Qualifier: "prefer",
		Key:       "CPU",
		Value:     "x64",
	}
	properties := map[string]string{
		"CPU": "x64",
	}
	res, err := evaluateConstraints([]model.ConstraintSpec{constraintA, constraintB}, properties)
	assert.Nil(t, err)
	assert.Equal(t, -2, res)
}
func TestEvaluateTargetComponentMatch(t *testing.T) {
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			{
				Qualifier: "must",
				Key:       "GPU",
				Value:     "T-4",
			},
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"GPU": "T-4",
		},
	}

	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, 0, score)
}
func TestEvaluateTargetComponentMismatch(t *testing.T) {
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			{
				Qualifier: "must",
				Key:       "GPU",
				Value:     "T-4",
			},
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"GPU": "T-5",
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, -1, score)
}
func TestEvaluateTargetComponentMatchMultiple(t *testing.T) {
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			{
				Qualifier: "must",
				Key:       "GPU",
				Value:     "T-4",
			},
			{
				Qualifier: "must",
				Key:       "MEMORY",
				Value:     "512G",
			},
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"GPU":    "T-4",
			"MEMORY": "512G",
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, 0, score)
}
func TestEvaluateTargetComponentMatchPartialMatch(t *testing.T) {
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			{
				Qualifier: "must",
				Key:       "GPU",
				Value:     "T-4",
			},
			{
				Qualifier: "must",
				Key:       "MEMORY",
				Value:     "512G",
			},
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"GPU":    "T-4",
			"MEMORY": "128G",
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, -1, score)
}
func TestEvaluateTargetComponentMatchAny(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			constraint,
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"CPU": "arm",
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, 0, score)
}
func TestEvaluateTargetComponentMatchAnyMiss(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Key:   "CPU",
		Value: "arm",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			constraint,
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"CPU": "dragon",
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, 0, score)
}
func TestEvaluateTargetComponentPreferButRejected(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "dragon",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "prefer",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			constraint,
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"CPU": "dragon",
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, -1, score)
}
func TestEvaluateTargetComponentMustButRejected(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Key:   "CPU",
		Value: "x64",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "dragon",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			constraint,
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"CPU": "dragon",
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, -1, score)
}
func TestEvaluateTargetComponentMultipleRejects(t *testing.T) {
	constraintA := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "turtle",
	}
	dataA, _ := json.Marshal(constraintA)
	constraintB := model.ConstraintSpec{
		Qualifier: "reject",
		Key:       "CPU",
		Value:     "dragon",
	}
	dataB, _ := json.Marshal(constraintB)
	constraint := model.ConstraintSpec{
		Qualifier: "must",
		Operator:  "any",
		Key:       "CPU",
		Values:    []string{string(dataA), string(dataB)},
	}
	component := model.ComponentSpec{
		Constraints: []model.ConstraintSpec{
			constraint,
		},
	}
	target := model.TargetSpec{
		Properties: map[string]string{
			"CPU": "dragon",
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, -1, score)
}

func TestTargetMismatchComponent(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"GPU": "T-5",
		},
	}
	target := model.TargetSpec{
		Constraints: []model.ConstraintSpec{
			{
				Qualifier: "must",
				Key:       "GPU",
				Value:     "T-4",
			},
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, -1, score)
}
func TestTargetRejectComponent(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"GPU": "T-4",
		},
	}
	target := model.TargetSpec{
		Constraints: []model.ConstraintSpec{
			{
				Qualifier: "reject",
				Key:       "GPU",
				Value:     "T-4",
			},
		},
	}
	score, err := evaluateTargetCompatibility(target, component)
	assert.Nil(t, err)
	assert.Equal(t, -1, score)
}

func TestReadStringWithOverrides(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "HIJ",
	}, "ABC", "")
	assert.Equal(t, "HIJ", val)
}
func TestReadStringWithNoOverrides(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"CDE": "HIJ",
	}, "ABC", "")
	assert.Equal(t, "DEF", val)
}
func TestReadStringOverrideOnly(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"CDE": "HIJ",
	}, "CDE", "")
	assert.Equal(t, "HIJ", val)
}

func TestReadStringMissWithDefault(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "HIJ",
	}, "DEF", "HE")
	assert.Equal(t, "HE", val)
}
func TestReadStringEmptyOverride(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "",
	}, "DEF", "")
	assert.Equal(t, "", val)
}

func TestFormatObjectEmpty(t *testing.T) {
	obj := new(interface{})
	val, err := FormatObject(obj, false, "", "")
	assert.Nil(t, err)
	assert.Equal(t, "null", string(val))
}
func TestFormatObjectEmptyDict(t *testing.T) {
	obj := map[string]interface{}{}
	val, err := FormatObject(obj, false, "", "")
	assert.Nil(t, err)
	assert.Equal(t, "{}", string(val))
}
func TestFormatObjectDictJson(t *testing.T) {
	obj := map[string]interface{}{
		"foo": "bar",
	}
	val, err := FormatObject(obj, false, "$.foo", "")
	assert.Nil(t, err)
	assert.Equal(t, "\"bar\"", string(val))
}
func TestFormatObjectDictYaml(t *testing.T) {
	obj := map[string]interface{}{
		"foo": "bar",
	}
	val, err := FormatObject(obj, false, "$.foo", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "bar\n", string(val))
}
