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

package helm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"log"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var sLog = logger.NewLogger("coa.runtime")

type HelmTargetProviderConfig struct {
	Name       string `json:"name"`
	ConfigType string `json:"configType,omitempty"`
	ConfigData string `json:"configData,omitempty"`
	Context    string `json:"context,omitempty"`
	InCluster  bool   `json:"inCluster"`
}
type HelmTargetProvider struct {
	Config          HelmTargetProviderConfig
	Context         *contexts.ManagerContext
	ListClient      *action.List
	InstallClient   *action.Install
	UpgradeClient   *action.Upgrade
	UninstallClient *action.Uninstall
}

func HelmTargetProviderConfigFromMap(properties map[string]string) (HelmTargetProviderConfig, error) {
	ret := HelmTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["configType"]; ok {
		ret.ConfigType = v
	}
	if v, ok := properties["configData"]; ok {
		ret.ConfigData = v
	}
	if v, ok := properties["context"]; ok {
		ret.Context = v
	}
	if v, ok := properties["inCluster"]; ok {
		val := v
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'inCluster' setting of Helm provider", v1alpha2.BadConfig)
			}
			ret.InCluster = bVal
		}
	}
	return ret, nil
}
func (i *HelmTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := HelmTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func (i *HelmTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Helm Target Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("  P (Helm Target): Init()")

	// convert config to HelmTargetProviderConfig type
	helmConfig, err := toHelmTargetProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Helm Target): expected HelmTargetProviderConfig: %+v", err)
		return err
	}

	i.Config = helmConfig

	var actionConfig *action.Configuration
	if i.Config.InCluster {
		settings := cli.New()
		actionConfig = new(action.Configuration)
		// TODO: $HELM_DRIVER	set the backend storage driver. Values are: configmap, secret, memory, sql. Do we need to handle this differently?
		if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Helm Target): failed to init: %+v", err)
			return err
		}
	} else {
		switch i.Config.ConfigType {
		case "bytes":
			if i.Config.ConfigData != "" {
				kConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Helm Target): failed to init with config bytes: %+v", err)
					return err
				}
				namespace := "default"
				actionConfig, err = getActionConfig(context.Background(), namespace, kConfig)
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Helm Target): failed to init with config bytes: %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Helm Target): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted value is: bytes", v1alpha2.BadConfig)
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Helm Target): %+v", err)
			return err
		}
	}
	i.ListClient = action.NewList(actionConfig)
	i.InstallClient = action.NewInstall(actionConfig)
	i.UninstallClient = action.NewUninstall(actionConfig)
	i.UpgradeClient = action.NewUpgrade(actionConfig)
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func getActionConfig(ctx context.Context, namespace string, config *rest.Config) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	cliConfig := genericclioptions.NewConfigFlags(false)
	cliConfig.APIServer = &config.Host
	cliConfig.BearerToken = &config.BearerToken
	cliConfig.Namespace = &namespace
	// Drop their rest.Config and just return inject own
	wrapper := func(*rest.Config) *rest.Config {
		return config
	}
	cliConfig.WithWrapConfigFn(wrapper)
	if err := actionConfig.Init(cliConfig, namespace, "secret", log.Printf); err != nil {
		return nil, err
	}
	return actionConfig, nil
}
func toHelmTargetProviderConfig(config providers.IProviderConfig) (HelmTargetProviderConfig, error) {
	ret := HelmTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *HelmTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	for _, dc := range desired {
		found := false
		differ := false
		for _, cc := range current {
			if cc.Name == dc.Name {
				found = true
				if !model.HasSameProperty(dc.Properties, cc.Properties, "helm.chart.name") {
					differ = true
					break
				}
				// only compare version if desired version is supplied
				if dc.Properties["helm.chart.version"] != "" {
					if !model.HasSameProperty(dc.Properties, cc.Properties, "helm.chart.version") {
						differ = true
						break
					}
				}
				// We skip helm.chart.repo as it's not reliably reconstructable
			}
		}
		if !found || differ {
			return true
		}
	}
	return false
}
func (i *HelmTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	for _, dc := range desired {
		for _, cc := range current {
			if cc.Name == dc.Name {
				return true
			}
		}
	}
	return false
}

