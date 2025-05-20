/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package target

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference/mock"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	reference := &MockDeviceReferenceProvider{}
	reporter := &MockReporter{}
	prob := &MockProb{}
	uploader := &MockUploader{}
	manager := TargetManager{
		ReferenceProvider: reference,
		Reporter:          reporter,
		ProbeProvider:     prob,
	}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.probe":     "MockProb",
			"providers.reference": "MockDeviceReferenceProvider",
			"providers.uploader":  "MockUploader",
			"providers.reporter":  "MockReporter",
		},
	}

	providers := make(map[string]providers.IProvider)
	providers["MockProb"] = prob
	providers["MockReporter"] = reporter
	providers["MockUploader"] = uploader
	providers["MockDeviceReferenceProvider"] = reference
	err := manager.Init(nil, config, providers)
	assert.Nil(t, err)
}

// write test case to create a TargetSpec using the manager
func TestBasic(t *testing.T) {
	provider := &mock.MockReferenceProvider{}
	provider.Init(mock.MockReferenceProvider{})
	manager := TargetManager{
		ReferenceProvider: provider,
	}
	assert.NotNil(t, manager)
	err := manager.Apply(context.Background(), model.TargetSpec{})
	assert.Nil(t, err)
	_, errGet := manager.Get(context.Background())
	assert.Nil(t, errGet)
	err = manager.Remove(context.Background(), model.TargetSpec{})
	assert.Nil(t, err)
	enabled := manager.Enabled()
	assert.False(t, enabled)
	errPoll := manager.Poll()
	assert.Equal(t, []error{}, errPoll)
	errRec := manager.Reconcil()
	assert.Nil(t, errRec)
}

func TestReport(t *testing.T) {
	provider := &MockDeviceReferenceProvider{}
	reporter := &MockReporter{}
	provider.Init(MockDeviceReferenceProvider{})
	manager := TargetManager{
		ReferenceProvider: provider,
		Reporter:          reporter,
	}
	errRep := manager.reportStatus("testDev", "default", "testTar", "testSnapshot", "active", "active", true, "testErr")
	assert.Equal(t, []error{}, errRep)
}

func TestPollAndUpload(t *testing.T) {
	provider := &MockDeviceReferenceProvider{}
	reporter := &MockReporter{}
	prob := &MockProb{}
	uploader := &MockUploader{}
	provider.Init(MockDeviceReferenceProvider{})
	manager := TargetManager{
		ReferenceProvider: provider,
		Reporter:          reporter,
		ProbeProvider:     prob,
		UploaderProvider:  uploader,
	}
	errPoll := manager.Poll()
	assert.NotNil(t, errPoll)
}

type MockReporter struct{}

func (r *MockReporter) Init(config providers.IProviderConfig) error {
	return nil
}
func (r *MockReporter) Report(id string, namespace string, group string, kind string, version string, properties map[string]string, overwrite bool) error {
	return nil
}

type MockProb struct{}

func (r *MockProb) Init(config providers.IProviderConfig) error {
	return nil
}
func (r *MockProb) Probe(user string, password string, ip string, name string) (map[string]string, error) {
	prob := map[string]string{
		"snapshot": "snapshot.txt",
	}
	return prob, nil
}

type MockUploader struct{}

func (r *MockUploader) Init(config providers.IProviderConfig) error {
	return nil
}
func (r *MockUploader) Upload(name string, data []byte) (string, error) {
	return "done", nil
}

type MockDeviceReferenceProvider struct {
	mock.MockReferenceProvider
}

func (m *MockDeviceReferenceProvider) List(labelSelector string, fieldSelector string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	properties := map[string]string{
		"user":     "user",
		"password": "password",
		"ip":       "ip",
	}
	metadata := make(map[string]interface{})
	metadata["name"] = "name"
	deviceSpec := DeviceSpec{
		Properties: properties,
	}
	device := Device{
		Object: Object{
			ApiVersion: "version",
			Kind:       "kind",
			Metadata:   metadata,
			Spec:       deviceSpec,
		},
	}
	devices := []Device{device}

	return devices, nil
}
