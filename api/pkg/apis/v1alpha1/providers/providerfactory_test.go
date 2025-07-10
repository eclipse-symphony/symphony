/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package providers

import (
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	catalogconfig "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/config/catalog"
	memorygraph "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph/memory"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/counter"
	symphonystage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/create"
	delaystage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/delay"
	httpstage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/http"
	liststage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/list"
	materialize "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/materialize"
	mockstage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/mock"
	patchstage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/patch"
	remotestage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/remote"
	scriptstage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/script"
	waitstage "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage/wait"
	k8sstate "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/states/k8s"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/adb"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/azure/adu"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/azure/iotedge"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/configmap"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/docker"
	targethttp "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/http"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/ingress"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/k8s"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/kubectl"
	tgtmock "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/mock"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/piccolo"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/proxy"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/script"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/staging"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/win10/sideload"
	mockconfig "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config/mock"
	mockledger "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/ledger/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/probe/rtsp"
	mempubsub "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	memoryqueue "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/queue/memory"
	cvref "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference/customvision"
	httpref "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference/http"
	k8sref "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference/k8s"
	httpreporter "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/http"
	k8sreporter "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/k8s"
	mocksecret "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/httpstate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/redisstate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/uploader/azure/blob"
	"github.com/stretchr/testify/assert"
)

