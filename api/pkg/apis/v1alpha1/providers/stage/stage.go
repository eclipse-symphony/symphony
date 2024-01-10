/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package stage

import (
	"context"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
)

type IStageProvider interface {
	// Return values: map[string]interface{} - outputs, bool - should the activation be paused (wait for a remote event), error
	Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error)
}

func ReadInputString(inputs map[string]interface{}, key string) string {
	if inputs == nil {
		return ""
	}
	if val, ok := inputs[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
