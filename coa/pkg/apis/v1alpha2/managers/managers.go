/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package managers

import (
	"context"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/ledger"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/probe"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/queue"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/uploader"
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
	v1alpha2.Terminable
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
	for _, p := range providers {
		if c, ok := p.(contexts.IWithManagerContext); ok {
			c.SetContext(m.Context)
		}
	}
	m.Context.Logger.Debugf(" M (%s): initalize manager type '%s'", config.Name, config.Type)
	return err
}
func GetQueueProvider(config ManagerConfig, providers map[string]providers.IProvider) (queue.IQueueProvider, error) {
	queueProviderName, ok := config.Properties[v1alpha2.ProviderQueue]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "queue provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[queueProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "queue provider is not supplied", v1alpha2.MissingConfig)
	}
	queueProvider, ok := provider.(queue.IQueueProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a queue provider", v1alpha2.BadConfig)
	}
	return queueProvider, nil
}

func (m *Manager) Shutdown(ctx context.Context) error {
	return nil
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
func GetLedgerProvider(config ManagerConfig, providers map[string]providers.IProvider) (ledger.ILedgerProvider, error) {
	ledgerProviderName, ok := config.Properties[v1alpha2.ProviderLedger]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "ledger provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[ledgerProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "ledger provider is not supplied", v1alpha2.MissingConfig)
	}
	ledgerProvider, ok := provider.(ledger.ILedgerProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a ledger provider", v1alpha2.BadConfig)
	}
	return ledgerProvider, nil
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
func GetExtConfigProvider(con ManagerConfig, providers map[string]providers.IProvider) (config.IExtConfigProvider, error) {
	configProviderName, ok := con.Properties[v1alpha2.ProvidersConfig]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "config provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[configProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "config provider is not supplied", v1alpha2.MissingConfig)
	}
	configProvider, ok := provider.(config.IExtConfigProvider)
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
		return nil, v1alpha2.NewCOAError(nil, "reference provider is not configured", v1alpha2.MissingConfig)
	}
	provider, ok := providers[referenceProviderName]
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "reference provider is not supplied", v1alpha2.MissingConfig)
	}
	referenceProvider, ok := provider.(reference.IReferenceProvider)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "supplied provider is not a reference provider", v1alpha2.BadConfig)
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
