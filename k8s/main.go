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
	"strings"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	zaplog "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	aiv1 "gopls-workspace/apis/ai/v1"
	configv1 "gopls-workspace/apis/config/v1"
	"gopls-workspace/apis/dynamicclient"
	fabricv1 "gopls-workspace/apis/fabric/v1"
	federationv1 "gopls-workspace/apis/federation/v1"
	commoncontainer "gopls-workspace/apis/model/v1"
	solutionv1 "gopls-workspace/apis/solution/v1"
	workflowv1 "gopls-workspace/apis/workflow/v1"
	"gopls-workspace/constants"

	monitorv1 "gopls-workspace/apis/monitor/v1"
	actioncontrollers "gopls-workspace/controllers/actions"
	aicontrollers "gopls-workspace/controllers/ai"
	fabriccontrollers "gopls-workspace/controllers/fabric"
	federationcontrollers "gopls-workspace/controllers/federation"
	monitorcontrollers "gopls-workspace/controllers/monitor"
	solutioncontrollers "gopls-workspace/controllers/solution"
	workflowcontrollers "gopls-workspace/controllers/workflow"
	//+kubebuilder:scaffold:imports
)

type LogMode string

const (
	Development LogMode = "development"
	Production  LogMode = "production"
)

// Set implements flag.Value.
func (l *LogMode) Set(s string) error {
	if l == nil {
		return flag.ErrHelp
	}

	logModeStr := strings.ToLower(s)
	switch logModeStr {
	case "development":
		*l = Development
	case "production":
		*l = Production
	default:
		return flag.ErrHelp
	}

	return nil
}

// String implements flag.Value.
func (l *LogMode) String() string {
	if l == nil {
		return ""
	}
	return string(*l)
}

func (l *LogMode) IsDevelopment() bool {
	if l == nil {
		return false
	}
	return *l == Development
}

