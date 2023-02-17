package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
