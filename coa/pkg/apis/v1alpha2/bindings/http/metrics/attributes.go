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

func Status(
	statusCode int,
	formatedStatusCode string,
) map[string]any {
	return map[string]any{
		"statusCode":         statusCode,
		"formatedStatusCode": formatedStatusCode,
	}
}

func mergeAttrs(attrs ...map[string]any) map[string]any {
	merged := make(map[string]any)
	for _, attr := range attrs {
		for k, v := range attr {
			merged[k] = v
		}
	}
	return merged
}
