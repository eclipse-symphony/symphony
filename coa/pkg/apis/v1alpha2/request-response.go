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