func (i *HelmTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Helm Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Infof("  P (Helm Target): getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	i.ListClient.Deployed = true
	results, err := i.ListClient.Run()
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Helm Target): failed to create Helm list client: %+v", err)
		return nil, err
	}

	desired := deployment.GetComponentSlice()

	ret := make([]model.ComponentSpec, 0)
	for _, component := range desired {
		for _, res := range results {
			if (deployment.Instance.Scope == "" || res.Namespace == deployment.Instance.Scope) && res.Name == component.Name {
				repo := ""
				if strings.HasPrefix(res.Chart.Metadata.Tags, "SYM:") { //we use this special metadata tag to remember the chart URL
					repo = res.Chart.Metadata.Tags[4:]
				}
				ret = append(ret, model.ComponentSpec{
					Name: res.Name,
					Type: "helm.v3",
					Properties: map[string]string{
						"helm.chart.version": res.Chart.Metadata.Version,
						"helm.chart.repo":    repo,
						"helm.chart.name":    res.Chart.Metadata.Name,
					},
				})
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}
func (i *HelmTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan("Helm Target Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	sLog.Infof("  P (Helm Target): deleting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := deployment.GetComponentSlice()
	for _, component := range components {
		if component.Type == "helm.v3" {
			_, err := i.UninstallClient.Run(component.Name)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Helm Target): failed to uninstall Helm chart: %+v", err)
				return err
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func downloadFile(url string, fileName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fileHandle, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fileHandle.Close()
	_, err = io.Copy(fileHandle, resp.Body)
	return err
}
func (i *HelmTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec) error {
	_, span := observability.StartSpan("Helm Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("  P (Helm Target): applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	injections := &model.ValueInjections{
		InstanceId: deployment.Instance.Name,
		SolutionId: deployment.Instance.Stages[0].Solution,
		TargetId:   deployment.Stages[0].ActiveTarget,
	}

	components := deployment.GetComponentSlice()

	for _, component := range components {
		if component.Type == "helm.v3" {
			regClient, err := registry.NewClient()
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Helm Target): failed to create registry client: %+v", err)
				return err
			}
			repo := model.ReadProperty(component.Properties, "helm.chart.repo", injections)
			version := model.ReadProperty(component.Properties, "helm.chart.version", injections)
			chartName := model.ReadProperty(component.Properties, "helm.chart.name", injections)

			if repo == "" {
				err = errors.New("component doesn't have helm.chart.repo property")
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Helm Target): component doesn't have helm.chart.repo property")
				return err
			}
			if chartName == "" {
				err = errors.New("component doesn't have helm.chart.name property")
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Helm Target): component doesn't have helm.chart.name property")
				return err
			}
			var fileName string
			if version == "" {
				fileName = fmt.Sprintf("%s.tgz", chartName)
			} else {
				fileName = fmt.Sprintf("%s-%s.tgz", chartName, version)
			}

			var pullRes *registry.PullResult
			if strings.HasSuffix(repo, ".tgz") && strings.HasPrefix(repo, "http") {
				err = downloadFile(repo, fileName)
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Helm Target): failed to download chart from repo: %+v", err)
					return err
				}
			} else {
				pullRes, err = regClient.Pull(fmt.Sprintf("%s:%s", repo, version), registry.PullOptWithChart(true))
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Helm Target): failed to pull chart from repo: %+v", err)
					return err
				}

				err = ioutil.WriteFile(fileName, pullRes.Chart.Data, 0644)
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Helm Target): failed to save chart: %+v", err)
					return err
				}
			}
			chart, err := loader.Load(fileName)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Helm Target): failed to load chart: %+v", err)
				return err
			}
			chart.Metadata.Tags = "SYM:" + repo //this is not used by Helm SDK, we use this to carry repo info
			if deployment.Instance.Scope == "" {
				i.InstallClient.Namespace = "default"
				i.UpgradeClient.Namespace = "default"
			} else {
				i.InstallClient.Namespace = deployment.Instance.Scope
				i.UpgradeClient.Namespace = deployment.Instance.Scope
			}
			i.InstallClient.ReleaseName = component.Name
			i.InstallClient.IsUpgrade = true

			//i.UpgradeClient.ReleaseName = component.Name
			i.UpgradeClient.Install = true

			properties := model.CollectPropertiesWithPrefix(component.Properties, "helm.values.", injections)
			intfaceProperties := map[string]interface{}{}
			for k, v := range properties {
				intfaceProperties[k] = v
			}
			//if _, err := i.InstallClient.Run(chart, nil); err != nil {
			if _, err := i.UpgradeClient.Run(component.Name, chart, intfaceProperties); err != nil {
				if _, err := i.InstallClient.Run(chart, intfaceProperties); err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Helm Target): failed to apply: %+v", err)
					return err
				}
			}
			os.Remove(fileName)
		}
	}

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
