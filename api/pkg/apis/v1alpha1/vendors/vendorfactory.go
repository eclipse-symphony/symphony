/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
)

type SymphonyVendorFactory struct {
}

func (c SymphonyVendorFactory) CreateVendor(config vendors.VendorConfig) (vendors.IVendor, error) {
	switch config.Type {
	case "vendors.echo":
		return &EchoVendor{}, nil
	case "vendors.solution":
		return &SolutionVendor{}, nil
	case "vendors.agent":
		return &AgentVendor{}, nil
	case "vendors.targets":
		return &TargetsVendor{}, nil
	case "vendors.instances":
		return &InstancesVendor{}, nil
	case "vendors.devices":
		return &DevicesVendor{}, nil
	case "vendors.solutions":
		return &SolutionsVendor{}, nil
	case "vendors.campaigns":
		return &CampaignsVendor{}, nil
	case "vendors.catalogs":
		return &CatalogsVendor{}, nil
	case "vendors.activations":
		return &ActivationsVendor{}, nil
	case "vendors.users":
		return &UsersVendor{}, nil
	case "vendors.jobs":
		return &JobVendor{}, nil
	case "vendors.stage":
		return &StageVendor{}, nil
	case "vendors.federation":
		return &FederationVendor{}, nil
	case "vendors.staging":
		return &StagingVendor{}, nil
	case "vendors.models":
		return &ModelsVendor{}, nil
	case "vendors.skills":
		return &SkillsVendor{}, nil
	case "vendors.settings":
		return &SettingsVendor{}, nil
	case "vendors.trails":
		return &TrailsVendor{}, nil
	case "vendors.backgroundjob":
		return &BackgroundJobVendor{}, nil
	default:
		return nil, nil //Can't throw errors as other factories may create it...
	}
}
