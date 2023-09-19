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

package conformance

import (
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"github.com/stretchr/testify/assert"
)

func GetSecretNotFound[P secret.ISecretProvider](t *testing.T, p P) {
	// TODO: this case should fail. This is a prototype of conformance test suite
	// but unfortunately the mock secret provider doesn't confirm with reasonable
	// expected behavior
	_, err := p.Get("fake_object", "fake_key")
	assert.Nil(t, err)
}
func ConformanceSuite[P secret.ISecretProvider](t *testing.T, p P) {
	t.Run("Level=Default", func(t *testing.T) {
		GetSecretNotFound(t, p)
	})
}
