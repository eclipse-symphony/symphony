/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rtsp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
)

type RTSPProbeProviderConfig struct {
	Name string `json:"name"`
}

func RTSPProbeProviderConfigFromMap(properties map[string]string) (RTSPProbeProviderConfig, error) {
	ret := RTSPProbeProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	return ret, nil
}

type RTSPProbeProvider struct {
	Config  RTSPProbeProviderConfig
	Context *contexts.ManagerContext
}

func (i *RTSPProbeProvider) InitWithMap(properties map[string]string) error {
	config, err := RTSPProbeProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (m *RTSPProbeProvider) ID() string {
	return m.Config.Name
}

func (a *RTSPProbeProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context
}

func (m *RTSPProbeProvider) Init(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toRTSPProbProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid RTSP probe provider config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	return nil
}

func toRTSPProbProviderConfig(config providers.IProviderConfig) (RTSPProbeProviderConfig, error) {
	ret := RTSPProbeProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	return ret, err
}
func (m *RTSPProbeProvider) Probe(user string, password string, ip string, name string) (map[string]string, error) {
	address, err := fixRtspUrl(ip, user, password)
	if err != nil {
		return nil, err
	}
	snapshotPath := "./"
	target := os.Getenv("SNAPSHOT_ROOT")
	if target != "" {
		snapshotPath = target
		if _, err := os.Stat(snapshotPath); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(snapshotPath, 0777)
			if err != nil {
				snapshotPath = "./"
			}
		}
	}

	snapshotPath = path.Join(snapshotPath, name+"-snapshot.jpg")

	cmd := exec.Command("ffmpeg", "-rtsp_transport", "tcp", "-i", address, "-f", "image2", "-frames", "1", snapshotPath, "-y")

	path := ""
	_, err = cmd.CombinedOutput()
	if err == nil {
		path, err = filepath.Abs(snapshotPath)
		if err != nil {
			return nil, err
		}
	}
	if path != "" {
		return map[string]string{
			"snapshot": path,
		}, nil
	} else {
		return make(map[string]string), nil
	}
}
func fixRtspUrl(address string, user string, password string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	scheme := u.Scheme
	userName := ""
	userPassword := ""
	if u.User != nil {
		userName = u.User.Username()
		userPassword, _ = u.User.Password()
	}
	if user != "" {
		userName = user
	}
	if password != "" {
		userPassword = password
	}
	if scheme != "rtsp" {
		scheme = "rtsp"
	}
	hostName := u.Hostname()
	if hostName == "" {
		hostName = address
	}
	port := u.Port()
	if port == "" || port == "554" {
		port = ""
	} else {
		port = ":" + port
	}
	uri := u.RequestURI()
	if uri == "/" {
		uri = ""
	}
	if userName != "" && userPassword != "" {
		return fmt.Sprintf("%s://%s:%s@%s%s%s", scheme, userName, userPassword, hostName, port, uri), nil
	} else {
		return fmt.Sprintf("%s://%s%s%s", scheme, hostName, port, uri), nil
	}
}
