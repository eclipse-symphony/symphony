/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
)

type TargetContainerState struct {
	ObjectMeta ObjectMeta             `json:"metadata,omitempty"`
	Spec       *TargetContainerSpec   `json:"spec,omitempty"`
	Status     *TargetContainerStatus `json:"status,omitempty"`
}

type TargetContainerSpec struct {
}

type TargetContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c TargetContainerSpec) DeepEquals(other IDeepEquals) (bool, error) {
	return true, nil
}

func (c TargetContainerState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(TargetContainerState)
	if !ok {
		return false, errors.New("parameter is not a TargetContainerState type")
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
