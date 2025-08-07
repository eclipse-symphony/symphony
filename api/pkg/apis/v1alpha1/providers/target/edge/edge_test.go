package edge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEdgeTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := EdgeProviderConfigFromMap(nil)
	assert.NotNil(t, err)

}
