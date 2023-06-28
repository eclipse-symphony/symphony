package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstraintDeepEquals(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	other := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	res, err := constraint.DeepEquals(other)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestConstraintDeepEqualsOneEmpty(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	res, err := constraint.DeepEquals(nil)
	assert.Errorf(t, err, "parameter is not a ConstraintSpec type")
	assert.False(t, res)
}

func TestConstraintDeepEqualsKeyNotMatch(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	other := ConstraintSpec{
		Key:       "os1",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	res, err := constraint.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConstraintDeepEqualsQualiferNotMatch(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	other := ConstraintSpec{
		Key:       "os",
		Qualifier: "must1",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	res, err := constraint.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConstraintDeepEqualsOperatorNotMatch(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	other := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals1",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	res, err := constraint.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConstraintDeepEqualsValueNotMatch(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	other := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS1",
		Values:    []string{"RTOS", "Linux"},
	}
	res, err := constraint.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConstraintDeepEqualsDifferentValueLengths(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS", "Linux"},
	}
	other := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Operator:  "equals",
		Value:     "RTOS",
		Values:    []string{"RTOS"},
	}
	res, err := constraint.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConstraintMatch(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Value:     "RTOS",
	}
	res := constraint.Match(map[string]string{
		"os":  "RTOS",
		"app": "rtos-demo",
	})
	assert.True(t, res)
}

func TestConstraintMisMatch(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "os",
		Qualifier: "must",
		Value:     "RTOS",
	}
	res := constraint.Match(map[string]string{
		"os":      "Linux",
		"app":     "rtos-demo",
		"runtime": "azure.iotedge",
	})
	assert.False(t, res)
}

func TestConstraintMatch2(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "runtime",
		Qualifier: "must",
		Value:     "azure.iotedge",
	}
	res := constraint.Match(map[string]string{
		"os":  "RTOS",
		"app": "rtos-demo",
	})
	assert.False(t, res)
}

func TestConstraintMisMatch2(t *testing.T) {
	constraint := ConstraintSpec{
		Key:       "runtime",
		Qualifier: "must",
		Value:     "azure.iotedge",
	}
	res := constraint.Match(map[string]string{
		"os":      "Linux",
		"app":     "rtos-demo",
		"runtime": "azure.iotedge",
	})
	assert.True(t, res)
}
