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

package managers

import (
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/campaigns"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/configs"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/devices"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/instances"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/jobs"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/models"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/reference"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/sites"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/skills"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/solutions"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/stage"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/staging"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/sync"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/target"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/targets"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/trails"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/users"
	cm "github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
)

type SymphonyManagerFactory struct {
	SingletonsCache map[string]cm.IManager
}

func (c *SymphonyManagerFactory) CreateManager(config cm.ManagerConfig) (cm.IManager, error) {
	if c.SingletonsCache == nil {
		c.SingletonsCache = make(map[string]cm.IManager)
	}
	if config.Properties["singleton"] == "true" {
		if c.SingletonsCache[config.Type] != nil {
			return c.SingletonsCache[config.Type], nil
		}
	}
	var manager cm.IManager
	switch config.Type {
	case "managers.symphony.solution":
		manager = &solution.SolutionManager{}
	case "managers.symphony.reference":
		manager = &reference.ReferenceManager{}
	case "managers.symphony.target":
		manager = &target.TargetManager{}
	case "managers.symphony.targets":
		manager = &targets.TargetsManager{}
	case "managers.symphony.devices":
		manager = &devices.DevicesManager{}
	case "managers.symphony.solutions":
		manager = &solutions.SolutionsManager{}
	case "managers.symphony.instances":
		manager = &instances.InstancesManager{}
	case "managers.symphony.users":
		manager = &users.UsersManager{}
	case "managers.symphony.jobs":
		manager = &jobs.JobsManager{}
	case "managers.symphony.campaigns":
		manager = &campaigns.CampaignsManager{}
	case "managers.symphony.catalogs":
		manager = &catalogs.CatalogsManager{}
	case "managers.symphony.activations":
		manager = &activations.ActivationsManager{}
	case "managers.symphony.stage":
		manager = &stage.StageManager{}
	case "managers.symphony.configs":
		manager = &configs.ConfigsManager{}
	case "managers.symphony.sites":
		manager = &sites.SitesManager{}
	case "managers.symphony.staging":
		manager = &staging.StagingManager{}
	case "managers.symphony.sync":
		manager = &sync.SyncManager{}
	case "managers.symphony.models":
		manager = &models.ModelsManager{}
	case "managers.symphony.skills":
		manager = &skills.SkillsManager{}
	case "managers.symphony.trails":
		manager = &trails.TrailsManager{}
	}
	if manager != nil && config.Properties["singleton"] == "true" {
		c.SingletonsCache[config.Type] = manager
	}
	return manager, nil
}
