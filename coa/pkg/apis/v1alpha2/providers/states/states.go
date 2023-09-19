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

package states

import (
	"context"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/yalp/jsonpath"
	//"encoding/json"
)

type StateEntry struct {
	ID   string      `json:"id"`
	Body interface{} `json:"body"`
	ETag string      `json:"etag,omitempty"`
}
type IStateProvider interface {
	Init(config providers.IProviderConfig) error
	Upsert(context.Context, UpsertRequest) (string, error)
	Delete(context.Context, DeleteRequest) error
	Get(context.Context, GetRequest) (StateEntry, error)
	List(context.Context, ListRequest) ([]StateEntry, string, error)
	SetContext(context *contexts.ManagerContext) error
}
type GetOption struct {
	Consistency string `json:"consistency"` //eventual or strong
}
type GetRequest struct {
	ID       string            `json:"id"`
	Metadata map[string]string `json:"metadata"`
	Options  GetOption         `json:"options,omitempty"`
}
type DeleteOption struct {
	Concurrency string `json:"concurency"` //concurrency
	Consistency string `json:"consistency` //eventual or strong
}
type DeleteRequest struct {
	ID       string            `json:"id"`
	ETag     *string           `json:"etag,omitempty"`
	Metadata map[string]string `json:"metadata"`
	Options  DeleteOption      `json:"options,omitempty"`
}
type UpsertOption struct {
	Concurrency string `json:"concurrency,omitempty"` //first-write, last-write
	Consistency string `json:"consistency"`           //eventual, strong
}
type UpsertRequest struct {
	Value    StateEntry        `json:"value"`
	ETag     *string           `json:"etag,omitempty"`
	Metadata map[string]string `json:"metadata"`
	Options  UpsertOption      `json:"options,omitempty"`
}
type ListRequest struct {
	FilterType       string            `json:"filterType"`
	Filter           string            `json:"filter"`
	FilterParameters map[string]string `json:"filterParameters"`
	Metadata         map[string]string `json:"metadata"`
}

func JsonPathMatch(jsonData interface{}, path string, target string) bool {
	// var data interface{}
	// if err := json.Unmarshal(jsonData, &data); err != nil {
	// 	return false
	// }
	res, err := jsonpath.Read(jsonData, path)
	if err != nil {
		return false
	}
	return res.(string) == target
}
