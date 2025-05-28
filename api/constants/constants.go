/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package constants

import (
	_ "embed"
	"os"
	"strconv"
)

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
	AzureDeleteOperationKey      = "management.azure.com/deleteOperation"
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

var LabelLengthUpperLimit = 64 // Default value if environment variable is not set

func init() {
	// Convert LABEL_LENGTH_UPPER_LIMIT environment variable to int if provided
	if envLimit := os.Getenv("LABEL_LENGTH_UPPER_LIMIT"); envLimit != "" {
		if val, err := strconv.Atoi(envLimit); err == nil {
			LabelLengthUpperLimit = val
		}
	}
}

func SystemReservedAnnotations() []string {
	return []string{
		AzureCorrelationIdKey,
		AzureCloudLocationKey,
		AzureEdgeLocationKey,
		AzureOperationIdKey,
		AzureDeleteOperationKey,
		AzureNameIdKey,
		AzureResourceIdKey,
		AzureSystemDataKey,
		AzureTenantIdKey,
		RunningAzureCorrelationIdKey,
		SummaryJobIdKey,
		OperationStartTimeKeyPostfix,
		GuidKey,
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
	RootResourceUid    = "rootResourceUid"
	ParentName         = "parentName"
	StatusMessage      = "statusMessage"
	Solution           = "solution"
	SolutionUid        = "solutionUid"
	Target             = "target"
	TargetUid          = "targetUid"
	Campaign           = "campaign"
	CampaignUid        = "campaignUid"
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
