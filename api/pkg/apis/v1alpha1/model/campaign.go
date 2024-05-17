/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
	"reflect"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type CampaignState struct {
	ObjectMeta ObjectMeta    `json:"metadata,omitempty"`
	Spec       *CampaignSpec `json:"spec,omitempty"`
}

type ActivationState struct {
	ObjectMeta ObjectMeta        `json:"metadata,omitempty"`
	Spec       *ActivationSpec   `json:"spec,omitempty"`
	Status     *ActivationStatus `json:"status,omitempty"`
}
type StageSpec struct {
	Name          string                 `json:"name,omitempty"`
	Contexts      string                 `json:"contexts,omitempty"`
	Provider      string                 `json:"provider,omitempty"`
	Config        interface{}            `json:"config,omitempty"`
	StageSelector string                 `json:"stageSelector,omitempty"`
	Inputs        map[string]interface{} `json:"inputs,omitempty"`
	HandleErrors  bool                   `json:"handleErrors,omitempty"`
	Schedule      *v1alpha2.ScheduleSpec `json:"schedule,omitempty"`
}

func (s StageSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherS, ok := other.(StageSpec)
	if !ok {
		return false, errors.New("parameter is not a StageSpec type")
	}

	if s.Name != otherS.Name {
		return false, nil
	}

	if s.Provider != otherS.Provider {
		return false, nil
	}

	if !reflect.DeepEqual(s.Config, otherS.Config) {
		return false, nil
	}

	if s.StageSelector != otherS.StageSelector {
		return false, nil
	}

	if !reflect.DeepEqual(s.Inputs, otherS.Inputs) {
		return false, nil
	}

	if !reflect.DeepEqual(s.Schedule, otherS.Schedule) {
		return false, nil
	}

	return true, nil
}

type ActivationStatus struct {
	Stage                string                 `json:"stage"`
	NextStage            string                 `json:"nextStage,omitempty"`
	Inputs               map[string]interface{} `json:"inputs,omitempty"`
	Outputs              map[string]interface{} `json:"outputs,omitempty"`
	Status               v1alpha2.State         `json:"status,omitempty"`
	StatusMessage        string                 `json:"statusMessage,omitempty"`
	ErrorMessage         string                 `json:"errorMessage,omitempty"`
	IsActive             bool                   `json:"isActive,omitempty"`
	ActivationGeneration string                 `json:"activationGeneration,omitempty"`
	UpdateTime           string                 `json:"updateTime,omitempty"`
}

type ActivationSpec struct {
	Campaign   string                 `json:"campaign,omitempty"`
	Stage      string                 `json:"stage,omitempty"`
	Inputs     map[string]interface{} `json:"inputs,omitempty"`
	Generation string                 `json:"generation,omitempty"`
}

func (c ActivationSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(ActivationSpec)
	if !ok {
		return false, errors.New("parameter is not a ActivationSpec type")
	}

	if c.Campaign != otherC.Campaign {
		return false, nil
	}

	if c.Stage != otherC.Stage {
		return false, nil
	}

	if !reflect.DeepEqual(c.Inputs, otherC.Inputs) {
		return false, nil
	}

	return true, nil
}
func (c ActivationState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(ActivationState)
	if !ok {
		return false, errors.New("parameter is not a ActivationState type")
	}

	equal, err := c.ObjectMeta.DeepEquals(otherC.ObjectMeta)
	if err != nil || !equal {
		return equal, err
	}

	equal, err = c.Spec.DeepEquals(*otherC.Spec)
	if err != nil || !equal {
		return equal, err
	}
	return true, nil
}

type CampaignSpec struct {
	FirstStage   string               `json:"firstStage,omitempty"`
	Stages       map[string]StageSpec `json:"stages,omitempty"`
	SelfDriving  bool                 `json:"selfDriving,omitempty"`
	Version      string               `json:"version,omitempty"`
	RootResource string               `json:"rootResource,omitempty"`
}

func (c CampaignSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CampaignSpec)
	if !ok {
		return false, errors.New("parameter is not a CampaignSpec type")
	}

	if c.FirstStage != otherC.FirstStage {
		return false, nil
	}

	if c.SelfDriving != otherC.SelfDriving {
		return false, nil
	}

	if len(c.Stages) != len(otherC.Stages) {
		return false, nil
	}

	for i, stage := range c.Stages {
		otherStage := otherC.Stages[i]

		if eq, err := stage.DeepEquals(otherStage); err != nil || !eq {
			return eq, err
		}
	}

	return true, nil
}
func (c CampaignState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CampaignState)
	if !ok {
		return false, errors.New("parameter is not a CampaignState type")
	}

	equal, err := c.ObjectMeta.DeepEquals(otherC.ObjectMeta)
	if err != nil || !equal {
		return equal, err
	}

	equal, err = c.Spec.DeepEquals(*otherC.Spec)
	if err != nil || !equal {
		return equal, err
	}

	return true, nil
}
