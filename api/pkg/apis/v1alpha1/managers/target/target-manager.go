/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package target

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/probe"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/uploader"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")
var lock sync.Mutex

type TargetManager struct {
	managers.Manager
	ReferenceProvider reference.IReferenceProvider
	ProbeProvider     probe.IProbeProvider
	UploaderProvider  uploader.IUploader
	Reporter          reporter.IReporter
}

type Device struct {
	Object Object
}
type Object struct {
	ApiVersion string                 `json:"apiVersion`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       DeviceSpec             `json:"spec"`
}
type DeviceSpec struct {
	Properties map[string]string `json:"properties"`
}

func (s *TargetManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		log.Errorf(" M (Target): failed to initialize manager %+v", err)
		return err
	}

	probeProvider, err := managers.GetProbeProvider(config, providers)
	if err == nil {
		s.ProbeProvider = probeProvider
	} else {
		log.Errorf(" M (Target): failed to get probe provider %+v", err)
		return err
	}

	referenceProvider, err := managers.GetReferenceProvider(config, providers)
	if err == nil {
		s.ReferenceProvider = referenceProvider
	} else {
		log.Errorf(" M (Target): failed to get reference provider %+v", err)
		return err
	}

	uploaderProvider, err := managers.GetUploaderProvider(config, providers)
	if err == nil {
		s.UploaderProvider = uploaderProvider
	} else {
		log.Errorf(" M (Target): failed to get upload provider %+v", err)
		return err
	}

	reporterProvider, err := managers.GetReporter(config, providers)
	if err == nil {
		s.Reporter = reporterProvider
	} else {
		log.Errorf(" M (Target): failed to get reporter %+v", err)
		return err
	}

	return nil
}

func (s *TargetManager) Apply(ctx context.Context, target model.TargetSpec) error {
	return nil
}
func (s *TargetManager) Get(ctx context.Context) (model.TargetSpec, error) {
	return model.TargetSpec{}, nil
}
func (s *TargetManager) Remove(ctx context.Context, target model.TargetSpec) error {
	return nil
}
func (s *TargetManager) Enabled() bool {
	return s.Config.Properties["poll.enabled"] == "true"
}
func (s *TargetManager) Poll() []error {
	target := s.ReferenceProvider.TargetID()
	log.Infof(" M (Target): Poll target- %s", target)

	ret, err := s.ReferenceProvider.List(target+"=true", "", "", model.FabricGroup, "devices", "v1", "v1alpha2.ReferenceK8sCRD")
	if err != nil {
		return []error{err}
	}
	jsonData, _ := json.Marshal(ret)
	devices := make([]Device, 0)
	json.Unmarshal(jsonData, &devices)
	log.Debugf("polling %d devices...", len(devices))
	errors := make([]error, 0)

	first := true
	for _, device := range devices {
		user := ""
		if u, ok := device.Object.Spec.Properties["user"]; ok {
			user = u
		}
		password := ""
		if p, ok := device.Object.Spec.Properties["password"]; ok {
			password = p
		}
		ip := ""
		if i, ok := device.Object.Spec.Properties["ip"]; ok {
			ip = i
		}
		name := device.Object.Metadata["name"].(string)
		namespace, ok := device.Object.Metadata["namespace"].(string)
		if !ok {
			namespace = "default"
		}
		if ip != "" {
			if user != "" && password != "" {
				log.Debugf("taking snapshot from rtsp://%s:%s@%s...", user, "<password>", strings.ReplaceAll(ip, "rtsp://", ""))
			} else {
				log.Debugf("taking snapshot from rtsp://%s...", strings.ReplaceAll(ip, "rtsp://", ""))
			}
			ret, err := s.ProbeProvider.Probe(user, password, ip, name)
			if err != nil {
				log.Debugf("failed to probe device: %s", err.Error())
				errors = append(errors, err)
				errors = append(errors, s.reportStatus(name, namespace, target, "", "disconnected", "disconnected", first, err.Error())...)
				continue
			}
			if v, ok := ret["snapshot"]; ok {
				file, err := os.Open(v)
				if err != nil {
					log.Debugf("failed to open local file: %s", err.Error())
					errors = append(errors, err)
					errors = append(errors, s.reportStatus(name, namespace, target, "", "connected", "connected", first, err.Error())...)
					continue
				}
				data, err := ioutil.ReadAll(file)
				if err != nil {
					log.Debugf("failed to read local file: %s", err.Error())
					errors = append(errors, err)
					errors = append(errors, s.reportStatus(name, namespace, target, "", "connected", "connected", first, err.Error())...)
					continue
				}
				fileName := filepath.Base(v)
				str, err := s.UploaderProvider.Upload(fileName, data)
				if err != nil {
					log.Debugf("failed to upload snapshot: %s", err.Error())
					errors = append(errors, err)
					errors = append(errors, s.reportStatus(name, namespace, target, "", "connected", "connected", first, err.Error())...)
					continue
				}
				log.Debugf("file is uploaded to %s", str)
				errors = append(errors, s.reportStatus(name, namespace, target, str, "connected", "connected", first, "")...)
			}
		} else {
			errors = append(errors, s.reportStatus(name, namespace, target, "", "disconnected", "disconnected", first, "device ip is not set")...)
		}
		first = false
	}
	for _, err := range errors {
		log.Errorf(" M (Target): polling error: %s", err.Error())
	}
	return errors
}
func (s *TargetManager) reportStatus(deviceName string, namespace string, targetName string, snapshot string, targetStatus string, deviceStatus string, overwrite bool, errStr string) []error {
	log.Infof(" M (Target): reportStatus deviceName- %s, targetName - %s, snapshot -%s targetStatus -%s, deviceStatus -%s, overwrite -%s", deviceName, targetName, snapshot, targetStatus, deviceStatus, overwrite)

	ret := make([]error, 0)
	report := make(map[string]string)
	report[targetName+".status"] = targetStatus
	if snapshot != "" {
		report["snapshot"] = snapshot
	}
	if errStr != "" {
		report[targetName+".err"] = errStr
	}
	err := s.Reporter.Report(deviceName, namespace, model.FabricGroup, "devices", "v1", report, false) //can't overwrite device state properties as other targets may be reporting as well
	if err != nil {
		log.Debugf("failed to report device status: %s", err.Error())
		ret = append(ret, err)
	}
	report = make(map[string]string)
	report[deviceName+".status"] = deviceStatus
	if errStr != "" {
		report[deviceName+".err"] = errStr
	}
	err = s.Reporter.Report(targetName, namespace, model.FabricGroup, "targets", "v1", report, overwrite)
	if err != nil {
		log.Debugf("failed to report target status: %s", err.Error())
		ret = append(ret, err)
	}

	return ret
}
func (s *TargetManager) Reconcil() []error {
	log.Debug("Rconciling....")
	return nil
}
