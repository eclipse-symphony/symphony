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

package model

import "errors"

// +kubebuilder:object:generate=true
type RouteSpec struct {
	Route      string            `json:"route"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
	Filters    []FilterSpec      `json:"filters,omitempty"`
}

func (c RouteSpec) DeepEquals(other IDeepEquals) (bool, error) { // avoid using reflect, which has performance problems
	otherC, ok := other.(RouteSpec)
	if !ok {
		return false, errors.New("parameter is not a RouteSpec type")
	}

	if c.Route != otherC.Route {
		return false, nil
	}

	if c.Type != otherC.Type {
		return false, nil
	}

	if !StringMapsEqual(c.Properties, otherC.Properties, nil) {
		return false, nil
	}

	if !SlicesEqual(c.Filters, otherC.Filters) {
		return false, nil
	}

	return true, nil
}
