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

package managers

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/probe"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/reference"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/reporter"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/stack"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/uploader"
)

type ProviderConfig struct {
	Type   string                    `json:"type"`
	Config providers.IProviderConfig `json:"config"`
}

type ManagerConfig struct {
	Name       string                    `json:"name"`
	Type       string                    `json:"type"`
	Properties map[string]string         `json:"properties"`
	Providers  map[string]ProviderConfig `json:"providers"`
}

type IManager interface {
	Init(context *contexts.VendorContext, config ManagerConfig, providers map[string]providers.IProvider) error
}

type ISchedulable interface {
	Poll() []error
	Reconcil() []error
	Enabled() bool
}

type IEntityManager interface {
	Init(context *contexts.VendorContext, config ManagerConfig, providers map[string]providers.IProvider) error
}

type IManagerFactroy interface {
	CreateManager(config ManagerConfig) (IManager, error)
}

type Manager struct {
	VendorContext *contexts.VendorContext
	Context       *contexts.ManagerContext
	Config        ManagerConfig
}

func (m *Manager) Init(context *contexts.VendorContext, config ManagerConfig, providers map[string]providers.IProvider) error {
	m.VendorContext = context
	m.Context = &contexts.ManagerContext{}
	m.Config = config
	err := m.Context.Init(m.VendorContext, nil)
	m.Context.Logger.Debugf(" M (%s): initalize manager type '%s'", config.Name, config.Type)
	return err
}
func GetStackProvider(config ManagerConfig, providers map[string]providers.IProvider) (stack.IStackProvider, error) {
	stackProviderName, ok := config.Properties[v1alpha2.ProvidersState]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "stack provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[stackProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "stack provider is not supplied", v1alpha2.MissingConfig)
	}
	stackProvider, ok := provider.(stack.IStackProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a stack provider", v1alpha2.BadConfig)
	}
	return stackProvider, nil
}
func GetStateProvider(config ManagerConfig, providers map[string]providers.IProvider) (states.IStateProvider, error) {
	stateProviderName, ok := config.Properties[v1alpha2.ProvidersState]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "state provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[stateProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "state provider is not supplied", v1alpha2.MissingConfig)
	}
	stateProvider, ok := provider.(states.IStateProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a state provider", v1alpha2.BadConfig)
	}
	return stateProvider, nil
}
func GetConfigProvider(con ManagerConfig, providers map[string]providers.IProvider) (config.IConfigProvider, error) {
	configProviderName, ok := con.Properties[v1alpha2.ProvidersConfig]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "config provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[configProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "config provider is not supplied", v1alpha2.MissingConfig)
	}
	configProvider, ok := provider.(config.IConfigProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a config provider", v1alpha2.BadConfig)
	}
	return configProvider, nil
}
func GetSecretProvider(con ManagerConfig, providers map[string]providers.IProvider) (secret.ISecretProvider, error) {
	configProviderName, ok := con.Properties[v1alpha2.ProvidersSecret]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "secret provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[configProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "secret provider is not supplied", v1alpha2.MissingConfig)
	}
	secretProvider, ok := provider.(secret.ISecretProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a secret provider", v1alpha2.BadConfig)
	}
	return secretProvider, nil
}
func GetProbeProvider(config ManagerConfig, providers map[string]providers.IProvider) (probe.IProbeProvider, error) {
	probeProviderName, ok := config.Properties[v1alpha2.ProvidersProbe]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "probe provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[probeProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "probe provider is not supplied", v1alpha2.MissingConfig)
	}
	probeProvider, ok := provider.(probe.IProbeProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a probe provider", v1alpha2.BadConfig)
	}
	return probeProvider, nil
}
func GetUploaderProvider(config ManagerConfig, providers map[string]providers.IProvider) (uploader.IUploader, error) {
	uploaderProviderName, ok := config.Properties[v1alpha2.ProvidersUploader]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "uploader provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[uploaderProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "uploader provider is not supplied", v1alpha2.MissingConfig)
	}
	uploaderProvider, ok := provider.(uploader.IUploader)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a uploader provider", v1alpha2.BadConfig)
	}
	return uploaderProvider, nil
}
func GetReferenceProvider(config ManagerConfig, providers map[string]providers.IProvider) (reference.IReferenceProvider, error) {
	referenceProviderName, ok := config.Properties[v1alpha2.ProvidersReference]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "uploader provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[referenceProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "uploader provider is not supplied", v1alpha2.MissingConfig)
	}
	referenceProvider, ok := provider.(reference.IReferenceProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a uploader provider", v1alpha2.BadConfig)
	}
	return referenceProvider, nil
}
func GetReporter(config ManagerConfig, providers map[string]providers.IProvider) (reporter.IReporter, error) {
	reporterName, ok := config.Properties[v1alpha2.ProvidersReporter]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "reporter provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[reporterName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "reporter provider is not supplied", v1alpha2.MissingConfig)
	}
	reporterProvider, ok := provider.(reporter.IReporter)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a reporter provider", v1alpha2.BadConfig)
	}
	return reporterProvider, nil
}
