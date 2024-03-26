/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"golang.org/x/exp/maps"
)

type ComponentResultSpec struct {
	Status  v1alpha2.State `json:"status"`
	Message string         `json:"message"`
}
type TargetResultSpec struct {
	Status           string                         `json:"status"`
	Message          string                         `json:"message,omitempty"`
	ComponentResults map[string]ComponentResultSpec `json:"components,omitempty"`
}
type SummarySpec struct {
	TargetCount          int                         `json:"targetCount"`
	SuccessCount         int                         `json:"successCount"`
	TargetResults        map[string]TargetResultSpec `json:"targets,omitempty"`
	SummaryMessage       string                      `json:"message,omitempty"`
	Skipped              bool                        `json:"skipped"`
	IsRemoval            bool                        `json:"isRemoval"`
	AllAssignedDeployed  bool                        `json:"allAssignedDeployed"`
	IsDeploymentFinished bool                        `json:"isDeploymentFinished"`
}
type SummaryResult struct {
	Summary        SummarySpec  `json:"summary"`
	Generation     string       `json:"generation"`
	Time           time.Time    `json:"time"`
	State          SummaryState `json:"state"`
	DeploymentHash string       `json:"deploymentHash"`
}

const (
	SummaryStatePending SummaryState = iota // Currently unused
	SummaryStateRunning                     // Should indicate that a reconcile operation is in progress
	SummaryStateDone                        // Should indicate that a reconcile operation has completed either successfully or unsuccessfully
)

type SummaryState int

func (s *SummarySpec) UpdateTargetResult(target string, spec TargetResultSpec) {
	// If the value in map is not set, set the value; otherwise merge it with the
	// existing spec
	if v, ok := s.TargetResults[target]; !ok {
		s.TargetResults[target] = spec
	} else {
		v.Status = spec.Status
		v.Message = spec.Message
		maps.Copy(v.ComponentResults, spec.ComponentResults)
		s.TargetResults[target] = v
	}
}
