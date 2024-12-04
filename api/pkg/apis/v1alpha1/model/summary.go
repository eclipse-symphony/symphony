/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"time"

	"golang.org/x/exp/maps"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
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
	TargetCount         int                         `json:"targetCount"`
	SuccessCount        int                         `json:"successCount"`
	PlannedDeployment   int                         `json:"plannedDeployment"`
	CurrentDeployed     int                         `json:"currentDeployed"`
	TargetResults       map[string]TargetResultSpec `json:"targets,omitempty"`
	SummaryMessage      string                      `json:"message,omitempty"`
	JobID               string                      `json:"jobID,omitempty"`
	Skipped             bool                        `json:"skipped"`
	IsRemoval           bool                        `json:"isRemoval"`
	AllAssignedDeployed bool                        `json:"allAssignedDeployed"`
	Removed             bool                        `json:"removed"`
	PlanState           PlanState                   `json:"planState"`
}
type SolutionManagerDeploymentState struct {
	Spec  DeploymentSpec  `json:"spec,omitempty"`
	State DeploymentState `json:"state,omitempty"`
}

type SummaryResult struct {
	Summary        SummarySpec  `json:"summary"`
	SummaryId      string       `json:"summaryid,omitempty"`
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
	if v, ok := s.TargetResults[target]; !ok {
		s.TargetResults[target] = spec
	} else {
		status := v.Status
		if spec.Status != "OK" {
			status = spec.Status
		}
		message := v.Message
		if spec.Message != "" {
			if message != "" {
				message += "; "
			}
			message += spec.Message
		}
		v.Status = status
		v.Message = message
		maps.Copy(v.ComponentResults, spec.ComponentResults)
		s.TargetResults[target] = v
	}
}

func (summary *SummaryResult) IsDeploymentFinished() bool {
	return summary.State == SummaryStateDone
}
