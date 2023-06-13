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
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	"github.com/google/uuid"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var sLog = logger.NewLogger("coa.runtime")

const (
	DEFAULT_NAMESPACE = "default"
	TEMP_CHART_DIR    = "/tmp/symphony/charts"
)

type (
	// HelmTargetProviderConfig is the configuration for the Helm provider
	HelmTargetProviderConfig struct {
		Name       string `json:"name"`
		ConfigType string `json:"configType,omitempty"`
		ConfigData string `json:"configData,omitempty"`
		Context    string `json:"context,omitempty"`
		InCluster  bool   `json:"inCluster"`
	}
	// HelmTargetProvider is the Helm provider
	HelmTargetProvider struct {
		Config          HelmTargetProviderConfig
		Context         *contexts.ManagerContext
		ListClient      *action.List
		InstallClient   *action.Install
		UpgradeClient   *action.Upgrade
		UninstallClient *action.Uninstall
	}
	// HelmProperty is the property for the Helm provider
	HelmProperty struct {
		Chart  HelmChartProperty      `json:"chart"`
		Values map[string]interface{} `json:"values,omitempty"`
	}
	// HelmChartProperty is the property for the Helm Charts
	HelmChartProperty struct {
		Repo    string `json:"repo"`
		Version string `json:"version"`
		Wait    bool   `json:"wait"`
	}
)

// HelmTargetProviderConfigFromMap converts a map to a HelmTargetProviderConfig
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

// InitWithMap initializes the HelmTargetProvider with a map
func (i *HelmTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := HelmTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}

	return i.Init(config)
}

// Init initializes the HelmTargetProvider
func (i *HelmTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan(
		"Helm Target Provider",
		context.Background(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Info("  P (Helm Target): Init()")

	err = initChartsDir()
	if err != nil {
		sLog.Errorf("  P (Helm Target): failed to init charts dir: %+v", err)
		return err
	}

	// convert config to HelmTargetProviderConfig type
	var helmConfig HelmTargetProviderConfig
	helmConfig, err = toHelmTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Helm Target): expected HelmTargetProviderConfig: %+v", err)
		return err
	}

	i.Config = helmConfig
	var actionConfig *action.Configuration
	if i.Config.InCluster {
		settings := cli.New()
		actionConfig = new(action.Configuration)
		// TODO: $HELM_DRIVER	set the backend storage driver. Values are: configmap, secret, memory, sql. Do we need to handle this differently?
		if err = actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
			sLog.Errorf("  P (Helm Target): failed to init: %+v", err)
			return err
		}
	} else {
		switch i.Config.ConfigType {
		case "bytes":
			if i.Config.ConfigData != "" {
				var kConfig *rest.Config
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.Errorf("  P (Helm Target): failed to init with config bytes: %+v", err)
					return err
				}

				namespace := DEFAULT_NAMESPACE
				actionConfig, err = getActionConfig(context.Background(), namespace, kConfig)
				if err != nil {
					sLog.Errorf("  P (Helm Target): failed to init with config bytes: %+v", err)
					return err
				}

			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				sLog.Errorf("  P (Helm Target): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted value is: bytes", v1alpha2.BadConfig)
			sLog.Errorf("  P (Helm Target): %+v", err)
			return err
		}
	}

	i.ListClient = action.NewList(actionConfig)
	i.InstallClient = action.NewInstall(actionConfig)
	i.UninstallClient = action.NewUninstall(actionConfig)
	i.UpgradeClient = action.NewUpgrade(actionConfig)
	return nil
}

// getActionConfig returns an action configuration
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

// toHelmTargetProviderConfig converts a generic IProviderConfig to a HelmTargetProviderConfig
func toHelmTargetProviderConfig(config providers.IProviderConfig) (HelmTargetProviderConfig, error) {
	ret := HelmTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}

// NeedsUpdate returns true if the desired state is different from the current state
func (i *HelmTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	for _, dc := range desired {
		desiredHelmChart, err := getHelmPropertyFromComponent(dc)
		if err != nil {
			sLog.Errorf("  P (Helm Provider): %+v", err)
			return true
		}
		found := false
		for _, cc := range current {
			if cc.Name == dc.Name && cc.Type == dc.Type && cc.Type == "helm.v3" {
				found = true
				currentHelmChart, err := getHelmPropertyFromComponent(cc)
				if err != nil {
					sLog.Errorf("  P (Helm Provider): %+v", err)
					return true
				}

				// check if helm chart needs an update
				if desiredHelmChart.Chart != currentHelmChart.Chart {
					sLog.Info("  P (Helm Provider): NeedsUpdate: returning true")
					return true
				}

				// check if helm values needs an update
				if isEmpty(desiredHelmChart.Values) && isEmpty(currentHelmChart.Values) {
					break
				}

				if !reflect.DeepEqual(desiredHelmChart.Values, currentHelmChart.Values) {
					sLog.Info("  P (Helm Provider): NeedsUpdate: returning true")
					return true
				}

			}
		}

		if !found {
			sLog.Info("  P (Helm Provider): NeedsUpdate: returning true")
			return true
		}
	}

	sLog.Info("  P (Helm Provider): NeedsUpdate: returning false")
	return false
}

// isEmpty returns true if the map is empty
func isEmpty(data map[string]interface{}) bool {
	if data == nil {
		return true
	}
	return len(data) == 0
}

// NeedsRemove returns true if the desired state is different from the current state
func (i *HelmTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	for _, dc := range desired {
		for _, cc := range current {
			if cc.Name == dc.Name {
				sLog.Info("  P (Helm Provider): NeedsRemove: returning true")
				return true
			}
		}
	}
	sLog.Info("  P (Helm Provider): NeedsUpdate: returning false")
	return false
}

