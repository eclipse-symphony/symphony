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

type (
	ConnectionSpec struct {
		Node  string `json:"node"`
		Route string `json:"route"`
	}
	EdgeSpec struct {
		Source ConnectionSpec `json:"source"`
		Target ConnectionSpec `json:"target"`
	}
)

func (c ConnectionSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherSpec, ok := other.(*ConnectionSpec)
	if !ok {
		return false, nil
	}
	if c.Node != otherSpec.Node {
		return false, nil
	}
	if c.Route != otherSpec.Route {
		return false, nil
	}
	return true, nil
}
func (c EdgeSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherSpec, ok := other.(*EdgeSpec)
	if !ok {
		return false, nil
	}
	equal, err := c.Source.DeepEquals(&otherSpec.Source)
	if err != nil {
		return false, err
	}
	if !equal {
		return false, nil
	}
	equal, err = c.Target.DeepEquals(&otherSpec.Target)
	if err != nil {
		return false, err
	}
	if !equal {
		return false, nil
	}
	return true, nil
}
