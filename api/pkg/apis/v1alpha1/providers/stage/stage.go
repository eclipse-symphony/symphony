/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package stage

import (
	"context"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
)

type IStageProvider interface {
	// Return values: map[string]interface{} - outputs, bool - should the activation be paused (wait for a remote event), error
	Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error)
}

type IProxyStageProvider interface {
	// Return values: map[string]interface{} - outputs, bool - should the activation be paused (wait for a remote event), error
	Process(ctx context.Context, mgrContext contexts.ManagerContext, activationdata v1alpha2.ActivationData) (map[string]interface{}, bool, error)
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

func GetNamespace(inputs map[string]interface{}) string {
	// objectNamespace is the namespace declared in the stage yaml, #1 priority
	// __namespace is the namespace where the stage is triggered, #2 priority
	objNamespace := ReadInputString(inputs, "objectNamespace")
	if objNamespace != "" {
		return objNamespace
	}
	return ReadInputString(inputs, "__namespace")
}
