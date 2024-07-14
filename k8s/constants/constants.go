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
	FullGroupName                = "symphony"
	AzureOperationIdKey          = "management.azure.com/operationId"
	AzureCorrelationIdKey        = "management.azure.com/correlationId"
	AzureResourceIdKey           = "management.azure.com/resourceId"
	AzureSystemDataKey           = "management.azure.com/systemData"
	AzureTenantIdKey             = "management.azure.com/tenantId"
	AzureLocationKey             = "management.azure.com/location"
	AzureCreaedByKey             = "createdBy"
	DefaultScope                 = "default"
	K8S                          = "symphony-k8s"
	OperationStartTimeKeyPostfix = FullGroupName + "/started-at"
	FinalizerPostfix             = FullGroupName + "/finalizer"
	ResourceSeperator            = "-v-"
	ActivityCategory_Activity    = "Activity"
	ActivityOperation_Write      = "Write"
	ActivityOperation_Read       = "Read"
	ActivityOperation_Delete     = "Delete"
)

// Environment variables keys
const (
	SymphonyAPIUrlEnvName = "SYMPHONY_API_URL"
	ConfigName            = "CONFIG_NAME"
	ApiCertEnvName        = "API_SERVING_CA"
)

// Eula Message
var (
	//go:embed eula.txt
	EulaMessage string
)
