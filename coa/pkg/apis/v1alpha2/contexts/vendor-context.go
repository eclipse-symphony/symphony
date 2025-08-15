/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package contexts

import (
	"fmt"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/bindings/mqtt"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	logger "github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

type VendorContext struct {
	Logger            logger.Logger
	PubsubProvider    pubsub.IPubSubProvider
	SiteInfo          v1alpha2.SiteInfo
	EvaluationContext *utils.EvaluationContext
	mqttBinding       *mqtt.MQTTBinding
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

func (vc *VendorContext) SetMQTTBinding(binding *mqtt.MQTTBinding) {
	fmt.Printf("Setting MQTTBinding: %v", binding)
	vc.mqttBinding = binding
}
func (vc *VendorContext) GetMQTTBinding() *mqtt.MQTTBinding {
	fmt.Printf("GetMQTTBinding called, returning: %v\n", vc.mqttBinding)
	return vc.mqttBinding
}
