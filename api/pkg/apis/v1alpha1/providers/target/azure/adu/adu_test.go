package adu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitWithNil(t *testing.T) {
	provider := ADUTargetProvider{}
	err := provider.Init(nil)
	assert.NotNil(t, err)
}
