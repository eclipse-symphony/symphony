/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package bindings

import (
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type IBinding interface {
	v1alpha2.Terminable
}
