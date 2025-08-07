/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package target

import (
	"context"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type ITargetProvider interface {
	Init(config providers.IProviderConfig) error
	// get validation rules
	GetValidationRule(ctx context.Context) model.ValidationRule
	// get current component states from a target. The desired state is passed in as a reference
	Get(ctx context.Context, reference model.TargetProviderGetReference) ([]model.ComponentSpec, error)
	// apply components to a target
	Apply(ctx context.Context, reference model.TargetProviderApplyReference) (map[string]model.ComponentResultSpec, error)
}