func TestCreateProvider(t *testing.T) {
	getTestMiniKubeEnabled := os.Getenv("TEST_MINIKUBE_ENABLED")
	testRedis := os.Getenv("TEST_REDIS")

	providerfactory := SymphonyProviderFactory{}
	provider, err := providerfactory.CreateProvider("providers.state.memory", memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*memorystate.MemoryStateProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping providers.state.k8s test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.state.k8s", k8sstate.K8sStateProviderConfig{})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8sstate.K8sStateProvider))
	}

	if testRedis == "" {
		t.Log("Skipping providers.state.redis test as TEST_REDIS is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.state.redis", redisstate.RedisStateProviderConfig{Host: "localhost:6379"})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*redisstate.RedisStateProvider))
	}

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping providers.config.k8scatalog test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.config.k8scatalog", k8sstate.K8sStateProviderConfig{})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8sstate.K8sStateProvider))
	}

	provider, err = providerfactory.CreateProvider("providers.state.http", httpstate.HttpStateProviderConfig{Url: "http://localhost:3500/v1.0/state/statestore"})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*httpstate.HttpStateProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping providers.reference.k8s test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.reference.k8s", k8sref.K8sReferenceProviderConfig{})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8sref.K8sReferenceProvider))
	}

	provider, err = providerfactory.CreateProvider("providers.reference.customvision", cvref.CustomVisionReferenceProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*cvref.CustomVisionReferenceProvider))

	provider, err = providerfactory.CreateProvider("providers.reference.http", httpref.HTTPReferenceProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*httpref.HTTPReferenceProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping providers.reporter.k8s test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.reporter.k8s", k8sreporter.K8sReporterConfig{})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8sreporter.K8sReporter))
	}

	provider, err = providerfactory.CreateProvider("providers.reporter.http", httpreporter.HTTPReporterConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*httpreporter.HTTPReporter))

	provider, err = providerfactory.CreateProvider("providers.probe.rtsp", rtsp.RTSPProbeProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*rtsp.RTSPProbeProvider))

	provider, err = providerfactory.CreateProvider("providers.uploader.azure.blob", blob.AzureBlobUploaderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*blob.AzureBlobUploader))

	provider, err = providerfactory.CreateProvider("providers.ledger.mock", mockledger.MockLedgerProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mockledger.MockLedgerProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.counter", counter.CounterStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*counter.CounterStageProvider))

	provider, err = providerfactory.CreateProvider("providers.target.azure.iotedge", iotedge.IoTEdgeTargetProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*iotedge.IoTEdgeTargetProvider))

	provider, err = providerfactory.CreateProvider("providers.target.azure.adu", adu.ADUTargetProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*adu.ADUTargetProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping providers.target.k8s test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.target.k8s", k8s.K8sTargetProviderConfig{ConfigType: "path"})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8s.K8sTargetProvider))
	}

	provider, err = providerfactory.CreateProvider("providers.target.docker", docker.DockerTargetProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*docker.DockerTargetProvider))

	provider, err = providerfactory.CreateProvider("providers.target.piccolo", piccolo.PiccoloTargetProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*piccolo.PiccoloTargetProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping providers.target.ingress test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.target.ingress", ingress.IngressTargetProviderConfig{ConfigType: "path"})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*ingress.IngressTargetProvider))
	}

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping providers.target.kubectl test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.target.kubectl", kubectl.KubectlTargetProviderConfig{ConfigType: "path"})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*kubectl.KubectlTargetProvider))
	}

	provider, err = providerfactory.CreateProvider("providers.target.staging", staging.StagingTargetProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*staging.StagingTargetProvider))

	provider, err = providerfactory.CreateProvider("providers.target.script", script.ScriptProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*script.ScriptProvider))

	provider, err = providerfactory.CreateProvider("providers.target.http", targethttp.HttpTargetProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*targethttp.HttpTargetProvider))

	provider, err = providerfactory.CreateProvider("providers.target.win10.sideload", sideload.Win10SideLoadProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*sideload.Win10SideLoadProvider))

	provider, err = providerfactory.CreateProvider("providers.target.adb", adb.AdbProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*adb.AdbProvider))

	provider, err = providerfactory.CreateProvider("providers.target.proxy", proxy.ProxyUpdateProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*proxy.ProxyUpdateProvider))

	provider, err = providerfactory.CreateProvider("providers.target.mock", tgtmock.MockTargetProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*tgtmock.MockTargetProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping providers.target.configmap test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = providerfactory.CreateProvider("providers.target.configmap", configmap.ConfigMapTargetProviderConfig{ConfigType: "path"})
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*configmap.ConfigMapTargetProvider))
	}

	provider, err = providerfactory.CreateProvider("providers.config.mock", mockconfig.MockConfigProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mockconfig.MockConfigProvider))

	provider, err = providerfactory.CreateProvider("providers.config.catalog", catalogconfig.CatalogConfigProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*catalogconfig.CatalogConfigProvider))

	provider, err = providerfactory.CreateProvider("providers.secret.mock", mocksecret.MockSecretProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mocksecret.MockSecretProvider))

	provider, err = providerfactory.CreateProvider("providers.pubsub.memory", mempubsub.InMemoryPubSubConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mempubsub.InMemoryPubSubProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.mock", mockstage.MockStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mockstage.MockStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.http", httpstage.HttpStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*httpstage.HttpStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.create", symphonystage.CreateStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*symphonystage.CreateStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.script", scriptstage.ScriptStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*scriptstage.ScriptStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.patch", patchstage.PatchStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*patchstage.PatchStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.list", liststage.ListStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*liststage.ListStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.remote", remotestage.RemoteStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*remotestage.RemoteStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.wait", waitstage.WaitStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*waitstage.WaitStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.delay", delaystage.DelayStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*delaystage.DelayStageProvider))

	provider, err = providerfactory.CreateProvider("providers.stage.materialize", materialize.MaterializeStageProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*materialize.MaterializeStageProvider))

	provider, err = providerfactory.CreateProvider("providers.queue.memory", memoryqueue.MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*memoryqueue.MemoryQueueProvider))

	provider, err = providerfactory.CreateProvider("providers.graph.memory", memorygraph.MemoryGraphProviderConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*memorygraph.MemoryGraphProvider))
}

