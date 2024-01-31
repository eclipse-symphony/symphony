/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
)

func TestCreateVendor(t *testing.T) {
	factory := SymphonyVendorFactory{}
	config := vendors.VendorConfig{}
	config.Type = "vendors.echo"
	vendor, err := factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*EchoVendor))

	config.Type = "vendors.solution"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*SolutionVendor))

	config.Type = "vendors.agent"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*AgentVendor))

	config.Type = "vendors.targets"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*TargetsVendor))

	config.Type = "vendors.instances"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*InstancesVendor))

	config.Type = "vendors.devices"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*DevicesVendor))

	config.Type = "vendors.solutions"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*SolutionsVendor))

	config.Type = "vendors.campaigns"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*CampaignsVendor))

	config.Type = "vendors.catalogs"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*CatalogsVendor))

	config.Type = "vendors.activations"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*ActivationsVendor))

	config.Type = "vendors.users"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*UsersVendor))

	config.Type = "vendors.jobs"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*JobVendor))

	config.Type = "vendors.stage"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*StageVendor))

	config.Type = "vendors.federation"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*FederationVendor))

	config.Type = "vendors.staging"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*StagingVendor))

	config.Type = "vendors.models"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*ModelsVendor))

	config.Type = "vendors.skills"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*SkillsVendor))

	config.Type = "vendors.settings"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*SettingsVendor))

	config.Type = "vendors.trails"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*TrailsVendor))

	config.Type = "vendors.backgroundjob"
	vendor, err = factory.CreateVendor(config)
	assert.Nil(t, err)
	assert.NotNil(t, vendor.(*BackgroundJobVendor))
}
