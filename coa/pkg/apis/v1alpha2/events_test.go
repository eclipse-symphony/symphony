/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduleErrorTimeZone(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2020-01-01",
		Time: "12:00:00PM",
		Zone: "XXX",
	}
	_, err := schedule.GetTime()
	assert.NotNil(t, err)
}

func TestScheduleTimeZone(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2020-01-01",
		Time: "12:00:00PM",
		Zone: "PST",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2020-01-01 12:00:00 -0800 PST", dt.String())
}

func TestScheduleTimeZoneDaylight(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2020-10-31",
		Time: "12:00:00PM",
		Zone: "PDT",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2020-10-31 12:00:00 -0700 PDT", dt.String()) //This is parsed as PDT because it is daylight savings time
}

func TestScheduleTimeZoneDaylight2(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2020-10-31",
		Time: "12:00:00PM",
		Zone: "PDT",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2020-10-31 12:00:00 -0700 PDT", dt.String())
}

func TestScheduleTimeZoneDaylight3(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2020-10-20",
		Time: "11:53:03AM",
		Zone: "PDT",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2020-10-20 11:53:03 -0700 PDT", dt.String())
}

func TestScheduleIANATimeZone(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2020-01-02",
		Time: "12:00:00PM",
		Zone: "America/Los_Angeles",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2020-01-02 12:00:00 -0800 PST", dt.String())
}

func TestScheduleEmpty(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2020-01-01",
		Time: "12:00:00PM",
		Zone: "", //this is equivalent to UTC
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2020-01-01 12:00:00 +0000 UTC", dt.String())
}

func TestScheduleUTC(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2020-01-01",
		Time: "12:00:00PM",
		Zone: "UTC",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2020-01-01 12:00:00 +0000 UTC", dt.String())
}

func TestScheduleShouldFire(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2023-10-17",
		Time: "12:00:00PM",
		Zone: "PDT",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2023-10-17 12:00:00 -0700 PDT", dt.String())
	fire, _ := schedule.ShouldFireNow()
	assert.True(t, fire)
}
func TestScheduleShouldFireUTC(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2023-10-20",
		Time: "9:48:00PM",
		Zone: "UTC",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2023-10-20 21:48:00 +0000 UTC", dt.String())
	fire, _ := schedule.ShouldFireNow()
	assert.True(t, fire)
}
func TestGetTime(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2023-10-20",
		Time: "3:53:00PM",
		Zone: "PDT",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2023-10-20 15:53:00 -0700 PDT", dt.String())
	assert.Equal(t, "2023-10-20 22:53:00 +0000 UTC", dt.In(time.UTC).String())
}
func TestScheduleShouldFirePDT(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2023-10-20",
		Time: "3:26:00PM",
		Zone: "PDT",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2023-10-20 15:26:00 -0700 PDT", dt.String())
	fire, _ := schedule.ShouldFireNow()
	assert.True(t, fire)
}

func TestScheduleShouldNotFire(t *testing.T) {
	schedule := ScheduleSpec{
		Date: "2073-10-17",
		Time: "12:00:00PM",
		Zone: "PDT",
	}
	dt, err := schedule.GetTime()
	assert.Nil(t, err)
	assert.Equal(t, "2073-10-17 12:00:00 -0700 PDT", dt.String())
	fire, _ := schedule.ShouldFireNow()
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
