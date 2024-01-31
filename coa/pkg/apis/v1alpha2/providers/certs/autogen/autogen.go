/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package autogen

import (
	"encoding/json"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	fasthttp "github.com/valyala/fasthttp"
)

var log = logger.NewLogger("coa.runtime")

type AutoGenCertProviderConfig struct {
	Name string `json:"name"`
}

type AutoGenCertProvider struct {
	Config AutoGenCertProviderConfig
}

func (w *AutoGenCertProvider) ID() string {
	return w.Config.Name
}
func (s *AutoGenCertProvider) SetContext(ctx contexts.ManagerContext) {
}
func (w *AutoGenCertProvider) Init(config providers.IProviderConfig) error {
	certConfig, err := toAutoGenCertProviderConfig(config)
	if err != nil {
		log.Errorf("  P (Autogen): failed to parse provider config %+v", err)
		return v1alpha2.NewCOAError(nil, "provided config is not a valid cert generation provider config", v1alpha2.InvalidArgument)
	}
	w.Config = certConfig
	return nil
}

func toAutoGenCertProviderConfig(config providers.IProviderConfig) (AutoGenCertProviderConfig, error) {
	ret := AutoGenCertProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (w *AutoGenCertProvider) GetCert(host string) ([]byte, []byte, error) {
	return fasthttp.GenerateTestCertificate(host)
}
