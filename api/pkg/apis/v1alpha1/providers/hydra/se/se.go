/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package se

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var aLog = logger.NewLogger("coa.runtime")

type DesiredState struct {
	Apps    []App    `json:"apps"`
	Devices []Device `json:"devices"`
}

type App struct {
	Version  string     `json:"version,omitempty"`
	Kind     string     `json:"kind"`
	Metadata Metadata   `json:"metadata"`
	Spec     AppSpec    `json:"spec"`
	Status   *AppStatus `json:"status,omitempty"`
}

type Metadata struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type AppSpec struct {
	Container ContainerSpec `json:"container"`
	Affinity  Affinity      `json:"affinity"`
}

type ContainerSpec struct {
	Name      string          `json:"name"`
	Image     string          `json:"image"`
	Networks  []Network       `json:"networks"`
	Resources ContainerLimits `json:"resources"`
}

type Network struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type ContainerLimits struct {
	Limits ResourceLimits `json:"limits"`
}

type ResourceLimits struct {
	Memory string `json:"memory"`
	CPUs   string `json:"cpus"`
}

type Affinity struct {
	PreferredHosts  []string `json:"preferredHosts"`
	AppAntiAffinity []string `json:"appAntiAffinity"`
}

type AppStatus struct {
	Status      string `json:"status"`
	TimeStamp   string `json:"timeStamp"`
	RunningHost string `json:"runningHost"`
}

type Device struct {
	Kind     string     `json:"kind"`
	Metadata Metadata   `json:"metadata"`
	Spec     DeviceSpec `json:"spec"`
}

type DeviceSpec struct {
	Addresses         []string           `json:"addresses"`
	Networks          []DeviceNetwork    `json:"networks,omitempty"`
	ContainerNetworks []ContainerNetwork `json:"containerNetworks"`
}

type DeviceNetwork struct {
	NICList        []string `json:"nicList"`
	NetName        string   `json:"netName"`
	NICName        string   `json:"nicName"`
	RedundancyMode string   `json:"redundancyMode"`
	IPv4           string   `json:"ipv4"`
	Gateway        string   `json:"gateway"`
}

type ContainerNetwork struct {
	Subnet    string `json:"subnet"`
	Gateway   string `json:"gateway"`
	NetworkID string `json:"networkId"`
	NICName   string `json:"nicName"`
	Type      string `json:"type"`
}

type SEProviderConfig struct {
	Name  string `json:"name"`
	HASet string `json:"haSet"`
}

type SEProvider struct {
	Config  SEProviderConfig
	Context *contexts.ManagerContext
}

func SEProviderConfigFromMap(properties map[string]string) (SEProviderConfig, error) {
	ret := SEProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["haSet"]; ok {
		ret.HASet = v
	}
	return ret, nil
}

