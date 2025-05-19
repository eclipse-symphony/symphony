/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package pubsub

import (
	"context"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IPubSubProvider interface {
	Init(config providers.IProviderConfig) error
	Publish(topic string, message v1alpha2.Event) error
	Subscribe(topic string, handler v1alpha2.EventHandler) error
	Cancel() context.CancelFunc
}
