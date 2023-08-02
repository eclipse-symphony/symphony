package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCampaignMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name: "name",
	}
	campaign2 := CampaignSpec{
		Name: "name",
	}
	equal, err := campaign1.DeepEquals(campaign2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestCampaignMatchOneEmpty(t *testing.T) {
	campaign1 := CampaignSpec{
		Name: "name",
	}
	res, err := campaign1.DeepEquals(nil)
	assert.Errorf(t, err, "parameter is not a CampaignSpec type")
	assert.False(t, res)
}

func TestCampaignRoleNotMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name: "name",
	}
	campaign2 := CampaignSpec{
		Name: "name1",
	}
	equal, err := campaign1.DeepEquals(campaign2)
	assert.Nil(t, err)
	assert.False(t, equal)
}
