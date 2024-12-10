/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
)

type Event struct {
	Metadata map[string]string `json:"metadata"`
	Body     interface{}       `json:"body"`
	Context  context.Context   `json:"-"`
}

func (e *Event) propagteDiagnosticLogContextToEventMetadata() {
	if e == nil || e.Context == nil {
		return
	}

	contexts.PropagteDiagnosticLogContextToMetadata(e.Context, e.Metadata)
}

func (e *Event) parseDiagnosticLogContextFromEventMetadata() {
	if e == nil || e.Metadata == nil {
		return
	}

	diagCtx := contexts.ParseDiagnosticLogContextFromMetadata(e.Metadata)
	e.Context = contexts.PatchDiagnosticLogContextToCurrentContext(diagCtx, e.Context)
}

func (e *Event) clearDiagnosticLogContextFromEventMetadata() {
	if e == nil {
		return
	}
	contexts.ClearDiagnosticLogContextFromMetadata(e.Metadata)
}

func (e *Event) propagateActivityLogContextToEventMetadata() {
	if e == nil || e.Context == nil {
		return
	}
	contexts.PropagateActivityLogContextToMetadata(e.Context, e.Metadata)
}

func (e *Event) parseActivityLogContextFromEventMatadata() {
	if e == nil || e.Metadata == nil {
		return
	}

	actCtx := contexts.ParseActivityLogContextFromMetadata(e.Metadata)
	e.Context = contexts.PatchActivityLogContextToCurrentContext(actCtx, e.Context)
}

func (e *Event) clearActivityLogContextFromEventMetadata() {
	if e == nil {
		return
	}
	contexts.ClearActivityLogContextFromMetadata(e.Metadata)
}

func (e *Event) DeepCopy() *Event {
	if e == nil {
		return nil
	}
	newEvent := *e
	newEvent.Metadata = make(map[string]string)
	for k, v := range e.Metadata {
		newEvent.Metadata[k] = v
	}
	if e.Body != nil {
		jsonBody, err := json.Marshal(e.Body)
		if err == nil {
			json.Unmarshal(jsonBody, &newEvent.Body)
		}
	}
	if e.Context != nil {
		actCtx, ok := e.Context.Value(contexts.ActivityLogContextKey).(*contexts.ActivityLogContext)
		if ok {
			actCtx = actCtx.DeepCopy()
			newEvent.Context = contexts.PatchActivityLogContextToCurrentContext(actCtx, newEvent.Context)
		}
		diagCtx, ok := e.Context.Value(contexts.DiagnosticLogContextKey).(*contexts.DiagnosticLogContext)
		if ok {
			diagCtx = diagCtx.DeepCopy()
			newEvent.Context = contexts.PatchDiagnosticLogContextToCurrentContext(diagCtx, newEvent.Context)
		}
	}
	return &newEvent
}

func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	e1 := e.DeepCopy()
	e1.propagateActivityLogContextToEventMetadata()
	e1.propagteDiagnosticLogContextToEventMetadata()
	return json.Marshal(&struct {
		Alias
	}{Alias: (Alias)(*e1)})
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event
	aux := &struct {
		Alias
	}{Alias: (Alias)(*e)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*e = Event(aux.Alias)
	e.parseActivityLogContextFromEventMatadata()
	e.parseDiagnosticLogContextFromEventMetadata()
	e.clearActivityLogContextFromEventMetadata()
	e.clearDiagnosticLogContextFromEventMetadata()
	return nil
}

func (e Event) DeepEquals(other Event) bool {
	return EventEquals(&e, &other)
}

func EventEquals(e1, e2 *Event) bool {
	if e1 == nil || e2 == nil {
		return e1 == e2
	}

	if len(e1.Metadata) != len(e2.Metadata) {
		return false
	}
	for k, v := range e1.Metadata {
		if e2.Metadata[k] != v {
			return false
		}
	}
	jsonBody1, err := json.Marshal(e1.Body)
	if err != nil {
		return false
	}
	jsonBody2, err := json.Marshal(e2.Body)
	if err != nil {
		return false
	}
	if string(jsonBody1) != string(jsonBody2) {
		return false
	}

	if (e1.Context == nil && e2.Context != nil) || (e1.Context != nil && e2.Context == nil) {
		return false
	}

	if e1.Context != nil && e2.Context != nil {
		diagCtx1, ok1 := e1.Context.Value(contexts.DiagnosticLogContextKey).(*contexts.DiagnosticLogContext)
		diagCtx2, ok2 := e2.Context.Value(contexts.DiagnosticLogContextKey).(*contexts.DiagnosticLogContext)
		if !ok1 || !ok2 {
			return false
		}

		if !contexts.DiagnosticLogContextEquals(diagCtx1, diagCtx2) {
			return false
		}

		actCtx1, ok1 := e1.Context.Value(contexts.ActivityLogContextKey).(*contexts.ActivityLogContext)
		actCtx2, ok2 := e2.Context.Value(contexts.ActivityLogContextKey).(*contexts.ActivityLogContext)
		if !ok1 || !ok2 {
			return false
		}

		if !contexts.ActivityLogContextEquals(actCtx1, actCtx2) {
			return false
		}
	}

	return true
}

func (e Event) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}

type EventHandler struct {
	Handler func(topic string, message Event) error
	// Group is used to distinguish different handlers for the same topic
	// Important: The Group name of an existing handler should NOT be modified.
	Group string
}

func EventShouldRetryWrapper(handler EventHandler, topic string, message Event) bool {
	err := handler.Handler(topic, message)
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
	Proxy                *ProxySpec                        `json:"proxy,omitempty"`
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
	JobId     string          `json:"id"`
	Scope     string          `json:"scope,omitempty"`
	Action    HeartBeatAction `json:"action,omitempty"`
	Time      time.Time       `json:"time,omitempty"`
	JobAction JobAction       `json:"jobaction"`
}

type ProxySpec struct {
	Provider string                 `json:"provider,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
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
