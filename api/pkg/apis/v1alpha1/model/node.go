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

// +kubebuilder:object:generate=true
type NodeSpec struct {
	Id             string            `json:"id"`
	NodeType       string            `json:"type"`
	Name           string            `json:"name"`
	Configurations map[string]string `json:"configurations,omitempty"`
	Inputs         []RouteSpec       `json:"inputs,omitempty"`
	Outputs        []RouteSpec       `json:"outputs,omitempty"`
	Model          string            `json:"model,omitempty"`
}

func (n NodeSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherSpec, ok := other.(NodeSpec)
	if !ok {
		return false, nil
	}
	if n.Id != otherSpec.Id {
		return false, nil
	}
	if n.NodeType != otherSpec.NodeType {
		return false, nil
	}
	if n.Name != otherSpec.Name {
		return false, nil
	}
	if n.Model != otherSpec.Model {
		return false, nil
	}
	if !StringMapsEqual(n.Configurations, otherSpec.Configurations, nil) {
		return false, nil
	}
	if !SlicesEqual(n.Inputs, otherSpec.Inputs) {
		return false, nil
	}
	if !SlicesEqual(n.Outputs, otherSpec.Outputs) {
		return false, nil
	}
	return true, nil
}
