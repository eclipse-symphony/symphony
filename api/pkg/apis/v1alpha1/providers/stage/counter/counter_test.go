/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package counter

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestSmpleCount(t *testing.T) {

	provider := CounterStageProvider{}
	err := provider.Init(provider.Config)
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo": 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), outputs["foo"])
}
func TestAccumulate(t *testing.T) {

	provider := CounterStageProvider{}
	err := provider.Init(provider.Config)
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo": 1,
	})
	outputs2, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo":     1,
		"__state": outputs["__state"],
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(2), outputs2["foo"])
}
func TestSmpleCountWithInitialValue(t *testing.T) {

	provider := CounterStageProvider{}
	err := provider.Init(provider.Config)
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo":      1,
		"foo.init": 5,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(6), outputs["foo"])
}
func TestAccumulateWithInitialValue(t *testing.T) {

	provider := CounterStageProvider{}
	err := provider.Init(provider.Config)
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo":      1,
		"foo.init": 5,
	})
	outputs2, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo":     1,
		"__state": outputs["__state"],
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(7), outputs2["foo"])
}
