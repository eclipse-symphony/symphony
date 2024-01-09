/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
)

type BackgroundJobVendor struct {
	vendors.Vendor
	// Add a new manager if you want to add another background job
	ActivationsCleanerManager *activations.ActivationsCleanupManager
}

func (s *BackgroundJobVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  s.Vendor.Version,
		Name:     "BackgroundJob",
		Producer: "Microsoft",
	}
}

func (o *BackgroundJobVendor) GetEndpoints() []v1alpha2.Endpoint {
	return []v1alpha2.Endpoint{}
}

func (s *BackgroundJobVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := s.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range s.Managers {
		if c, ok := m.(*activations.ActivationsCleanupManager); ok {
			s.ActivationsCleanerManager = c
		}
		// Load a new manager if you want to add another background job
	}
	if s.ActivationsCleanerManager != nil {
		log.Info("ActivationsCleanupManager is enabled")
	} else {
		log.Info("ActivationsCleanupManager is disabled")
	}
	return nil
}
