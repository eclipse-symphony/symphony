//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

// Test config
const (
	TEST_TIMEOUT = "30m"
)

type FaultTestCase struct {
	testCase  string
	podLabel  string
	fault     string
	faultType string
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
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["api"],
			fault:     "onQueueError",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["api"],
			fault:     "beforeProviders",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["api"],
			fault:     "beforeDeploymentError",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["api"],
			fault:     "afterDeploymentError",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["api"],
			fault:     "beforeConcludeSummary",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["api"],
			fault:     "beforeConcludeSummary",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["k8s"],
			fault:     "beforePollingResult",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["k8s"],
			fault:     "afterPollingResult",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["k8s"],
			fault:     "beforeQueueJob",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["solutionUpdate"],
			podLabel:  PodLabels["k8s"],
			fault:     "afterQueueJob",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["workflowMaterialize"],
			podLabel:  PodLabels["api"],
			fault:     "afterMaterializeOnce",
			faultType: DefaultFaultType,
		},
		{
			testCase:  TestCases["workflowMaterialize"],
			podLabel:  PodLabels["api"],
			fault:     "afterProvider",
			faultType: DefaultFaultType,
		},
		// afterPublishTrigger fault test cannot pass now because of dedup issue in activation
		// {
		// 	testCase:  TestCases["workflowMaterialize"],
		// 	podLabel:  PodLabels["api"],
		// 	fault:     "afterPublishTrigger",
		// 	faultType: DefaultFaultType,
		// },
		{
			testCase:  TestCases["workflowMaterialize"],
			podLabel:  PodLabels["api"],
			fault:     "afterRunTrigger",
			faultType: DefaultFaultType,
		},
	}

	LocalPortForward = "22381"

	DefaultFaultType = "100.0%panic"
)
