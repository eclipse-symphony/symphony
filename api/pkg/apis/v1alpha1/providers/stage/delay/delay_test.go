/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package delay

import (
	"context"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestDelayInitFromVendorMap(t *testing.T) {
	provider := DelayStageProvider{}
	input := map[string]string{
		"id": "test",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	assert.Equal(t, "test", provider.Config.ID)
}
func TestDelayProcess(t *testing.T) {
	provider := DelayStageProvider{}
	err := provider.InitWithMap(map[string]string{})
	assert.Nil(t, err)
	dt1 := time.Now()
	outputs, _, _ := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"delay": 1,
	})
	dt2 := time.Now()
	assert.Equal(t, v1alpha2.OK, outputs[v1alpha2.StatusOutput])
	assert.GreaterOrEqual(t, dt2.Sub(dt1).Seconds(), 1.0)

	dt1 = time.Now()
	outputs, _, _ = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"delay": int32(1),
	})
	dt2 = time.Now()
	assert.Equal(t, v1alpha2.OK, outputs[v1alpha2.StatusOutput])
	assert.GreaterOrEqual(t, dt2.Sub(dt1).Seconds(), 1.0)

	dt1 = time.Now()
	outputs, _, _ = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"delay": int64(1),
	})
	dt2 = time.Now()
	assert.Equal(t, v1alpha2.OK, outputs[v1alpha2.StatusOutput])
	assert.GreaterOrEqual(t, dt2.Sub(dt1).Seconds(), 1.0)

	dt1 = time.Now()
	outputs, _, _ = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"delay": "1s",
	})
	dt2 = time.Now()
	assert.Equal(t, v1alpha2.OK, outputs[v1alpha2.StatusOutput])
	assert.GreaterOrEqual(t, dt2.Sub(dt1).Seconds(), 1.0)

	dt1 = time.Now()
	outputs, _, _ = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"delay": "1",
	})
	dt2 = time.Now()
	assert.Equal(t, v1alpha2.OK, outputs[v1alpha2.StatusOutput])
	assert.GreaterOrEqual(t, dt2.Sub(dt1).Seconds(), 1.0)
}

func TestDelayProcessFailedCase(t *testing.T) {
	provider := DelayStageProvider{}
	err := provider.InitWithMap(map[string]string{})
	assert.Nil(t, err)

	outputs, _, _ := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"delay": "abc",
	})
	assert.Equal(t, v1alpha2.InternalError, outputs[v1alpha2.StatusOutput])
}
