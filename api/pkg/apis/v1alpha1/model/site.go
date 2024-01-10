/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type SiteState struct {
	Id       string            `json:"id"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Spec     *SiteSpec         `json:"spec,omitempty"`
	Status   *SiteStatus       `json:"status,omitempty"`
}
type TargetStatus struct {
	State  v1alpha2.State `json:"state,omitempty"`
	Reason string         `json:"reason,omitempty"`
}
type InstanceStatus struct {
	State  v1alpha2.State `json:"state,omitempty"`
	Reason string         `json:"reason,omitempty"`
}

// +kubebuilder:object:generate=true
type SiteStatus struct {
	IsOnline         bool                      `json:"isOnline,omitempty"`
	TargetStatuses   map[string]TargetStatus   `json:"targetStatuses,omitempty"`
	InstanceStatuses map[string]InstanceStatus `json:"instanceStatuses,omitempty"`
	LastReported     string                    `json:"lastReported,omitempty"`
}

// +kubebuilder:object:generate=true
type SiteSpec struct {
	Name       string            `json:"name,omitempty"`
	IsSelf     bool              `json:"isSelf,omitempty"`
	PublicKey  string            `json:"secretHash,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

func (s SiteSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherS, ok := other.(SiteSpec)
	if !ok {
		return false, errors.New("parameter is not a SiteSpec type")
	}

	if s.Name != otherS.Name {
		return false, nil
	}

	if s.PublicKey != otherS.PublicKey {
		return false, nil
	}

	return true, nil
}
