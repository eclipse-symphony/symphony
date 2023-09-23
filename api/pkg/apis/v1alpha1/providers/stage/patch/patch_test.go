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

package patch

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

func TestPatchSolution(t *testing.T) {
	testPatchSolution := os.Getenv("TEST_PATCH_SOLUTION")
	if testPatchSolution != "yes" {
		t.Skip("Skipping becasue TEST_PATCH_SOLUTION is missing or not set to 'yes'")
	}
	provider := PatchStageProvider{}
	err := provider.Init(PatchStageProviderConfig{
		BaseUrl:  "http://localhost:8082/v1alpha2/",
		User:     "admin",
		Password: "",
	})

	provider.SetContext(&contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), map[string]interface{}{
		"objectType":  "solution",
		"objectName":  "sample-redis",
		"patchSource": "test-config",
		"component":   "sample-redis",
		"property":    "deep",
		"subKey":      "key",
		"dedupKey":    "foo",
		"patchAction": "remove",
	})
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, "OK", outputs["status"])
}
