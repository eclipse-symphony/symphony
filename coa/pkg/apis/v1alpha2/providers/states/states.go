/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package states

import (
	"context"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	//"encoding/json"
)

const (
	FirstWrite = "first-write"
	LastWrite  = "last-write"
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
	Concurrency      string `json:"concurrency,omitempty"` //first-write, last-write
	Consistency      string `json:"consistency"`           //eventual, strong
	UpdateStatusOnly bool   `json:"updateStatusOnly,omitempty"`
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

func GetObjectState(ctx context.Context, stateProvider IStateProvider, resourceType validation.ResourceType, name string, namespace string) (interface{}, error) {
	group, version, resource, kind := validation.GetResourceMetadata(resourceType)

	object, err := stateProvider.Get(ctx, GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     group,
			"version":   version,
			"resource":  resource,
			"kind":      kind,
		},
	})
	if err != nil {
		return nil, err
	}
	return object.Body, nil
}

func formatLabelSelector(labels map[string]string) string {
	var parts []string
	for key, value := range labels {
		parts = append(parts, key+"="+value)
	}
	if len(parts) > 0 {
		return strings.Join(parts, ",")
	}
	return ""
}

func ListObjectStateWithLabels(ctx context.Context, stateProvider IStateProvider, resourceType validation.ResourceType, namespace string, labels map[string]string, count int64) ([]interface{}, error) {
	group, version, resource, kind := validation.GetResourceMetadata(resourceType)

	list, _, err := stateProvider.List(ctx, ListRequest{
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     group,
			"version":   version,
			"resource":  resource,
			"kind":      kind,
		},
		FilterType:  "label",
		FilterValue: formatLabelSelector(labels),
	})
	if err != nil {
		return nil, err
	}

	var objectStateList []interface{}
	for _, item := range list {
		objectStateList = append(objectStateList, item.Body)
	}
	return objectStateList, nil
}

func GetObjectStateWithUniqueName(ctx context.Context, stateProvider IStateProvider, resourceType validation.ResourceType, displayName string, namespace string) (interface{}, error) {
	objectList, err := ListObjectStateWithLabels(ctx, stateProvider, resourceType, namespace, map[string]string{constants.DisplayName: utils.ConvertStringToValidLabel(displayName)}, 1)
	if err != nil {
		return nil, err
	}
	if len(objectList) > 0 {
		return objectList[0], nil
	}
	return nil, v1alpha2.NewCOAError(nil, string(resourceType)+" not found", v1alpha2.NotFound)
}
