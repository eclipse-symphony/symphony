/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
)

type InstanceContainerState struct {
	ObjectMeta ObjectMeta               `json:"metadata,omitempty"`
	Spec       *InstanceContainerSpec   `json:"spec,omitempty"`
	Status     *InstanceContainerStatus `json:"status,omitempty"`
}

type InstanceContainerSpec struct {
}

type InstanceContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c InstanceContainerSpec) DeepEquals(other IDeepEquals) (bool, error) {
	return true, nil
}

func (c InstanceContainerState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(InstanceContainerState)
	if !ok {
		return false, errors.New("parameter is not a InstanceContainerState type")
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
