/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/k8s/projectors"
	utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	log                      = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

const (
	ENV_NAME      string = "SYMPHONY_AGENT_ADDRESS"
	SINGLE_POD    string = "single-pod"
	SERVICES      string = "services"
	SERVICES_NS   string = "ns-services"
	SERVICES_HNS  string = "hns-services" //TODO: future versions
	componentName string = "P (K8s Target Provider)"
	loggerName    string = "providers.target.k8s"
	k8s           string = "k8s"
)

type K8sTargetProviderConfig struct {
	Name                 string `json:"name"`
	ConfigType           string `json:"configType,omitempty"`
	ConfigData           string `json:"configData,omitempty"`
	Context              string `json:"context,omitempty"`
	InCluster            bool   `json:"inCluster"`
	Projector            string `json:"projector,omitempty"`
	DeploymentStrategy   string `json:"deploymentStrategy,omitempty"`
	DeleteEmptyNamespace bool   `json:"deleteEmptyNamespace"`
	RetryCount           int    `json:"retryCount"`
	RetryIntervalInSec   int    `json:"retryIntervalInSec"`
}

type K8sTargetProvider struct {
	Config        K8sTargetProviderConfig
	Context       *contexts.ManagerContext
	Client        kubernetes.Interface
	DynamicClient dynamic.Interface
}

func K8sTargetProviderConfigFromMap(properties map[string]string) (K8sTargetProviderConfig, error) {
	ret := K8sTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["configType"]; ok {
		ret.ConfigType = v
	}
	if ret.ConfigType == "" {
		ret.ConfigType = "path"
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
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'inCluster' setting of K8s reference provider", v1alpha2.BadConfig)
			}
			ret.InCluster = bVal
		}
	}
	if v, ok := properties["deploymentStrategy"]; ok && v != "" {
		if v != SERVICES && v != SINGLE_POD && v != SERVICES_NS {
			return ret, v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid deployment strategy. Expected: %s (default), %s or %s", SINGLE_POD, SERVICES, SERVICES_NS), v1alpha2.BadConfig)
		}
		ret.DeploymentStrategy = v
	} else {
		ret.DeploymentStrategy = SINGLE_POD
	}
	if v, ok := properties["deleteEmptyNamespace"]; ok {
		val := v
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'deleteEmptyNamespace' setting of K8s reference provider", v1alpha2.BadConfig)
			}
			ret.DeleteEmptyNamespace = bVal
		}
	}
	if v, ok := properties["retryCount"]; ok && v != "" {
		ival, err := strconv.Atoi(v)
		if err != nil {
			return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'retryCount' setting of K8s reference provider", v1alpha2.BadConfig)
		}
		ret.RetryCount = ival
	} else {
		ret.RetryCount = 3
	}
	if v, ok := properties["retryIntervalInSec"]; ok && v != "" {
		ival, err := strconv.Atoi(v)
		if err != nil {
			return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'retryInterval' setting of K8s reference provider", v1alpha2.BadConfig)
		}
		ret.RetryIntervalInSec = ival
	} else {
		ret.RetryIntervalInSec = 2
	}
	return ret, nil
}
func (i *K8sTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := K8sTargetProviderConfigFromMap(properties)
	if err != nil {
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init", componentName), v1alpha2.InitFailed)
	}
	return i.Init(config)
}
func (s *K8sTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *K8sTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"K8s Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfoCtx(ctx, "  P (K8s Target): Init()")

	updateConfig, err := toK8sTargetProviderConfig(config)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (K8s Target): expected K8sTargetProviderConfig - %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to convert to K8sTargetProviderConfig", componentName), v1alpha2.InitFailed)
	}
	i.Config = updateConfig
	var kConfig *rest.Config
	kConfig, err = i.getKubernetesConfig()
	if err != nil {
		log.ErrorfCtx(ctx, "  P (K8s Target): failed to get the cluster config: %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get kubernetes config", componentName), v1alpha2.InitFailed)
	}

	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (K8s Target): failed to create a new clientset: %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create kubernetes client", componentName), v1alpha2.InitFailed)
	}

	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (K8s Target): failed to create a discovery client: %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create dynamic client", componentName), v1alpha2.InitFailed)
	}

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				log.ErrorfCtx(ctx, "  P (K8s Target): failed to create metrics: %+v", err)
			}
		}
	})

	return err
}

