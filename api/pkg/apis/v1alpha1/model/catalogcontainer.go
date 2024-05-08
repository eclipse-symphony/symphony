/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
)

// TODO: all state objects should converge to this paradigm: id, spec and status
type CatalogContainerState struct {
	ObjectMeta ObjectMeta              `json:"metadata,omitempty"`
	Spec       *CatalogContainerSpec   `json:"spec,omitempty"`
	Status     *CatalogContainerStatus `json:"status,omitempty"`
}

type CatalogContainerSpec struct {
}

type CatalogContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c CatalogContainerSpec) DeepEquals(other IDeepEquals) (bool, error) {
	return true, nil
}

func (c CatalogContainerState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CatalogContainerState)
	if !ok {
		return false, errors.New("parameter is not a CatalogContainerState type")
	}

	equal, err := c.ObjectMeta.DeepEquals(otherC.ObjectMeta)
	if err != nil || !equal {
		return equal, err
	}

	equal, err = c.Spec.DeepEquals(*otherC.Spec)
	if err != nil || !equal {
		return equal, err
	}

	return true, nil
}
