/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestKubeconfigRejectsHostCredentialSources(t *testing.T) {
	cases := []struct {
		field  string
		mutate func(*clientcmdapi.AuthInfo, *clientcmdapi.Cluster)
	}{
		{"exec", func(auth *clientcmdapi.AuthInfo, _ *clientcmdapi.Cluster) {
			auth.Exec = &clientcmdapi.ExecConfig{}
		}},
		{"tokenFile", func(auth *clientcmdapi.AuthInfo, _ *clientcmdapi.Cluster) {
			auth.TokenFile = "must-not-be-read"
		}},
		{"auth-provider", func(auth *clientcmdapi.AuthInfo, _ *clientcmdapi.Cluster) {
			auth.AuthProvider = &clientcmdapi.AuthProviderConfig{Name: "oidc"}
		}},
		{"client-certificate", func(auth *clientcmdapi.AuthInfo, _ *clientcmdapi.Cluster) {
			auth.ClientCertificate = "must-not-be-read"
			auth.ClientCertificateData = []byte("embedded")
		}},
		{"client-key", func(auth *clientcmdapi.AuthInfo, _ *clientcmdapi.Cluster) {
			auth.ClientKey = "must-not-be-read"
			auth.ClientKeyData = []byte("embedded")
		}},
		{"certificate-authority", func(_ *clientcmdapi.AuthInfo, cluster *clientcmdapi.Cluster) {
			cluster.CertificateAuthority = "must-not-be-read"
			cluster.CertificateAuthorityData = []byte("embedded")
		}},
	}
	for _, tc := range cases {
		for _, inactive := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/inactive=%t", tc.field, inactive), func(t *testing.T) {
				config := clientcmdapi.Config{
					Clusters:       map[string]*clientcmdapi.Cluster{"safe": {Server: "https://127.0.0.1"}},
					AuthInfos:      map[string]*clientcmdapi.AuthInfo{"safe": {Token: "token"}},
					Contexts:       map[string]*clientcmdapi.Context{"safe": {Cluster: "safe", AuthInfo: "safe"}},
					CurrentContext: "safe",
				}
				auth, cluster := config.AuthInfos["safe"], config.Clusters["safe"]
				if inactive {
					auth, cluster = &clientcmdapi.AuthInfo{}, &clientcmdapi.Cluster{}
					config.AuthInfos["unused"] = auth
					config.Clusters["unused"] = cluster
				}
				tc.mutate(auth, cluster)
				data, err := clientcmd.Write(config)
				require.NoError(t, err)
				path := filepath.Join(t.TempDir(), "config")
				require.NoError(t, os.WriteFile(path, data, 0600))
				for _, load := range []func() (*rest.Config, error){
					func() (*rest.Config, error) { return RESTConfigFromKubeConfig(data) },
					func() (*rest.Config, error) { return RESTConfigFromKubeConfigFile(path) },
				} {
					config, err := load()
					require.Nil(t, config)
					require.ErrorContains(t, err, fmt.Sprintf("field %q is not allowed", tc.field))
					require.Equal(t, v1alpha2.BadConfig, v1alpha2.GetErrorState(err))
					require.NotContains(t, err.Error(), "must-not-be-read")
				}
			})
		}
	}
}

func TestKubeconfigEmbeddedCredentials(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer server.Close()
	certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	key, err := x509.MarshalPKCS8PrivateKey(server.TLS.Certificates[0].PrivateKey)
	require.NoError(t, err)
	privateKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key})
	for name, auth := range map[string]*clientcmdapi.AuthInfo{
		"token":       {Token: "embedded-token"},
		"basic":       {Username: "user", Password: "password"},
		"certificate": {ClientCertificateData: certificate, ClientKeyData: privateKey},
	} {
		t.Run(name, func(t *testing.T) {
			data, err := clientcmd.Write(clientcmdapi.Config{
				Clusters:       map[string]*clientcmdapi.Cluster{"test": {Server: server.URL, CertificateAuthorityData: certificate}},
				AuthInfos:      map[string]*clientcmdapi.AuthInfo{"test": auth},
				Contexts:       map[string]*clientcmdapi.Context{"test": {Cluster: "test", AuthInfo: "test", Namespace: "example"}},
				CurrentContext: "test",
			})
			require.NoError(t, err)
			config, err := RESTConfigFromKubeConfig(data)
			require.NoError(t, err)
			require.Equal(t, certificate, config.CAData)
			require.Equal(t, auth.Token, config.BearerToken)
			require.Equal(t, auth.Username, config.Username)
			require.Equal(t, auth.Password, config.Password)
			require.Equal(t, auth.ClientCertificateData, config.CertData)
			require.Equal(t, auth.ClientKeyData, config.KeyData)
			client, err := rest.HTTPClientFor(config)
			require.NoError(t, err)
			response, err := client.Get(server.URL)
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, http.StatusOK, response.StatusCode)
		})
	}
}

func TestDenyExecCredentialPluginsDefenseInDepth(t *testing.T) {
	config := &rest.Config{
		Host: "https://127.0.0.1",
		ExecProvider: &clientcmdapi.ExecConfig{
			Command:         "symphony-test-credential-plugin",
			APIVersion:      "client.authentication.k8s.io/v1",
			InteractiveMode: clientcmdapi.NeverExecInteractiveMode,
			PluginPolicy: clientcmdapi.PluginPolicy{
				PolicyType: clientcmdapi.PluginPolicyAllowlist,
				Allowlist:  []clientcmdapi.AllowlistEntry{{Name: "symphony-test-credential-plugin"}},
			},
		},
	}
	client, err := rest.HTTPClientFor(DenyExecCredentialPlugins(config))
	require.NoError(t, err)
	_, err = client.Get(config.Host)
	require.ErrorContains(t, err, `policy set to "DenyAll"`)
}
