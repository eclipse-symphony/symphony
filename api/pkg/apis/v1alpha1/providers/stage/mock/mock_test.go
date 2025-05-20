/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestMockInitFromMap(t *testing.T) {
	provider := MockStageProvider{}
	input := map[string]string{
		"id": "test",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
}

func TestMockProcess(t *testing.T) {
	provider := MockStageProvider{}
	input := map[string]string{
		"id": "test",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	output, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo": "2",
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(3), output["foo"])

	output, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo": "",
	})
	assert.Nil(t, err)
	assert.Equal(t, int(1), output["foo"])
}
