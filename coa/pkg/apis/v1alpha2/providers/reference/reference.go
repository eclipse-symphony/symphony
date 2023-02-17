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

package reference

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IReferenceProvider interface {
	ID() string
	TargetID() string
	Init(config providers.IProviderConfig) error
	Get(id string, namespace string, group string, kind string, version string, ref string) (interface{}, error)
	List(labelSelector string, fieldSelector string, namespace string, group string, kind string, version string, ref string) (interface{}, error)
	SetContext(context *contexts.ManagerContext) error
	ReferenceType() string
	Reconfigure(config providers.IProviderConfig) error
}
