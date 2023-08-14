/*
Copyright 2022 The COA Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
	default:
		return nil, nil //Can't throw errors as other factories may create it...
	}
}