func TestCreateProviderForTargetRole(t *testing.T) {
	getTestMiniKubeEnabled := os.Getenv("TEST_MINIKUBE_ENABLED")

	targetState := model.TargetState{
		Spec: &model.TargetSpec{
			DisplayName: "target",
			Topologies: []model.TopologySpec{
				{
					Bindings: []model.BindingSpec{
						{
							Role:     "memorystate",
							Provider: "providers.state.memory",
							Config:   map[string]string{},
						},
						{
							Role:     "k8sstate",
							Provider: "providers.state.k8s",
							Config:   map[string]string{},
						},
						{
							Role:     "redisstate",
							Provider: "providers.state.redis",
							Config:   map[string]string{},
						},
						{
							Role:     "k8scatalog",
							Provider: "providers.config.k8scatalog",
							Config:   map[string]string{},
						},
						{
							Role:     "httpstate",
							Provider: "providers.state.http",
							Config: map[string]string{
								"url": "http://localhost:3500/v1.0/state/statestore",
							},
						},
						{
							Role:     "k8sref",
							Provider: "providers.reference.k8s",
							Config:   map[string]string{},
						},
						{
							Role:     "cvref",
							Provider: "providers.reference.customvision",
							Config: map[string]string{
								"key": "fakekey",
							},
						},
						{
							Role:     "httpref",
							Provider: "providers.reference.http",
							Config:   map[string]string{},
						},
						{
							Role:     "mockledger",
							Provider: "providers.ledger.mock",
							Config:   map[string]string{},
						},
						{
							Role:     "counter",
							Provider: "providers.stage.counter",
							Config:   map[string]string{},
						},
						{
							Role:     "k8s",
							Provider: "providers.target.k8s",
							Config:   map[string]string{},
						},
						{
							Role:     "docker",
							Provider: "providers.target.docker",
							Config:   map[string]string{},
						},
						{
							Role:     "piccolo",
							Provider: "providers.target.piccolo",
							Config:   map[string]string{},
						},
						{
							Role:     "ingress",
							Provider: "providers.target.ingress",
							Config: map[string]string{
								"configType": "path",
							},
						},
						{
							Role:     "kubectl",
							Provider: "providers.target.kubectl",
							Config: map[string]string{
								"configType": "path",
							},
						},
						{
							Role:     "staging",
							Provider: "providers.target.staging",
							Config: map[string]string{
								"targetName": "target1",
							},
						},
						{
							Role:     "script",
							Provider: "providers.target.script",
							Config: map[string]string{
								"applyScript":  "mock-apply.sh",
								"removeScript": "mock-remove.sh",
								"getScript":    "mock-get.sh",
								"scriptFolder": "",
							},
						},
						{
							Role:     "http",
							Provider: "providers.target.http",
							Config:   map[string]string{},
						},
						{
							Role:     "win10sideload",
							Provider: "providers.target.win10.sideload",
							Config:   map[string]string{},
						},
						{
							Role:     "adb",
							Provider: "providers.target.adb",
							Config:   map[string]string{},
						},
						{
							Role:     "proxy",
							Provider: "providers.target.proxy",
							Config: map[string]string{
								"name":      "proxy",
								"serverUrl": "",
							},
						},
						{
							Role:     "configmap",
							Provider: "providers.target.configmap",
							Config: map[string]string{
								"configType": "path",
							},
						},
						{
							Role:     "mock",
							Provider: "providers.target.mock",
							Config:   map[string]string{},
						},
						{
							Role:     "azureiotedge",
							Provider: "providers.target.azure.iotedge",
							Config: map[string]string{
								"name":       "iot-edge",
								"keyName":    "iothubowner",
								"key":        "fakekey",
								"iotHub":     "fakenet",
								"apiVersion": "fakeversion",
								"deviceName": "s8c-vm",
							},
						},
						{
							Role:     "azureadu",
							Provider: "providers.target.azure.adu",
							Config: map[string]string{
								"name":               "adu",
								"tenantId":           "faketenant",
								"clientId":           "fakeclient",
								"clientSecret":       "fakesecret",
								"aduAccountEndpoint": "fakeendpoint",
								"aduAccountInstance": "fakeinstance",
								"aduGroup":           "fakegroup",
							},
						},
						{
							Role:     "mockconfig",
							Provider: "providers.config.mock",
							Config:   map[string]string{},
						},
						{
							Role:     "catalogconfig",
							Provider: "providers.config.catalog",
							Config: map[string]string{
								"baseUrl":  "fakeuri",
								"user":     "fake",
								"password": "",
							},
						},
						{
							Role:     "mocksecret",
							Provider: "providers.secret.mock",
							Config:   map[string]string{},
						},
						{
							Role:     "memoryqueue",
							Provider: "providers.queue.memory",
							Config:   map[string]string{},
						},
						{
							Role:     "memorygraph",
							Provider: "providers.graph.memory",
							Config:   map[string]string{},
						},
						{
							Role:     "mockstage",
							Provider: "providers.stage.mock",
							Config:   map[string]string{},
						},
						{
							Role:     "httpstage",
							Provider: "providers.stage.http",
							Config: map[string]string{
								"url":    "fakeurl",
								"method": "GET",
							},
						},
						{
							Role:     "createstage",
							Provider: "providers.stage.create",
							Config: map[string]string{
								"baseUrl":       "fakeUrl",
								"user":          "admin",
								"password":      "",
								"wait.count":    "1",
								"wait.interval": "1",
							},
						},
						{
							Role:     "scriptstage",
							Provider: "providers.stage.script",
							Config: map[string]string{
								"script": "",
							},
						},
						{
							Role:     "patchstage",
							Provider: "providers.stage.patch",
							Config: map[string]string{
								"baseUrl":  "fakeUrl",
								"user":     "admin",
								"password": "",
							},
						},
						{
							Role:     "liststage",
							Provider: "providers.stage.list",
							Config: map[string]string{
								"baseUrl":  "fakeUrl",
								"user":     "admin",
								"password": "",
							},
						},
						{
							Role:     "remotestage",
							Provider: "providers.stage.remote",
							Config: map[string]string{
								"baseUrl":  "fakeUrl",
								"user":     "admin",
								"password": "",
							},
						},
						{
							Role:     "waitstage",
							Provider: "providers.stage.wait",
							Config: map[string]string{
								"baseUrl":  "fakeUrl",
								"user":     "admin",
								"password": "",
							},
						},
						{
							Role:     "delaystage",
							Provider: "providers.stage.delay",
							Config: map[string]string{
								"baseUrl":  "fakeUrl",
								"user":     "admin",
								"password": "",
							},
						},
						{
							Role:     "materializestage",
							Provider: "providers.stage.materialize",
							Config: map[string]string{
								"baseUrl":  "fakeUrl",
								"user":     "admin",
								"password": "",
							},
						},
						{
							Role:     "mempubsub",
							Provider: "providers.pubsub.memory",
							Config:   map[string]string{},
						},
						{
							Role:     "httpreporter",
							Provider: "providers.reporter.http",
							Config:   map[string]string{},
						},
						{
							Role:     "k8sreporter",
							Provider: "providers.reporter.k8s",
							Config:   map[string]string{},
						},
					},
				},
			},
		},
	}

	provider, err := CreateProviderForTargetRole(nil, "memorystate", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*memorystate.MemoryStateProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping k8sstate test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = CreateProviderForTargetRole(nil, "k8sstate", targetState, nil)
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8sstate.K8sStateProvider))
	}

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping k8scatalog test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = CreateProviderForTargetRole(nil, "k8scatalog", targetState, nil)
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8sstate.K8sStateProvider))
	}

	provider, err = CreateProviderForTargetRole(nil, "httpstate", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*httpstate.HttpStateProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping k8sref test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = CreateProviderForTargetRole(nil, "k8sref", targetState, nil)
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8sref.K8sReferenceProvider))
	}

	provider, err = CreateProviderForTargetRole(nil, "cvref", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*cvref.CustomVisionReferenceProvider))

	provider, err = CreateProviderForTargetRole(nil, "httpref", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*httpref.HTTPReferenceProvider))

	provider, err = CreateProviderForTargetRole(nil, "mockledger", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mockledger.MockLedgerProvider))

	provider, err = CreateProviderForTargetRole(nil, "counter", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*counter.CounterStageProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping k8s test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = CreateProviderForTargetRole(nil, "k8s", targetState, nil)
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8s.K8sTargetProvider))
	}

	provider, err = CreateProviderForTargetRole(nil, "docker", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*docker.DockerTargetProvider))

	provider, err = CreateProviderForTargetRole(nil, "piccolo", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*piccolo.PiccoloTargetProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping ingress test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = CreateProviderForTargetRole(nil, "ingress", targetState, nil)
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*ingress.IngressTargetProvider))
	}

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping kubectl test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = CreateProviderForTargetRole(nil, "kubectl", targetState, nil)
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*kubectl.KubectlTargetProvider))
	}

	provider, err = CreateProviderForTargetRole(nil, "staging", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*staging.StagingTargetProvider))

	provider, err = CreateProviderForTargetRole(nil, "script", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*script.ScriptProvider))

	provider, err = CreateProviderForTargetRole(nil, "http", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*targethttp.HttpTargetProvider))

	provider, err = CreateProviderForTargetRole(nil, "win10sideload", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*sideload.Win10SideLoadProvider))

	provider, err = CreateProviderForTargetRole(nil, "adb", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*adb.AdbProvider))

	provider, err = CreateProviderForTargetRole(nil, "proxy", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*proxy.ProxyUpdateProvider))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping configmap test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = CreateProviderForTargetRole(nil, "configmap", targetState, nil)
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*configmap.ConfigMapTargetProvider))
	}

	provider, err = CreateProviderForTargetRole(nil, "mock", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*tgtmock.MockTargetProvider))

	provider, err = CreateProviderForTargetRole(nil, "azureiotedge", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*iotedge.IoTEdgeTargetProvider))

	provider, err = CreateProviderForTargetRole(nil, "azureadu", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*adu.ADUTargetProvider))

	provider, err = CreateProviderForTargetRole(nil, "mockconfig", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mockconfig.MockConfigProvider))

	provider, err = CreateProviderForTargetRole(nil, "catalogconfig", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*catalogconfig.CatalogConfigProvider))

	provider, err = CreateProviderForTargetRole(nil, "mocksecret", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mocksecret.MockSecretProvider))

	provider, err = CreateProviderForTargetRole(nil, "memoryqueue", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*memoryqueue.MemoryQueueProvider))

	provider, err = CreateProviderForTargetRole(nil, "memorygraph", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*memorygraph.MemoryGraphProvider))

	provider, err = CreateProviderForTargetRole(nil, "mockstage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mockstage.MockStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "httpstage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*httpstage.HttpStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "createstage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*symphonystage.CreateStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "scriptstage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*scriptstage.ScriptStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "patchstage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*patchstage.PatchStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "liststage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*liststage.ListStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "remotestage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*remotestage.RemoteStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "waitstage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*waitstage.WaitStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "delaystage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*delaystage.DelayStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "materializestage", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*materialize.MaterializeStageProvider))

	provider, err = CreateProviderForTargetRole(nil, "mempubsub", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*mempubsub.InMemoryPubSubProvider))

	provider, err = CreateProviderForTargetRole(nil, "httpreporter", targetState, nil)
	assert.Nil(t, err)
	assert.NotNil(t, *provider.(*httpreporter.HTTPReporter))

	if getTestMiniKubeEnabled == "" {
		t.Log("Skipping k8sreporter test as TEST_MINIKUBE_ENABLED is not set")
	} else {
		provider, err = CreateProviderForTargetRole(nil, "k8sreporter", targetState, nil)
		assert.Nil(t, err)
		assert.NotNil(t, *provider.(*k8sreporter.K8sReporter))
	}
}
