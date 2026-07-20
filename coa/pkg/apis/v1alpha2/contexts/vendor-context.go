/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package contexts

import (
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	logger "github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

// SecurityPolicy defines server-side restrictions on outbound URL access for providers.
// It is populated by the SecurityPolicyVendor and propagated to all managers and providers
// via VendorContext. Providers read it through their ManagerContext to enforce allow-lists
// and exclusive-mode restrictions without exposing these settings in user-editable CRDs.
type SecurityPolicy struct {
	// AllowedIPRanges is a list of CIDR ranges or plain IP addresses that are
	// explicitly permitted, overriding the default SSRF deny list. Plain IPs are
	// treated as host routes (/32 for IPv4, /128 for IPv6).
	AllowedIPRanges []string `json:"allowedIPRanges,omitempty"`
	// AllowListExclusive, when true, requires the target host to resolve to an address
	// within AllowedIPRanges. All other addresses are rejected, including public IPs.
	AllowListExclusive bool `json:"allowListExclusive,omitempty"`
}

type VendorContext struct {
	Logger            logger.Logger
	PubsubProvider    pubsub.IPubSubProvider
	SiteInfo          v1alpha2.SiteInfo
	EvaluationContext *utils.EvaluationContext
	SecurityPolicy    *SecurityPolicy
}

func (v *VendorContext) Init(p pubsub.IPubSubProvider) error {
	v.Logger = logger.NewLogger("coa.runtime")
	v.PubsubProvider = p
	return nil
}

func (v *VendorContext) Publish(feed string, event v1alpha2.Event) error {
	if v.PubsubProvider != nil {
		return v.PubsubProvider.Publish(feed, event)
	}
	return nil
}

func (v *VendorContext) Subscribe(feed string, handler v1alpha2.EventHandler) error {
	if v.PubsubProvider != nil {
		return v.PubsubProvider.Subscribe(feed, handler)
	}
	return nil
}