func (i *SEProvider) InitWithMap(properties map[string]string) error {
	config, err := SEProviderConfigFromMap(properties)
	if err != nil {
		aLog.Errorf("  P (SE Hydra): expected SEProviderConfig: %+v", err)
		return err
	}
	return i.Init(config)
}
func (s *SEProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *SEProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("SE Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	aLog.InfoCtx(ctx, "  P (SE Hydra): Init()")

	updateConfig, err := toSEProviderConfig(config)
	if err != nil {
		aLog.ErrorfCtx(ctx, "  P (SE Hydra): expected SEProviderConfig: %+v", err)
		return errors.New("expected SEProviderConfig")
	}
	i.Config = updateConfig
	return nil
}

func toSEProviderConfig(config providers.IProviderConfig) (SEProviderConfig, error) {
	ret := SEProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *SEProvider) GetArtifact(system string, artifacts model.ArtifactPack) ([]byte, error) {
	return nil, nil
}

func (i *SEProvider) SetArtifact(system string, artifact []byte) (model.ArtifactPack, error) {
	ctx, span := observability.StartSpan("SE Provider", context.TODO(), &map[string]string{
		"method": "SetArtifact",
	})
	defer observ_utils.CloseSpanWithError(span, nil)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, nil)

	var desiredState DesiredState
	err := json.Unmarshal(artifact, &desiredState)
	if err != nil {
		aLog.ErrorfCtx(ctx, "  P (SE Hydra): SetArtifact failed - %s", err.Error())
		return model.ArtifactPack{}, err
	}

	ret := model.ArtifactPack{
		Targets:            []model.TargetState{},
		SolutionContainers: []model.SolutionContainerState{},
		Solutions:          []model.SolutionState{},
		Instances:          []model.InstanceState{},
	}

	lastSeenDeviceKind := ""
	for _, device := range desiredState.Devices {
		if lastSeenDeviceKind == "" {
			lastSeenDeviceKind = device.Kind
		} else if device.Kind != lastSeenDeviceKind {
			err := v1alpha2.NewCOAError(nil, "devices must be of the same kind", v1alpha2.InvalidArgument)
			aLog.ErrorfCtx(ctx, "  P (SE Hydra): SetArtifact failed - %s", err.Error())
			return model.ArtifactPack{}, err
		}
		target := model.TargetState{
			ObjectMeta: model.ObjectMeta{
				Name: device.Metadata.Name,
			},
			Spec: &model.TargetSpec{},
		}
		for k, v := range device.Metadata.Labels {
			if target.ObjectMeta.Labels == nil {
				target.ObjectMeta.Labels = make(map[string]string)
			}
			target.ObjectMeta.Labels[k] = v
		}
		target.ObjectMeta.Labels["kind"] = device.Kind
		if i.Config.HASet != "" {
			target.ObjectMeta.Labels["haSet"] = i.Config.HASet
		}
		target.Spec.DisplayName = device.Metadata.Name
		target.Spec.Properties = make(map[string]string)
		err = setPropertyValue(target.Spec.Properties, "addresses", device.Spec.Addresses)
		if err != nil {
			aLog.ErrorfCtx(ctx, "  P (SE Hydra): SetArtifact failed to set addresses - %s", err.Error())
			return model.ArtifactPack{}, err
		}
		err = setPropertyValue(target.Spec.Properties, "networks", device.Spec.Networks)
		if err != nil {
			aLog.ErrorfCtx(ctx, "  P (SE Hydra): SetArtifact failed to set networks - %s", err.Error())
			return model.ArtifactPack{}, err
		}
		err = setPropertyValue(target.Spec.Properties, "containerNetworks", device.Spec.ContainerNetworks)
		if err != nil {
			aLog.ErrorfCtx(ctx, "  P (SE Hydra): SetArtifact failed to set containerNetworks - %s", err.Error())
			return model.ArtifactPack{}, err
		}
		target.Spec.Topologies = []model.TopologySpec{
			{
				Bindings: []model.BindingSpec{
					{
						Role:     "container",
						Provider: "providers.target.se",
						Config:   map[string]string{},
					},
				},
			},
		}
		ret.Targets = append(ret.Targets, target)
	}

	selector := model.TargetSelector{
		LabelSelector: map[string]string{
			"haSet": i.Config.HASet,
			"kind":  lastSeenDeviceKind,
		},
	}
	selectorData, _ := json.Marshal(selector)

	if i.Config.HASet != "" {
		haTarget := model.TargetState{
			ObjectMeta: model.ObjectMeta{
				Name: i.Config.HASet,
				Labels: map[string]string{
					"haSet": i.Config.HASet,
					"kind":  "group",
				},
			},
			Spec: &model.TargetSpec{
				Components: []model.ComponentSpec{},
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "instance",
								Provider: "providers.target.group",
								Config: map[string]string{
									"user":           "admin",
									"password":       "",
									"targetSelector": string(selectorData),
								},
							},
						},
					},
				},
			},
		}
		ret.Targets = append(ret.Targets, haTarget)

		solutionContainer := model.SolutionContainerState{
			ObjectMeta: model.ObjectMeta{
				Name: i.Config.HASet,
				Labels: map[string]string{
					"haSet": i.Config.HASet,
					"kind":  "group",
				},
			},
			Spec: &model.SolutionContainerSpec{},
		}

		ret.SolutionContainers = append(ret.SolutionContainers, solutionContainer)

		solution := model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Name: i.Config.HASet + "-v-v1",
			},
			Spec: &model.SolutionSpec{
				RootResource: i.Config.HASet,
			},
		}

		for _, app := range desiredState.Apps {
			component := model.ComponentSpec{
				Name: app.Metadata.Name,
				Type: app.Kind,
				Metadata: map[string]string{
					"name": app.Metadata.Name,
				},
				Properties: map[string]interface{}{
					"version": app.Version,
					"image":   app.Spec.Container.Image,
					"affinity": map[string]interface{}{
						"preferredHosts":  app.Spec.Affinity.PreferredHosts,
						"appAntiAffinity": app.Spec.Affinity.AppAntiAffinity,
					},
					"containerNetworks": app.Spec.Container.Networks,
					"resources": map[string]interface{}{
						"limits": map[string]interface{}{
							"memory": app.Spec.Container.Resources.Limits.Memory,
							"cpus":   app.Spec.Container.Resources.Limits.CPUs,
						},
					},
				},
			}
			for k, v := range app.Metadata.Labels {
				if component.Metadata == nil {
					component.Metadata = make(map[string]string)
				}
				component.Metadata["labels."+k] = v
			}
			solution.Spec.Components = append(solution.Spec.Components, component)
		}

		ret.Solutions = append(ret.Solutions, solution)

		instance := model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: i.Config.HASet,
			},
			Spec: &model.InstanceSpec{
				DisplayName: i.Config.HASet,
				Solution:    i.Config.HASet + ":v1",
				Target: model.TargetSelector{
					LabelSelector: map[string]string{
						"haSet": i.Config.HASet,
						"kind":  "group",
					},
				},
			},
		}
		ret.Instances = append(ret.Instances, instance)
	}

	return ret, nil
}

func setPropertyValue(properties map[string]string, key string, obj interface{}) error {
	value, err := objectToString(obj)
	if err != nil {
		return err
	}
	properties[key] = value
	return nil
}

func objectToString(obj interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
