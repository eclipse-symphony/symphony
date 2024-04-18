/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rtsp

import (
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := RTSPProbeProvider{}
	err := provider.Init(RTSPProbeProviderConfig{
		Name: "test",
	})
	assert.Nil(t, err)

	properties := map[string]string{
		"name": "test",
	}
	assert.Nil(t, err)
	err = provider.InitWithMap(properties)
	assert.Nil(t, err)

}
func TestProbe(t *testing.T) {
	rtspAddr := os.Getenv("TEST_RTSP")
	if rtspAddr == "" {
		t.Skip("Skipping because TEST_RTSP enviornment variable is not set")
	}
	provider := RTSPProbeProvider{}
	err := provider.Init(RTSPProbeProviderConfig{
		Name: "test",
	})
	assert.Nil(t, err)
	ret, err := provider.Probe("", "", rtspAddr, "abc")
	assert.Nil(t, err)
	_, ok := ret["snapshot"]
	assert.True(t, ok)
}
func TestParseRTSPAddress(t *testing.T) {
	addr, err := fixRtspUrl("rtsp://20.212.158.240:1234/1.mkv", "", "")
	assert.Nil(t, err)
	assert.Equal(t, "rtsp://20.212.158.240:1234/1.mkv", addr)
}
func TestParseRTSPAddressNoPort(t *testing.T) {
	addr, err := fixRtspUrl("rtsp://20.212.158.240/1.mkv", "", "")
	assert.Nil(t, err)
	assert.Equal(t, "rtsp://20.212.158.240/1.mkv", addr)
}
func TestParseRTSPAddressPort554(t *testing.T) {
	addr, err := fixRtspUrl("rtsp://20.212.158.240:554/1.mkv", "", "")
	assert.Nil(t, err)
	assert.Equal(t, "rtsp://20.212.158.240/1.mkv", addr)
}
func TestParseRTSPCustomPortNoPath(t *testing.T) {
	addr, err := fixRtspUrl("rtsp://20.212.158.240:1234", "", "")
	assert.Nil(t, err)
	assert.Equal(t, "rtsp://20.212.158.240:1234", addr)
}
func TestParseRTSPAllCustom(t *testing.T) {
	addr, err := fixRtspUrl("rtsp://20.212.158.240:1234/file.mp4", "admin", "pass")
	assert.Nil(t, err)
	assert.Equal(t, "rtsp://admin:pass@20.212.158.240:1234/file.mp4", addr)
}

func TestID(t *testing.T) {
	provider := RTSPProbeProvider{}
	err := provider.Init(RTSPProbeProviderConfig{
		Name: "test",
	})
	assert.Nil(t, err)

	id := provider.ID()
	assert.Equal(t, "test", id)
}

func TestContext(t *testing.T) {
	provider := RTSPProbeProvider{}
	err := provider.Init(RTSPProbeProviderConfig{
		Name: "test",
	})
	assert.Nil(t, err)

	provider.SetContext(&contexts.ManagerContext{
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "test",
		},
	})
	assert.Equal(t, "test", provider.Context.SiteInfo.SiteId)
}