func (l *LogMode) IsUndefined() bool {
	if l == nil {
		return true
	}
	return *l != Development && *l != Production
}

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
	utilruntime.Must(monitorv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	fmt.Println(constants.EulaMessage)
	fmt.Println()

	time.Sleep(10 * time.Millisecond) // sleep 10ms to make sure license print at first and won't be mixed with other logs

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var pollIntervalString string
	var reconcileIntervalString string
	var deleteTimeOutString string
	var metricsConfigFile string
	var logsConfigFile string
	var disableWebhooksServer bool
	var deleteSyncDelayString string
	var logMode LogMode

	flag.StringVar(&metricsConfigFile, "metrics-config-file", "", "The path to the otel metrics config file.")
	flag.StringVar(&logsConfigFile, "logs-config-file", "", "The path to the otel logs config file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&disableWebhooksServer, "disable-webhooks-server", false, "Whether to disable webhooks server endpoints. ")
	flag.StringVar(&pollIntervalString, "poll-interval", "10s", "The interval in seconds to poll the target and instance status during reconciliation.")
	flag.StringVar(&reconcileIntervalString, "reconcile-interval", "30m", "The interval in seconds to reconcile the target and instance status.")
	// Honor OSS changes: use 1m instead of 5m for delete-timeout
	flag.StringVar(&deleteTimeOutString, "delete-timeout", "30m", "The timeout in seconds to wait for the target and instance deletion.")
	// Add new settings for delete sync delay
	flag.StringVar(&deleteSyncDelayString, "delete-sync-delay", "0s", "The delay in seconds to wait for the status sync back in delete operations.")
	flag.Var(&logMode, "log-mode", "The log mode. Options are development or production.")

	if logMode.IsUndefined() {
		logMode = Development
	}

	// Create a custom EncoderConfig
	encoderConfig := zapcore.EncoderConfig{
		CallerKey:     "caller",
		TimeKey:       "time",
		LevelKey:      "level",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeCaller:  zapcore.FullCallerEncoder,
	}

	// Create a JSON encoder with the custom EncoderConfig
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	opts := zap.Options{
		Development: logMode.IsDevelopment(),
		TimeEncoder: zapcore.ISO8601TimeEncoder,
		ZapOpts:     []zaplog.Option{zaplog.AddCaller()},
		Encoder:     jsonEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	var err error
	// crtl zap logger
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	// api logrus logger
	loggerOptions := logger.DefaultOptions()
	// align with zap logger
	// zap logger won't use json format in development mode
	// loggerOptions.JSONFormatEnabled = !logMode.IsDevelopment()
	loggerOptions.JSONFormatEnabled = true
	logLevel := "debug"
	if !logMode.IsDevelopment() {
		logLevel = "info"
	}
	err = loggerOptions.SetOutputLevel(logLevel)
	if err != nil {
		setupLog.Error(err, "unable to set log level in logrus")
		os.Exit(1)
	}
	if err = logger.ApplyOptionsToLoggers(&loggerOptions); err != nil {
		setupLog.Error(err, "unable to apply log options to logrus")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()

	options := ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "33405cb8.symphony",
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

	if logsConfigFile != "" {
		obs, err := initLogs(logsConfigFile)
		if err != nil {
			setupLog.Error(err, "unable to initialize logs")
			os.Exit(1)
		}
		defer shutdownLogs(obs)
	}

	apiClientOptions := []utils.ApiClientOption{
		utils.WithCertAuth(apiCertPath),
	}

	if utils.ShouldUseSATokens() {
		apiClientOptions = append(apiClientOptions, utils.WithServiceAccountToken())
	} else {
		apiClientOptions = append(apiClientOptions, utils.WithUserPassword(ctx))
	}

	apiClient, err := utils.NewApiClient(
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
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		ApiClient: apiClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Activation")
		os.Exit(1)
	}
	if err = (&solutioncontrollers.InstanceQueueingReconciler{
		InstanceReconciler: solutioncontrollers.InstanceReconciler{
			Client:                 mgr.GetClient(),
			Scheme:                 mgr.GetScheme(),
			ReconciliationInterval: reconcileInterval,
			DeleteTimeOut:          deleteTimeOut,
			PollInterval:           pollInterval,
			DeleteSyncDelay:        deleteSyncDelay,
			ApiClient:              apiClient,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create instance queueing controller", "controller", "Instance")
		os.Exit(1)
	}

	if err = (&solutioncontrollers.InstancePollingReconciler{
		InstanceReconciler: solutioncontrollers.InstanceReconciler{
			Client:                 mgr.GetClient(),
			Scheme:                 mgr.GetScheme(),
			ReconciliationInterval: reconcileInterval,
			DeleteTimeOut:          deleteTimeOut,
			PollInterval:           pollInterval,
			DeleteSyncDelay:        deleteSyncDelay,
			ApiClient:              apiClient,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create instance polling controller", "controller", "Instance")
		os.Exit(1)
	}

	if err = (&fabriccontrollers.TargetPollingReconciler{
		TargetReconciler: fabriccontrollers.TargetReconciler{
			Client:                 mgr.GetClient(),
			Scheme:                 mgr.GetScheme(),
			ReconciliationInterval: reconcileInterval,
			DeleteTimeOut:          deleteTimeOut,
			PollInterval:           pollInterval,
			DeleteSyncDelay:        deleteSyncDelay,
			ApiClient:              apiClient,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create target polling controller", "controller", "Target")
		os.Exit(1)
	}
	if err = (&fabriccontrollers.TargetQueueingReconciler{
		TargetReconciler: fabriccontrollers.TargetReconciler{
			Client:                 mgr.GetClient(),
			Scheme:                 mgr.GetScheme(),
			ReconciliationInterval: reconcileInterval,
			DeleteTimeOut:          deleteTimeOut,
			PollInterval:           pollInterval,
			DeleteSyncDelay:        deleteSyncDelay,
			ApiClient:              apiClient,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create target queueing controller", "controller", "Target")
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
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		ApiClient: apiClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Catalog")
		os.Exit(1)
	}
	if err = (&actioncontrollers.CatalogEvalReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		ApiClient: apiClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Catalog")
		os.Exit(1)
	}
	if !disableWebhooksServer {
		dynamicclient.SetClient(mgr.GetConfig())
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
		if err = (&workflowv1.Activation{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Activation")
			os.Exit(1)
		}
		if err = (&federationv1.Catalog{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Catalog")
			os.Exit(1)
		}
		if err = (&federationv1.CatalogEvalExpression{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CatalogEvalExpression")
			os.Exit(1)
		}
		if err = (&workflowv1.Campaign{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Campaign")
			os.Exit(1)
		}
		if err = (&monitorv1.Diagnostic{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Diagnostic")
			os.Exit(1)
		}
		if err = commoncontainer.InitCommonContainerWebHook(mgr); err != nil {
			setupLog.Error(err, "unable to Init Common Conainer", "webhook", "Common Conainer")
			os.Exit(1)
		}
		if err = commoncontainer.SetupWebhookWithManager(mgr, &workflowv1.CampaignContainer{}); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CampaignContainer")
			os.Exit(1)
		}
		if err = commoncontainer.SetupWebhookWithManager(mgr, &federationv1.CatalogContainer{}); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CatalogContainer")
			os.Exit(1)
		}
		if err = commoncontainer.SetupWebhookWithManager(mgr, &solutionv1.SolutionContainer{}); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "SolutionContainer")
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
	if err = (&workflowcontrollers.CampaignContainerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CampaignContainer")
		os.Exit(1)
	}
	if err = (&monitorcontrollers.DiagnosticReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Diagnostic")
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

func initLogs(configPath string) (*observability.Observability, error) {
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
	if err := obs.InitLog(config); err != nil {
		return nil, err
	}

	return &obs, nil
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

func shutdownLogs(obs *observability.Observability) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	obs.Shutdown(ctx)
}

func shutdownMetrics(obs *observability.Observability) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	obs.Shutdown(ctx)
}
