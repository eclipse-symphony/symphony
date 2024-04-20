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
	AzureCorrelationId           = "management.azure.com/correlationId"
	DefaultScope                 = "default"
	K8S                          = "symphony-k8s"
	OperationStartTimeKeyPostfix = FullGroupName + "/started-at"
	FinalizerPostfix             = FullGroupName + "/finalizer"
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