func (i *K8sTargetProvider) getKubernetesConfig() (*rest.Config, error) {
	if i.Config.InCluster {
		return rest.InClusterConfig()
	}

	switch i.Config.ConfigType {
	case "path":
		return i.getConfigFromPath()
	case "bytes":
		return i.getConfigFromBytes()
	default:
		return nil, v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and bytes", v1alpha2.BadConfig)
	}
}

func (i *K8sTargetProvider) getConfigFromPath() (*rest.Config, error) {
	if i.Config.ConfigData == "" {
		home := homedir.HomeDir()
		if home == "" {
			return nil, v1alpha2.NewCOAError(nil, "can't locate home directory to read default kubernetes config file. To run in cluster, set inCluster to true", v1alpha2.BadConfig)
		}
		i.Config.ConfigData = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
}

func (i *K8sTargetProvider) getConfigFromBytes() (*rest.Config, error) {
	if i.Config.ConfigData == "" {
		return nil, v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
	}
	return clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
}

func toK8sTargetProviderConfig(config providers.IProviderConfig) (K8sTargetProviderConfig, error) {
	ret := K8sTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	//ret.ConfigPath = providers.LoadEnv(ret.ConfigPath)
	return ret, err
}

func (i *K8sTargetProvider) getDeployment(ctx context.Context, namespace string, name string) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "getDeployment",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (K8s Target Provider): getDeployment scope - %s, name - %s", namespace, name)

	if namespace == "" {
		namespace = "default"
	}

	deployment, err := i.Client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return nil, nil
		}
		log.ErrorfCtx(ctx, "  P (K8s Target Provider): getDeployment %s failed - %s", name, err.Error())
		return nil, err
	}
	components, err := deploymentToComponents(ctx, *deployment)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (K8s Target Provider): getDeployment failed - %s", err.Error())
		return nil, err
	}
	return components, nil
}
func (i *K8sTargetProvider) fillServiceMeta(ctx context.Context, namespace string, name string, component model.ComponentSpec) error {
	if namespace == "" {
		namespace = "default"
	}
	svc, err := i.Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if component.Metadata == nil {
		component.Metadata = make(map[string]string)
	}
	portData, _ := json.Marshal(svc.Spec.Ports)
	component.Metadata["service.ports"] = string(portData)
	component.Metadata["service.type"] = string(svc.Spec.Type)
	if svc.ObjectMeta.Name != name {
		component.Metadata["service.name"] = svc.ObjectMeta.Name
	}
	if component.Metadata["service.type"] == "LoadBalancer" {
		component.Metadata["service.loadBalancerIP"] = svc.Spec.LoadBalancerIP
	}
	for k, v := range svc.ObjectMeta.Annotations {
		component.Metadata["service.annotation."+k] = v
	}
	return nil
}
func (i *K8sTargetProvider) Get(ctx context.Context, dep model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (K8s Target Provider): getting artifacts: %s - %s", dep.Instance.Spec.Scope, dep.Instance.ObjectMeta.Name)

	var components []model.ComponentSpec

	switch i.Config.DeploymentStrategy {
	case "", SINGLE_POD:
		components, err = i.getDeployment(ctx, dep.Instance.Spec.Scope, dep.Instance.ObjectMeta.Name)
		if err != nil {
			log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to get - %s", err.Error())
			err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get components from deployment spec", componentName), v1alpha2.GetComponentSpecFailed)
			return nil, err
		}
	case SERVICES, SERVICES_NS:
		components = make([]model.ComponentSpec, 0)
		scope := dep.Instance.Spec.Scope
		if i.Config.DeploymentStrategy == SERVICES_NS {
			scope = dep.Instance.ObjectMeta.Name
		}
		slice := dep.GetComponentSlice()
		for _, component := range slice {
			var cComponents []model.ComponentSpec
			cComponents, err = i.getDeployment(ctx, scope, component.Name)
			if err != nil {
				log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to get deployment: %s", err.Error())
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get components from deployment spec", componentName), v1alpha2.GetComponentSpecFailed)
				return nil, err
			}
			if len(cComponents) > 1 {
				log.DebugfCtx(ctx, "  P (K8s Target Provider): can't read multiple components %s", err.Error())
				err = v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: can't read multiple components when %s strategy or %s strategy is used", componentName, SERVICES, SERVICES_NS), v1alpha2.GetComponentSpecFailed)
				return nil, err
			}
			if len(cComponents) == 1 {
				serviceName := cComponents[0].Name

				if cComponents[0].Metadata != nil {
					if v, ok := cComponents[0].Metadata["service.name"]; ok && v != "" {
						serviceName = v
					}
				}
				if cComponents[0].Metadata == nil {
					cComponents[0].Metadata = make(map[string]string)
				}

				err = i.fillServiceMeta(ctx, scope, serviceName, cComponents[0])
				if err != nil {
					log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to get: %s", err.Error())
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to fill service meta data", componentName), v1alpha2.GetComponentSpecFailed)
					return nil, err
				}
				components = append(components, cComponents...)
			}
		}
	}

	return components, nil
}
func (i *K8sTargetProvider) removeService(ctx context.Context, namespace string, serviceName string) error {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "removeService",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (K8s Target Provider): removeService namespace - %s, serviceName - %s", namespace, serviceName)

	if namespace == "" {
		namespace = "default"
	}

	svc, err := i.Client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err == nil && svc != nil {
		foregroundDeletion := metav1.DeletePropagationForeground
		err = i.Client.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{PropagationPolicy: &foregroundDeletion})
		if err != nil {
			if !k8s_errors.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}
func (i *K8sTargetProvider) removeDeployment(ctx context.Context, namespace string, name string) error {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "removeDeployment",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (K8s Target Provider): removeDeployment namespace - %s, name - %s", namespace, name)

	if namespace == "" {
		namespace = "default"
	}

	foregroundDeletion := metav1.DeletePropagationForeground
	logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Starting remove deployment under namespace - %s, name - %s", namespace, name)
	err = i.Client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &foregroundDeletion})
	if err != nil {
		if !k8s_errors.IsNotFound(err) {
			return err
		}
	}
	logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Triggered remove deployment under namespace - %s, name - %s", namespace, name)

	return nil
}
func (i *K8sTargetProvider) removeNamespace(ctx context.Context, namespace string, retryCount int, retryIntervalInSec int) error {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "removeNamespace",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (K8s Target Provider): removeNamespace namespace - %s", namespace)

	_, err = i.Client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if namespace == "" || namespace == "default" {
		return nil
	}

	resourceCount := make(map[string]int)
	count := 0
	for {
		count++
		podList, _ := i.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		if len(podList.Items) == 0 || count == retryCount {
			resourceCount["pod"] = len(podList.Items)
			break
		}
		time.Sleep(time.Second * time.Duration(retryIntervalInSec))
	}

	deploymentList, err := i.Client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["deployment"] = len(deploymentList.Items)

	serviceList, err := i.Client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["service"] = len(serviceList.Items)

	replicasetList, err := i.Client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["replicaset"] = len(replicasetList.Items)

	statefulsetList, err := i.Client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["statefulset"] = len(statefulsetList.Items)

	daemonsetList, err := i.Client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["daemonset"] = len(daemonsetList.Items)

	jobList, err := i.Client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["job"] = len(jobList.Items)

	isEmpty := true
	for resource, count := range resourceCount {
		if count != 0 {
			log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to delete %s namespace as resource %s is not empty", namespace, resource)
			isEmpty = false
			break
		}
	}

	if isEmpty {
		logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Starting remove namespace - %s", namespace)
		err = i.Client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Triggered remove namespace - %s", namespace)
	}
	return nil
}
func (i *K8sTargetProvider) createNamespace(ctx context.Context, namespace string) error {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "createNamespace",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (K8s Target Provider): removeDeployment namespace - %s", namespace)

	if namespace == "" || namespace == "default" {
		return nil
	}
	_, err = i.Client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})

	if err != nil {
		if k8s_errors.IsNotFound(err) {
			logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Starting create namespace - %s", namespace)
			_, err = i.Client.CoreV1().Namespaces().Create(ctx, &apiv1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Triggered create namespace - %s", namespace)
		} else {
			return err
		}
	}
	return nil
}
func (i *K8sTargetProvider) upsertDeployment(ctx context.Context, namespace string, name string, deployment *v1.Deployment) error {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "upsertDeployment",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (K8s Target Provider): upsertDeployment namespace - %s, name - %s", namespace, name)

	if namespace == "" {
		namespace = "default"
	}

	existing, err := i.Client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}
	if k8s_errors.IsNotFound(err) {
		logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Starting create deployment under namespace - %s, name - %s", namespace, name)
		_, err = i.Client.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	} else {
		deployment.ResourceVersion = existing.ResourceVersion
		logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Starting update deployment under namespace - %s, name - %s", namespace, name)
		_, err = i.Client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	}
	if err != nil {
		return err
	}
	return nil
}
func (i *K8sTargetProvider) upsertService(ctx context.Context, namespace string, name string, service *apiv1.Service) error {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "upsertService",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (K8s Target Provider): upsertService namespace - %s, name - %s", namespace, name)

	if namespace == "" {
		namespace = "default"
	}

	existing, err := i.Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}
	if k8s_errors.IsNotFound(err) {
		logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Starting create service under namespace - %s, name - %s", namespace, name)
		_, err = i.Client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	} else {
		service.ResourceVersion = existing.ResourceVersion
		logger.GetUserAuditsLogger().InfofCtx(ctx, "  P (K8s Target Provider): Starting update service under namespace - %s, name - %s", namespace, name)
		_, err = i.Client.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{})
	}
	if err != nil {
		return err
	}
	return nil
}
func (i *K8sTargetProvider) deployComponents(ctx context.Context, namespace string, name string, metadata map[string]string, components []model.ComponentSpec, projector IK8sProjector, instanceName string) error {
	var err error = nil
	log.InfofCtx(ctx, "  P (K8s Target Provider): deployComponents namespace - %s, name - %s", namespace, name)

	if namespace == "" {
		namespace = "default"
	}

	deployment, err := componentsToDeployment(ctx, namespace, name, metadata, components, instanceName)
	if projector != nil {
		err = projector.ProjectDeployment(namespace, name, metadata, components, deployment)
		if err != nil {
			log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to project deployment: %s", err.Error())
			return err
		}
	}
	if err != nil {
		log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to apply: %s", err.Error())
		return err
	}
	service, err := metadataToService(ctx, namespace, name, metadata)
	if err != nil {
		log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to apply (convert): %s", err.Error())
		return err
	}
	if projector != nil {
		err = projector.ProjectService(namespace, name, metadata, service)
		if err != nil {
			log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to project service: %s", err.Error())
			return err
		}
	}

	log.DebugCtx(ctx, "  P (K8s Target Provider): checking namespace")
	err = i.createNamespace(ctx, namespace)
	if err != nil {
		log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to create namespace: %s", err.Error())
		return err
	}

	log.DebugCtx(ctx, "  P (K8s Target Provider): creating deployment")
	err = i.upsertDeployment(ctx, namespace, name, deployment)
	if err != nil {
		log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to apply (API): %s", err.Error())
		return err
	}

	if service != nil {
		log.DebugCtx(ctx, "  P (K8s Target Provider): creating service")
		err = i.upsertService(ctx, namespace, service.Name, service)
		if err != nil {
			log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to apply (service): %s", err.Error())
			return err
		}
	}
	return nil
}
func (i *K8sTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: i.Config.DeploymentStrategy == SERVICES,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{model.ContainerImage},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "container.*", IgnoreCase: true, SkipIfMissing: false},
				{Name: "env.*", IgnoreCase: true, SkipIfMissing: true},
			},
		},
		SidecarValidationRule: model.ComponentValidationRule{
			RequiredProperties: []string{model.ContainerImage},
			OptionalProperties: []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "container.*", IgnoreCase: true, SkipIfMissing: false},
				{Name: "env.*", IgnoreCase: true, SkipIfMissing: true},
			},
		},
	}
}
func (i *K8sTargetProvider) Apply(ctx context.Context, dep model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, "  P (K8s Target Provider): applying artifacts: %s - %s", dep.Instance.Spec.Scope, dep.Instance.ObjectMeta.Name)

	functionName := observ_utils.GetFunctionName()
	applyTime := time.Now().UTC()
	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (K8s Target Provider): failed to validate components, error: %v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: the rule validation failed", componentName), v1alpha2.ValidateFailed)
		providerOperationMetrics.ProviderOperationErrors(
			k8s,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.UpdateOperationType,
			v1alpha2.ValidateFailed.String(),
		)
		return nil, err
	}
	if isDryRun {
		return nil, nil
	}

	ret := step.PrepareResultMap()

	projector, err := createProjector(i.Config.Projector)
	if err != nil {
		log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to create projector: %s", err.Error())
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to create projector", componentName), v1alpha2.CreateProjectorFailed)
		providerOperationMetrics.ProviderOperationErrors(
			k8s,
			functionName,
			metrics.K8SProjectorOperation,
			metrics.UpdateOperationType,
			v1alpha2.CreateProjectorFailed.String(),
		)
		return ret, err
	}

	switch i.Config.DeploymentStrategy {
	case "", SINGLE_POD:
		updated := step.GetUpdatedComponents()
		if len(updated) > 0 {
			err = i.deployComponents(ctx, dep.Instance.Spec.Scope, dep.Instance.ObjectMeta.Name, dep.Instance.Spec.Metadata, components, projector, dep.Instance.ObjectMeta.Name)
			if err != nil {
				log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to apply components: %s", err.Error())
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to deploy components", componentName), v1alpha2.K8sDeploymentFailed)
				providerOperationMetrics.ProviderOperationErrors(
					k8s,
					functionName,
					metrics.K8SDeploymentOperation,
					metrics.UpdateOperationType,
					v1alpha2.K8sDeploymentFailed.String(),
				)
				return ret, err
			}
			providerOperationMetrics.ProviderOperationLatency(
				applyTime,
				k8s,
				metrics.K8SDeploymentOperation,
				metrics.UpdateOperationType,
				functionName,
			)
		}
		deleteTime := time.Now().UTC()
		deleted := step.GetDeletedComponents()
		if len(deleted) > 0 {
			serviceName := dep.Instance.ObjectMeta.Name
			if v, ok := dep.Instance.Spec.Metadata["service.name"]; ok && v != "" {
				serviceName = v
			}
			err = i.removeService(ctx, dep.Instance.Spec.Scope, serviceName)
			if err != nil {
				log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to remove service: %s", err.Error())
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to remove k8s service", componentName), v1alpha2.K8sRemoveServiceFailed)
				providerOperationMetrics.ProviderOperationErrors(
					k8s,
					functionName,
					metrics.K8SRemoveServiceOperation,
					metrics.DeleteOperationType,
					v1alpha2.K8sRemoveServiceFailed.String(),
				)
				return ret, err
			}
			err = i.removeDeployment(ctx, dep.Instance.Spec.Scope, dep.Instance.ObjectMeta.Name)
			if err != nil {
				log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to remove deployment: %s", err.Error())
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to remove k8s deployment", componentName), v1alpha2.K8sRemoveDeploymentFailed)
				providerOperationMetrics.ProviderOperationErrors(
					k8s,
					functionName,
					metrics.K8SRemoveDeploymentOperation,
					metrics.DeleteOperationType,
					v1alpha2.K8sRemoveDeploymentFailed.String(),
				)
				return ret, err
			}
			if i.Config.DeleteEmptyNamespace {
				err = i.removeNamespace(ctx, dep.Instance.Spec.Scope, i.Config.RetryCount, i.Config.RetryIntervalInSec)
				if err != nil {
					log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to remove namespace: %s", err.Error())
				}
			}
			providerOperationMetrics.ProviderOperationLatency(
				deleteTime,
				k8s,
				metrics.K8SDeploymentOperation,
				metrics.DeleteOperationType,
				functionName,
			)
		}
	case SERVICES, SERVICES_NS:
		updated := step.GetUpdatedComponents()
		if len(updated) > 0 {
			scope := dep.Instance.Spec.Scope
			if i.Config.DeploymentStrategy == SERVICES_NS {
				scope = dep.Instance.ObjectMeta.Name
			}
			for _, component := range components {
				if dep.Instance.Spec.Metadata != nil {
					if v, ok := dep.Instance.Spec.Metadata[ENV_NAME]; ok && v != "" {
						if component.Metadata == nil {
							component.Metadata = make(map[string]string)
						}
						component.Metadata[ENV_NAME] = v
					}
				}
				err = i.deployComponents(ctx, scope, component.Name, component.Metadata, []model.ComponentSpec{component}, projector, dep.Instance.ObjectMeta.Name)
				if err != nil {
					log.DebugfCtx(ctx, "  P (K8s Target Provider): failed to apply components: %s", err.Error())
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to deploy components", componentName), v1alpha2.K8sDeploymentFailed)
					providerOperationMetrics.ProviderOperationErrors(
						k8s,
						functionName,
						metrics.K8SDeploymentOperation,
						metrics.UpdateOperationType,
						v1alpha2.K8sDeploymentFailed.String(),
					)
					return ret, err
				}
			}
			providerOperationMetrics.ProviderOperationLatency(
				applyTime,
				k8s,
				metrics.K8SDeploymentOperation,
				metrics.UpdateOperationType,
				functionName,
			)
		}
		deleteTime := time.Now().UTC()
		deleted := step.GetDeletedComponents()
		if len(deleted) > 0 {
			scope := dep.Instance.Spec.Scope
			if i.Config.DeploymentStrategy == SERVICES_NS {
				scope = dep.Instance.ObjectMeta.Name
			}
			for _, component := range deleted {
				serviceName := component.Name
				if component.Metadata != nil {
					if v, ok := component.Metadata["service.name"]; ok {
						serviceName = v
					}
				}
				err = i.removeService(ctx, scope, serviceName)
				if err != nil {
					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.DeleteFailed,
						Message: err.Error(),
					}
					log.DebugfCtx(ctx, "P (K8s Target Provider): failed to remove service: %s", err.Error())
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to remove k8s service", componentName), v1alpha2.K8sRemoveServiceFailed)
					providerOperationMetrics.ProviderOperationErrors(
						k8s,
						functionName,
						metrics.K8SRemoveServiceOperation,
						metrics.DeleteOperationType,
						v1alpha2.K8sRemoveServiceFailed.String(),
					)
					return ret, err
				}
				err = i.removeDeployment(ctx, scope, component.Name)
				if err != nil {
					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.DeleteFailed,
						Message: err.Error(),
					}
					log.DebugfCtx(ctx, "P (K8s Target Provider): failed to remove deployment: %s", err.Error())
					err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to remove k8s deployment", componentName), v1alpha2.K8sRemoveDeploymentFailed)
					providerOperationMetrics.ProviderOperationErrors(
						k8s,
						functionName,
						metrics.K8SRemoveDeploymentOperation,
						metrics.DeleteOperationType,
						v1alpha2.K8sRemoveDeploymentFailed.String(),
					)
					return ret, err
				}
				if i.Config.DeleteEmptyNamespace {
					err = i.removeNamespace(ctx, dep.Instance.Spec.Scope, i.Config.RetryCount, i.Config.RetryIntervalInSec)
					if err != nil {
						log.DebugfCtx(ctx, "P (K8s Target Provider): failed to remove namespace: %s", err.Error())
					}
				}
			}
			providerOperationMetrics.ProviderOperationLatency(
				deleteTime,
				k8s,
				metrics.K8SDeploymentOperation,
				metrics.DeleteOperationType,
				functionName,
			)
		}
	}
	err = nil
	return ret, nil
}
func deploymentToComponents(ctx context.Context, deployment v1.Deployment) ([]model.ComponentSpec, error) {
	components := make([]model.ComponentSpec, 0)
	for _, c := range deployment.Spec.Template.Spec.Containers {
		key := fmt.Sprintf("%s.sidecar_of", c.Name)
		if deployment.Spec.Template.ObjectMeta.Labels[key] != "" {
			// Skip sidecar containers for now
			continue
		}
		component := makeComponentSpec(c)
		components = append(components, component)
	}

	for _, c := range deployment.Spec.Template.Spec.Containers {
		key := fmt.Sprintf("%s.sidecar_of", c.Name)
		componentName := deployment.Spec.Template.ObjectMeta.Labels[key]
		if componentName != "" {
			for i, component := range components {
				if component.Name == componentName {
					sidecar := makeComponentSpec(c)
					components[i].Sidecars = append(components[i].Sidecars, convertComponentSpecToSidecar(sidecar))
				}
			}
		}
	}
	componentsJson, _ := json.Marshal(components)
	log.DebugfCtx(ctx, "  P (K8s Target Provider): deploymentToComponents - components: %s", string(componentsJson))
	return components, nil
}
func convertComponentSpecToSidecar(c model.ComponentSpec) model.SidecarSpec {
	sidecar := model.SidecarSpec{
		Name:       c.Name,
		Type:       c.Type,
		Properties: c.Properties,
	}
	return sidecar
}
func makeComponentSpec(c apiv1.Container) model.ComponentSpec {
	component := model.ComponentSpec{
		Name:       c.Name,
		Properties: make(map[string]interface{}),
	}
	component.Properties[model.ContainerImage] = c.Image
	policy := string(c.ImagePullPolicy)
	if policy != "" {
		component.Properties["container.imagePullPolicy"] = policy
	}
	if len(c.Ports) > 0 {
		ports, _ := json.Marshal(c.Ports)
		component.Properties["container.ports"] = string(ports)
	}
	if len(c.Args) > 0 {
		args, _ := json.Marshal(c.Args)
		component.Properties["container.args"] = string(args)
	}
	if len(c.Command) > 0 {
		commands, _ := json.Marshal(c.Command)
		component.Properties["container.commands"] = string(commands)
	}
	resources, _ := json.Marshal(c.Resources)
	if string(resources) != "{}" {
		component.Properties["container.resources"] = string(resources)
	}
	if len(c.VolumeMounts) > 0 {
		volumeMounts, _ := json.Marshal(c.VolumeMounts)
		component.Properties["container.volumeMounts"] = string(volumeMounts)
	}
	if len(c.Env) > 0 {
		for _, e := range c.Env {
			component.Properties["env."+e.Name] = e.Value
		}
	}
	return component
}
func metadataToService(ctx context.Context, namespace string, name string, metadata map[string]string) (*apiv1.Service, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	servicePorts := make([]apiv1.ServicePort, 0)

	if v, ok := metadata["service.ports"]; ok && v != "" {
		log.DebugfCtx(ctx, "  P (K8s Target Provider): metadataToService - service ports: %s", v)
		e := json.Unmarshal([]byte(v), &servicePorts)
		if e != nil {
			log.ErrorfCtx(ctx, "  P (K8s Target Provider): metadataToService - unmarshal: %v", e)
			return nil, e
		}
	} else {
		return nil, nil
	}

	serviceName := utils.ReadString(metadata, "service.name", name)
	serviceType := utils.ReadString(metadata, "service.type", "ClusterIP")

	service := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: apiv1.ServiceSpec{
			Type:  apiv1.ServiceType(serviceType),
			Ports: servicePorts,
			Selector: map[string]string{
				"app": name,
			},
		},
	}
	if _, ok := metadata["service.loadBalancerIP"]; ok {
		service.Spec.LoadBalancerIP = utils.ReadString(metadata, "service.loadBalancerIP", "")
	}
	annotations := utils.CollectStringMap(metadata, "service.annotation.")
	if len(annotations) > 0 {
		service.ObjectMeta.Annotations = make(map[string]string)
		for k, v := range annotations {
			service.ObjectMeta.Annotations[k[19:]] = v
		}
	}
	return &service, nil
}
func int32Ptr(i int32) *int32 { return &i }

