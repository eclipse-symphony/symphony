/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package rtsp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := RTSPProbeProvider{}
	err := provider.Init(RTSPProbeProviderConfig{
		Name: "test",
	})
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
