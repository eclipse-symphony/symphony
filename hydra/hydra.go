/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package hydra

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

type AppPackageDescription struct {
	Type    string
	Path    string
	Version string
}

type SolutionReader interface {
	Parse(appPackage AppPackageDescription) (model.SolutionState, error)
}
