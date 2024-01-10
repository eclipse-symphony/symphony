/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package security

import (
	"context"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")
var lock sync.Mutex

type SecurityManager struct {
}

func (s *SecurityManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	return nil
}

func (s *SecurityManager) Apply(ctx context.Context, target model.TargetSpec) error {
	return nil
}
func (s *SecurityManager) Get(ctx context.Context) (model.TargetSpec, error) {
	return model.TargetSpec{}, nil
}
func (s *SecurityManager) Remove(ctx context.Context, target model.TargetSpec) error {
	return nil
}
func (s *SecurityManager) Enabled() bool {
	return false
}
func (s *SecurityManager) Poll() []error {
	return nil
}
func (s *SecurityManager) Reconcil() []error {
	return nil
}
