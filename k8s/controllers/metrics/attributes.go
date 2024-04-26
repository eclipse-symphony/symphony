/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

// Deployment gets common logging attributes for a deployment.
func Deployment(
	reconciliationType ReconciliationType,
	reconciliationResult ReconciliationResult,
	resourceType ResourceType,
	operationStatus OperationStatus,
	chartVersion string,
) map[string]any {
	return map[string]any{
		"reconciliationType":   reconciliationType,
		"reconciliationResult": reconciliationResult,
		"resourceType":         resourceType,
		"operationStatus":      operationStatus,
		"chartVersion":         chartVersion,
	}
}
