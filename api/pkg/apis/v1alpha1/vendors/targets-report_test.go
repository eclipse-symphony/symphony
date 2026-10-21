/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/status"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	mqttbinding "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/bindings/mqtt"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	mqttreporter "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/mqtt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func targetReportRequest(body string) v1alpha2.COARequest {
	return v1alpha2.COARequest{
		Context: context.Background(), Method: "POST", Route: "report", Body: []byte(body),
		Parameters: map[string]string{
			"id": "edge-01", "namespace": "factory", "group": model.FabricGroup,
			"kind": "targets", "version": "v1", "overwrite": "false",
		},
	}
}

func TestTargetsReport(t *testing.T) {
	vendor := createTargetsVendor()
	ctx := context.Background()
	require.NoError(t, vendor.TargetsManager.UpsertState(ctx, "edge-01", model.TargetState{
		ObjectMeta: model.ObjectMeta{Name: "edge-01", Namespace: "factory"},
		Spec:       &model.TargetSpec{DisplayName: "My edge agent"},
	}))
	for _, body := range []string{`{"agent.site":"factory-a"}`, `{"agent.connected":"true"}`} {
		assert.Equal(t, v1alpha2.OK, vendor.onReport(targetReportRequest(body)).State)
	}
	target, err := vendor.TargetsManager.GetState(ctx, "edge-01", "factory")
	require.NoError(t, err)
	assert.Equal(t, "My edge agent", target.Spec.DisplayName)
	assert.Equal(t, "factory-a", target.Status.Properties["agent.site"])
	assert.Equal(t, "true", target.Status.Properties["agent.connected"])
	assert.False(t, target.Status.LastModified.IsZero())
	request := targetReportRequest("{}")
	request.Parameters["namespace"] = "missing"
	assert.Equal(t, v1alpha2.NotFound, vendor.onReport(request).State)
}

func TestTargetsReportInvalidRequests(t *testing.T) {
	vendor := createTargetsVendor()
	for _, body := range []string{"invalid", "null", "[]", `{"value":null}`, `{"value":1}`, `{"":""}`} {
		assert.Equal(t, v1alpha2.BadRequest, vendor.onReport(targetReportRequest(body)).State, body)
	}
	for key, value := range map[string]string{"id": "", "group": "other", "kind": "secrets", "version": "v2", "overwrite": "true"} {
		request := targetReportRequest("{}")
		request.Parameters[key] = value
		assert.Equal(t, v1alpha2.BadRequest, vendor.onReport(request).State, key)
	}
	request := targetReportRequest("{}")
	request.Method = "GET"
	assert.Equal(t, v1alpha2.MethodNotAllowed, vendor.onReport(request).State)
}

func TestTargetsReportMQTTStandalone(t *testing.T) {
	address := startStatusTestBroker(t)
	vendor := createTargetsVendor()
	require.NoError(t, vendor.TargetsManager.UpsertState(context.Background(), "edge-01", model.TargetState{
		ObjectMeta: model.ObjectMeta{Name: "edge-01", Namespace: "factory"},
		Spec:       &model.TargetSpec{DisplayName: "Standalone agent"},
	}))
	bindingConfig := mqttbinding.MQTTBindingConfig{
		BrokerAddress: address, ClientID: "standalone-receiver",
		RequestTopic: "standalone/request", ResponseTopic: "standalone/response", TimeoutSeconds: 2,
	}
	binding := &mqttbinding.MQTTBinding{}
	require.NoError(t, binding.Launch(bindingConfig, vendor.GetEndpoints()))
	t.Cleanup(func() { require.NoError(t, binding.Shutdown(context.Background())) })
	reporterConfig := mqttreporter.MQTTReporterConfig{MQTTBindingConfig: bindingConfig, Route: "report"}
	reporterConfig.ClientID = "standalone-reporter"
	reporter := &mqttreporter.MQTTReporter{}
	require.NoError(t, reporter.Init(reporterConfig))
	manager := &status.StatusManager{}
	require.NoError(t, manager.Init(nil, managers.ManagerConfig{Properties: map[string]string{
		"target.name": "edge-01", "target.namespace": "factory", "providers.reporter": "mqtt",
		"report.properties": `{"agent.site":"factory-a"}`,
	}}, map[string]providers.IProvider{"mqtt": reporter}))
	t.Cleanup(func() { require.NoError(t, manager.Shutdown(context.Background())) })
	require.NoError(t, manager.Report(context.Background()))
	target, err := vendor.TargetsManager.GetState(context.Background(), "edge-01", "factory")
	require.NoError(t, err)
	assert.Equal(t, "Standalone agent", target.Spec.DisplayName)
	assert.Equal(t, "factory-a", target.Status.Properties["agent.site"])
	assert.NotEmpty(t, target.Status.Properties[status.ReportedAtProperty])
}
