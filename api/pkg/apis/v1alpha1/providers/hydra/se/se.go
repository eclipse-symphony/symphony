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
	Apps    []App    `json:"Apps"`
	Devices []Device `json:"Devices"`
}

type App struct {
	Version  string     `json:"Version,omitempty"`
	Kind     string     `json:"Kind"`
	Metadata Metadata   `json:"Metadata"`
	Spec     AppSpec    `json:"Spec"`
	Status   *AppStatus `json:"Status,omitempty"`
	Deleted  bool       `json:"Deleted,omitempty"`
}

type Metadata struct {
	Name    string            `json:"Name"`
	Labels  map[string]string `json:"Labels"`
	Uuid    string            `json:"Uuid"`
	OwnerId string            `json:"OwnerId"`
}

type AppSpec struct {
	Container ContainerSpec `json:"Container"`
	Affinity  Affinity      `json:"Affinity"`
	HasBlob   bool          `json:"HasBlob"`
	Blob      []interface{} `json:"Blob"`
	DataCase  int           `json:"DataCase"`
}

type ContainerSpec struct {
	Name      string          `json:"Name"`
	Image     string          `json:"Image"`
	Networks  []Network       `json:"Networks"`
	Resources ContainerLimits `json:"Resources"`
}

type Network struct {
	Ipv4      string `json:"Ipv4"`
	Ipv6      string `json:"Ipv6"`
	NetworkId string `json:"NetworkId"`
}

type ContainerLimits struct {
	Limits ResourceLimits `json:"Limits"`
}

type ResourceLimits struct {
	Memory string `json:"Memory"`
	CPUs   string `json:"Cpus"`
}

type Affinity struct {
	PreferredHosts  []string `json:"PreferredHosts"`
	AppAntiAffinity []string `json:"AppAntiAffinity"`
}

type AppStatus struct {
	Status          string `json:"Status"`
	TimeStamp       string `json:"TimeStamp"`
	RunningHost     string `json:"RunningHost"`
	InterlinkStatus string `json:"InterlinkStatus"`
}

type Device struct {
	Kind        string        `json:"Kind"`
	Metadata    Metadata      `json:"Metadata"`
	Spec        DeviceSpec    `json:"Spec"`
	Status      *DeviceStatus `json:"Status,omitempty"`
	TimeSetting interface{}   `json:"TimeSetting"`
	Deleted     bool          `json:"Deleted"`
}

type DeviceSpec struct {
	Addresses              []string           `json:"Addresses"`
	Networks               []DeviceNetwork    `json:"Networks,omitempty"`
	ContainerNetworks      []ContainerNetwork `json:"ContainerNetworks"`
	ReservedAppInterlinkIp string             `json:"ReservedAppInterlinkIp"`
}

type DeviceNetwork struct {
	NicList        []string `json:"NicList"`
	NetName        string   `json:"NetName"`
	NicName        string   `json:"NicName"`
	RedundancyMode string   `json:"RedundancyMode"`
	Ipv4           string   `json:"Ipv4"`
	Gateway        string   `json:"Gateway"`
}

type ContainerNetwork struct {
	Subnet    string `json:"Subnet"`
	Gateway   string `json:"Gateway"`
	NetworkID string `json:"NetworkId"`
	NicName   string `json:"NicName"`
	Type      string `json:"Type"`
}

type DeviceStatus struct {
	Status              string        `json:"Status"`
	TimeStamp           string        `json:"TimeStamp"`
	RunningAppInstances []interface{} `json:"RunningAppInstances"`
	InterlinkStatus     string        `json:"InterlinkStatus"`
	NtpStatus           string        `json:"NtpStatus"`
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
		StateSelector: map[string]string{
			"probed": "true",
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
