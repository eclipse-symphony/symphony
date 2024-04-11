/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateString(t *testing.T) {
	stateStringMap := map[State]string{
		OK:                            "OK",
		Accepted:                      "Accepted",
		BadRequest:                    "Bad Request",
		Unauthorized:                  "Unauthorized",
		NotFound:                      "Not Found",
		MethodNotAllowed:              "Method Not Allowed",
		Conflict:                      "Conflict",
		InternalError:                 "Internal Error",
		BadConfig:                     "Bad Config",
		MissingConfig:                 "Missing Config",
		InvalidArgument:               "Invalid Argument",
		APIRedirect:                   "API Redirect",
		FileAccessError:               "File Access Error",
		SerializationError:            "Serialization Error",
		DeleteRequested:               "Delete Requested",
		UpdateFailed:                  "Update Failed",
		DeleteFailed:                  "Delete Failed",
		ValidateFailed:                "Validate Failed",
		Updated:                       "Updated",
		Deleted:                       "Deleted",
		Running:                       "Running",
		Paused:                        "Paused",
		Done:                          "Done",
		Delayed:                       "Delayed",
		Untouched:                     "Untouched",
		NotImplemented:                "Not Implemented",
		InitFailed:                    "Init Failed",
		CreateActionConfigFailed:      "Create Action Config Failed",
		HelmActionFailed:              "Helm Action Failed",
		GetComponentSpecFailed:        "Get Component Spec Failed",
		CreateProjectorFailed:         "Create Projector Failed",
		K8sRemoveServiceFailed:        "Remove K8s Service Failed",
		K8sRemoveDeploymentFailed:     "Remove K8s Deployment Failed",
		K8sDeploymentFailed:           "K8s Deployment Failed",
		ReadYamlFailed:                "Read Yaml Failed",
		ApplyYamlFailed:               "Apply Yaml Failed",
		ReadResourcePropertyFailed:    "Read Resource Property Failed",
		ApplyResourceFailed:           "Apply Resource Failed",
		DeleteYamlFailed:              "Delete Yaml Failed",
		DeleteResourceFailed:          "Delete Resource Failed",
		CheckResourceStatusFailed:     "Check Resource Status Failed",
		ApplyScriptFailed:             "Apply Script Failed",
		RemoveScriptFailed:            "Remove Script Failed",
		YamlResourcePropertyNotFound:  "Yaml or Resource Property Not Found",
		GetHelmPropertyFailed:         "Get Helm Property Failed",
		HelmChartPullFailed:           "Helm Chart Pull Failed",
		HelmChartLoadFailed:           "Helm Chart Load Failed",
		HelmChartApplyFailed:          "Helm Chart Apply Failed",
		HelmChartUninstallFailed:      "Helm Chart Uninstall Failed",
		SolutionGetFailed:             "Solution does not exist",
		TargetCandidatesNotFound:      "Target does not exist",
		TargetListGetFailed:           "Target list does not exist",
		ObjectInstanceCoversionFailed: "Object to Instance conversion failed",
		TimedOut:                      "Timed Out",
		TargetPropertyNotFound:        "Target Property Not Found",
	}
	for state, expected := range stateStringMap {
		assert.Equal(t, expected, state.String())
	}

	assert.Equal(t, "Unknown State: 12001", State(12001).String())
}
