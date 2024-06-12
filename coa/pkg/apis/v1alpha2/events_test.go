/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScheduleShouldFire(t *testing.T) {
	activationData := ActivationData{
		Schedule: "2023-10-17T12:00:00-07:00",
	}
	fire, _ := activationData.ShouldFireNow()
	assert.True(t, fire)
}
func TestScheduleShouldFireUTC(t *testing.T) {
	activationData := ActivationData{
		Schedule: "2023-10-20T21:48:00Z",
	}
	fire, _ := activationData.ShouldFireNow()
	assert.True(t, fire)
}

func TestScheduleShouldNotFire(t *testing.T) {
	activationData := ActivationData{
		Schedule: "2073-10-17T12:00:00-07:00",
	}
	fire, _ := activationData.ShouldFireNow()
	assert.False(t, fire) // This should remain false for the next 50 years, so I guess we'll have to update this test in 2073
}

// TODO: This test works only in PST timezone, need to fix it for all time zones
// func TestScheduleLocal(t *testing.T) {
// 	schedule := ScheduleSpec{
// 		Date: "2020-01-01",
// 		Time: "12:00PM",
// 		Zone: "Local",
// 	}
// 	dt, err := schedule.GetTime()
// 	assert.Nil(t, err)
// 	assert.Equal(t, "2020-01-01 12:00:00 -0800 PST", dt.String())
// }