func componentsToDeployment(ctx context.Context, scope string, name string, metadata map[string]string, components []model.ComponentSpec, instanceName string) (*v1.Deployment, error) {
	deployment := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.DeploymentSpec{
			Replicas: int32Ptr(utils.ReadInt32(metadata, "deployment.replicas", 1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{},
				},
			},
		},
	}

	for _, c := range components {
		container, err := createContainerSpec(c.Name, c.Properties, metadata)
		if err != nil {
			return nil, err
		}
		deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, *container)
		if len(c.Sidecars) > 0 {
			for _, sidecar := range c.Sidecars {
				container, err := createContainerSpec(sidecar.Name, sidecar.Properties, metadata)
				if err != nil {
					return nil, err
				}
				if deployment.Spec.Template.ObjectMeta.Labels == nil {
					deployment.Spec.Template.ObjectMeta.Labels = make(map[string]string)
				}
				key := fmt.Sprintf("%s.sidecar_of", sidecar.Name)
				deployment.Spec.Template.ObjectMeta.Labels[key] = c.Name
				deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, *container)
			}
		}
	}
	if v, ok := metadata["deployment.imagePullSecrets"]; ok && v != "" {
		secrets := make([]apiv1.LocalObjectReference, 0)
		e := json.Unmarshal([]byte(v), &secrets)
		if e != nil {
			return nil, e
		}
		deployment.Spec.Template.Spec.ImagePullSecrets = secrets
	}
	if v, ok := metadata["pod.volumes"]; ok && v != "" {
		volumes := make([]apiv1.Volume, 0)
		e := json.Unmarshal([]byte(v), &volumes)
		if e != nil {
			return nil, e
		}
		deployment.Spec.Template.Spec.Volumes = volumes
	}
	if v, ok := metadata["deployment.nodeSelector"]; ok && v != "" {
		selector := make(map[string]string)
		e := json.Unmarshal([]byte(v), &selector)
		if e != nil {
			return nil, e
		}
		deployment.Spec.Template.Spec.NodeSelector = selector
	}

	data, _ := json.Marshal(deployment)
	log.DebugCtx(ctx, string(data))

	return &deployment, nil
}

