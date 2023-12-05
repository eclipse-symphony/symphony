/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package contexts

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	logger "github.com/azure/symphony/coa/pkg/logger"
)

type ManagerContext struct {
	Logger         logger.Logger
	PubsubProvider pubsub.IPubSubProvider
	SiteInfo       v1alpha2.SiteInfo
	VencorContext  *VendorContext
}

func (v *ManagerContext) Init(c *VendorContext, p pubsub.IPubSubProvider) error {
	if c != nil {
		v.Logger = c.Logger
	} else {
		v.Logger = logger.NewLogger("coa.runtime")
	}
	if c == nil {
		v.PubsubProvider = p
	} else {
		v.PubsubProvider = c.PubsubProvider
	}
	if c != nil {
		v.SiteInfo = c.SiteInfo
	}
	if c != nil {
		v.VencorContext = c
	}
	return nil
}

func (v *ManagerContext) Publish(feed string, event v1alpha2.Event) error {
	if v.PubsubProvider != nil {
		return v.PubsubProvider.Publish(feed, event)
	}
	return nil
}

func (v *ManagerContext) Subscribe(feed string, handler v1alpha2.EventHandler) error {
	if v.PubsubProvider != nil {
		return v.PubsubProvider.Subscribe(feed, handler)
	}
	return nil
}

type IWithManagerContext interface {
	SetContext(ctx *ManagerContext)
}
