/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestCallRemoteProcessor(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)
		json.NewEncoder(w).Encode(model.StageStatus{
			Status: v1alpha2.Done,
			Outputs: map[string]interface{}{
				"foo": "bar",
			},
		})
	}))
	defer server.Close()

	client, err := NewApiClient(context.Background(), server.URL)
	assert.Nil(t, err)

	status, err := client.CallRemoteProcessor(context.Background(), v1alpha2.ActivationData{
		CampaignVersion: "campaignversion-v1",
		Activation:      "activation1",
		Stage:           "stage1",
		Proxy: &v1alpha2.ProxySpec{
			Config: map[string]interface{}{
				"baseUrl": "http://should-not-be-forwarded/",
			},
		},
	}, "", "")
	assert.Nil(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/processor", gotPath)
	// the proxy config must be stripped before the event is sent to the remote site
	_, hasProxy := gotBody["proxy"]
	assert.False(t, hasProxy)
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, "bar", status.Outputs["foo"])
}

func TestReportActivationStatus(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody model.ActivationStatus
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewApiClient(context.Background(), server.URL)
	assert.Nil(t, err)

	err = client.ReportActivationStatus(context.Background(), "activation1", model.ActivationStatus{
		ActivationGeneration: "1",
		Status:               v1alpha2.Done,
		StatusMessage:        "done",
	}, "", "")
	assert.Nil(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/activations/status/activation1", gotPath)
	assert.Equal(t, "done", gotBody.StatusMessage)
}

func TestReportActivationStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewApiClient(context.Background(), server.URL)
	assert.Nil(t, err)

	err = client.ReportActivationStatus(context.Background(), "activation1", model.ActivationStatus{}, "", "")
	assert.NotNil(t, err)
}

func TestGetCampaignVersion(t *testing.T) {
	var gotPath, gotNamespace string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotNamespace = r.URL.Query().Get("namespace")
		json.NewEncoder(w).Encode(model.CampaignVersionState{
			ObjectMeta: model.ObjectMeta{
				Name:      "campaignversion-v1",
				Namespace: "default",
			},
		})
	}))
	defer server.Close()

	client, err := NewApiClient(context.Background(), server.URL)
	assert.Nil(t, err)

	campaignversion, err := client.GetCampaignVersion(context.Background(), "campaignversion-v1", "default", "", "")
	assert.Nil(t, err)
	assert.Equal(t, "/campaignversions/campaignversion-v1", gotPath)
	assert.Equal(t, "default", gotNamespace)
	assert.Equal(t, "campaignversion-v1", campaignversion.ObjectMeta.Name)
}

func TestGetSolutionVersionsForAllNamespaces(t *testing.T) {
	var gotPath, gotRawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRawQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode([]model.SolutionVersionState{
			{ObjectMeta: model.ObjectMeta{Name: "solutionversion-v1"}},
			{ObjectMeta: model.ObjectMeta{Name: "solutionversion-v2"}},
		})
	}))
	defer server.Close()

	client, err := NewApiClient(context.Background(), server.URL)
	assert.Nil(t, err)

	solutionversions, err := client.GetSolutionVersionsForAllNamespaces(context.Background(), "", "")
	assert.Nil(t, err)
	assert.Equal(t, "/solutionversions", gotPath)
	// all-namespaces listing must not carry a namespace filter
	assert.Equal(t, "", gotRawQuery)
	assert.Equal(t, 2, len(solutionversions))
	assert.Equal(t, "solutionversion-v1", solutionversions[0].ObjectMeta.Name)
}
