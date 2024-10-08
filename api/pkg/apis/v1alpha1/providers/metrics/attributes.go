/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

// Target gets common logging attributes for a target.
func Target(
	providerType string,
	functionName string,
	operation string,
	operationType string,
	errorCode string,
) map[string]any {
	return map[string]any{
		"providerType":  providerType,
		"functionName":  functionName,
		"operation":     operation,
		"operationType": operationType,
		"errorCode":     errorCode,
	}
}

func TargetWithoutErrorCode(
	providerType string,
	functionName string,
	operation string,
	operationType string,
) map[string]any {
	return map[string]any{
		"providerType":  providerType,
		"functionName":  functionName,
		"operation":     operation,
		"operationType": operationType,
	}
}
