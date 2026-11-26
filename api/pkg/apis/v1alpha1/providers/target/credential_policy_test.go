/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package target_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/configmap"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/helm"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/ingress"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/k8s"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/kubectl"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestTargetCredentialPluginPolicy(t *testing.T) {
	listSecrets := func(client kubernetes.Interface) error {
		_, err := client.CoreV1().Secrets("default").List(context.Background(), metav1.ListOptions{})
		return err
	}
	providers := []struct {
		name  string
		modes []string
		run   func(map[string]string) error
	}{
		{"kubectl", []string{"inline", "path"}, func(config map[string]string) error {
			provider := &kubectl.KubectlTargetProvider{}
			if err := provider.InitWithMap(config); err != nil {
				return err
			}
			return listSecrets(provider.Client)
		}},
		{"k8s", []string{"bytes", "path"}, func(config map[string]string) error {
			provider := &k8s.K8sTargetProvider{}
			if err := provider.InitWithMap(config); err != nil {
				return err
			}
			return listSecrets(provider.Client)
		}},
		{"configmap", []string{"inline", "path"}, func(config map[string]string) error {
			provider := &configmap.ConfigMapTargetProvider{}
			if err := provider.InitWithMap(config); err != nil {
				return err
			}
			return listSecrets(provider.Client)
		}},
		{"ingress", []string{"inline", "path"}, func(config map[string]string) error {
			provider := &ingress.IngressTargetProvider{}
			if err := provider.InitWithMap(config); err != nil {
				return err
			}
			return listSecrets(provider.Client)
		}},
		{"helm", []string{"bytes", "environment"}, func(config map[string]string) error {
			provider := &helm.HelmTargetProvider{}
			if err := provider.InitWithMap(config); err != nil {
				return err
			}
			_, err := provider.Get(context.Background(), model.DeploymentSpec{
				Instance: model.InstanceState{Spec: &model.InstanceSpec{Scope: "default"}},
			}, nil)
			return err
		}},
	}
	for _, provider := range providers {
		for _, mode := range provider.modes {
			for _, credential := range []string{"exec", "tokenFile", "auth-provider", "client-certificate", "client-key", "certificate-authority", "token"} {
				t.Run(provider.name+"/"+mode+"/"+credential, func(t *testing.T) {
					var requests atomic.Int32
					server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests.Add(1)
						if r.Header.Get("Authorization") != "Bearer test-token" {
							t.Errorf("unexpected authorization header: %q", r.Header.Get("Authorization"))
						}
						w.Header().Set("Content-Type", "application/json")
						if r.URL.Path == "/version" {
							fmt.Fprint(w, `{"major":"1","minor":"35","gitVersion":"v1.35.0"}`)
						} else {
							fmt.Fprint(w, `{"apiVersion":"v1","kind":"SecretList","items":[]}`)
						}
					}))
					defer server.Close()

					marker := filepath.Join(t.TempDir(), "plugin-executed")
					auth := &clientcmdapi.AuthInfo{Token: "test-token"}
					cluster := &clientcmdapi.Cluster{Server: server.URL, InsecureSkipTLSVerify: true}
					forbiddenFile := filepath.Join(t.TempDir(), "must-not-be-read")
					switch credential {
					case "exec":
						command, err := os.Executable()
						require.NoError(t, err)
						auth = &clientcmdapi.AuthInfo{Exec: &clientcmdapi.ExecConfig{
							Command:         command,
							Args:            []string{"-test.run=^TestCredentialPluginProcess$"},
							APIVersion:      "client.authentication.k8s.io/v1",
							InteractiveMode: clientcmdapi.NeverExecInteractiveMode,
							Env: []clientcmdapi.ExecEnvVar{{
								Name: "SYMPHONY_TEST_CREDENTIAL_PLUGIN_MARKER", Value: marker,
							}},
						}}
					case "tokenFile":
						auth = &clientcmdapi.AuthInfo{TokenFile: forbiddenFile}
					case "auth-provider":
						auth = &clientcmdapi.AuthInfo{AuthProvider: &clientcmdapi.AuthProviderConfig{Name: "oidc"}}
					case "client-certificate":
						auth = &clientcmdapi.AuthInfo{ClientCertificate: forbiddenFile}
					case "client-key":
						auth = &clientcmdapi.AuthInfo{ClientKey: forbiddenFile}
					case "certificate-authority":
						cluster.CertificateAuthority = forbiddenFile
					}
					data, err := clientcmd.Write(clientcmdapi.Config{
						Kind: "Config", APIVersion: "v1",
						Clusters:       map[string]*clientcmdapi.Cluster{"test": cluster},
						AuthInfos:      map[string]*clientcmdapi.AuthInfo{"test": auth},
						Contexts:       map[string]*clientcmdapi.Context{"test": {Cluster: "test", AuthInfo: "test"}},
						CurrentContext: "test",
					})
					require.NoError(t, err)
					config := map[string]string{"configType": mode, "configData": string(data)}
					if provider.name == "helm" && mode == "bytes" && credential == "token" {
						ambient, err := clientcmd.Write(clientcmdapi.Config{
							Clusters:       map[string]*clientcmdapi.Cluster{"ambient": cluster},
							AuthInfos:      map[string]*clientcmdapi.AuthInfo{"ambient": {TokenFile: forbiddenFile}},
							Contexts:       map[string]*clientcmdapi.Context{"ambient": {Cluster: "ambient", AuthInfo: "ambient"}},
							CurrentContext: "ambient",
						})
						require.NoError(t, err)
						path := filepath.Join(t.TempDir(), "ambient-config")
						require.NoError(t, os.WriteFile(path, ambient, 0600))
						t.Setenv("KUBECONFIG", path)
					}
					if mode == "path" || mode == "environment" {
						path := filepath.Join(t.TempDir(), "config")
						require.NoError(t, os.WriteFile(path, data, 0600))
						config["configData"] = path
						if mode == "environment" {
							t.Setenv("KUBECONFIG", path)
							config["inCluster"] = "true"
						}
					}
					err = provider.run(config)
					if credential != "token" {
						require.ErrorContains(t, err, fmt.Sprintf("field %q is not allowed", credential))
						require.Zero(t, requests.Load())
					} else {
						require.NoError(t, err)
						require.Positive(t, requests.Load())
					}
					_, err = os.Stat(marker)
					require.True(t, os.IsNotExist(err), "credential plugin must not execute")
				})
			}
		}
	}
}

// The test binary itself acts as a harmless credential plugin if the policy regresses.
func TestCredentialPluginProcess(t *testing.T) {
	marker := os.Getenv("SYMPHONY_TEST_CREDENTIAL_PLUGIN_MARKER")
	if marker == "" {
		return
	}
	if err := os.WriteFile(marker, []byte("executed"), 0600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(`{"apiVersion":"client.authentication.k8s.io/v1","kind":"ExecCredential","status":{"token":"test-token"}}`)
	os.Exit(0)
}
