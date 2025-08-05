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
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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
		Targets: []model.TargetState{},
	}

	for _, device := range desiredState.Devices {
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
						Role:     "instance",
						Provider: "providers.target.se",
						Config:   map[string]string{},
					},
				},
			},
		}
		ret.Targets = append(ret.Targets, target)
	}

	if i.Config.HASet != "" {
		for _, target := range ret.Targets {
			if target.Spec.Properties == nil {
				target.Spec.Properties = make(map[string]string)
			}
			if strings.Contains(target.ObjectMeta.Name, "spare") {
				if _, ok := target.Spec.Properties["ha-sets"]; !ok {
					target.Spec.Properties["ha-sets"] = i.Config.HASet
					target.Spec.Properties["role"] = "spare"
				}
			} else {
				if _, ok := target.Spec.Properties["ha-set"]; !ok {
					target.Spec.Properties["ha-set"] = i.Config.HASet
					target.Spec.Properties["role"] = "member"
				}
			}
		}
		haTarget := model.TargetState{
			ObjectMeta: model.ObjectMeta{
				Name: i.Config.HASet,
			},
			Spec: &model.TargetSpec{
				Components: []model.ComponentSpec{
					{
						Name: "ha-set",
						Type: "group",
						Properties: map[string]interface{}{
							"targetPropertySelector": map[string]string{
								"ha-set": i.Config.HASet,
								"role":   "member",
							},
							"targetStateSelector": map[string]string{
								"status": "Succeeded",
							},
							"sparePropertySelector": map[string]string{
								"ha-set": i.Config.HASet,
								"role":   "spare",
							},
							"spareStateSelector": map[string]string{
								"status": "Succeeded",
							},
							"minMatchCount": 2,
							"maxMatchCount": 2,
							"lowMatchAction": map[string]interface{}{
								"sparePatch": map[string]interface{}{
									"ha-sets": "~REMOVE",
									"ha-set":  i.Config.HASet,
									"role":    "member",
								},
								"targetPatch": map[string]interface{}{
									"ha-set":  "~REMOVE",
									"ha-sets": "~COPY_ha-set",
									"role":    "spare",
								},
							},
						},
					},
				},
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "group",
								Provider: "providers.target.group",
								Config:   map[string]string{},
							},
						},
					},
				},
			},
		}
		ret.Targets = append(ret.Targets, haTarget)
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
