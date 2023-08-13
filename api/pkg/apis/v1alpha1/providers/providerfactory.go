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

package providers

import (
	"fmt"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	symphonystage "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage/create"
	httpstage "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage/http"
	liststage "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage/list"
	mockstage "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage/mock"
	remotestage "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage/remote"
	k8sstate "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/states/k8s"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/adb"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/arcextension"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/azure/adu"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/azure/iotedge"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/docker"
	extendedlocation "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/extendedlocation"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/helm"
	targethttp "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/http"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/k8s"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/kubectl"
	tgtmock "github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/mock"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/mqtt"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/proxy"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/script"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/staging"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/win10/sideload"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	cp "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	mockconfig "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config/mock"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/probe/rtsp"
	mempubsub "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	cvref "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/reference/customvision"
	httpref "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/reference/http"
	k8sref "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/reference/k8s"
	httpreporter "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/reporter/http"
	k8sreporter "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/reporter/k8s"
	mocksecret "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	memorystack "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/stack/memory"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states/httpstate"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/uploader/azure/blob"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
)

type SymphonyProviderFactory struct {
}

// CreateProviders initializes the config for the providers from the vendor config
func (c SymphonyProviderFactory) CreateProviders(config vendors.VendorConfig) (map[string]map[string]cp.IProvider, error) {
	ret := make(map[string]map[string]cp.IProvider)
	for _, m := range config.Managers {
		ret[m.Name] = make(map[string]cp.IProvider)
		for k, p := range m.Providers {
			provider, err := c.CreateProvider(p.Type, p.Config)
			if err != nil {
				return ret, err
			}
			if provider != nil {
				ret[m.Name][k] = provider
			}
		}
	}
	return ret, nil
}

func (s SymphonyProviderFactory) CreateProvider(providerType string, config cp.IProviderConfig) (cp.IProvider, error) {
	var err error
	switch providerType {
	case "providers.state.memory":
		mProvider := &memorystate.MemoryStateProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.state.k8s":
		mProvider := &k8sstate.K8sStateProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.state.http":
		mProvider := &httpstate.HttpStateProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.reference.k8s":
		mProvider := &k8sref.K8sReferenceProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.reference.customvision":
		mProvider := &cvref.CustomVisionReferenceProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.reference.http":
		mProvider := &httpref.HTTPReferenceProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.reporter.k8s":
		mProvider := &k8sreporter.K8sReporter{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.reporter.http":
		mProvider := &httpreporter.HTTPReporter{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.probe.rtsp":
		mProvider := &rtsp.RTSPProbeProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.uploader.azure.blob":
		mProvider := &blob.AzureBlobUploader{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.azure.iotedge":
		mProvider := &iotedge.IoTEdgeTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.azure.adu":
		mProvider := &adu.ADUTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.k8s":
		mProvider := &k8s.K8sTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.docker":
		mProvider := &docker.DockerTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.kubectl":
		mProvider := &kubectl.KubectlTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.arcextension":
		mProvider := &arcextension.ArcExtensionTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.staging":
		mProvider := &staging.StagingTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.script":
		mProvider := &script.ScriptProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.http":
		mProvider := &targethttp.HttpTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.win10.sideload":
		mProvider := &sideload.Win10SideLoadProvider{}
		err := mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.adb":
		mProvider := &adb.AdbProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.proxy":
		mProvider := &proxy.ProxyUpdateProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.mqtt":
		mProvider := &mqtt.MQTTTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.target.mock":
		mProvider := &tgtmock.MockTargetProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.config.mock":
		mProvider := &mockconfig.MockConfigProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.secret.mock":
		mProvider := &mocksecret.MockSecretProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.pubsub.memory":
		mProvider := &mempubsub.InMemoryPubSubProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.stage.mock":
		mProvider := &mockstage.MockStageProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.stage.http":
		mProvider := &httpstage.HttpStageProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.stage.create":
		mProvider := &symphonystage.CreateStageProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.stage.list":
		mProvider := &liststage.ListStageProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.stage.remote":
		mProvider := &remotestage.RemoteStageProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	case "providers.stack.memory":
		mProvider := &memorystack.MemoryStackProvider{}
		err = mProvider.Init(config)
		if err == nil {
			return mProvider, nil
		}
	}
	return nil, err //TODO: in current design, factory doesn't return errors on unrecognized provider types as there could be other factories. We may want to change this.
}
func CreateProviderForTargetRole(role string, target model.TargetSpec, override cp.IProvider) (cp.IProvider, error) {
	for _, topology := range target.Topologies {
		for _, binding := range topology.Bindings {
			testRole := role
			if role == "" || role == "container" {
				testRole = "instance"
			}
			if binding.Role == testRole {
				switch binding.Provider {
				case "providers.target.azure.iotedge":
					provider := &iotedge.IoTEdgeTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.azure.adu":
					provider := &adu.ADUTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.k8s":
					provider := &k8s.K8sTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.arcextension":
					provider := &arcextension.ArcExtensionTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.docker":
					provider := &docker.DockerTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.kubectl":
					provider := &kubectl.KubectlTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.staging":
					provider := &staging.StagingTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.script":
					provider := &script.ScriptProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.http":
					provider := &targethttp.HttpTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.win10.sideload":
					provider := &sideload.Win10SideLoadProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.adb":
					provider := &adb.AdbProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.proxy":
					if override == nil {
						provider := &proxy.ProxyUpdateProvider{}
						err := provider.InitWithMap(binding.Config)
						if err != nil {
							return nil, err
						}
						return provider, nil
					} else {
						return override, nil
					}
				case "providers.target.mqtt":
					if override == nil {
						provider := &mqtt.MQTTTargetProvider{}
						err := provider.InitWithMap(binding.Config)
						if err != nil {
							return nil, err
						}
						return provider, nil
					} else {
						return override, nil
					}
				case "providers.state.memory":
					provider := &memorystate.MemoryStateProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.state.k8s":
					provider := &k8sstate.K8sStateProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.state.http":
					provider := &httpstate.HttpStateProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.reference.k8s":
					provider := &k8sref.K8sReferenceProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.reference.customvision":
					provider := &cvref.CustomVisionReferenceProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.reference.http":
					provider := &httpref.HTTPReferenceProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.reporter.k8s":
					provider := &k8sreporter.K8sReporter{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.reporter.http":
					provider := &httpreporter.HTTPReporter{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.helm":
					provider := &helm.HelmTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.extendedlocation":
					provider := &extendedlocation.ExtendedLocationTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.target.mock":
					provider := &tgtmock.MockTargetProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.config.mock":
					provider := &mockconfig.MockConfigProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.secret.mock":
					provider := &mocksecret.MockSecretProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.stage.mock":
					provider := &mockstage.MockStageProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.stage.remote":
					provider := &remotestage.RemoteStageProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
				case "providers.stage.http":
					provider := &httpstage.HttpStageProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.stage.create":
					provider := &symphonystage.CreateStageProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
					return provider, nil
				case "providers.stage.list":
					provider := &liststage.ListStageProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
				case "providers.stack.memory":
					provider := &memorystack.MemoryStackProvider{}
					err := provider.InitWithMap(binding.Config)
					if err != nil {
						return nil, err
					}
				}

			}
		}
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("target doesn't have a '%s' role defined", role), v1alpha2.BadConfig)
}
