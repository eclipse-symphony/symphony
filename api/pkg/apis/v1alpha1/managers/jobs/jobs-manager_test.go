/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package jobs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestHandleEvent(t *testing.T) {
	testInstanceId := os.Getenv("TEST_INSTANCE_ID")
	if testInstanceId == "" {
		t.Skip("Skipping becasue TEST_INSTANCE_ID is missing")
	}
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := JobsManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "state",
			"baseUrl":         "http://localhost:8082/v1alpha2/",
			"password":        "",
			"user":            "admin",
			"interval":        "#15",
		},
	}, map[string]providers.IProvider{
		"state": stateProvider,
	})
	assert.Nil(t, err)
	errs := manager.HandleJobEvent(context.Background(), v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": "instance",
		},
		Body: v1alpha2.JobData{
			Id:     testInstanceId,
			Action: "UPDATE",
		},
	})
	assert.Nil(t, errs)
}
func TestHandleJobEvent(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	ts := InitializeMockSymphonyAPI()
	defer ts.Close()

	jobManager := JobsManager{}
	err := jobManager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "state",
			"baseUrl":         ts.URL + "/",
			"password":        "",
			"user":            "admin",
			"interval":        "#15",
		},
	}, map[string]providers.IProvider{
		"state": stateProvider,
	})
	assert.Nil(t, err)

	errs := jobManager.HandleJobEvent(context.Background(), v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": "instance",
		},
		Body: v1alpha2.JobData{
			Id:     "instance1",
			Action: "UPDATE",
		},
	})
	assert.Nil(t, errs)
	instance, err := stateProvider.Get(context.Background(), states.GetRequest{ID: "i_instance1"})
	assert.Nil(t, err)
	assert.NotNil(t, instance)

	errs = jobManager.HandleJobEvent(context.Background(), v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": "target",
		},
		Body: v1alpha2.JobData{
			Id:     "target1",
			Action: "UPDATE",
		},
	})
	assert.Nil(t, errs)
	target, err := stateProvider.Get(context.Background(), states.GetRequest{ID: "t_target1"})
	assert.Nil(t, err)
	assert.NotNil(t, target)
}

func TestHandleScheduleEvent(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	jobManager := JobsManager{}
	err := jobManager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "state",
			"baseUrl":         "http://localhost:8082/v1alpha2/",
			"password":        "",
			"user":            "admin",
			"interval":        "#15",
		},
	}, map[string]providers.IProvider{
		"state": stateProvider,
	})
	assert.Nil(t, err)
	jobManager.HandleScheduleEvent(context.Background(), v1alpha2.Event{
		Body: v1alpha2.ActivationData{Campaign: "campaign1", Activation: "activation1"},
	})

	schedule, err := stateProvider.Get(context.Background(), states.GetRequest{ID: "sch_campaign1-activation1"})
	assert.Nil(t, err)
	assert.NotNil(t, schedule)
}

func TestHandleheartbeatEvent(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	jobManager := JobsManager{}
	err := jobManager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "state",
			"baseUrl":         "http://localhost:8082/v1alpha2/",
			"password":        "",
			"user":            "admin",
			"interval":        "#15",
		},
	}, map[string]providers.IProvider{
		"state": stateProvider,
	})
	assert.Nil(t, err)
	jobManager.HandleHeartBeatEvent(context.Background(), v1alpha2.Event{
		Body: v1alpha2.HeartBeatData{JobId: "instance1"},
	})

	heartbeat, err := stateProvider.Get(context.Background(), states.GetRequest{ID: "h_instance1"})
	assert.Nil(t, err)
	assert.NotNil(t, heartbeat)
}

func TestPoll(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	ts := InitializeMockSymphonyAPI()
	defer ts.Close()
	jobManager := JobsManager{}
	err := jobManager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state":  "state",
			"baseUrl":          ts.URL + "/",
			"password":         "",
			"user":             "admin",
			"interval":         "#15",
			"poll.enabled":     "true",
			"schedule.enabled": "true",
		},
	}, map[string]providers.IProvider{
		"state": stateProvider,
	})
	assert.Nil(t, err)
	jobManager.HandleScheduleEvent(context.Background(), v1alpha2.Event{
		Body: v1alpha2.ActivationData{Campaign: "campaign1", Activation: "activation1", Schedule: &v1alpha2.ScheduleSpec{Time: "03:04:05PM", Date: "2006-01-02"}},
	})
	enabled := jobManager.Enabled()
	assert.True(t, enabled)
	errlist := jobManager.Poll()
	assert.Nil(t, errlist)
}

func TestDelayOrSkipJobPoll(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	ts := InitializeMockSymphonyAPI()
	defer ts.Close()
	jobManager := JobsManager{}
	err := jobManager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state":  "state",
			"baseUrl":          ts.URL + "/",
			"password":         "",
			"user":             "admin",
			"interval":         "#15",
			"poll.enabled":     "true",
			"schedule.enabled": "true",
		},
	}, map[string]providers.IProvider{
		"state": stateProvider,
	})
	assert.Nil(t, err)
	jobManager.HandleHeartBeatEvent(context.Background(), v1alpha2.Event{
		Body: v1alpha2.HeartBeatData{JobId: "instance1", Time: time.Now().Add(-time.Hour)},
	})
	err = jobManager.DelayOrSkipJob(context.Background(), "instances", v1alpha2.JobData{Id: "instance1", Action: "UPDATE"})
	assert.Nil(t, err)

	jobManager.HandleHeartBeatEvent(context.Background(), v1alpha2.Event{
		Body: v1alpha2.HeartBeatData{JobId: "instance1", Time: time.Now()},
	})
	err = jobManager.DelayOrSkipJob(context.Background(), "instances", v1alpha2.JobData{Id: "instance1", Action: "UPDATE"})
	assert.NotNil(t, err)
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
		case "/instances/instance1":
			response = model.InstanceState{
				Id:        "instance1",
				Namespace: "default",
				Spec: &model.InstanceSpec{
					Name:     "instance1",
					Solution: "solution1",
				},
			}
		case "/instances":
			response = []model.InstanceState{{
				Id:        "instance1",
				Namespace: "default",
				Spec: &model.InstanceSpec{
					Name:     "instance1",
					Solution: "solution1",
				},
			}}
		case "/targets/registry":
			response = []model.TargetState{{
				Id:        "target1",
				Namespace: "default",
				Spec: &model.TargetSpec{
					DisplayName: "target1",
				},
			}}
		case "/targets/registry/target1":
			response = model.TargetState{
				Id:        "target1",
				Namespace: "default",
				Spec: &model.TargetSpec{
					DisplayName: "target1",
				},
			}
		case "/solutions/solution1":
			response = model.SolutionState{
				Id:        "solution1",
				Namespace: "default",
				Spec:      &model.SolutionSpec{},
			}
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
