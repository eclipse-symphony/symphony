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

package model

import (
	"errors"
	"reflect"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
)

type CampaignState struct {
	Id   string        `json:"id"`
	Spec *CampaignSpec `json:"spec,omitempty"`
}

type ActivationState struct {
	Id       string            `json:"id"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Spec     *ActivationSpec   `json:"spec,omitempty"`
	Status   *ActivationStatus `json:"status,omitempty"`
}

type StageSpec struct {
	Name          string                 `json:"name,omitempty"`
	Contexts      string                 `json:"contexts,omitempty"`
	Provider      string                 `json:"provider,omitempty"`
	Config        interface{}            `json:"config,omitempty"`
	StageSelector string                 `json:"stageSelector,omitempty"`
	Inputs        map[string]interface{} `json:"inputs,omitempty"`
	HandleErrors  bool                   `json:"handleErrors,omitempty"`
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
	return true, nil
}

type ActivationStatus struct {
	Stage                string                 `json:"stage"`
	NextStage            string                 `json:"nextStage,omitempty"`
	Inputs               map[string]interface{} `json:"inputs,omitempty"`
	Outputs              map[string]interface{} `json:"outputs,omitempty"`
	Status               v1alpha2.State         `json:"status,omitempty"`
	ErrorMessage         string                 `json:"errorMessage,omitempty"`
	IsActive             bool                   `json:"isActive,omitempty"`
	ActivationGeneration string                 `json:"activationGeneration,omitempty"`
}

type ActivationSpec struct {
	Campaign   string                 `json:"campaign,omitempty"`
	Name       string                 `json:"name,omitempty"`
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

	if c.Name != otherC.Name {
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

type CampaignSpec struct {
	Name        string               `json:"name,omitempty"`
	FirstStage  string               `json:"firstStage,omitempty"`
	Stages      map[string]StageSpec `json:"stages,omitempty"`
	SelfDriving bool                 `json:"selfDriving,omitempty"`
}

func (c CampaignSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CampaignSpec)
	if !ok {
		return false, errors.New("parameter is not a CampaignSpec type")
	}

	if c.Name != otherC.Name {
		return false, nil
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
