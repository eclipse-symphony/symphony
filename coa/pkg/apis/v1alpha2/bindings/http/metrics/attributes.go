/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

// Deployment gets common logging attributes for a deployment.
func Deployment(
	operation string,
	operationType string,
) map[string]any {
	return map[string]any{
		"operation":     operation,
		"operationType": operationType,
	}
}

func Error(
	errorCode string,
) map[string]any {
	return map[string]any{
		"errorCode": errorCode,
	}
}
