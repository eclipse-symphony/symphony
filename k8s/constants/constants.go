//go:build !azure
// +build !azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package constants

import _ "embed"

const (
	FullGroupName            = "symphony"
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
	ActivationOperationNamePrefix        = "activations.workflow." + FullGroupName
	CatalogOperationNamePrefix           = "catalogs.federation." + FullGroupName
	CatalogEvalOperationNamePrefix       = "catalogevalexpression.federation." + FullGroupName
	CampaignOperationNamePrefix          = "campaigns.workflow." + FullGroupName
	CampaignContainerOperationNamePrefix = "campaigncontainers.workflow." + FullGroupName
	DiagnosticsOperationNamePrefix       = "diagnostics.monitor." + FullGroupName
)

// system annotations, reserved and should not be modified by client.
const (
	AzureCorrelationIdKey        = "management.azure.com/correlationId"
	AzureEdgeLocationKey         = "management.azure.com/customLocation"
	AzureOperationIdKey          = "management.azure.com/operationId"
	AzureResourceIdKey           = "management.azure.com/resourceId"
	AzureSystemDataKey           = "management.azure.com/systemData"
	AzureTenantIdKey             = "management.azure.com/tenantId"
	RunningAzureCorrelationIdKey = "management.azure.com/runningCorrelationId"
	SummaryJobIdKey              = "SummaryJobIdKey"
	OperationStartTimeKeyPostfix = FullGroupName + "/started-at" // instance/target
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
