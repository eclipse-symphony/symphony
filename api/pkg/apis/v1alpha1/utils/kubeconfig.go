/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"fmt"
	"sort"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ValidateKubeconfig checks every entry before client-go can read host credentials.
func ValidateKubeconfig(config clientcmdapi.Config) error {
	users := make([]string, 0, len(config.AuthInfos))
	for name := range config.AuthInfos {
		users = append(users, name)
	}
	sort.Strings(users)
	for _, name := range users {
		auth := config.AuthInfos[name]
		if auth == nil {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("kubeconfig user %q is empty", name), v1alpha2.BadConfig)
		}
		field := ""
		switch {
		case auth.Exec != nil:
			field = "exec"
		case auth.TokenFile != "":
			field = "tokenFile"
		case auth.AuthProvider != nil:
			field = "auth-provider"
		case auth.ClientCertificate != "":
			field = "client-certificate"
		case auth.ClientKey != "":
			field = "client-key"
		}
		if field != "" {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("kubeconfig user %q: field %q is not allowed", name, field), v1alpha2.BadConfig)
		}
	}
	clusters := make([]string, 0, len(config.Clusters))
	for name := range config.Clusters {
		clusters = append(clusters, name)
	}
	sort.Strings(clusters)
	for _, name := range clusters {
		cluster := config.Clusters[name]
		if cluster == nil {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("kubeconfig cluster %q is empty", name), v1alpha2.BadConfig)
		}
		if cluster.CertificateAuthority != "" {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("kubeconfig cluster %q: field %q is not allowed", name, "certificate-authority"), v1alpha2.BadConfig)
		}
	}
	return nil
}

func RESTConfigFromKubeConfig(data []byte) (*rest.Config, error) {
	loader, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, v1alpha2.NewCOAError(err, "failed to parse kubeconfig", v1alpha2.BadConfig)
	}
	return RESTConfigFromClientConfig(loader)
}

func RESTConfigFromKubeConfigFile(path string) (*rest.Config, error) {
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: path},
		&clientcmd.ConfigOverrides{},
	)
	return RESTConfigFromClientConfig(loader)
}

func RESTConfigFromClientConfig(loader clientcmd.ClientConfig) (*rest.Config, error) {
	raw, err := loader.RawConfig()
	if err != nil {
		return nil, v1alpha2.NewCOAError(err, "failed to load kubeconfig", v1alpha2.BadConfig)
	}
	if err := ValidateKubeconfig(raw); err != nil {
		return nil, err
	}
	config, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}
	return DenyExecCredentialPlugins(config), nil
}

// DenyExecCredentialPlugins prevents kubeconfigs from executing programs in the API process.
func DenyExecCredentialPlugins(config *rest.Config) *rest.Config {
	if config.ExecProvider != nil {
		config.ExecProvider.PluginPolicy = clientcmdapi.PluginPolicy{
			PolicyType: clientcmdapi.PluginPolicyDenyAll,
		}
	}
	return config
}
