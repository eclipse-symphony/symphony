/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package config

import (
	"context"

	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IConfigProvider interface {
	Init(config providers.IProviderConfig) error
	Read(ctx context.Context, object string, field string, localContext interface{}) (interface{}, error)
	ReadObject(ctx context.Context, object string, localContext interface{}) (map[string]interface{}, error)
	Set(ctx context.Context, object string, field string, value interface{}) error
	SetObject(ctx context.Context, object string, value map[string]interface{}) error
	Remove(ctx context.Context, object string, field string) error
	RemoveObject(ctx context.Context, object string) error
}

type IExtConfigProvider interface {
	Get(object string, field string, overrides []string, localContext interface{}) (interface{}, error)
	GetObject(object string, overrides []string, localContext interface{}) (map[string]interface{}, error)
}
