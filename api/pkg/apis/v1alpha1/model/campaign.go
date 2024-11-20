/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"time"

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
	Proxy         *v1alpha2.ProxySpec    `json:"proxy,omitempty"`
	Schedule      string                 `json:"schedule,omitempty"`
}

// UnmarshalJSON customizes the JSON unmarshalling for StageSpec
func (s *StageSpec) UnmarshalJSON(data []byte) error {
	type Alias StageSpec
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// validate if Schedule meet RFC 3339
	if s.Schedule != "" {
		if _, err := time.Parse(time.RFC3339, s.Schedule); err != nil {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid timestamp format: %v", err), v1alpha2.BadConfig)
		}
	}
	return nil
}

// MarshalJSON customizes the JSON marshalling for StageSpec
func (s StageSpec) MarshalJSON() ([]byte, error) {
	type Alias StageSpec
	if s.Schedule != "" {
		if _, err := time.Parse(time.RFC3339, s.Schedule); err != nil {
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid timestamp format: %v", err), v1alpha2.BadConfig)
		}
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&s),
	})
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
	if s.Proxy == nil && otherS.Proxy != nil {
		return false, nil
	}
	if s.Proxy != nil && otherS.Proxy == nil {
		return false, nil
	}
	if s.Proxy != nil && otherS.Proxy != nil {
		if !reflect.DeepEqual(s.Proxy.Provider, otherS.Proxy.Provider) {
			return false, nil
		}
	}
	return true, nil
}

type ActivationStatus struct {
	ActivationGeneration string         `json:"activationGeneration,omitempty"`
	UpdateTime           string         `json:"updateTime,omitempty"`
	Status               v1alpha2.State `json:"status,omitempty"`
	StatusMessage        string         `json:"statusMessage,omitempty"`
	StageHistory         []StageStatus  `json:"stageHistory,omitempty"`
}
type StageStatus struct {
	Stage         string                 `json:"stage,omitempty"`
	NextStage     string                 `json:"nextStage,omitempty"`
	Inputs        map[string]interface{} `json:"inputs,omitempty"`
	Outputs       map[string]interface{} `json:"outputs,omitempty"`
	Status        v1alpha2.State         `json:"status,omitempty"`
	IsActive      bool                   `json:"isActive,omitempty"`
	StatusMessage string                 `json:"statusMessage,omitempty"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
}

type ActivationSpec struct {
	Campaign string                 `json:"campaign,omitempty"`
	Stage    string                 `json:"stage,omitempty"`
	Inputs   map[string]interface{} `json:"inputs,omitempty"`
}

func (c ActivationSpec) Hash() (string, error) {
	hasher := sha256.New()

	// Write the simple fields to the hasher
	writeStringHash(hasher, c.Campaign)
	writeStringHash(hasher, c.Stage)

	// Sort the map keys to ensure consistent order
	keys := make([]string, 0, len(c.Inputs))
	for key := range c.Inputs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Write the sorted map entries to the hasher
	for _, key := range keys {
		valueBytes, err := json.Marshal(c.Inputs[key])
		if err != nil {
			return "", err
		}
		writeStringHash(hasher, key)
		hasher.Write(valueBytes)
	}

	// Get the final hash result
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func writeStringHash(writer io.Writer, value string) {
	fmt.Fprintf(writer, "<%s>", value)
}

func (c ActivationSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(ActivationSpec)
	if !ok {
		return false, errors.New("parameter is not a ActivationSpec type")
	}

	if c.Campaign != otherC.Campaign {
		return false, errors.New("campaign doesn't match")
	}

	if c.Stage != otherC.Stage {
		return false, errors.New("stage doesn't match")
	}

	if !reflect.DeepEqual(c.Inputs, otherC.Inputs) {
		return false, errors.New("inputs doesn't match")
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
