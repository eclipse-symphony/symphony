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

package managers

import (
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/campaigns"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/devices"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/instances"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/jobs"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/reference"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/solutions"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/stage"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/target"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/targets"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/users"
	cm "github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
)

type SymphonyManagerFactory struct {
}

func (c SymphonyManagerFactory) CreateManager(config cm.ManagerConfig) (cm.IManager, error) {
	switch config.Type {
	case "managers.symphony.solution":
		return &solution.SolutionManager{}, nil
	case "managers.symphony.reference":
		return &reference.ReferenceManager{}, nil
	case "managers.symphony.target":
		return &target.TargetManager{}, nil
	case "managers.symphony.targets":
		return &targets.TargetsManager{}, nil
	case "managers.symphony.devices":
		return &devices.DevicesManager{}, nil
	case "managers.symphony.solutions":
		return &solutions.SolutionsManager{}, nil
	case "managers.symphony.instances":
		return &instances.InstancesManager{}, nil
	case "managers.symphony.users":
		return &users.UsersManager{}, nil
	case "managers.symphony.jobs":
		return &jobs.JobsManager{}, nil
	case "managers.symphony.campaigns":
		return &campaigns.CampaignsManager{}, nil
	case "managers.symphony.catalogs":
		return &catalogs.CatalogsManager{}, nil
	case "managers.symphony.activations":
		return &activations.ActivationsManager{}, nil
	case "managers.symphony.stage":
		return &stage.StageManager{}, nil
	default:
		return nil, nil //TBD: can't throw errors here as other manages may pick up creation process
	}
}
