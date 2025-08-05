/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package hydra

import (
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var lock sync.Mutex

var log = logger.NewLogger("coa.runtime")

type HydraManager struct {
	managers.Manager
}

func (s *HydraManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	return nil
}

func (s *HydraManager) GetArtifacts(system string, artifacts model.ArtifactPack) ([]interface{}, error) {
	return nil, nil
}

func (s *HydraManager) SetArtifacts(system string, artifacts []interface{}) (model.ArtifactPack, error) {
	return model.ArtifactPack{}, nil
}
