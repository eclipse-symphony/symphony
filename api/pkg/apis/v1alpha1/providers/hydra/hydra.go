/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package hydra

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IHydraProvider interface {
	Init(config providers.IProviderConfig) error
	GetArtifact(objType string, artifacts model.ArtifactPack) ([]byte, error)
	SetArtifact(objType string, jsonData []byte) (model.ArtifactPack, error)
}
