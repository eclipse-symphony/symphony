/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/status"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
)

type StatusVendor struct {
	vendors.Vendor
}

func (v *StatusVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, ps map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	if config.LoopInterval <= 0 {
		return v1alpha2.NewCOAError(nil, "status vendor requires a positive loopInterval in seconds", v1alpha2.BadConfig)
	}
	if err := v.Vendor.Init(config, factories, ps, pubsubProvider); err != nil {
		return err
	}
	for _, manager := range v.Managers {
		if _, ok := manager.(*status.StatusManager); ok {
			return nil
		}
	}
	return v1alpha2.NewCOAError(nil, "status manager is not supplied", v1alpha2.MissingConfig)
}

func (v *StatusVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  v.Version,
		Name:     "Status",
		Producer: "Microsoft",
	}
}

func (v *StatusVendor) GetEndpoints() []v1alpha2.Endpoint {
	return nil
}
