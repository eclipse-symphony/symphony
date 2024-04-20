/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

// Deployment gets common logging attributes for a deployment.
func Deployment(
	opeartion string,
	operationType string,
) map[string]any {
	return map[string]any{
		"operation":     opeartion,
		"operationType": operationType,
	}
}
