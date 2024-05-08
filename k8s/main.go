//go:build !azure
// +build !azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	aiv1 "gopls-workspace/apis/ai/v1"
	configv1 "gopls-workspace/apis/config/v1"
	fabricv1 "gopls-workspace/apis/fabric/v1"
	federationv1 "gopls-workspace/apis/federation/v1"
	solutionv1 "gopls-workspace/apis/solution/v1"
	workflowv1 "gopls-workspace/apis/workflow/v1"
	"gopls-workspace/constants"

	aicontrollers "gopls-workspace/controllers/ai"
	fabriccontrollers "gopls-workspace/controllers/fabric"
	federationcontrollers "gopls-workspace/controllers/federation"
	solutioncontrollers "gopls-workspace/controllers/solution"
	workflowcontrollers "gopls-workspace/controllers/workflow"
	//+kubebuilder:scaffold:imports
)

var (
	scheme      = runtime.NewScheme()
	setupLog    = ctrl.Log.WithName("setup")
	apiCertPath = os.Getenv(constants.ApiCertEnvName)
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(solutionv1.AddToScheme(scheme))
	utilruntime.Must(fabricv1.AddToScheme(scheme))
	utilruntime.Must(aiv1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	utilruntime.Must(workflowv1.AddToScheme(scheme))
	utilruntime.Must(federationv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var configFile string
	var pollIntervalString string
	var reconcileIntervalString string
	var deleteTimeOutString string
	var metricsConfigFile string
	var disableWebhooksServer bool
	var deleteSyncDelayString string

	flag.StringVar(&metricsConfigFile, "metrics-config-file", "", "The path to the otel metrics config file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&configFile, "config", "", "The controller will laod its initial configuration from this file. "+
		"Omit this flag to use the default configuration value. "+
		"Command-line flags override configuration from this file.")
	flag.BoolVar(&disableWebhooksServer, "disable-webhooks-server", false, "Whether to disable webhooks server endpoints. ")
	flag.StringVar(&pollIntervalString, "poll-interval", "10s", "The interval in seconds to poll the target and instance status during reconciliation.")
	flag.StringVar(&reconcileIntervalString, "reconcile-interval", "30m", "The interval in seconds to reconcile the target and instance status.")
	// Honor OSS changes: use 1m instead of 5m for delete-timeout
	flag.StringVar(&deleteTimeOutString, "delete-timeout", "1m", "The timeout in seconds to wait for the target and instance deletion.")
	// Add new settings for delete sync delay
	flag.StringVar(&deleteSyncDelayString, "delete-sync-delay", "0s", "The delay in seconds to wait for the status sync back in delete operations.")

	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	fmt.Println(constants.EulaMessage)
	fmt.Println()

	ctx := ctrl.SetupSignalHandler()
	var err error
	ctrlConfig := configv1.ProjectConfig{}
	options := ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "33405cb8.symphony",
	}
	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if metricsConfigFile != "" {
		obs, err := initMetrics(metricsConfigFile)
		if err != nil {
			setupLog.Error(err, "unable to initialize metrics")
			os.Exit(1)
		}
		defer shutdownMetrics(obs)
	}

	apiClientOptions := []utils.ApiClientOption{
		utils.WithCertAuth(apiCertPath),
	}

	if utils.ShouldUseSATokens() {
		apiClientOptions = append(apiClientOptions, utils.WithServiceAccountToken())
	} else {
		apiClientOptions = append(apiClientOptions, utils.WithUserPassword(ctx, "admin", ""))
	}

	apiClient, err := utils.NewAPIClient(
		ctx,
		utils.GetSymphonyAPIAddressBase(),
		apiClientOptions...,
	)
	if err != nil {
		setupLog.Error(err, "unable to create api client")
		os.Exit(1)
	}

	pollInterval, err := time.ParseDuration(pollIntervalString)
	if err != nil {
		setupLog.Error(err, "unable to parse poll interval")
		os.Exit(1)
	}

	reconcileInterval, err := time.ParseDuration(reconcileIntervalString)
	if err != nil {
		setupLog.Error(err, "unable to parse reconcile interval")
		os.Exit(1)
	}

	deleteTimeOut, err := time.ParseDuration(deleteTimeOutString)
	if err != nil {
		setupLog.Error(err, "unable to parse delete timeout")
		os.Exit(1)
	}

	deleteSyncDelay, err := time.ParseDuration(deleteSyncDelayString)
	if err != nil {
		setupLog.Error(err, "unable to parse delete sync delay")
		os.Exit(1)
	}

	if err = (&solutioncontrollers.SolutionReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Solution")
		os.Exit(1)
	}
	if err = (&workflowcontrollers.CampaignReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Campaign")
		os.Exit(1)
	}
	if err = (&workflowcontrollers.ActivationReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Activation")
		os.Exit(1)
	}
	if err = (&solutioncontrollers.InstanceReconciler{
		Client:                 mgr.GetClient(),
		Scheme:                 mgr.GetScheme(),
		ReconciliationInterval: reconcileInterval,
		DeleteTimeOut:          deleteTimeOut,
		PollInterval:           pollInterval,
		DeleteSyncDelay:        deleteSyncDelay,
		ApiClient:              apiClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Instance")
		os.Exit(1)
	}
	if err = (&fabriccontrollers.TargetReconciler{
		Client:                 mgr.GetClient(),
		Scheme:                 mgr.GetScheme(),
		ReconciliationInterval: reconcileInterval,
		DeleteTimeOut:          deleteTimeOut,
		PollInterval:           pollInterval,
		DeleteSyncDelay:        deleteSyncDelay,
		ApiClient:              apiClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Target")
		os.Exit(1)
	}
	if err = (&fabriccontrollers.DeviceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Device")
		os.Exit(1)
	}
	if err = (&aicontrollers.ModelReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Model")
		os.Exit(1)
	}
	if err = (&aicontrollers.SkillReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Skill")
		os.Exit(1)
	}
	if err = (&aicontrollers.SkillPackageReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SkillPackage")
		os.Exit(1)
	}
	if err = (&federationcontrollers.SiteReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Site")
		os.Exit(1)
	}
	if err = (&federationcontrollers.CatalogReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Catalog")
		os.Exit(1)
	}
	if !disableWebhooksServer {
		if err = (&fabricv1.Device{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Device")
			os.Exit(1)
		}
		if err = (&fabricv1.Target{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Target")
			os.Exit(1)
		}
		if err = (&solutionv1.Solution{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Solution")
			os.Exit(1)
		}
		if err = (&solutionv1.Instance{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Instance")
			os.Exit(1)
		}
		if err = (&aiv1.Model{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Model")
			os.Exit(1)
		}
		if err = (&aiv1.Skill{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Skill")
			os.Exit(1)
		}
		if err = (&federationv1.Catalog{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Catalog")
			os.Exit(1)
		}
	}
	if err = (&solutioncontrollers.SolutionContainerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SolutionContainer")
		os.Exit(1)
	}
	if err = (&federationcontrollers.CatalogContainerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CatalogContainer")
		os.Exit(1)
	}
	if err = (&fabriccontrollers.TargetContainerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TargetContainer")
		os.Exit(1)
	}
	if err = (&workflowcontrollers.CampaignContainerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CampaignContainer")
		os.Exit(1)
	}
	if err = (&solutioncontrollers.InstanceContainerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "InstanceContainer")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func initMetrics(configPath string) (*observability.Observability, error) {
	// Read file content and parse int ObservabilityConfig
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config observability.ObservabilityConfig

	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, err
	}

	obs := observability.New(constants.K8S)
	if err := obs.InitMetric(config); err != nil {
		return nil, err
	}

	return &obs, nil
}

func shutdownMetrics(obs *observability.Observability) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	obs.Shutdown(ctx)
}
