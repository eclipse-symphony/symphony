/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package states

import (
	"context"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
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
	GetLatest(context.Context, GetRequest) (StateEntry, error)
	List(context.Context, ListRequest) ([]StateEntry, string, error)
	SetContext(context *contexts.ManagerContext)
}
type GetOption struct {
	Consistency string `json:"consistency"` //eventual or strong
}
type GetRequest struct {
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
	Options  GetOption              `json:"options,omitempty"`
}
type DeleteOption struct {
	Concurrency string `json:"concurency"` //concurrency
	Consistency string `json:"consistency` //eventual or strong
}
type DeleteRequest struct {
	ID       string                 `json:"id"`
	ETag     *string                `json:"etag,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
	Options  DeleteOption           `json:"options,omitempty"`
}
type UpsertOption struct {
	Concurrency     string `json:"concurrency,omitempty"` //first-write, last-write
	Consistency     string `json:"consistency"`           //eventual, strong
	UpdateStateOnly bool   `json:"updateStateOnly,omitempty"`
}
type UpsertRequest struct {
	Value    StateEntry             `json:"value"`
	ETag     *string                `json:"etag,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
	Options  UpsertOption           `json:"options,omitempty"`
}
type ListRequest struct {
	FilterType  string                 `json:"filterType"`
	FilterValue string                 `json:"filterValue"`
	Metadata    map[string]interface{} `json:"metadata"`
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
