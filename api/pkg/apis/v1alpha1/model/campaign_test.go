package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCampaignMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name: "name",
	}
	targetRef := TargetRefSpec{
		Name: "name",
	}
	equal, err := campaign1.DeepEquals(targetRef)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestCampaignMatchOneEmpty(t *testing.T) {
	campaign1 := CampaignSpec{
		Name: "name",
	}
	res, err := campaign1.DeepEquals(nil)
	assert.Errorf(t, err, "parameter is not a TargetRefSpec type")
	assert.False(t, res)
}

func TestCampaignRoleNotMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name: "name",
	}
	targetRef := TargetRefSpec{
		Name: "name1",
	}
	equal, err := campaign1.DeepEquals(targetRef)
	assert.Nil(t, err)
	assert.False(t, equal)
}
