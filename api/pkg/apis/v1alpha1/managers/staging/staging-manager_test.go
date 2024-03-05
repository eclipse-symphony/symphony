/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package staging

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	memoryqueue "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/queue/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestPoll(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	queueProvider := &memoryqueue.MemoryQueueProvider{}
	queueProvider.Init(memoryqueue.MemoryQueueProviderConfig{})

	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	manager := StagingManager{
		StateProvider: stateProvider,
		QueueProvider: queueProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  ts.URL + "/",
				Username: "admin",
				Password: "",
			},
		},
	}
	queueProvider.Enqueue("site-job-queue", "fake")
	errList := manager.Poll()
	assert.Nil(t, errList)

	jobData, err := queueProvider.Dequeue("fake")
	assert.Nil(t, err)
	assert.NotNil(t, jobData)
	assert.Equal(t, "catalog1", jobData.(v1alpha2.JobData).Id)
	assert.Equal(t, v1alpha2.JobUpdate, jobData.(v1alpha2.JobData).Action)

	item, err := stateProvider.Get(context.Background(), states.GetRequest{
		ID: "fake-catalog1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, item)
}

func TestHandleJobEvent(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	queueProvider := &memoryqueue.MemoryQueueProvider{}
	queueProvider.Init(memoryqueue.MemoryQueueProviderConfig{})

	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	manager := StagingManager{
		StateProvider: stateProvider,
		QueueProvider: queueProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  ts.URL + "/",
				Username: "admin",
				Password: "",
			},
		},
	}
	err := manager.HandleJobEvent(context.Background(), v1alpha2.Event{
		Metadata: map[string]string{
			"site": "fake",
		},
		Body: v1alpha2.JobData{
			Id:     "catalog1",
			Action: v1alpha2.JobUpdate,
		},
	})
	assert.Nil(t, err)

	jobData, err := queueProvider.Dequeue("fake")
	assert.Nil(t, err)
	assert.NotNil(t, jobData)
	assert.Equal(t, "catalog1", jobData.(v1alpha2.JobData).Id)
	assert.Equal(t, v1alpha2.JobUpdate, jobData.(v1alpha2.JobData).Action)

	site, err := queueProvider.Dequeue("site-job-queue")
	assert.Nil(t, err)
	assert.NotNil(t, "fake", site.(string))
}
func TestGetABatchForSite(t *testing.T) {
	queueProvider := &memoryqueue.MemoryQueueProvider{}
	queueProvider.Init(memoryqueue.MemoryQueueProviderConfig{})

	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	manager := StagingManager{
		StateProvider: stateProvider,
		QueueProvider: queueProvider,
	}

	queueProvider.Enqueue("fake", v1alpha2.JobData{
		Id:     "catalog1",
		Action: v1alpha2.JobUpdate,
	})
	queueProvider.Enqueue("fake", v1alpha2.JobData{
		Id:     "catalog2",
		Action: v1alpha2.JobUpdate,
	})
	jobs, err := manager.GetABatchForSite("fake", 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(jobs))
	assert.Equal(t, "catalog1", jobs[0].Id)
	assert.Equal(t, v1alpha2.JobUpdate, jobs[0].Action)

	jobs, err = manager.GetABatchForSite("fake", 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(jobs))
	assert.Equal(t, "catalog2", jobs[0].Id)
	assert.Equal(t, v1alpha2.JobUpdate, jobs[0].Action)
}

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func InitializeMockSymphonyAPI() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		log.Info("Mock Symphony API called", "path", r.URL.Path)
		switch r.URL.Path {
		case "/catalogs/registry":
			response = []model.CatalogState{{
				ObjectMeta: model.ObjectMeta{
					Name: "catalog1",
				},
				Spec: &model.CatalogSpec{
					Generation: "1",
					ParentName: "fakeparent",
				},
				Status: &model.CatalogStatus{
					Properties: map[string]string{},
				},
			}}
		default:
			response = AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}

		json.NewEncoder(w).Encode(response)
	}))
	return ts
}
