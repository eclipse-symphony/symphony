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

package autogen

import (
	"encoding/json"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	fasthttp "github.com/valyala/fasthttp"
)

type AutoGenCertProviderConfig struct {
	Name string `json:"name"`
}

type AutoGenCertProvider struct {
	Config AutoGenCertProviderConfig
}

func (w *AutoGenCertProvider) ID() string {
	return w.Config.Name
}
func (s *AutoGenCertProvider) SetContext(ctx contexts.ManagerContext) error {
	return v1alpha2.NewCOAError(nil, "Auto cert generation provider doesn't support manager context", v1alpha2.InternalError)
}
func (w *AutoGenCertProvider) Init(config providers.IProviderConfig) error {
	certConfig, err := toAutoGenCertProviderConfig(config)
	if err != nil {
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
