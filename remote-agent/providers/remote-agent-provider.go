/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package providers

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	loggerName = "providers.target.remote-agent"
	maxRetries = 3
	retryDelay = 5 * time.Second
	agentName  = "remote-agent"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
	state                    = "active"
)

type RemoteAgentProviderConfig struct {
	Version        string `json:"version,omitempty"`
	PublicCertPath string `json:"publicCertPath,omitempty"`
	PrivateKeyPath string `json:"privateKeyPath,omitempty"`
	BaseUrl        string `json:"baseUrl,omitempty"`
	ConfigPath     string `json:"configPath,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	TargetName     string `json:"targetName,omitempty"`
	TopologyPath   string `json:"topologyPath,omitempty"`
	ExecDir        string `json:"execDir,omitempty"`
}

type RemoteAgentProvider struct {
	Config RemoteAgentProviderConfig
	Client *http.Client
}

func RemoteAgentProviderConfigFromMap(properties map[string]string) (RemoteAgentProviderConfig, error) {
	ret := RemoteAgentProviderConfig{}
	if v, ok := properties["version"]; ok {
		ret.Version = v
	}
	if v, ok := properties["publicCertPath"]; ok {
		ret.PublicCertPath = v
	}
	if v, ok := properties["privateKeyPath"]; ok {
		ret.PrivateKeyPath = v
	}
	if v, ok := properties["baseUrl"]; ok {
		ret.BaseUrl = v
	}
	if v, ok := properties["configPath"]; ok {
		ret.ConfigPath = v
	}
	if v, ok := properties["namespace"]; ok {
		ret.Namespace = v
	}
	if v, ok := properties["targetName"]; ok {
		ret.TargetName = v
	}
	if v, ok := properties["topologyPath"]; ok {
		ret.TopologyPath = v
	}
	return ret, nil
}

func (i *RemoteAgentProvider) InitWithMap(properties map[string]string) error {
	config, err := RemoteAgentProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (Remote Agent Provider): expected ScriptProviderConfig: %+v", err)
		return err
	}
	return i.Init(config)
}

func (i *RemoteAgentProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("Remote Agent Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Script Target): Init()")

	updateConfig, err := toRemoteAgentConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): expected RemoteAgentProviderConfig - %+v", err)
		err = errors.New("expected RemoteAgentProviderConfig")
		return err
	}
	i.Config = updateConfig

	return err
}

func toRemoteAgentConfig(config providers.IProviderConfig) (RemoteAgentProviderConfig, error) {
	ret := RemoteAgentProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *RemoteAgentProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Remote Agent Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	//sLog.InfofCtx(ctx, "  P (Remote Agent Provider): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)
	notAfter, err := i.getCertificateExpirationOrThumbPrint(i.Config.PublicCertPath, "expiration")
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to get certificate expiration: %+v. Path is : %s", err, i.Config.PublicCertPath)
		return nil, err
	}
	for _, ref := range references {
		ref.Component.Properties = map[string]interface{}{
			"state":                 state,
			"version":               i.Config.Version,
			"lastConnected":         time.Now().UTC().Format(time.RFC3339),
			"certificateExpiration": notAfter,
		}
		ret = append(ret, ref.Component)
	}

	return ret, nil
}

func (i *RemoteAgentProvider) getCertificateExpirationOrThumbPrint(certPath string, kind string) (string, error) {
	certPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "", fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	if kind == "thumbprint" {
		thumbprint := sha1.Sum(cert.Raw)
		return hex.EncodeToString(thumbprint[:]), nil
	} else {
		return cert.NotAfter.Format(time.RFC3339), nil
	}
}

func (i *RemoteAgentProvider) composeComponentResultSpec(state v1alpha2.State, err error) model.ComponentResultSpec {
	if err == nil {
		return model.ComponentResultSpec{
			Status:  state,
			Message: "Succeeded",
		}
	} else {
		return model.ComponentResultSpec{
			Status:  state,
			Message: err.Error(),
		}
	}
}

func (i *RemoteAgentProvider) generateAgentStatus(ctx context.Context) (string, error) {
	notAfter, err := i.getCertificateExpirationOrThumbPrint(i.Config.PublicCertPath, "expiration")
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to get certificate expiration: %+v. Path is : %s", err, i.Config.PublicCertPath)
		return "", err
	}

	status := map[string]string{
		"state":                 state,
		"version":               i.Config.Version,
		"lastConnected":         time.Now().UTC().Format(time.RFC3339),
		"certificateExpiration": notAfter,
	}

	statusBytes, err := json.Marshal(status)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to marshal status: %+v", err)
		return "", err
	}
	return string(statusBytes), nil
}

func downloadFile(client *http.Client, url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (i *RemoteAgentProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Remote Agent Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Remote Agent Provider): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := map[string]model.ComponentResultSpec{}
	components := step.GetComponents()
	for _, c := range components {
		action, ok := c.Properties["action"].(string)
		if !ok {
			sLog.InfofCtx(ctx, "  P (Remote Agent Provider): There is no action. Report status back.")
			agentStatus, err := i.generateAgentStatus(ctx)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to generate agent status: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			ret[c.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.OK,
				Message: agentStatus,
			}
			continue
		}
		switch action {
		case "upgrade":
			// check the upgraded version
			version, ok := c.Properties["version"].(string)
			if !ok {
				err = fmt.Errorf("missing version parameter in component %s", c.Name)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			sLog.InfofCtx(ctx, "  P (Remote Agent Provider): The remote agent version is %s.\n Upgrading it to: %s.", i.Config.Version, version)

			if i.Config.Version == version {
				sLog.InfofCtx(ctx, "  P (Remote Agent Provider): The two versions are identical. No need to upgrade.")
				agentStatus, err := i.generateAgentStatus(ctx)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to generate agent status: %+v", err)
					ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
					continue
				}
				ret[c.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.OK,
					Message: agentStatus,
				}
				continue
			}
			// download the new agent binary
			newBinaryPath := filepath.Join(i.Config.ExecDir, "new-binary")
			err = downloadFile(i.Client, fmt.Sprintf("%s/files/%s", i.Config.BaseUrl, agentName), newBinaryPath)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to create temp file: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}

			// Replace the current binary with the new binary
			execPath, err := os.Executable()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to get executable path: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}

			err = os.Rename(newBinaryPath, execPath)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to replace binary: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}

			// Change the mode of the execPath to add execute permissions
			err = os.Chmod(execPath, 0755)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to replace binary: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}

			// Restart the process
			cmd := exec.Command(execPath, fmt.Sprintf("-config=%s", i.Config.ConfigPath),
				fmt.Sprintf("-client-key=%s", i.Config.PrivateKeyPath),
				fmt.Sprintf("-client-cert=%s", i.Config.PublicCertPath),
				fmt.Sprintf("-target-name=%s", i.Config.TargetName),
				fmt.Sprintf("-namespace=%s", i.Config.Namespace),
				fmt.Sprintf("-topology=%s", i.Config.TopologyPath),
			)
			pid, err := restartTheProcessWithRetry(cmd)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to restart process: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			} else {
				sLog.InfofCtx(ctx, "  P (Remote Agent Provider): restarted process with PID %d", pid)
				os.Exit(0)
			}
		case "secretrotation":
			// check if the target needs SR
			thumbprint, err := i.getCertificateExpirationOrThumbPrint(i.Config.PublicCertPath, "thumbprint")
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to get certificate thumbprint: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			sLog.InfofCtx(ctx, "  P (Remote Agent Provider): certificate thumbprint %s for %s from local.", c.Name, thumbprint)
			upstreamThumb, ok := c.Properties["thumbprint"].(string)
			if !ok {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): missing thumbprint parameter in component %s", c.Name)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			sLog.InfofCtx(ctx, "  P (Remote Agent Provider): certificate thumbprint %s for %s from upstream.", c.Name, upstreamThumb)
			if thumbprint == upstreamThumb {
				sLog.InfofCtx(ctx, "  P (Remote Agent Provider): The two certs are identical. No need to SR.")
				agentStatus, err := i.generateAgentStatus(ctx)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to generate agent status: %+v", err)
					ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
					continue
				}
				ret[c.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.OK,
					Message: agentStatus,
				}
				continue
			}

			// call the secret rotation api
			srUrl := fmt.Sprintf("%s/targets/secretrotate/%s?namespace=%s", i.Config.BaseUrl, step.Target, i.Config.Namespace)
			req, err := http.NewRequest("POST", srUrl, nil)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to create secret rotation request: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := i.Client.Do(req)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to call secret rotation", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to call secret rotation: %s", resp.Status)
				err = fmt.Errorf("failed to call secret rotation: %s", resp.Status)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			// parse resp body to get the new cert
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to read response body: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}

			var result map[string]string
			err = json.Unmarshal(body, &result)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to unmarshal response body: %+v. The body is %s", err, body)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			public, ok := result["public"]
			if !ok {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to get public cert from response body")
				err = fmt.Errorf("failed to get public cert from response body")
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			private, ok := result["private"]
			if !ok {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to get private cert from response body")
				err = fmt.Errorf("failed to get private cert from response body")
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}

			if public == "" || private == "" {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to get public or private cert from response body")
				err = fmt.Errorf("failed to get public or private cert from response body")
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}

			// write the new cert to the cert file
			err = ioutil.WriteFile(i.Config.PublicCertPath, []byte(formatPEM(public, "public")), 0644)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to write new cert to file: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}

			err = ioutil.WriteFile(i.Config.PrivateKeyPath, []byte(formatPEM(private, "private")), 0644)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to write new key to file: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			//update the http client with the new cert
			cert, err := tls.LoadX509KeyPair(i.Config.PublicCertPath, i.Config.PrivateKeyPath)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to create new cert: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			i.Client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			}
			agentStatus, err := i.generateAgentStatus(ctx)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Remote Agent Provider): failed to generate agent status: %+v", err)
				ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
				continue
			}
			ret[c.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.OK,
				Message: agentStatus,
			}
			continue
		case "log":
		default:
			err = fmt.Errorf("invalid action parameter in component %s", c.Name)
			ret[c.Name] = i.composeComponentResultSpec(v1alpha2.UpdateFailed, err)
		}
	}
	return ret, nil
}
func restartTheProcessWithRetry(cmd *exec.Cmd) (int, error) {
	//return 0, err if the process is not started
	// return pid, nil if the process is started
	var pid int
	var err error
	for i := 0; i < maxRetries; i++ {
		sLog.Infof("  P (Remote Agent Provider): running command %s", cmd.String())
		err = cmd.Start()
		if err == nil {
			pid = cmd.Process.Pid
			break
		}
		time.Sleep(retryDelay)
	}
	return pid, err
}

func (*RemoteAgentProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
		},
	}
}

func formatPEM(cert string, kind string) string {
	var pemHeader, pemFooter string
	if kind == "public" {
		pemHeader = "-----BEGIN CERTIFICATE-----"
		pemFooter = "-----END CERTIFICATE-----"
	}
	if kind == "private" {
		pemHeader = "-----BEGIN RSA PRIVATE KEY-----"
		pemFooter = "-----END RSA PRIVATE KEY-----"
	}

	// Remove any existing headers and footers
	cert = strings.Replace(cert, pemHeader, "", -1)
	cert = strings.Replace(cert, pemFooter, "", -1)
	// remove any space at the beginning or end of the cert
	cert = strings.TrimSpace(cert)
	// Encode the certificate with line breaks
	var buffer bytes.Buffer
	buffer.WriteString(pemHeader + "\n")
	parts := strings.Split(cert, " ")
	for i := 0; i < len(parts); i++ {
		buffer.WriteString(parts[i] + "\n")
	}
	buffer.WriteString(pemFooter)

	return buffer.String()
}
