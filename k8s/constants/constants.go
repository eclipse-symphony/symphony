/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package constants

import (
	_ "embed"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
)

const (
	FullGroupName            = api_constants.FullGroupName
	AzureLocationKey         = "management.azure.com/location" // diagnostic
	AzureCreatedByKey        = "createdBy"                     // systemDataMap
	DefaultScope             = "default"                       // Namespace
	K8S                      = "symphony-k8s"                  // observable
	FinalizerPostfix         = FullGroupName + "/finalizer"    // finalizerName, instance/target
	ResourceSeperator        = "-v-"
	ReferenceSeparator       = ":"
	ActivityOperation_Write  = "Write"
	ActivityOperation_Read   = "Read"
	ActivityOperation_Delete = "Delete"

	SolutionContainerOperationNamePrefix = "solutioncontainers.solution." + FullGroupName
	SolutionOperationNamePrefix          = "solutions.solution." + FullGroupName
	TargetOperationNamePrefix            = "targets.fabric." + FullGroupName
	InstanceOperationNamePrefix          = "instances.solution." + FullGroupName
	InstanceHistoryOperationNamePrefix   = "instancehistories.solution." + FullGroupName
	ActivationOperationNamePrefix        = "activations.workflow." + FullGroupName
	CatalogOperationNamePrefix           = "catalogs.federation." + FullGroupName
	CatalogEvalOperationNamePrefix       = "catalogevalexpression.federation." + FullGroupName
	CampaignOperationNamePrefix          = "campaigns.workflow." + FullGroupName
	CampaignContainerOperationNamePrefix = "campaigncontainers.workflow." + FullGroupName
	DiagnosticsOperationNamePrefix       = "diagnostics.monitor." + FullGroupName
)

// system annotations, reserved and should not be modified by client.
const (
	AzureCorrelationIdKey        = api_constants.AzureCorrelationIdKey
	AzureEdgeLocationKey         = api_constants.AzureEdgeLocationKey
	AzureOperationIdKey          = api_constants.AzureOperationIdKey
	AzureResourceIdKey           = api_constants.AzureResourceIdKey
	AzureSystemDataKey           = api_constants.AzureSystemDataKey
	AzureTenantIdKey             = api_constants.AzureTenantIdKey
	RunningAzureCorrelationIdKey = api_constants.RunningAzureCorrelationIdKey
	SummaryJobIdKey              = api_constants.SummaryJobIdKey
	OperationStartTimeKeyPostfix = api_constants.OperationStartTimeKeyPostfix // instance/target
)

// Environment variables keys
const (
	SymphonyAPIUrlEnvName = "SYMPHONY_API_URL"
	ConfigName            = "CONFIG_NAME"
	ApiCertEnvName        = "API_SERVING_CA"
	DeploymentFinalizer   = "DEPLOYMENT_FINALIZER"
)

// Eula Message
var (
	//go:embed eula.txt
	EulaMessage string
)
