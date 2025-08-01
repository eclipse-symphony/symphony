/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package helm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils/metahelper"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/google/uuid"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

const (
	defaultNamespace = "default"
	tempChartDir     = "/tmp/symphony/charts"
	helmDriver       = "secret"
	helm             = "helm"
	providerName     = "P (Helm Target)"
	loggerName       = "providers.target.helm"
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
		Config        HelmTargetProviderConfig
		Context       *contexts.ManagerContext
		MetaPopulator metahelper.MetaPopulator
	}
	// HelmProperty is the property for the Helm chart
	HelmProperty struct {
		Chart       HelmChartProperty      `json:"chart"`
		Values      map[string]interface{} `json:"values,omitempty"`
		ReleaseName string                 `json:"releaseName,omitempty"`
	}
	// HelmChartProperty is the property for the Helm Charts
	HelmChartProperty struct {
		Repo     string `json:"repo"`
		Name     string `json:"name,omitempty"`
		Version  string `json:"version"`
		Wait     bool   `json:"wait"`
		Timeout  string `json:"timeout,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
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
		sLog.Errorf("  P (ConfigMap Target): expected HelmTargetProviderConfigFromMap: %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init", providerName), v1alpha2.InitFailed)
	}

	return i.Init(config)
}

func (s *HelmTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

// Init initializes the HelmTargetProvider
func (i *HelmTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"Helm Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfoCtx(ctx, "  P (Helm Target): Init()")

	i.MetaPopulator, err = metahelper.NewMetaPopulator(metahelper.WithDefaultPopulators())
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to create meta populator: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create meta populator", providerName), v1alpha2.InitFailed)
		sLog.ErrorCtx(ctx, err)
		return err
	}

	err = initChartsDir()
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to init charts dir: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init charts dir", providerName), v1alpha2.InitFailed)
		return err
	}

	// convert config to HelmTargetProviderConfig type
	var helmConfig HelmTargetProviderConfig
	helmConfig, err = toHelmTargetProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): expected HelmTargetProviderConfig: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to convert to HelmTargetProviderConfig", providerName), v1alpha2.InitFailed)
		return err
	}

	i.Config = helmConfig

	// validate config
	_, err = i.createActionConfig(context.Background(), defaultNamespace)
	if err != nil {
		return err
	}

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to create metrics: %+v", err)
			}
		}
	})

	return err
}

func (i *HelmTargetProvider) createActionConfig(ctx context.Context, namespace string) (*action.Configuration, error) {
	var actionConfig *action.Configuration
	if namespace == "" {
		namespace = constants.DefaultScope
	}
	sLog.DebugfCtx(ctx, "  P (Helm Target): creating action config for namespace %s", namespace)
	var err error
	if i.Config.InCluster {
		settings := cli.New()
		settings.SetNamespace(namespace)
		actionConfig = new(action.Configuration)
		// TODO: $HELM_DRIVER	set the backend storage driver. Values are: configmap, secret, memory, sql. Do we need to handle this differently?
		if err = actionConfig.Init(settings.RESTClientGetter(), namespace, helmDriver, sLog.Debugf); err != nil {
			sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to init: %+v", err)
			err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init action config", providerName), v1alpha2.CreateActionConfigFailed)
			return nil, err
		}
	} else {
		switch i.Config.ConfigType {
		case "bytes":
			if i.Config.ConfigData != "" {
				var kConfig *rest.Config
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to get RestConfig: %+v", err)
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get RestConfig", providerName), v1alpha2.CreateActionConfigFailed)
					return nil, err
				}

				actionConfig, err = getActionConfig(context.Background(), namespace, kConfig)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to get ActionConfig: %+v", err)
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get ActionConfig", providerName), v1alpha2.CreateActionConfigFailed)
					return nil, err
				}

			} else {
				err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: config data is not supplied", providerName), v1alpha2.CreateActionConfigFailed)
				sLog.ErrorCtx(ctx, err)
				return nil, err
			}
		default:
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: unrecognized config type, accepted value is: bytes", providerName), v1alpha2.CreateActionConfigFailed)
			sLog.ErrorCtx(ctx, err)
			return nil, err
		}
	}
	return actionConfig, nil
}

// getActionConfig returns an action configuration
func getActionConfig(ctx context.Context, namespace string, config *rest.Config) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	cliConfig := genericclioptions.NewConfigFlags(false)
	cliConfig.APIServer = &config.Host
	cliConfig.BearerToken = &config.BearerToken
	cliConfig.Namespace = &namespace
	// Drop their rest.Config and just return inject own
	cliConfig.WithWrapConfigFn(func(*rest.Config) *rest.Config {
		return config
	})

	if err := actionConfig.Init(cliConfig, namespace, helmDriver, sLog.Debugf); err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to init: %+v", err)
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

// Get returns the list of components for a given deployment
func (i *HelmTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"Helm Target Provider",
		ctx,
		&map[string]string{
			"method": "Get",
		},
	)
	var err error
	var actionConfig *action.Configuration
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Helm Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)
	actionConfig, err = i.createActionConfig(ctx, deployment.Instance.Spec.Scope)
	if err != nil {
		sLog.ErrorCtx(ctx, err)
		return nil, err
	}
	listClient := action.NewList(actionConfig)
	listClient.Deployed = true
	var results []*release.Release
	results, err = listClient.Run()
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to create Helm list client: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create Helm list client", providerName), v1alpha2.HelmActionFailed)
		return nil, err
	}

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		helmProp, err := getHelmPropertyFromComponent(component.Component)
		releaseName := GetReleaseName(component.Component, helmProp)
		for _, res := range results {
			if (deployment.Instance.Spec.Scope == "" || res.Namespace == deployment.Instance.Spec.Scope) && res.Name == releaseName {
				repo := ""
				name := ""
				if strings.HasPrefix(res.Chart.Metadata.Tags, "SYM-REPO:") { //we use this special metadata tag to remember the chart URL
					parts := strings.Split(res.Chart.Metadata.Tags, ";")
					if len(parts) != 2 {
						sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to parse chart metadata tags: %+v", res.Chart.Metadata.Tags)
						err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to parse chart metadata tags", providerName), v1alpha2.HelmActionFailed)
						return nil, err
					}
					repo = parts[0][9:]
					name = parts[1][9:]
				}
				ret = append(ret, model.ComponentSpec{
					Name: component.Component.Name,
					Type: "helm.v3",
					Properties: map[string]interface{}{
						"releaseName": res.Name,
						"chart": map[string]string{
							"repo":    repo,
							"name":    name,
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

// GetValidationRule returns the validation rule for this provider
func (*HelmTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{"chart"},
			OptionalProperties:    []string{"values"},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "chart", IgnoreCase: false, SkipIfMissing: true}, //TODO: deep change detection on interface{}
				{Name: "values", PropChanged: propChange},
			},
		},
	}
}

func propChange(old, new interface{}) bool {
	// scenarios where either is an empty map and the other is nil count as no change
	if isEmpty(old) && isEmpty(new) {
		return false
	}
	return !reflect.DeepEqual(old, new)
}

func isEmpty(values interface{}) bool {
	if values == nil {
		return true
	}
	valueMap, ok := values.(map[string]interface{})
	if ok {
		return len(valueMap) == 0
	}
	return false
}

// check if this uri is a downloadable uri
func isDownloadableUri(uri string) bool {
	// check if the uri has suffix of .tgz or .tar.gz after removing the query parameter
	uri = strings.Split(uri, "?")[0]
	return strings.HasSuffix(uri, ".tgz") || strings.HasSuffix(uri, ".tar.gz")
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
func (i *HelmTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan(
		"Helm Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Helm Target): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := utils.GetFunctionName()
	startTime := time.Now().UTC()
	defer providerOperationMetrics.ProviderOperationLatency(
		startTime,
		helm,
		metrics.ApplyOperation,
		metrics.ApplyOperationType,
		functionName,
	)

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to validate components: %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			helm,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.ApplyOperationType,
			v1alpha2.ValidateFailed.String(),
		)

		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: the rule validation failed", providerName), v1alpha2.ValidateFailed)
		return nil, err
	}

	if isDryRun {
		sLog.DebugCtx(ctx, "  P (Helm Target): dryRun is enabled, skipping apply")
		return nil, nil
	}

	ret := step.PrepareResultMap()

	var actionConfig *action.Configuration
	actionConfig, err = i.createActionConfig(ctx, deployment.Instance.Spec.Scope)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to create action config: %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			helm,
			functionName,
			metrics.HelmActionConfigOperation,
			metrics.ApplyOperationType,
			v1alpha2.CreateActionConfigFailed.String(),
		)
		return ret, err
	}

	for _, component := range step.Components {
		var helmProp *HelmProperty
		helmProp, err = getHelmPropertyFromComponent(component.Component)
		releaseName := GetReleaseName(component.Component, helmProp)
		if component.Action == model.ComponentUpdate {
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to get Helm properties: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get helm properties", providerName), v1alpha2.GetHelmPropertyFailed)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				providerOperationMetrics.ProviderOperationErrors(
					helm,
					functionName,
					metrics.HelmPropertiesOperation,
					metrics.ApplyOperationType,
					v1alpha2.GetHelmPropertyFailed.String(),
				)

				return ret, err
			}

			var fileName string
			fileName, err = i.pullChart(ctx, &helmProp.Chart)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to pull chart: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to pull chart", providerName), v1alpha2.HelmActionFailed)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				providerOperationMetrics.ProviderOperationErrors(
					helm,
					functionName,
					metrics.PullChartOperation,
					metrics.ApplyOperationType,
					v1alpha2.HelmChartPullFailed.String(),
				)

				return ret, err
			}
			defer os.Remove(fileName)

			var chart *chart.Chart
			chart, err = loader.Load(fileName)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to load chart: %+v", err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to load chart", providerName), v1alpha2.HelmActionFailed)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				providerOperationMetrics.ProviderOperationErrors(
					helm,
					functionName,
					metrics.LoadChartOperation,
					metrics.ApplyOperationType,
					v1alpha2.HelmChartLoadFailed.String(),
				)

				return ret, err
			}

			chart.Metadata.Tags = "SYM-REPO:" + helmProp.Chart.Repo + ";SYM-NAME:" + helmProp.Chart.Name //this is not used by Helm SDK, we use this to carry repo info

			postRender := &PostRenderer{
				instance:  deployment.Instance,
				populator: i.MetaPopulator,
			}
			installClient, err := configureInstallClient(ctx, releaseName, &helmProp.Chart, &deployment, actionConfig, postRender)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to config helm install client: %+v", err)
				return nil, err
			}
			upgradeClient, err := configureUpgradeClient(ctx, &helmProp.Chart, &deployment, actionConfig, postRender)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to config helm upgrade client: %+v", err)
				return nil, err
			}

			releaseExists, err := checkReleaseExists(ctx, actionConfig, releaseName)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): Error checking if chart exists: %+v", err)
				return nil, err
			}
			utils.EmitUserAuditsLogs(ctx, "  P (Helm Target): Applying chart, releaseName: %s, defined in component: %s, chart: {repo: %s, name: %s, version: %s}, namespace: %s", releaseName, component.Component.Name, helmProp.Chart.Repo, helmProp.Chart.Name, helmProp.Chart.Version, deployment.Instance.Spec.Scope)
			if releaseExists {
				sLog.InfofCtx(ctx, "  P (Helm Target): Chart upgrade started. Details - Release Name: %s, Component Name: %s", releaseName, component.Component.Name)
				if _, err = upgradeClient.Run(releaseName, chart, helmProp.Values); err != nil {
					sLog.InfofCtx(ctx, "  P (Helm Target): failed to upgrade: %+v", err)
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to upgrade chart", providerName), v1alpha2.HelmActionFailed)
					ret[component.Component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.UpdateFailed,
						Message: err.Error(),
					}
					providerOperationMetrics.ProviderOperationErrors(
						helm,
						functionName,
						metrics.HelmChartOperation,
						metrics.ApplyOperationType,
						v1alpha2.HelmChartApplyFailed.String(),
					)
					return ret, err
				}
			} else {
				sLog.InfofCtx(ctx, "  P (Helm Target): Chart installation started. Details - Release Name: %s, Component Name: %s", releaseName, component.Component.Name)
				if _, err := installClient.Run(chart, helmProp.Values); err != nil {
					sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to install: %+v", err)
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to install chart", providerName), v1alpha2.HelmActionFailed)
					ret[component.Component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.UpdateFailed,
						Message: err.Error(),
					}
					providerOperationMetrics.ProviderOperationErrors(
						helm,
						functionName,
						metrics.HelmChartOperation,
						metrics.ApplyOperationType,
						v1alpha2.HelmChartApplyFailed.String(),
					)
					return ret, err
				}
				sLog.InfofCtx(ctx, "  P (Helm Target): Chart installation completed successfully. Details - Release Name: %s, Component Name: %s", releaseName, component.Component.Name)
			}

			sLog.InfofCtx(ctx, "  P (Helm Target): apply chart successfully. Details - Release Name: %s, Component Name: %s", releaseName, component.Component.Name)
			ret[component.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Updated,
				Message: fmt.Sprintf("No error. %s has been updated. Release Name: %s", component.Component.Name, releaseName),
			}
		} else {
			switch component.Component.Type {
			case "helm.v3":
				uninstallClient, err := configureUninstallClient(ctx, &helmProp.Chart, &deployment, actionConfig)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to configure uninstall client: %+v", err)
					return nil, err
				}
				utils.EmitUserAuditsLogs(ctx, "  P (Helm Target): Uninstalling chart, releaseName: %s, defined in component: %s, namespace: %s", releaseName, component.Component.Name, deployment.Instance.Spec.Scope)
				_, err = uninstallClient.Run(releaseName)
				if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
					sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to uninstall Helm chart: %+v", err)
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to uninstall chart", providerName), v1alpha2.HelmActionFailed)
					ret[component.Component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.DeleteFailed,
						Message: err.Error(),
					}
					providerOperationMetrics.ProviderOperationErrors(
						helm,
						functionName,
						metrics.HelmChartOperation,
						metrics.ApplyOperationType,
						v1alpha2.HelmChartUninstallFailed.String(),
					)

					return ret, err
				}

				sLog.InfofCtx(ctx, "  P (Helm Target): Chart uninstalled successfully. Details - Release Name: %s, Component Name: %s", releaseName, component.Component.Name)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.Deleted,
					Message: "",
				}
			default:
				sLog.ErrorfCtx(ctx, "  P (Helm Target): Failed to uninstall helm chart as %v is an invalid helm version", component.Component.Type)
			}
		}
	}

	return ret, nil
}

func (i *HelmTargetProvider) pullChart(ctx context.Context, chart *HelmChartProperty) (fileName string, err error) {
	fileName = fmt.Sprintf("%s/%s.tgz", tempChartDir, uuid.New().String())

	utils.EmitUserAuditsLogs(ctx, "   P (Helm Target): Starting pulling chart, repo - %s, name - %s, version - %s", chart.Repo, chart.Name, chart.Version)
	if strings.HasPrefix(chart.Repo, "http") {
		var chartPath string
		if isDownloadableUri(chart.Repo) {
			chartPath = chart.Repo
		} else {
			sLog.InfoCtx(ctx, "   P (Helm Target): artifact is hosted in public repo. Attempting to pull without basic auth")
			chartPath, err = repo.FindChartInRepoURL(chart.Repo, chart.Name, chart.Version, "", "", "", getter.All(&cli.EnvSettings{}))
			if err != nil {
				sLog.ErrorfCtx(ctx, "   P (Helm Target): failed to find helm chart in repo: %+v", err)
				return "", err
			}
		}

		err = downloadFile(chartPath, fileName)
		if err != nil {
			sLog.ErrorfCtx(ctx, "   P (Helm Target): failed to download chart from repo: %+v", err)
			return "", err
		}
	} else {
		var pullRes *registry.PullResult

		// Helm provider supports oci-based registry. Symphony manifest supports it in two formats.
		// 1. with oci prefix, e.g. oci://myregistry.azurecr.io/mychart:1.0.0 (https://helm.sh/docs/topics/registries/#oci-feature-deprecation-and-behavior-changes-with-v370)
		// 2. without oci prefix, e.g. myregistry.azurecr.io/mychart:1.0.0 (backwards compatibility with existing symphony behavior)
		// However, registry.Client doesn't like the reference to be prefixed with "oci://"
		// so we trim it here if it exists
		if chart.Username != "" && chart.Password != "" {
			sLog.InfoCtx(ctx, "  P (Helm Target): registry username and password provided. Attempting to pull using basic auth")
			pullRes, err = pullOCIChartWithBasicAuth(ctx, chart.Repo, chart.Version, chart.Username, chart.Password)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to pull chart from repo using basic auth: %+v", err)
				return "", err
			}
		} else {
			pullRes, err = pullOCIChart(ctx, chart.Repo, chart.Version)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Helm Target): got error pulling chart from repo: %+v", err)
				host, herr := getHostFromOCIRef(chart.Repo)
				if herr != nil {
					sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to get host from oci ref: %+v", herr)
					return "", herr
				}
				if isUnauthorized(err) {
					sLog.InfoCtx(ctx, "  P (Helm Target): pulling chart from repo failed with unauthorized error.")
					if isAzureContainerRegistry(host) {
						sLog.InfoCtx(ctx, "  P (Helm Target): artifact is hosted in ACR. Attempting to login to ACR")
						err = loginToACR(ctx, host)
						if err != nil {
							sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to login to ACR: %+v", err)
							return "", err
						}
						sLog.InfoCtx(ctx, "  P (Helm Target): successfully logged in to ACR. Now retrying to pull chart from repo")

						pullRes, err = pullOCIChart(ctx, chart.Repo, chart.Version)
						if err != nil {
							sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to pull chart from repo after login in: %+v", err)
							return "", err
						}
					}
				} else {
					sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to get host from oci ref and it is not because of access issue: %+v", err)
					return "", err
				}
			}
		}

		err = os.WriteFile(fileName, pullRes.Chart.Data, 0644)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to save chart: %+v", err)
			return
		}
	}
	return fileName, nil
}

func pullOCIChart(ctx context.Context, repo, version string) (*registry.PullResult, error) {
	client, err := registry.NewClient()
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to create registry client: %+v", err)
		return nil, err
	}

	pullRes, err := client.Pull(fmt.Sprintf("%s:%s", strings.TrimPrefix(repo, "oci://"), version), registry.PullOptWithChart(true))
	if err != nil {
		return nil, err
	}

	return pullRes, nil
}

func pullOCIChartWithBasicAuth(ctx context.Context, repo, version, username, password string) (*registry.PullResult, error) {
	client, err := registry.NewClient()
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to create registry client: %+v", err)
		return nil, err
	}
	host, herr := getHostFromOCIRef(repo)
	if herr != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to get host from oci ref: %+v", herr)
		return nil, herr
	}

	err = client.Login(host, registry.LoginOptBasicAuth(username, password))
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to login with basic auth: %+v", err)
		return nil, err
	}

	pullRes, err := client.Pull(fmt.Sprintf("%s:%s", strings.TrimPrefix(repo, "oci://"), version), registry.PullOptWithChart(true))
	if err != nil {
		return nil, err
	}

	return pullRes, nil
}

func configureInstallClient(ctx context.Context, name string, componentProps *HelmChartProperty, deployment *model.DeploymentSpec, config *action.Configuration, postRenderer postrender.PostRenderer) (*action.Install, error) {
	sLog.InfofCtx(ctx, "  P (Helm Target): start configuring install client in the namespace %s", deployment.Instance.Spec.Scope)
	installClient := action.NewInstall(config)
	installClient.ReleaseName = name
	if deployment.Instance.Spec.Scope == "" {
		installClient.Namespace = constants.DefaultScope
	} else {
		installClient.Namespace = deployment.Instance.Spec.Scope
	}

	installClient.Wait = componentProps.Wait
	if componentProps.Timeout != "" {
		duration, err := convertTimeout(ctx, componentProps.Timeout)
		if err != nil {
			return nil, err
		}
		installClient.Timeout = duration
	}

	installClient.IsUpgrade = true
	installClient.CreateNamespace = true
	installClient.PostRenderer = postRenderer
	// We can't add labels to the release in the current version of the helm client.
	// This should added when we upgrade to helm ^3.13.1
	return installClient, nil
}

func checkReleaseExists(ctx context.Context, config *action.Configuration, releaseName string) (bool, error) {
	sLog.InfofCtx(ctx, "  P (Helm Target): begin to check release exists %s", releaseName)

	if releaseName == "" {
		return false, v1alpha2.NewCOAError(nil, "Release name is required", v1alpha2.BadConfig)
	}

	client := action.NewHistory(config)
	client.Max = 1
	releases, err := client.Run(releaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return false, nil
		}
		return false, v1alpha2.NewCOAError(err, fmt.Sprintf("check release %s failed", releaseName), v1alpha2.HelmActionFailed)
	}

	if len(releases) > 0 {
		return true, nil
	}

	return false, nil
}

func configureUpgradeClient(ctx context.Context, componentProps *HelmChartProperty, deployment *model.DeploymentSpec, config *action.Configuration, postRenderer postrender.PostRenderer) (*action.Upgrade, error) {
	sLog.InfofCtx(ctx, "  P (Helm Target): start configuring upgrade client in the namespace %s", deployment.Instance.Spec.Scope)
	upgradeClient := action.NewUpgrade(config)
	upgradeClient.Wait = componentProps.Wait
	if componentProps.Timeout != "" {
		duration, err := convertTimeout(ctx, componentProps.Timeout)
		if err != nil {
			return nil, err
		}
		upgradeClient.Timeout = duration
	}
	if deployment.Instance.Spec.Scope == "" {
		upgradeClient.Namespace = constants.DefaultScope
	} else {
		upgradeClient.Namespace = deployment.Instance.Spec.Scope
	}
	upgradeClient.ResetValues = true
	upgradeClient.Install = false
	upgradeClient.PostRenderer = postRenderer
	// We can't add labels to the release in the current version of the helm client.
	// This should added when we upgrade to helm ^3.13.1
	return upgradeClient, nil
}

func configureUninstallClient(ctx context.Context, componentProps *HelmChartProperty, deployment *model.DeploymentSpec, config *action.Configuration) (*action.Uninstall, error) {
	sLog.InfofCtx(ctx, "  P (Helm Target): start configuring uninstall client. Details - componentProps.Name: %s, Namespace: %s", componentProps.Name, deployment.Instance.Spec.Scope)
	uninstallClient := action.NewUninstall(config)
	uninstallClient.Wait = componentProps.Wait
	if componentProps.Timeout != "" {
		duration, err := convertTimeout(ctx, componentProps.Timeout)
		if err != nil {
			return nil, err
		}
		uninstallClient.Timeout = duration
	}
	return uninstallClient, nil
}

func convertTimeout(ctx context.Context, timeout string) (time.Duration, error) {
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): failed to parse timeout duration: %v", err)
		err = v1alpha2.NewCOAError(err, "failed to parse timeout duration", v1alpha2.GetComponentPropsFailed)
		return 0, err
	}
	if duration < 0 {
		sLog.ErrorfCtx(ctx, "  P (Helm Target): Timeout is negative: %s", timeout)
		err = v1alpha2.NewCOAError(err, "target provider timeout can not be negative", v1alpha2.GetComponentPropsFailed)
		return 0, err
	}
	return duration, nil
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
	if _, err := os.Stat(tempChartDir); os.IsNotExist(err) {
		err = os.MkdirAll(tempChartDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetReleaseName retrieves the release name for a given component.
func GetReleaseName(component model.ComponentSpec, helmProp *HelmProperty) string {
	if helmProp != nil && helmProp.ReleaseName != "" {
		return helmProp.ReleaseName
	}
	return component.Name
}
