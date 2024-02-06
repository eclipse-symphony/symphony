/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestCampaignMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name:        "name",
		FirstStage:  "list",
		SelfDriving: true,
		Stages: map[string]StageSpec{
			"list": {
				Name:          "list",
				Provider:      "providers.stage.list",
				StageSelector: "wait-sync",
				Config: map[string]interface{}{
					"baseUrl":  "http://symphony-service:8080/v1alpha2/",
					"user":     "admin",
					"password": "",
				},
				Inputs: map[string]interface{}{
					"objectType":  "sites",
					"namesObject": true,
				},
			},
		},
	}
	campaign2 := CampaignSpec{
		Name:        "name",
		FirstStage:  "list",
		SelfDriving: true,
		Stages: map[string]StageSpec{
			"list": {
				Name:          "list",
				Provider:      "providers.stage.list",
				StageSelector: "wait-sync",
				Config: map[string]interface{}{
					"baseUrl":  "http://symphony-service:8080/v1alpha2/",
					"user":     "admin",
					"password": "",
				},
				Inputs: map[string]interface{}{
					"objectType":  "sites",
					"namesObject": true,
				},
			},
		},
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
	assert.EqualError(t, err, "parameter is not a CampaignSpec type")
	assert.False(t, res)
}

func TestCampaignNameNotMatch(t *testing.T) {
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

func TestCampaignFirstStageNotMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name:       "name",
		FirstStage: "list",
	}
	campaign2 := CampaignSpec{
		Name:       "name",
		FirstStage: "list1",
	}
	equal, err := campaign1.DeepEquals(campaign2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestCampaignSelfDrivingNotMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name:        "name",
		FirstStage:  "list",
		SelfDriving: true,
	}
	campaign2 := CampaignSpec{
		Name:        "name",
		FirstStage:  "list",
		SelfDriving: false,
	}
	equal, err := campaign1.DeepEquals(campaign2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestCampaignStagesLengthNotMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name:        "name",
		FirstStage:  "mock1",
		SelfDriving: true,
		Stages: map[string]StageSpec{
			"mock1": {
				Name:     "mock1",
				Provider: "providers.stage.mock",
			},
		},
	}
	campaign2 := CampaignSpec{
		Name:        "name",
		FirstStage:  "mock1",
		SelfDriving: true,
		Stages: map[string]StageSpec{
			"mock1": {
				Name:     "mock1",
				Provider: "providers.stage.mock",
			},
			"mock2": {
				Name:     "mock2",
				Provider: "providers.stage.mock",
			},
		},
	}
	equal, err := campaign1.DeepEquals(campaign2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestCampaignStagesNotMatch(t *testing.T) {
	campaign1 := CampaignSpec{
		Name:        "name",
		FirstStage:  "mock1",
		SelfDriving: true,
		Stages: map[string]StageSpec{
			"mock1": {
				Name:     "mock1",
				Provider: "providers.stage.mock",
			},
		},
	}
	campaign2 := CampaignSpec{
		Name:        "name",
		FirstStage:  "mock1",
		SelfDriving: true,
		Stages: map[string]StageSpec{
			"mock1": {
				Name:     "mock1",
				Provider: "providers.stage.mockv2",
			},
		},
	}
	equal, err := campaign1.DeepEquals(campaign2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestStageNotMatch(t *testing.T) {
	stage1 := StageSpec{
		Name: "mock1",
	}
	stage2 := StageSpec{
		Name: "mock1",
	}

	// name not match
	stage2.Name = "mock2"
	equal, err := stage1.DeepEquals(stage2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// provider not match
	stage2.Name = "mock1"
	stage1.Provider = "providers.stage.mock"
	stage2.Provider = "providers.stage.mockv2"
	equal, err = stage1.DeepEquals(stage2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// stageSelector not match
	stage2.Provider = "providers.stage.mock"
	stage1.StageSelector = "wait-sync"
	stage2.StageSelector = "wait-syncv2"
	equal, err = stage1.DeepEquals(stage2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// config not match
	stage2.StageSelector = "wait-sync"
	stage1.Config = map[string]interface{}{
		"baseUrl":  "http://symphony-service:8080/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	stage2.Config = map[string]interface{}{
		"baseUrl":  "http://symphony-service:8888/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	equal, err = stage1.DeepEquals(stage2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// inputs not match
	stage2.Config = map[string]interface{}{
		"baseUrl":  "http://symphony-service:8080/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	stage1.Inputs = map[string]interface{}{
		"objectType":  "sites",
		"namesObject": true,
	}
	stage2.Inputs = map[string]interface{}{
		"objectType":  "sites",
		"namesObject": false,
	}
	equal, err = stage1.DeepEquals(stage2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// schedule not match
	stage2.Inputs = map[string]interface{}{
		"objectType":  "sites",
		"namesObject": true,
	}
	stage1.Schedule = &v1alpha2.ScheduleSpec{
		Date: "2020-10-31",
		Time: "12:00:00PM",
		Zone: "PDT",
	}
	stage2.Schedule = &v1alpha2.ScheduleSpec{
		Date: "2020-10-31",
		Time: "12:00:00PM",
		Zone: "PST",
	}
	equal, err = stage1.DeepEquals(stage2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestStageMatchOneEmpty(t *testing.T) {
	stage1 := StageSpec{
		Name: "name",
	}
	res, err := stage1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a StageSpec type")
	assert.False(t, res)
}

func TestActivationMatchOneEmpty(t *testing.T) {
	activation1 := ActivationSpec{
		Name: "name",
	}
	res, err := activation1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a ActivationSpec type")
	assert.False(t, res)
}

func TestActivationMatch(t *testing.T) {
	activation1 := ActivationSpec{
		Name:     "multisite-deploy",
		Campaign: "site-apps",
		Stage:    "deploy",
		Inputs: map[string]interface{}{
			"site": "site1",
		},
	}
	activation2 := ActivationSpec{
		Name:     "multisite-deploy",
		Campaign: "site-apps",
		Stage:    "deploy",
		Inputs: map[string]interface{}{
			"site": "site1",
		},
	}
	equal, err := activation1.DeepEquals(activation2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestActivationNotMatch(t *testing.T) {
	activation1 := ActivationSpec{
		Name: "multisite-deploy",
	}
	activation2 := ActivationSpec{
		Name: "multisite-deploy2",
	}

	// name not match
	equal, err := activation1.DeepEquals(activation2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// compaign not match
	activation2.Name = "multisite-deploy"
	activation1.Campaign = "site-apps"
	activation2.Campaign = "site-apps2"
	equal, err = activation1.DeepEquals(activation2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// stage not match
	activation2.Campaign = "site-apps"
	activation1.Stage = "deploy"
	activation2.Stage = "deploy2"
	equal, err = activation1.DeepEquals(activation2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// inputs not match
	activation2.Stage = "deploy"
	activation1.Inputs = map[string]interface{}{
		"site": "site1",
	}
	activation2.Inputs = map[string]interface{}{
		"site": "site2",
	}
	equal, err = activation1.DeepEquals(activation2)
	assert.Nil(t, err)
	assert.False(t, equal)
}
