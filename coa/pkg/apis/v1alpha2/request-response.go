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

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
)

type ContextKey string

const (
	COAFastHTTPContextKey ContextKey = "coa-fasthttp-context"
)

type COARequest struct {
	Context     context.Context   `json:"-"`
	Method      string            `json:"method"`
	Route       string            `json:"route"`
	ContentType string            `json:"contentType"`
	Body        []byte            `json:"body"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

func (in *COARequest) propagteDiagnosticLogContextToCOARequestMetadata() {
	if in == nil || in.Context == nil {
		return
	}
	contexts.PropagteDiagnosticLogContextToMetadata(in.Context, in.Metadata)
}

func (in *COARequest) parseDiagnosticLogContextFromCOARequestMetadata() {
	if in == nil || in.Metadata == nil {
		return
	}

	diagCtx := contexts.ParseDiagnosticLogContextFromMetadata(in.Metadata)
	in.Context = contexts.PatchDiagnosticLogContextToCurrentContext(diagCtx, in.Context)
}

func (in *COARequest) clearDiagnosticLogContextFromCOARequestMetadata() {
	if in == nil {
		return
	}
	contexts.ClearDiagnosticLogContextFromMetadata(in.Metadata)
}

func (in *COARequest) propagateActivityLogContextToCOARequestMetadata() {
	if in == nil || in.Context == nil {
		return
	}
	contexts.PropagateActivityLogContextToMetadata(in.Context, in.Metadata)
}

func (in *COARequest) parseActivityLogContextFromCOARequestMatadata() {
	if in == nil || in.Metadata == nil {
		return
	}
	actCtx := contexts.ParseActivityLogContextFromMetadata(in.Metadata)
	in.Context = contexts.PatchActivityLogContextToCurrentContext(actCtx, in.Context)
}

func (in *COARequest) clearActivityLogContextFromCOARequestMetadata() {
	if in == nil {
		return
	}
	contexts.ClearActivityLogContextFromMetadata(in.Metadata)
}

func (in *COARequest) DeepCopyInto(out *COARequest) {
	if in == nil {
		out = nil
		return
	}

	if out == nil {
		out = new(COARequest)
	}

	*out = *in
	out.Context = in.Context //TODO: Is this okay?
	out.Method = in.Method
	out.Route = in.Route
	out.ContentType = in.ContentType
	if in.Body != nil {
		out.Body = make([]byte, len(in.Body))
		copy(out.Body, in.Body)
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

func (in *COARequest) DeepCopy() *COARequest {
	if in == nil {
		return nil
	}
	out := new(COARequest)
	in.DeepCopyInto(out)
	return out
}

func (in COARequest) MarshalJSON() ([]byte, error) {
	type Alias COARequest
	in1 := new(COARequest)
	in.DeepCopyInto(in1)
	in1.propagateActivityLogContextToCOARequestMetadata()
	in1.propagteDiagnosticLogContextToCOARequestMetadata()
	return json.Marshal(&struct {
		Alias
	}{Alias: (Alias)(*in1)})
}

func (in *COARequest) UnmarshalJSON(data []byte) error {
	type Alias COARequest
	aux := &struct {
		Alias
	}{Alias: (Alias)(*in)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*in = COARequest(aux.Alias)
	in.parseActivityLogContextFromCOARequestMatadata()
	in.parseDiagnosticLogContextFromCOARequestMetadata()
	in.clearActivityLogContextFromCOARequestMetadata()
	in.clearDiagnosticLogContextFromCOARequestMetadata()
	return nil
}

func (in COARequest) DeepEquals(other COARequest) bool {
	return COARequestEquals(&in, &other)
}

func COARequestEquals(a, b *COARequest) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.Method != b.Method {
		return false
	}

	if a.Route != b.Route {
		return false
	}

	if a.ContentType != b.ContentType {
		return false
	}

	if string(a.Body) != string(b.Body) {
		return false
	}

	if len(a.Metadata) != len(b.Metadata) {
		return false
	}

	for k, v := range a.Metadata {
		if b.Metadata[k] != v {
			return false
		}
	}

	if len(a.Parameters) != len(b.Parameters) {
		return false
	}

	for k, v := range a.Parameters {
		if b.Parameters[k] != v {
			return false
		}
	}

	if (a.Context == nil && b.Context != nil) || (a.Context != nil && b.Context == nil) {
		return false
	}

	if a.Context != nil && b.Context != nil {
		diagCtx1, ok1 := a.Context.Value(contexts.DiagnosticLogContextKey).(*contexts.DiagnosticLogContext)
		diagCtx2, ok2 := b.Context.Value(contexts.DiagnosticLogContextKey).(*contexts.DiagnosticLogContext)
		if !ok1 || !ok2 {
			return false
		}

		if !contexts.DiagnosticLogContextEquals(diagCtx1, diagCtx2) {
			return false
		}

		actCtx1, ok1 := a.Context.Value(contexts.ActivityLogContextKey).(*contexts.ActivityLogContext)
		actCtx2, ok2 := b.Context.Value(contexts.ActivityLogContextKey).(*contexts.ActivityLogContext)
		if !ok1 || !ok2 {
			return false
		}

		if !contexts.ActivityLogContextEquals(actCtx1, actCtx2) {
			return false
		}
	}

	return true
}

type COAResponse struct {
	ContentType string            `json:"contentType"`
	Body        []byte            `json:"body"`
	State       State             `json:"state"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	RedirectUri string            `json:"redirectUri,omitempty"`
}

func (c COAResponse) String() string {
	return string(c.Body)
}

func (c COAResponse) Println() {
	fmt.Println(string(c.Body))
}

type COAHandler func(COARequest) COAResponse

type Endpoint struct {
	Methods    []string
	Version    string
	Route      string
	Handler    COAHandler
	Parameters []string
}

func (e Endpoint) GetPath() string {
	path := fmt.Sprintf("/%s/%s", e.Version, e.Route)
	for _, p := range e.Parameters {
		path += "/{" + p + "}"
	}

	return path
}
