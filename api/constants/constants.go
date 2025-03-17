/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package constants

import _ "embed"

// Eula Message
var (
	//go:embed eula.txt
	EulaMessage string
)

const (
	FullGroupName       = "symphony"
	TargetRuntimePrefix = "target-runtime"

	// system annotations, reserved and should not be modified by client.
	AzureCorrelationIdKey        = "management.azure.com/correlationId"
	AzureEdgeLocationKey         = "management.azure.com/customLocation"
	AzureCloudLocationKey        = "management.azure.com/location"
	AzureOperationIdKey          = "management.azure.com/operationId"
	AzureNameIdKey               = "management.azure.com/azureName"
	AzureResourceIdKey           = "management.azure.com/resourceId"
	AzureSystemDataKey           = "management.azure.com/systemData"
	AzureTenantIdKey             = "management.azure.com/tenantId" // Not used
	GuidKey                      = "Guid"
	RunningAzureCorrelationIdKey = "management.azure.com/runningCorrelationId"
	SummaryJobIdKey              = "SummaryJobIdKey"
	OperationStartTimeKeyPostfix = FullGroupName + "/started-at" // instance/target

	ProviderName = "management.azure.com/provider-name"
)

func SystemReservedAnnotations() []string {
	return []string{
		AzureCorrelationIdKey,
		AzureCloudLocationKey,
		AzureEdgeLocationKey,
		AzureOperationIdKey,
		AzureNameIdKey,
		AzureResourceIdKey,
		AzureSystemDataKey,
		AzureTenantIdKey,
		RunningAzureCorrelationIdKey,
		SummaryJobIdKey,
		OperationStartTimeKeyPostfix,
	}
}

func SystemReservedLabels() []string {
	return []string{
		Campaign,
		DisplayName,
		ProviderName,
		ManagerMetaKey,
		ParentName,
		RootResource,
		Solution,
		StagedTarget,
		StatusMessage,
		Target,
	}
}

const (
	DefaultScope = "default"
	SATokenPath  = "/var/run/secrets/tokens/symphony-api-token"
	// These constants need to be in a shared package.
	GroupPrefix        = "symphony"
	ManagerMetaKey     = GroupPrefix + "/managed-by"
	InstanceMetaKey    = GroupPrefix + "/instance"
	ResourceSeperator  = "-v-"
	ReferenceSeparator = ":"
	DisplayName        = "displayName"
	RootResource       = "rootResource"
	ParentName         = "parentName"
	StatusMessage      = "statusMessage"
	Solution           = "solution"
	Target             = "target"
	Campaign           = "campaign"
	StagedTarget       = "staged_target"
)

// Environment variables keys
const (
	SymphonyCertEnvName           = "SYMPHONY_ROOT_CA"
	SATokenPathName               = "SA_TOKEN_PATH"
	ApiCertEnvName                = "API_SERVING_CA"
	UseServiceAccountTokenEnvName = "USE_SERVICE_ACCOUNT_TOKENS"
	SymphonyAPIUrlEnvName         = "SYMPHONY_API_URL"
	API                           = "symphony-api"
	EmitTimeFieldInUserLogs       = "EMIT_TIME_FIELD_IN_USER_LOGS"
)

const (
	Generation string = "generation"
	Status     string = "status"
)
