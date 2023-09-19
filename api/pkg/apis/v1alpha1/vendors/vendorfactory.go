/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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
	default:
		return nil, nil //Can't throw errors as other factories may create it...
	}
}
