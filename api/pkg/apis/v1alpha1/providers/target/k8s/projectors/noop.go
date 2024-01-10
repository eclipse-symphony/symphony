/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package projectors

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

type NoOpProjector struct {
}

func (p *NoOpProjector) ProjectDeployment(scope string, name string, metadata map[string]string, components []model.ComponentSpec, deployment *v1.Deployment) error {
	return nil
}
func (p *NoOpProjector) ProjectService(scope string, name string, metadata map[string]string, service *apiv1.Service) error {
	if name == "error" {
		return v1alpha2.NewCOAError(nil, "throw error project service", v1alpha2.BadConfig)
	}
	return nil
}
