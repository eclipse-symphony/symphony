/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"fmt"
	"sort"
	"strings"
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

func (s *SummarySpec) GenerateStatusMessage() string {
	if s.AllAssignedDeployed {
		return ""
	}

	errorMessage := "Failed to deploy"
	if s.SummaryMessage != "" {
		errorMessage += fmt.Sprintf(": %s", s.SummaryMessage)
	}
	errorMessage += ". "

	// Get target names and sort them
	targetNames := make([]string, 0, len(s.TargetResults))
	for target := range s.TargetResults {
		targetNames = append(targetNames, target)
	}
	sort.Strings(targetNames)

	// Build target errors in sorted order
	targetErrors := make([]string, 0, len(targetNames))
	for _, target := range targetNames {
		result := s.TargetResults[target]
		targetError := fmt.Sprintf("%s: \"%s\"", target, result.Message)

		// Get component names and sort them too for consistency
		componentNames := make([]string, 0, len(result.ComponentResults))
		for component := range result.ComponentResults {
			componentNames = append(componentNames, component)
		}
		sort.Strings(componentNames)

		// Add component results in sorted order
		for _, component := range componentNames {
			componentResult := result.ComponentResults[component]
			targetError += fmt.Sprintf(" (%s.%s: %s)", target, component, componentResult.Message)
		}
		targetErrors = append(targetErrors, targetError)
	}

	return errorMessage + fmt.Sprintf("Detailed status: %s", strings.Join(targetErrors, ", "))
}
