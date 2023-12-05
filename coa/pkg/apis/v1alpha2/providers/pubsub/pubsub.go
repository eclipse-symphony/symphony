/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package pubsub

import (
	v1alpha2 "github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IPubSubProvider interface {
	Init(config providers.IProviderConfig) error
	Publish(topic string, message v1alpha2.Event) error
	Subscribe(topic string, handler v1alpha2.EventHandler) error
}
