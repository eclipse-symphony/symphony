/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package staging

import (
	"context"
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/queue"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

var log = logger.NewLogger("coa.runtime")

type StagingManager struct {
	managers.Manager
	QueueProvider queue.IQueueProvider
	StateProvider states.IStateProvider
	apiClient     utils.ApiClient
}

const Site_Job_Queue = "site-job-queue"

func (s *StagingManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	queueProvider, err := managers.GetQueueProvider(config, providers)
	if err == nil {
		s.QueueProvider = queueProvider
	} else {
		return err
	}
	stateprovider, err := managers.GetVolatileStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	s.apiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
	return nil
}
func (s *StagingManager) Enabled() bool {
	return s.Config.Properties["poll.enabled"] == "true"
}
func (s *StagingManager) Poll() []error {
	ctx, span := observability.StartSpan("Staging Manager", context.TODO(), &map[string]string{
		"method": "Poll",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.Debug(" M (Staging): Polling...")
	if s.QueueProvider.Size(context.TODO(), Site_Job_Queue) == 0 {
		return nil
	}
	var site interface{}
	site, err = s.QueueProvider.Dequeue(context.TODO(), Site_Job_Queue)
	if err != nil {
		log.Errorf(" M (Staging): Failed to poll: %s", err.Error())
		return []error{err}
	}
	siteId := utils.FormatAsString(site)
	var catalogs []model.CatalogState
	catalogs, err = s.apiClient.GetCatalogs(ctx, "",
		s.VendorContext.SiteInfo.CurrentSite.Username,
		s.VendorContext.SiteInfo.CurrentSite.Password)
	if err != nil {
		log.Errorf(" M (Staging): Failed to get catalogs: %s", err.Error())
		return []error{err}
	}
	for _, catalog := range catalogs {
		cacheId := siteId + "-" + catalog.ObjectMeta.Name
		getRequest := states.GetRequest{
			ID: cacheId,
			Metadata: map[string]interface{}{
				"version":   "v1",
				"group":     model.FederationGroup,
				"resource":  "catalogs",
				"namespace": catalog.ObjectMeta.Namespace,
			},
		}
		var entry states.StateEntry
		entry, err = s.StateProvider.Get(ctx, getRequest)
		if err == nil && entry.Body != nil && entry.Body.(string) == catalog.ObjectMeta.ETag {
			continue
		}
		if err != nil && !utils.IsNotFound(err) {
			log.Errorf(" M (Staging): Failed to get catalog %s: %s", catalog.ObjectMeta.Name, err.Error())
		}
		s.QueueProvider.Enqueue(context.TODO(), siteId, v1alpha2.JobData{
			Id:     catalog.ObjectMeta.Name,
			Action: v1alpha2.JobUpdate,
			Body:   catalog,
		})

		// TODO: clean up the catalog synchronization status for multi-site
		_, err = s.StateProvider.Upsert(ctx, states.UpsertRequest{
			Value: states.StateEntry{
				ID:   cacheId,
				Body: catalog.ObjectMeta.ETag,
			},
			Metadata: map[string]interface{}{
				"version":   "v1",
				"group":     model.FederationGroup,
				"resource":  "catalogs",
				"namespace": catalog.ObjectMeta.Namespace,
			},
		})
		if err != nil {
			log.Errorf(" M (Staging): Failed to record catalog %s: %s", catalog.ObjectMeta.Name, err.Error())
		}
	}
	return nil
}
func (s *StagingManager) Reconcil() []error {
	return nil
}

func (s *StagingManager) HandleJobEvent(ctx context.Context, event v1alpha2.Event) error {
	_, span := observability.StartSpan("Staging Manager", ctx, &map[string]string{
		"method": "HandleJobEvent",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	var job v1alpha2.JobData
	jData, _ := json.Marshal(event.Body)
	err = json.Unmarshal(jData, &job)
	if err != nil {
		err = v1alpha2.NewCOAError(nil, "event body is not a job", v1alpha2.BadRequest)
		return err
	}
	s.QueueProvider.Enqueue(context.TODO(), Site_Job_Queue, event.Metadata["site"])
	_, err = s.QueueProvider.Enqueue(context.TODO(), event.Metadata["site"], job)
	return err
}
func (s *StagingManager) GetABatchForSite(site string, count int) ([]v1alpha2.JobData, error) {
	//TODO: this should return a group of jobs as optimization
	s.QueueProvider.Enqueue(context.TODO(), Site_Job_Queue, site)
	if s.QueueProvider.Size(context.TODO(), site) == 0 {
		return nil, nil
	}
	items := []v1alpha2.JobData{}
	itemCount := 0
	for {
		queueElement, err := s.QueueProvider.Dequeue(context.TODO(), site)
		if err != nil {
			return nil, err
		}
		if job, ok := queueElement.(v1alpha2.JobData); ok {
			items = append(items, job)
			itemCount++
		} else {
			s.QueueProvider.Enqueue(context.TODO(), site, queueElement)
		}
		if itemCount == count || s.QueueProvider.Size(context.TODO(), site) == 0 {
			break
		}
	}
	return items, nil
}
