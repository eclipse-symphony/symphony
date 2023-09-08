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

package localfile

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

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
func (s *LocalCertFileProvider) SetContext(ctx contexts.ManagerContext) error {
	return v1alpha2.NewCOAError(nil, "Local cert file provider doesn't support manager context", v1alpha2.InternalError)
}
func (w *LocalCertFileProvider) Init(config providers.IProviderConfig) error {
	certConfig, err := toLocalCertFileProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid local cert file provider config", v1alpha2.InvalidArgument)
	}
	w.Config = certConfig
	return nil
}

func toLocalCertFileProviderConfig(config providers.IProviderConfig) (LocalCertFileProviderConfig, error) {
	ret := LocalCertFileProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (w *LocalCertFileProvider) GetCert(host string) ([]byte, []byte, error) {
	certFile, err := os.Open(w.Config.CertFile)
	if err != nil {
		return nil, nil, v1alpha2.NewCOAError(err, "failed to read certificate file", v1alpha2.InternalError)
	}
	certData, err := ioutil.ReadAll(certFile)
	if err != nil {
		return nil, nil, v1alpha2.NewCOAError(err, "failed to read certificate file", v1alpha2.InternalError)
	}
	keyFile, err := os.Open(w.Config.KeyFile)
	if err != nil {
		return nil, nil, v1alpha2.NewCOAError(err, "failed to read key file", v1alpha2.InternalError)
	}
	keyData, err := ioutil.ReadAll(keyFile)
	if err != nil {
		return nil, nil, v1alpha2.NewCOAError(err, "failed to read key file", v1alpha2.InternalError)
	}
	return certData, keyData, nil
}