func createContainerSpec(name string, properties map[string]interface{}, metadata map[string]string) (*apiv1.Container, error) {
	ports := make([]apiv1.ContainerPort, 0)
	if v, ok := properties["container.ports"].(string); ok && v != "" {
		e := json.Unmarshal([]byte(v), &ports)
		if e != nil {
			return nil, e
		}
	}
	container := &apiv1.Container{
		Name:            name,
		Image:           utils.FormatAsString(properties[model.ContainerImage]),
		Ports:           ports,
		ImagePullPolicy: apiv1.PullPolicy(utils.ReadStringFromMapCompat(properties, "container.imagePullPolicy", "Always")),
	}
	if v, ok := properties["container.args"]; ok && v != "" {
		args := make([]string, 0)
		e := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &args)
		if e != nil {
			return nil, e
		}
		container.Args = args
	}
	if v, ok := properties["container.commands"]; ok && v != "" {
		cmds := make([]string, 0)
		e := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &cmds)
		if e != nil {
			return nil, e
		}
		container.Command = cmds
	}
	if v, ok := properties["container.resources"]; ok && v != "" {
		res := apiv1.ResourceRequirements{}
		e := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &res)
		if e != nil {
			return nil, e
		}
		container.Resources = res
	}
	if v, ok := properties["container.volumeMounts"]; ok && v != "" {
		mounts := make([]apiv1.VolumeMount, 0)
		e := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &mounts)
		if e != nil {
			return nil, e
		}
		container.VolumeMounts = mounts
	}
	for k, v := range properties {
		// Transitioning from map[string]string to map[string]interface{}
		// for now we'll assume that all relevant values are strings till we
		// refactor the code to handle the new format
		sv := fmt.Sprintf("%v", v)
		if strings.HasPrefix(k, "env.") {
			if container.Env == nil {
				container.Env = make([]apiv1.EnvVar, 0)
			}
			container.Env = append(container.Env, apiv1.EnvVar{
				Name:  k[4:],
				Value: sv,
			})
		}
	}
	agentName := metadata[ENV_NAME]
	if agentName != "" {
		if container.Env == nil {
			container.Env = make([]apiv1.EnvVar, 0)
		}
		container.Env = append(container.Env, apiv1.EnvVar{
			Name:  ENV_NAME,
			Value: agentName + ".default.svc.cluster.local", //agent is currently always installed under deault
		})
	}
	return container, nil
}

func createProjector(projector string) (IK8sProjector, error) {
	switch projector {
	case "noop":
		return &projectors.NoOpProjector{}, nil
	case "":
		return nil, nil
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("project type '%s' is unsupported", projector), v1alpha2.BadConfig)
}

type IK8sProjector interface {
	ProjectDeployment(scope string, name string, metadata map[string]string, components []model.ComponentSpec, deployment *v1.Deployment) error
	ProjectService(scope string, name string, metadata map[string]string, service *apiv1.Service) error
}
