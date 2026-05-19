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
type CatalogState struct {
	ObjectMeta ObjectMeta              `json:"metadata,omitempty"`
	Spec       *CatalogSpec   `json:"spec,omitempty"`
	Status     *CatalogStatus `json:"status,omitempty"`
}

type CatalogSpec struct {
}

type CatalogStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c CatalogSpec) DeepEquals(other IDeepEquals) (bool, error) {
	return true, nil
}

func (c CatalogState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CatalogState)
	if !ok {
		return false, errors.New("parameter is not a CatalogState type")
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
