/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package reference

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IReferenceProvider interface {
	ID() string
	TargetID() string
	Init(config providers.IProviderConfig) error
	Get(id string, namespace string, group string, kind string, version string, ref string) (interface{}, error)
	List(labelSelector string, fieldSelector string, namespace string, group string, kind string, version string, ref string) (interface{}, error)
	SetContext(context *contexts.ManagerContext)
	ReferenceType() string
	Reconfigure(config providers.IProviderConfig) error
}
