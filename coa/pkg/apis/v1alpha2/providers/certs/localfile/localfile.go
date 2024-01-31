/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package localfile

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type LocalCertFileProviderConfig struct {
	Name     string `json:"name"`
	CertFile string `json:"cert"`
	KeyFile  string `json:"key"`
}

type LocalCertFileProvider struct {
	Config LocalCertFileProviderConfig
}

func (w *LocalCertFileProvider) ID() string {
	return w.Config.Name
}
func (s *LocalCertFileProvider) SetContext(ctx contexts.ManagerContext) {
}
func (w *LocalCertFileProvider) Init(config providers.IProviderConfig) error {
	certConfig, err := toLocalCertFileProviderConfig(config)
	if err != nil {
		log.Errorf("  P (Localfile): expect LocalCertFileProviderConfig %+v", err)
		return v1alpha2.NewCOAError(nil, "provided config is not a valid local cert file provider config", v1alpha2.InvalidArgument)
	}
	w.Config = certConfig
	return nil
}

func toLocalCertFileProviderConfig(config providers.IProviderConfig) (LocalCertFileProviderConfig, error) {
	ret := LocalCertFileProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		log.Errorf("  P (Localfile): failed to marshall config %+v", err)
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (w *LocalCertFileProvider) GetCert(host string) ([]byte, []byte, error) {
	certFile, err := os.Open(w.Config.CertFile)
	if err != nil {
		log.Errorf("  P (Localfile): failed to open certificate file %+v", err)
		return nil, nil, v1alpha2.NewCOAError(err, "failed to read certificate file", v1alpha2.InternalError)
	}
	certData, err := ioutil.ReadAll(certFile)
	if err != nil {
		log.Errorf("  P (Localfile): failed to read certificate file %+v", err)
		return nil, nil, v1alpha2.NewCOAError(err, "failed to read certificate file", v1alpha2.InternalError)
	}
	keyFile, err := os.Open(w.Config.KeyFile)
	if err != nil {
		log.Errorf("  P (Localfile): failed to open key file %+v", err)
		return nil, nil, v1alpha2.NewCOAError(err, "failed to read key file", v1alpha2.InternalError)
	}
	keyData, err := ioutil.ReadAll(keyFile)
	if err != nil {
		log.Errorf("  P (Localfile): failed to read key file %+v", err)
		return nil, nil, v1alpha2.NewCOAError(err, "failed to read key file", v1alpha2.InternalError)
	}
	return certData, keyData, nil
}
