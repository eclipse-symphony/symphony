/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package projectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
)

func TestProjectDeployment(t *testing.T) {
	projector := &NoOpProjector{}
	deployment := appsv1.Deployment{}
	err := projector.ProjectDeployment("default", "name", nil, nil, &deployment)
	assert.Nil(t, err)
}

func ProjectService(t *testing.T) {
	projector := &NoOpProjector{}
	err := projector.ProjectService("default", "name", nil, nil)
	assert.Nil(t, err)
}

func ProjectServiceError(t *testing.T) {
	projector := &NoOpProjector{}
	err := projector.ProjectService("default", "error", nil, nil)
	assert.NotNil(t, err)
}
