/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package managers

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/campaigncontainers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/campaigns"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogcontainers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/configs"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/devices"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/instances"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/jobs"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/models"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/reference"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/sites"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/skills"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutioncontainers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/staging"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/sync"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/targets"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/trails"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/users"
	cm "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
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
	case "managers.symphony.solutioncontainers":
		manager = &solutioncontainers.SolutionContainersManager{}
	case "managers.symphony.instances":
		manager = &instances.InstancesManager{}
	case "managers.symphony.users":
		manager = &users.UsersManager{}
	case "managers.symphony.jobs":
		manager = &jobs.JobsManager{}
	case "managers.symphony.campaigns":
		manager = &campaigns.CampaignsManager{}
	case "managers.symphony.campaigncontainers":
		manager = &campaigncontainers.CampaignContainersManager{}
	case "managers.symphony.catalogs":
		manager = &catalogs.CatalogsManager{}
	case "managers.symphony.catalogcontainers":
		manager = &catalogcontainers.CatalogContainersManager{}
	case "managers.symphony.activations":
		manager = &activations.ActivationsManager{}
	case "managers.symphony.activationscleanup":
		manager = &activations.ActivationsCleanupManager{}
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
