/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package secret

import (
	"context"

	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type ISecretProvider interface {
	Init(config providers.IProviderConfig) error
	Read(ctx context.Context, name string, field string, localContext interface{}) (string, error)
}

type IExtSecretProvider interface {
	Get(ctx context.Context, name string, field string, localContext interface{}) (string, error)
}
