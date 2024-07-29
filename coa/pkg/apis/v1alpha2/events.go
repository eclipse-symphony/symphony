/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"encoding/json"
	"fmt"
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

func EventShouldRetryWrapper(handler EventHandler, topic string, message Event) bool {
	err := handler(topic, message)
	if err != nil {
		return IsRetriableErr(err)
	}

	return false
}

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
	Schedule             string                            `json:"schedule,omitempty"`
	NeedsReport          bool                              `json:"needsReport,omitempty"`
}

// UnmarshalJSON customizes the JSON unmarshalling for ActivationData
func (s *ActivationData) UnmarshalJSON(data []byte) error {
	type Alias ActivationData
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// validate if Schedule meet RFC 3339
	if s.Schedule != "" {
		if _, err := time.Parse(time.RFC3339, s.Schedule); err != nil {
			return fmt.Errorf("invalid timestamp format: %v", err)
		}
	}
	return nil
}

// MarshalJSON customizes the JSON marshalling for ActivationData
func (s ActivationData) MarshalJSON() ([]byte, error) {
	type Alias ActivationData
	if s.Schedule != "" {
		if _, err := time.Parse(time.RFC3339, s.Schedule); err != nil {
			return nil, fmt.Errorf("invalid timestamp format: %v", err)
		}
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&s),
	})
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

func (s ActivationData) ShouldFireNow() (bool, error) {
	dt, err := time.Parse(time.RFC3339, s.Schedule)
	if err != nil {
		return false, err
	}
	dtNow := time.Now().UTC()
	dtUTC := dt.In(time.UTC)
	return dtUTC.Before(dtNow), nil
}

type InputOutputData struct {
	Inputs  map[string]interface{}            `json:"inputs,omitempty"`
	Outputs map[string]map[string]interface{} `json:"outputs,omitempty"`
}
