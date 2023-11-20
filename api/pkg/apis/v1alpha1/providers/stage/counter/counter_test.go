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

package counter

import (
	"context"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestSmpleCount(t *testing.T) {

	provider := CounterStageProvider{}
	err := provider.Init(CounterStageProvider{})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo": 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), outputs["foo"])
}
func TestAccumulate(t *testing.T) {

	provider := CounterStageProvider{}
	err := provider.Init(CounterStageProvider{})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo": 1,
	})
	outputs2, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo":     1,
		"__state": outputs["__state"],
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(2), outputs2["foo"])
}
func TestSmpleCountWithInitialValue(t *testing.T) {

	provider := CounterStageProvider{}
	err := provider.Init(CounterStageProvider{})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo":      1,
		"foo.init": 5,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(6), outputs["foo"])
}
func TestAccumulateWithInitialValue(t *testing.T) {

	provider := CounterStageProvider{}
	err := provider.Init(CounterStageProvider{})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo":      1,
		"foo.init": 5,
	})
	outputs2, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"foo":     1,
		"__state": outputs["__state"],
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(7), outputs2["foo"])
}
