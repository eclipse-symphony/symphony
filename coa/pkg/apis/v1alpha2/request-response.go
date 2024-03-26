/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"context"
	"fmt"
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

func (in *COARequest) DeepCopyInto(out *COARequest) {
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
