/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package status

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingReporter struct {
	id, namespace, group, kind, version string
	properties                          map[string]string
	overwrite                           bool
	calls                               int
	err                                 error
	closed                              bool
}

func (r *recordingReporter) Init(providers.IProviderConfig) error { return nil }
func (r *recordingReporter) Report(id, namespace, group, kind, version string, properties map[string]string, overwrite bool) error {
	r.id, r.namespace, r.group, r.kind, r.version = id, namespace, group, kind, version
	r.properties, r.overwrite = properties, overwrite
	r.calls++
	return r.err
}
func (r *recordingReporter) Shutdown(context.Context) error {
	r.closed = true
	return nil
}

func managerConfig(bag string) managers.ManagerConfig {
	return managers.ManagerConfig{
		Type: "managers.symphony.status",
		Properties: map[string]string{
			"target.name":        "edge-01",
			"providers.reporter": "reporter",
			"poll.enabled":       "true",
			"report.properties":  bag,
		},
	}
}

func TestStatusReportEvaluatesLocallyEveryTick(t *testing.T) {
	t.Setenv("SYMPHONY_STATUS_TEST_SITE", "factory-a")
	config := managerConfig(`{"site":"$env:SYMPHONY_STATUS_TEST_SITE","description":"site-${{ $property('site') }}","answer":"${{ 6 * 7 }}","connected":"true"}`)
	r := &recordingReporter{}
	s := &StatusManager{}
	require.NoError(t, s.Init(nil, config, map[string]providers.IProvider{"reporter": r}))
	assert.True(t, s.Enabled())
	before := time.Now().UTC()
	require.Empty(t, s.Poll())
	assert.Equal(t, "edge-01", r.id)
	assert.Equal(t, "default", r.namespace)
	assert.Equal(t, "fabric.symphony", r.group)
	assert.Equal(t, "targets", r.kind)
	assert.Equal(t, "v1", r.version)
	assert.False(t, r.overwrite)
	assert.Equal(t, "factory-a", r.properties["site"])
	assert.Equal(t, "site-factory-a", r.properties["description"])
	assert.Equal(t, "42", r.properties["answer"])
	assert.Equal(t, "true", r.properties["connected"])
	reportedAt, err := time.Parse(time.RFC3339Nano, r.properties[ReportedAtProperty])
	require.NoError(t, err)
	assert.False(t, reportedAt.Before(before))
	t.Setenv("SYMPHONY_STATUS_TEST_SITE", "factory-b")
	require.Empty(t, s.Poll())
	assert.Equal(t, "factory-b", r.properties["site"])
	assert.Equal(t, "site-factory-b", r.properties["description"])
	assert.Equal(t, 2, r.calls)
	assert.Equal(t, "$env:SYMPHONY_STATUS_TEST_SITE", s.bag["site"])
	assert.Empty(t, s.Reconcil())
	require.NoError(t, s.Shutdown(context.Background()))
	assert.True(t, r.closed)
}

func TestStatusReportPreservesEnvironmentData(t *testing.T) {
	t.Setenv("SYMPHONY_STATUS_TEST_VALUE", "${{ $invalid() }}")
	t.Setenv("SYMPHONY_STATUS_TEST_EMPTY", "")
	r := &recordingReporter{}
	s := &StatusManager{}
	require.NoError(t, s.Init(nil, managerConfig(`{"value":"$env:SYMPHONY_STATUS_TEST_VALUE","empty":"","emptyEnv":"$env:SYMPHONY_STATUS_TEST_EMPTY"}`), map[string]providers.IProvider{"reporter": r}))
	require.Empty(t, s.Poll())
	assert.Equal(t, "${{ $invalid() }}", r.properties["value"])
	assert.Contains(t, r.properties, "empty")
	assert.Empty(t, r.properties["empty"])
	assert.Contains(t, r.properties, "emptyEnv")
	assert.Empty(t, r.properties["emptyEnv"])
}

func TestStatusManagerConfiguration(t *testing.T) {
	for _, bag := range []string{"", "null", "[]", `{"nested":{}}`, `{"n":1}`, `{"null":null}`, `{"":""}`, `{"agent.reportedAt":"override"}`} {
		t.Run(bag, func(t *testing.T) {
			s := &StatusManager{}
			assert.Error(t, s.Init(nil, managerConfig(bag), map[string]providers.IProvider{"reporter": &recordingReporter{}}))
		})
	}
	for _, field := range []string{"target.name", "target.namespace", "poll.enabled"} {
		t.Run(field, func(t *testing.T) {
			config := managerConfig("{}")
			if field == "poll.enabled" {
				config.Properties[field] = "invalid"
			} else {
				config.Properties[field] = ""
			}
			assert.Error(t, (&StatusManager{}).Init(nil, config, map[string]providers.IProvider{"reporter": &recordingReporter{}}))
		})
	}
	s := &StatusManager{}
	assert.Error(t, s.Init(nil, managerConfig("{}"), nil))
	t.Setenv("SYMPHONY_STATUS_TEST_TARGET", "edge-02")
	t.Setenv("SYMPHONY_STATUS_TEST_NAMESPACE", "edge")
	config := managerConfig("{}")
	config.Properties["target.name"] = "$env:SYMPHONY_STATUS_TEST_TARGET"
	config.Properties["target.namespace"] = "$env:SYMPHONY_STATUS_TEST_NAMESPACE"
	config.Properties["poll.enabled"] = "false"
	r := &recordingReporter{}
	require.NoError(t, s.Init(nil, config, map[string]providers.IProvider{"reporter": r}))
	assert.False(t, s.Enabled())
	require.NoError(t, s.Report(context.Background()))
	assert.Equal(t, "edge-02", r.id)
	assert.Equal(t, "edge", r.namespace)
	assert.Len(t, r.properties, 1)
}

func TestStatusReportErrorsAreNotPublished(t *testing.T) {
	for _, bag := range []string{
		`{"bad":"$env:SYMPHONY_STATUS_VARIABLE_NOT_SET"}`,
		`{"bad":"${{ $invalid() }}"}`,
		`{"bad":"${{ $property('missing') }}"}`,
		`{"bad":"${{ }}"}`,
		`{"a":"${{ $property('b') }}","b":"${{ $property('a') }}"}`,
	} {
		t.Run(bag, func(t *testing.T) {
			r := &recordingReporter{}
			s := &StatusManager{}
			require.NoError(t, s.Init(nil, managerConfig(bag), map[string]providers.IProvider{"reporter": r}))
			require.Len(t, s.Poll(), 1)
			assert.Zero(t, r.calls)
		})
	}
	r := &recordingReporter{err: errors.New("report failed")}
	s := &StatusManager{}
	require.NoError(t, s.Init(nil, managerConfig("{}"), map[string]providers.IProvider{"reporter": r}))
	errs := s.Poll()
	require.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], r.err)
}
