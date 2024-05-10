/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package sync

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	manager := SyncManager{}
	managerVendorContext := &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
		Logger: logger.NewLogger("coa.runtime"),
	}
	err := manager.Init(managerVendorContext, managers.ManagerConfig{}, nil)
	assert.Nil(t, err)
}

func TestInitWithoutSiteIdFail(t *testing.T) {
	manager := SyncManager{}
	managerVendorContext := &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		Logger:            logger.NewLogger("coa.runtime"),
	}
	err := manager.Init(managerVendorContext, managers.ManagerConfig{}, nil)
	assert.NotNil(t, err)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "siteId is required", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestEnabled(t *testing.T) {
	manager := SyncManager{}
	managerVendorContext := &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
		Logger: logger.NewLogger("coa.runtime"),
	}
	err := manager.Init(managerVendorContext, managers.ManagerConfig{
		Properties: map[string]string{
			"sync.enabled": "true",
		},
	}, nil)
	assert.Nil(t, err)

	assert.True(t, manager.Enabled())
}

func TestReconcil(t *testing.T) {
	manager := SyncManager{}
	managerVendorContext := &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
		Logger: logger.NewLogger("coa.runtime"),
	}
	err := manager.Init(managerVendorContext, managers.ManagerConfig{
		Properties: map[string]string{
			"sync.enabled": "true",
		},
	}, nil)
	assert.Nil(t, err)

	errs := manager.Reconcil()
	assert.Nil(t, errs)
}

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func InitiazlizeMockSymphonyAPI(siteId string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		fmt.Println("Mock Symphony API called", "path", r.URL.Path)
		switch r.URL.Path {
		case "/federation/sync/" + siteId:
			response = model.SyncPackage{
				Jobs: []v1alpha2.JobData{
					{
						Id:     "job1",
						Action: v1alpha2.JobUpdate,
					},
				},
				Catalogs: []model.CatalogState{
					{
						ObjectMeta: model.ObjectMeta{
							Name: "catalog1",
						},
						Spec: &model.CatalogSpec{
							Type: "Instance",
							Properties: map[string]interface{}{
								"foo": "bar",
							},
							Metadata: map[string]string{
								"foo": "bar",
							},
						},
					},
				},
				Origin: "batch-origin",
			}
		case "/users/auth":
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

func TestPoll(t *testing.T) {
	siteId := "fake"
	ts := InitiazlizeMockSymphonyAPI(siteId)
	defer ts.Close()
	_, err := url.Parse(ts.URL)
	assert.Nil(t, err)

	manager := SyncManager{}
	vendorContext := &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: siteId,
			ParentSite: v1alpha2.SiteConnection{
				BaseUrl:  ts.URL + "/",
				Username: "admin",
				Password: "",
			},
		},
		Logger: logger.NewLogger("coa.runtime"),
	}
	vendorContext.PubsubProvider = &memory.InMemoryPubSubProvider{}
	vendorContext.PubsubProvider.Init(memory.InMemoryPubSubConfig{})
	err = manager.Init(vendorContext, managers.ManagerConfig{
		Properties: map[string]string{
			"sync.enabled": "true",
		},
	}, nil)
	assert.Nil(t, err)

	sig1 := make(chan int)
	sig2 := make(chan int)
	catalog1 := model.CatalogState{}
	job1 := v1alpha2.JobData{}
	// validate that the sync package was published
	catalogCnt := 0
	vendorContext.Subscribe("catalog-sync", func(topic string, event v1alpha2.Event) error {
		catalogCnt++
		jobData := event.Body.(v1alpha2.JobData)
		catalog1 = jobData.Body.(model.CatalogState)
		sig1 <- 1
		return nil
	})
	jobCount := 0
	vendorContext.Subscribe("remote-job", func(topic string, event v1alpha2.Event) error {
		jobCount++
		job1 = event.Body.(v1alpha2.JobData)
		sig2 <- 1
		return nil
	})

	errs := manager.Poll()
	assert.Nil(t, errs)

	<-sig1
	<-sig2
	assert.Equal(t, 1, catalogCnt)
	assert.Equal(t, 1, jobCount)
	assert.Equal(t, "catalog1", catalog1.ObjectMeta.Name)
	assert.Equal(t, "job1", job1.Id)
}