// Get returns the list of components for a given deployment
func (i *HelmTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan(
		"Helm Target Provider",
		ctx,
		&map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Infof("  P (Helm Target): getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)
	i.ListClient.Deployed = true
	var results []*release.Release
	results, err = i.ListClient.Run()
	if err != nil {
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
					Properties: map[string]interface{}{
						"chart": map[string]string{
							"repo":    repo,
							"version": res.Chart.Metadata.Version,
						},
						"values": res.Config,
					},
				})
			}
		}
	}

	return ret, nil
}

// Remove deletes the artifacts for a given deployment
func (i *HelmTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan(
		"Helm Target Provider",
		ctx,
		&map[string]string{
			"method": "Remove",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Infof("  P (Helm Target): deleting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := deployment.GetComponentSlice()
	for _, component := range components {
		if component.Type == "helm.v3" {
			_, err = i.UninstallClient.Run(component.Name)
			if err != nil {
				sLog.Errorf("  P (Helm Target): failed to uninstall Helm chart: %+v", err)
				return err
			}
		}
	}

	return nil
}

// GetValidationRule returns the validation rule for this provider
func (*HelmTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{"chart"},
		OptionalProperties:    []string{"values"},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}

// downloadFile will download a url to a local file. It's efficient because it will
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

// Apply deploys the helm chart for a given deployment
func (i *HelmTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, isDryRun bool) error {
	_, span := observability.StartSpan(
		"Helm Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Infof("  P (Helm Target): applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := deployment.GetComponentSlice()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		return err
	}

	if isDryRun {
		return nil
	}

	for _, component := range components {

		var helmProp *HelmProperty
		helmProp, err = getHelmPropertyFromComponent(component)
		if err != nil {
			sLog.Errorf("  P (Helm Target): failed to get Helm properties: %+v", err)
			return err
		}

		fileName, err := i.pullChart(&helmProp.Chart)
		if err != nil {
			sLog.Errorf("  P (Helm Target): failed to pull chart: %+v", err)
			return err
		}
		defer os.Remove(fileName)

		var chart *chart.Chart
		chart, err = loader.Load(fileName)
		if err != nil {
			sLog.Errorf("  P (Helm Target): failed to load chart: %+v", err)
			return err
		}

		chart.Metadata.Tags = "SYM:" + helmProp.Chart.Repo //this is not used by Helm SDK, we use this to carry repo info
		i.configureUpsertClients(component.Name, &helmProp.Chart, &deployment)

		if _, err = i.UpgradeClient.Run(component.Name, chart, helmProp.Values); err != nil {
			if _, err = i.InstallClient.Run(chart, helmProp.Values); err != nil {
				sLog.Errorf("  P (Helm Target): failed to apply: %+v", err)
				return err
			}
		}

	}

	return nil
}

func (i *HelmTargetProvider) pullChart(chart *HelmChartProperty) (fileName string, err error) {
	fileName = fmt.Sprintf("%s/%s.tgz", TEMP_CHART_DIR, uuid.New().String())

	var pullRes *registry.PullResult
	if strings.HasSuffix(chart.Repo, ".tgz") && strings.HasPrefix(chart.Repo, "http") {
		err = downloadFile(chart.Repo, fileName)
		if err != nil {
			sLog.Errorf("  P (Helm Target): failed to download chart from repo: %+v", err)
			return "", err
		}
	} else {
		var regClient *registry.Client
		regClient, err = registry.NewClient()
		if err != nil {
			sLog.Errorf("  P (Helm Target): failed to create registry client: %+v", err)
			return
		}

		pullRes, err = regClient.Pull(fmt.Sprintf("%s:%s", chart.Repo, chart.Version), registry.PullOptWithChart(true))
		if err != nil {
			sLog.Errorf("  P (Helm Target): failed to pull chart from repo: %+v", err)
			return
		}

		err = ioutil.WriteFile(fileName, pullRes.Chart.Data, 0644)
		if err != nil {
			sLog.Errorf("  P (Helm Target): failed to save chart: %+v", err)
			return
		}
	}
	return fileName, nil
}

func (i *HelmTargetProvider) configureUpsertClients(name string, componentProps *HelmChartProperty, deployment *model.DeploymentSpec) {
	if deployment.Instance.Scope == "" {
		i.InstallClient.Namespace = DEFAULT_NAMESPACE
		i.UpgradeClient.Namespace = DEFAULT_NAMESPACE
	} else {
		i.InstallClient.Namespace = deployment.Instance.Scope
		i.UpgradeClient.Namespace = deployment.Instance.Scope
	}

	i.InstallClient.Wait = componentProps.Wait
	i.UpgradeClient.Wait = componentProps.Wait
	i.InstallClient.CreateNamespace = true
	i.InstallClient.ReleaseName = name
	i.InstallClient.IsUpgrade = true
	i.UpgradeClient.Install = true
}

func getHelmPropertyFromComponent(component model.ComponentSpec) (*HelmProperty, error) {
	ret := HelmProperty{}
	data, err := json.Marshal(component.Properties)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}

	return validateProps(&ret)
}

func validateProps(props *HelmProperty) (*HelmProperty, error) {
	if props.Chart.Repo == "" {
		return nil, errors.New("chart repo is required")
	}

	return props, nil
}

func initChartsDir() error {
	if _, err := os.Stat(TEMP_CHART_DIR); os.IsNotExist(err) {
		err = os.MkdirAll(TEMP_CHART_DIR, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
