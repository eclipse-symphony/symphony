/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package config

import (
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IConfigProvider interface {
	Init(config providers.IProviderConfig) error
	Read(object string, field string, localContext interface{}) (interface{}, error)
	ReadObject(object string, localContext interface{}) (map[string]interface{}, error)
	Set(object string, field string, value interface{}) error
	SetObject(object string, value map[string]interface{}) error
	Remove(object string, field string) error
	RemoveObject(object string) error
}

type IExtConfigProvider interface {
	Get(object string, field string, overrides []string, localContext interface{}) (interface{}, error)
	GetObject(object string, overrides []string, localContext interface{}) (map[string]interface{}, error)
}
