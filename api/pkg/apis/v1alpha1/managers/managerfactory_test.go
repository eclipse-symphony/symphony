/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package managers

import (
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/campaigns"
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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/staging"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/sync"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/targets"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/trails"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/users"
	cm "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/stretchr/testify/assert"
)

func testCreateManager[T cm.IManager](t *testing.T, config cm.ManagerConfig) {
	symphonyManagerFactory := SymphonyManagerFactory{}

	manager, err := symphonyManagerFactory.CreateManager(config)
	assert.Nil(t, err)
	assert.NotNil(t, manager)
	manager = manager.(T)
	assert.NotNil(t, manager)
}

func TestCreateManager(t *testing.T) {
	testCreateManager[*solution.SolutionManager](t, getSolutionManagerConfig())
	testCreateManager[*reference.ReferenceManager](t, getReferenceManagerConfig())
	testCreateManager[*target.TargetManager](t, getTargetManagerConfig())
	testCreateManager[*targets.TargetsManager](t, getTargetsManagerConfig())
	testCreateManager[*devices.DevicesManager](t, getDevicesManagerConfig())
	testCreateManager[*solutions.SolutionsManager](t, getSolutionsManagerConfig())
	testCreateManager[*instances.InstancesManager](t, getInstancesManagerConfig())
	testCreateManager[*users.UsersManager](t, getUsersManagerConfig())
	testCreateManager[*jobs.JobsManager](t, getJobsManagerConfig())
	testCreateManager[*campaigns.CampaignsManager](t, getCampaignsManagerConfig())
	testCreateManager[*catalogs.CatalogsManager](t, getCatalogsManagerConfig())
	testCreateManager[*activations.ActivationsManager](t, getActivationsManagerConfig())
	testCreateManager[*activations.ActivationsCleanupManager](t, getActivationsCleanupManagerConfig())
	testCreateManager[*stage.StageManager](t, getStageManagerConfig())
	testCreateManager[*configs.ConfigsManager](t, getConfigsManagerConfig())
	testCreateManager[*sites.SitesManager](t, getSitesManagerConfig())
	testCreateManager[*staging.StagingManager](t, getStagingManagerConfig())
	testCreateManager[*sync.SyncManager](t, getSyncManagerConfig())
	testCreateManager[*models.ModelsManager](t, getModelsManagerConfig())
	testCreateManager[*skills.SkillsManager](t, getSkillsManagerConfig())
	testCreateManager[*trails.TrailsManager](t, getTrailsManagerConfig())
}

func getSolutionManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.solution",
		Properties: map[string]string{
			"providers.volatilestate": "mem-state",
			"providers.config":        "mock-config",
			"providers.secret":        "mock-secret",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
			"mock-config": {
				Type: "providers.config.mock",
			},
			"mock-secret": {
				Type: "providers.secret.mock",
			},
		},
	}
}

func getReferenceManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.reference",
		Properties: map[string]string{
			"providers.reference":     "http-reference",
			"providers.volatilestate": "memory",
			"providers.reporter":      "http-reporter",
		},
		Providers: map[string]cm.ProviderConfig{
			"memory": {
				Type: "providers.symphony.state",
			},
			"http-reference": {
				Type: "providers.reference.http",
				Config: map[string]string{
					"url": "http://localhost:8080",
				},
			},
			"http-reporter": {
				Type: "providers.reporter.http",
				Config: map[string]string{
					"url": "http://localhost:8080",
				},
			},
		},
	}
}

func getTargetManagerConfig() cm.ManagerConfig {
	// symphony-agent.json
	return cm.ManagerConfig{
		Type: "managers.symphony.target",
		Properties: map[string]string{
			"providers.probe":     "rtsp-probe",
			"providers.reference": "http-reference",
			"providers.uploader":  "azure-uploader",
			"providers.reporter":  "http-reporter",
			"poll.enabled":        "true",
		},
		Providers: map[string]cm.ProviderConfig{
			"rtsp-probe": {
				Type: "providers.probe.rtsp",
			},
			"http-reference": {
				Type: "providers.reference.http",
				Config: map[string]string{
					"url":    "http://localhost:8080",
					"target": "target_name",
				},
			},
			"http-reporter": {
				Type: "providers.reporter.http",
				Config: map[string]string{
					"url": "http://localhost:8080",
				},
			},
			"azure-uploader": {
				Type: "providers.uploader.azure.blob",
				Config: map[string]string{
					"account":   "account",
					"container": "container",
				},
			},
		},
	}
}

func getTargetsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.targets",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getDevicesManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.devices",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getSolutionsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.solutions",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getInstancesManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.instances",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getUsersManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.users",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getJobsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.jobs",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
			"baseUrl":                   "http://localhost/",
			"user":                      "admin",
			"password":                  "",
			"interval":                  "#15",
			"poll.enabled":              "true",
			"schedule.enabled":          "true",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getCampaignsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.campaigns",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
			"singleton":                 "true",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getCatalogsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.catalogs",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
			"singleton":                 "true",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getActivationsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.activations",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
			"singleton":                 "true",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getActivationsCleanupManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.activationscleanup",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
			"singleton":                 "true",
			"RetentionInMinutes":        "1440",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getStageManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.stage",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
			"baseUrl":                   "http://localhost:8082/v1alpha2/",
			"user":                      "admin",
			"password":                  "",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.symphony.state",
			},
		},
	}
}

func getConfigsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.configs",
		Properties: map[string]string{
			"singleton": "true",
		},
		Providers: map[string]cm.ProviderConfig{
			"catalog": {
				Type: "providers.config.catalog",
				Config: map[string]string{
					"baseUrl":  "http://localhost:8082/v1alpha2/",
					"user":     "admin",
					"password": "",
				},
			},
		},
	}
}

func getSitesManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.sites",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.state.memory",
			},
		},
	}
}

func getStagingManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.staging",
		Properties: map[string]string{
			"poll.enabled":              "true",
			"interval":                  "#15",
			"providers.queue":           "memory-queue",
			"providers.persistentstate": "memory-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"memory-state": {
				Type: "providers.state.memory",
			},
			"memory-queue": {
				Type: "providers.queue.memory",
			},
		},
	}
}

func getSyncManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.sync",
		Properties: map[string]string{
			"baseUrl":      "http://localhost:8080/v1alpha2/",
			"user":         "admin",
			"password":     "",
			"interval":     "#15",
			"sync.enabled": "true",
		},
	}
}

func getModelsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.models",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.state.memory",
			},
		},
	}
}

func getSkillsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.skills",
		Properties: map[string]string{
			"providers.persistentstate": "mem-state",
		},
		Providers: map[string]cm.ProviderConfig{
			"mem-state": {
				Type: "providers.state.memory",
			},
		},
	}
}

func getTrailsManagerConfig() cm.ManagerConfig {
	// symphony-api-no-k8s.json
	return cm.ManagerConfig{
		Type: "managers.symphony.trails",
		Providers: map[string]cm.ProviderConfig{
			"mock": {
				Type: "providers.ledger.mock",
			},
		},
	}
}
