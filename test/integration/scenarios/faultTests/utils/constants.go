/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

const (
	TEST_TIMEOUT      = "30m"
	LocalPortForward  = "22381"
	InjectFaultEnvKey = "InjectFaultCommand"
	DeleteFaultEnvKey = "DeleteFaultCommand"
	PodEnvKey         = "InjectPodLabel"
)

type FaultTestCase struct {
	TestCase  string
	PodLabel  string
	Fault     string
	FaultType string
}

var (
	TestCases = map[string]string{
		"solutionUpdate":      "./solution/update/verify/...",
		"workflowMaterialize": "./workflow/materialize/verify/...",
	}

	PodLabels = map[string]string{
		"api": "app=symphony-api",
		"k8s": "control-plane=symphony-controller-manager",
	}
	Faults = []FaultTestCase{
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["api"],
			Fault:     "onQueueError",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["api"],
			Fault:     "beforeProviders",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["api"],
			Fault:     "beforeDeploymentError",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["api"],
			Fault:     "afterDeploymentError",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["api"],
			Fault:     "beforeConcludeSummary",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["api"],
			Fault:     "beforeConcludeSummary",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["k8s"],
			Fault:     "beforePollingResult",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["k8s"],
			Fault:     "afterPollingResult",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["k8s"],
			Fault:     "beforeQueueJob",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["solutionUpdate"],
			PodLabel:  PodLabels["k8s"],
			Fault:     "afterQueueJob",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["workflowMaterialize"],
			PodLabel:  PodLabels["api"],
			Fault:     "afterMaterializeOnce",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["workflowMaterialize"],
			PodLabel:  PodLabels["api"],
			Fault:     "afterProvider",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["workflowMaterialize"],
			PodLabel:  PodLabels["api"],
			Fault:     "afterPublishTrigger",
			FaultType: DefaultFaultType,
		},
		{
			TestCase:  TestCases["workflowMaterialize"],
			PodLabel:  PodLabels["api"],
			Fault:     "afterRunTrigger",
			FaultType: DefaultFaultType,
		},
	}

	DefaultFaultType = "100.0%panic"
)
