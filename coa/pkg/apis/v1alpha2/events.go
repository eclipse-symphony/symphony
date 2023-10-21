/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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

type JobData struct {
	Id     string      `json:"id"`
	Action string      `json:"action"`
	Body   interface{} `json:"body,omitempty"`
}
type ActivationData struct {
	Campaign             string                            `json:"campaign"`
	Activation           string                            `json:"activation"`
	ActivationGeneration string                            `json:"activationGeneration"`
	Stage                string                            `json:"stage"`
	Inputs               map[string]interface{}            `json:"inputs,omitempty"`
	Outputs              map[string]map[string]interface{} `json:"outputs,omitempty"`
	Provider             string                            `json:"provider,omitempty"`
	Config               interface{}                       `json:"config,omitempty"`
	TriggeringStage      string                            `json:"triggeringStage,omitempty"`
	Schedule             *ScheduleSpec                     `json:"schedule,omitempty"`
}
type HeartBeatData struct {
	JobId  string    `json:"id"`
	Action string    `json:"action"`
	Time   time.Time `json:"time"`
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
