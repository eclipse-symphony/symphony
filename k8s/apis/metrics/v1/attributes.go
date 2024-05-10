/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

// Deployment gets common logging attributes for a deployment.
func Deployment(
	validationType string,
	validationResult string,
	resourceType string,
) map[string]any {
	return map[string]any{
		"validationType":   validationType,
		"validationResult": validationResult,
		"resourceType":     resourceType,
	}
}
