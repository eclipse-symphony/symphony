/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"encoding/json"
	"time"
)

type Event struct {
	Metadata map[string]string `json:"metadata"`
	Body     interface{}       `json:"body"`
}

func (e Event) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}

type EventHandler func(topic string, message Event) error

type JobAction string

const (
	JobUpdate JobAction = "UPDATE"
	JobDelete JobAction = "DELETE"
	JobRun    JobAction = "RUN"
)

type JobData struct {
	Id     string      `json:"id"`
	Scope  string      `json:"scope,omitempty"`
	Action JobAction   `json:"action"`
	Body   interface{} `json:"body,omitempty"`
	Data   []byte      `json:"data"`
}
type ActivationData struct {
	Campaign             string                            `json:"campaign"`
	Namespace            string                            `json:"namespace,omitempty"`
	Activation           string                            `json:"activation"`
	ActivationGeneration string                            `json:"activationGeneration"`
	Stage                string                            `json:"stage"`
	Inputs               map[string]interface{}            `json:"inputs,omitempty"`
	Outputs              map[string]map[string]interface{} `json:"outputs,omitempty"`
	Provider             string                            `json:"provider,omitempty"`
	Config               interface{}                       `json:"config,omitempty"`
	TriggeringStage      string                            `json:"triggeringStage,omitempty"`
	Schedule             *ScheduleSpec                     `json:"schedule,omitempty"`
	NeedsReport          bool                              `json:"needsReport,omitempty"`
}

type HeartBeatAction string

const (
	HeartBeatUpdate HeartBeatAction = "UPDATE"
	HeartBeatDelete HeartBeatAction = "DELETE"
)

type HeartBeatData struct {
	JobId  string          `json:"id"`
	Scope  string          `json:"scope,omitempty"`
	Action HeartBeatAction `json:"action"`
	Time   time.Time       `json:"time"`
}
type ScheduleSpec struct {
	Date string `json:"date"`
	Time string `json:"time"`
	Zone string `json:"zone"`
}

func (s ScheduleSpec) ShouldFireNow() (bool, error) {
	dt, err := s.GetTime()
	if err != nil {
		return false, err
	}
	dtNow := time.Now().UTC()
	dtUTC := dt.In(time.UTC)
	return dtUTC.Before(dtNow), nil
}
func (s ScheduleSpec) GetTime() (time.Time, error) {
	dt, err := parseTimeWithZone(s.Time, s.Date, s.Zone)
	if err != nil {
		return time.Time{}, err
	}
	return dt, nil
}

func parseTimeWithZone(timeStr string, dateStr string, zoneStr string) (time.Time, error) {
	dtStr := dateStr + " " + timeStr

	switch zoneStr {
	case "LOCAL":
		zoneStr = ""
	case "PST", "PDT":
		zoneStr = "America/Los_Angeles"
	case "EST", "EDT":
		zoneStr = "America/New_York"
	case "CST", "CDT":
		zoneStr = "America/Chicago"
	case "MST", "MDT":
		zoneStr = "America/Denver"
	}

	loc, err := time.LoadLocation(zoneStr)
	if err != nil {
		return time.Time{}, err
	}

	dt, err := time.ParseInLocation("2006-01-02 3:04:05PM", dtStr, loc)
	if err != nil {
		return time.Time{}, err
	}

	return dt, nil
}

type InputOutputData struct {
	Inputs  map[string]interface{}            `json:"inputs,omitempty"`
	Outputs map[string]map[string]interface{} `json:"outputs,omitempty"`
}
