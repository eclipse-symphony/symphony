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

package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
}
func TestGetEmpty(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	_, err = provider.Read("obj", "field")
	assert.NotNil(t, err)
}
func TestGet(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider.Set("obj", "field", "obj::field")
	val, err := provider.Read("obj", "field")
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)
}
func TestGetObject(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider.SetObject("obj", map[string]string{"field": "obj::field"})
	val, err := provider.ReadObject("obj")
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"field": "obj::field"}, val)
}
func TestDelete(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider.Set("obj", "field", "obj::field")
	err = provider.Remove("obj", "field")
	assert.Nil(t, err)
	_, err = provider.Read("obj", "field")
	assert.NotNil(t, err)
}
func TestDeleteObject(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider.SetObject("obj", map[string]string{"field": "obj::field"})
	err = provider.RemoveObject("obj")
	assert.Nil(t, err)
	_, err = provider.ReadObject("obj")
	assert.NotNil(t, err)
}
