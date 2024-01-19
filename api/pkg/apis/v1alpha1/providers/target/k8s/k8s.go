/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/k8s/projectors"
	utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"go.opentelemetry.io/otel/trace"
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

var log = logger.NewLogger("coa.runtime")

const (
	ENV_NAME     string = "SYMPHONY_AGENT_ADDRESS"
	SINGLE_POD   string = "single-pod"
	SERVICES     string = "services"
	SERVICES_NS  string = "ns-services"
	SERVICES_HNS string = "hns-services" //TODO: future versions
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
		return err
	}
	return i.Init(config)
}
func (s *K8sTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *K8sTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan(
		"K8s Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Info("  P (K8s Target): Init()")

	updateConfig, err := toK8sTargetProviderConfig(config)
	if err != nil {
		log.Errorf("  P (K8s Target): expected K8sTargetProviderConfig - %+v", err)
		return errors.New("expected K8sTargetProviderConfig")
	}
	i.Config = updateConfig
	var kConfig *rest.Config
	if i.Config.InCluster {
		kConfig, err = rest.InClusterConfig()
	} else {
		switch i.Config.ConfigType {
		case "path":
			if i.Config.ConfigData == "" {
				if home := homedir.HomeDir(); home != "" {
					i.Config.ConfigData = filepath.Join(home, ".kube", "config")
				} else {
					err = v1alpha2.NewCOAError(nil, "can't locate home direction to read default kubernetes config file, to run in cluster, set inCluster config setting to true", v1alpha2.BadConfig)
					log.Errorf("  P (K8s Target): %+v", err)
					return err
				}
			}
			kConfig, err = clientcmd.BuildConfigFromFlags("", i.Config.ConfigData)
		case "bytes":
			if i.Config.ConfigData != "" {
				kConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(i.Config.ConfigData))
				if err != nil {
					log.Errorf("  P (K8s Target): failed to get RESTconfg:  %+v", err)
					return err
				}
			} else {
				err = v1alpha2.NewCOAError(nil, "config data is not supplied", v1alpha2.BadConfig)
				log.Errorf("  P (K8s Target): %+v", err)
				return err
			}
		default:
			err = v1alpha2.NewCOAError(nil, "unrecognized config type, accepted values are: path and inline", v1alpha2.BadConfig)
			log.Errorf("  P (K8s Target): %+v", err)
			return err
		}
	}
	if err != nil {
		log.Errorf("  P (K8s Target): failed to get the cluster config: %+v", err)
		return err
	}
	i.Client, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		log.Errorf("  P (K8s Target): failed to create a new clientset: %+v", err)
		return err
	}
	i.DynamicClient, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		log.Errorf("  P (K8s Target): failed to create a discovery client: %+v", err)
		return err
	}
	return nil
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

func (i *K8sTargetProvider) getDeployment(ctx context.Context, scope string, name string) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "getDeployment",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof("  P (K8s Target Provider): getDeployment scope - %s, name - %s, traceId: %s", scope, name, span.SpanContext().TraceID().String())

	deployment, err := i.Client.AppsV1().Deployments(scope).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return nil, nil
		}
		log.Errorf("  P (K8s Target Provider): getDeployment %s failed - %s", name, err.Error())
		return nil, err
	}
	components, err := deploymentToComponents(*deployment)
	if err != nil {
		log.Errorf("  P (K8s Target Provider): getDeployment failed - %s", err.Error())
		return nil, err
	}
	return components, nil
}
func (i *K8sTargetProvider) fillServiceMeta(ctx context.Context, scope string, name string, component model.ComponentSpec) error {
	svc, err := i.Client.CoreV1().Services(scope).Get(ctx, name, metav1.GetOptions{})
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
	log.Infof("  P (K8s Target Provider): getting artifacts: %s - %s, traceId: %s", dep.Instance.Scope, dep.Instance.Name, span.SpanContext().TraceID().String())

	var components []model.ComponentSpec

	switch i.Config.DeploymentStrategy {
	case "", SINGLE_POD:
		components, err = i.getDeployment(ctx, dep.Instance.Scope, dep.Instance.Name)
		if err != nil {
			log.Debugf("  P (K8s Target Provider): failed to get - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return nil, err
		}
	case SERVICES, SERVICES_NS:
		components = make([]model.ComponentSpec, 0)
		scope := dep.Instance.Scope
		if i.Config.DeploymentStrategy == SERVICES_NS {
			scope = dep.Instance.Name
		}
		slice := dep.GetComponentSlice()
		for _, component := range slice {
			var cComponents []model.ComponentSpec
			cComponents, err = i.getDeployment(ctx, scope, component.Name)
			if err != nil {
				log.Debugf("  P (K8s Target Provider) - failed to get: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
				return nil, err
			}
			if len(cComponents) > 1 {
				err = v1alpha2.NewCOAError(nil, fmt.Sprintf("can't read multiple components when %s strategy or %s strategy is used", SERVICES, SERVICES_NS), v1alpha2.InternalError)
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
					log.Debugf("failed to get: %s", err.Error())
					return nil, err
				}
				components = append(components, cComponents...)
			}
		}
	}

	return components, nil
}
func (i *K8sTargetProvider) removeService(ctx context.Context, scope string, serviceName string) error {
	_, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "removeService",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof("  P (K8s Target Provider): removeService scope - %s, serviceName - %s", scope, serviceName)

	svc, err := i.Client.CoreV1().Services(scope).Get(ctx, serviceName, metav1.GetOptions{})
	if err == nil && svc != nil {
		foregroundDeletion := metav1.DeletePropagationForeground
		err = i.Client.CoreV1().Services(scope).Delete(ctx, serviceName, metav1.DeleteOptions{PropagationPolicy: &foregroundDeletion})
		if err != nil {
			if !k8s_errors.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}
func (i *K8sTargetProvider) removeDeployment(ctx context.Context, scope string, name string) error {
	_, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "removeDeployment",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof("  P (K8s Target Provider): removeDeployment scope - %s, name - %s", scope, name)

	foregroundDeletion := metav1.DeletePropagationForeground
	err = i.Client.AppsV1().Deployments(scope).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &foregroundDeletion})
	if err != nil {
		if !k8s_errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}
func (i *K8sTargetProvider) removeNamespace(ctx context.Context, scope string, retryCount int, retryIntervalInSec int) error {
	_, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "removeNamespace",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof("  P (K8s Target Provider): removeNamespace scope - %s, traceId: %s", scope, span.SpanContext().TraceID().String())

	_, err = i.Client.CoreV1().Namespaces().Get(ctx, scope, metav1.GetOptions{})
	if err != nil {
		return err
	}

	resourceCount := make(map[string]int)
	count := 0
	for {
		count++
		podList, _ := i.Client.CoreV1().Pods(scope).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		if len(podList.Items) == 0 || count == retryCount {
			resourceCount["pod"] = len(podList.Items)
			break
		}
		time.Sleep(time.Second * time.Duration(retryIntervalInSec))
	}

	deploymentList, err := i.Client.AppsV1().Deployments(scope).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["deployment"] = len(deploymentList.Items)

	serviceList, err := i.Client.CoreV1().Services(scope).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["service"] = len(serviceList.Items)

	replicasetList, err := i.Client.AppsV1().ReplicaSets(scope).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["replicaset"] = len(replicasetList.Items)

	statefulsetList, err := i.Client.AppsV1().StatefulSets(scope).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["statefulset"] = len(statefulsetList.Items)

	daemonsetList, err := i.Client.AppsV1().DaemonSets(scope).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["daemonset"] = len(daemonsetList.Items)

	jobList, err := i.Client.BatchV1().Jobs(scope).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resourceCount["job"] = len(jobList.Items)

	isEmpty := true
	for resource, count := range resourceCount {
		if count != 0 {
			log.Debugf("  P (K8s Target Provider): failed to delete %s namespace as resource %s is not empty", scope, resource)
			isEmpty = false
			break
		}
	}

	if isEmpty {
		err = i.Client.CoreV1().Namespaces().Delete(ctx, scope, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
func (i *K8sTargetProvider) createNamespace(ctx context.Context, scope string) error {
	_, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "createNamespace",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof("  P (K8s Target Provider): removeDeployment scope - %s", scope)

	_, err = i.Client.CoreV1().Namespaces().Get(ctx, scope, metav1.GetOptions{})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			_, err = i.Client.CoreV1().Namespaces().Create(ctx, &apiv1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: scope,
				},
			}, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
func (i *K8sTargetProvider) upsertDeployment(ctx context.Context, scope string, name string, deployment *v1.Deployment) error {
	_, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "upsertDeployment",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof("  P (K8s Target Provider): upsertDeployment scope - %s, name - %s, traceId: %s", scope, name, span.SpanContext().TraceID().String())

	existing, err := i.Client.AppsV1().Deployments(scope).Get(ctx, name, metav1.GetOptions{})
	if err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}
	if k8s_errors.IsNotFound(err) {
		_, err = i.Client.AppsV1().Deployments(scope).Create(ctx, deployment, metav1.CreateOptions{})
	} else {
		deployment.ResourceVersion = existing.ResourceVersion
		_, err = i.Client.AppsV1().Deployments(scope).Update(ctx, deployment, metav1.UpdateOptions{})
	}
	if err != nil {
		return err
	}
	return nil
}
func (i *K8sTargetProvider) upsertService(ctx context.Context, scope string, name string, service *apiv1.Service) error {
	_, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "upsertService",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof("  P (K8s Target Provider): upsertService scope - %s, name - %s, traceId: %s", scope, name, span.SpanContext().TraceID().String())

	existing, err := i.Client.CoreV1().Services(scope).Get(ctx, name, metav1.GetOptions{})
	if err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}
	if k8s_errors.IsNotFound(err) {
		_, err = i.Client.CoreV1().Services(scope).Create(ctx, service, metav1.CreateOptions{})
	} else {
		service.ResourceVersion = existing.ResourceVersion
		_, err = i.Client.CoreV1().Services(scope).Update(ctx, service, metav1.UpdateOptions{})
	}
	if err != nil {
		return err
	}
	return nil
}
func (i *K8sTargetProvider) deployComponents(ctx context.Context, span trace.Span, scope string, name string, metadata map[string]string, components []model.ComponentSpec, projector IK8sProjector, instanceName string) error {
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof("  P (K8s Target Provider): deployComponents scope - %s, name - %s, traceId: %s", scope, name, span.SpanContext().TraceID().String())

	deployment, err := componentsToDeployment(scope, name, metadata, components, instanceName)
	if projector != nil {
		err = projector.ProjectDeployment(scope, name, metadata, components, deployment)
		if err != nil {
			log.Debugf("  P (K8s Target Provider): failed to project deployment: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return err
		}
	}
	if err != nil {
		log.Debugf("  P (K8s Target Provider): failed to apply: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
		return err
	}
	service, err := metadataToService(scope, name, metadata)
	if err != nil {
		log.Debugf("  P (K8s Target Provider): failed to apply (convert): %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
		return err
	}
	if projector != nil {
		err = projector.ProjectService(scope, name, metadata, service)
		if err != nil {
			log.Debugf("  P (K8s Target Provider): failed to project service: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return err
		}
	}

	log.Debug("  P (K8s Target Provider): checking namespace")
	err = i.createNamespace(ctx, scope)
	if err != nil {
		log.Debugf("  P (K8s Target Provider): failed to create namespace: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
		return err
	}

	log.Debug("  P (K8s Target Provider): creating deployment")
	err = i.upsertDeployment(ctx, scope, name, deployment)
	if err != nil {
		log.Debugf("  P (K8s Target Provider): failed to apply (API): %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
		return err
	}

	if service != nil {
		log.Debug("  P (K8s Target Provider): creating service")
		err = i.upsertService(ctx, scope, service.Name, service)
		if err != nil {
			log.Debugf("  P (K8s Target Provider): failed to apply (service): %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return err
		}
	}
	return nil
}
func (*K8sTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{model.ContainerImage},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
		ChangeDetectionProperties: []model.PropertyDesc{
			{Name: model.ContainerImage, IgnoreCase: true, SkipIfMissing: false},
			{Name: "env.*", IgnoreCase: true, SkipIfMissing: true},
		},
	}
}
func (i *K8sTargetProvider) Apply(ctx context.Context, dep model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("K8s Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Infof("  P (K8s Target Provider): applying artifacts: %s - %s, traceId: %s", dep.Instance.Scope, dep.Instance.Name, span.SpanContext().TraceID().String())

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		log.Errorf("  P (K8s Target Provider): failed to validate components, error: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	if isDryRun {
		return nil, nil
	}

	ret := step.PrepareResultMap()

	projector, err := createProjector(i.Config.Projector)
	if err != nil {
		log.Debugf("  P (K8s Target Provider): failed to create projector: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
		return ret, err
	}

	switch i.Config.DeploymentStrategy {
	case "", SINGLE_POD:
		updated := step.GetUpdatedComponents()
		if len(updated) > 0 {
			err = i.deployComponents(ctx, span, dep.Instance.Scope, dep.Instance.Name, dep.Instance.Metadata, components, projector, dep.Instance.Name)
			if err != nil {
				log.Debugf("  P (K8s Target Provider): failed to apply components: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
				return ret, err
			}
		}
		deleted := step.GetDeletedComponents()
		if len(deleted) > 0 {
			serviceName := dep.Instance.Name
			if v, ok := dep.Instance.Metadata["service.name"]; ok && v != "" {
				serviceName = v
			}
			err = i.removeService(ctx, dep.Instance.Scope, serviceName)
			if err != nil {
				log.Debugf("  P (K8s Target Provider): failed to remove service: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
				return ret, err
			}
			err = i.removeDeployment(ctx, dep.Instance.Scope, dep.Instance.Name)
			if err != nil {
				log.Debugf("  P (K8s Target Provider): failed to remove deployment: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
				return ret, err
			}
			if i.Config.DeleteEmptyNamespace {
				err = i.removeNamespace(ctx, dep.Instance.Scope, i.Config.RetryCount, i.Config.RetryIntervalInSec)
				if err != nil {
					log.Debugf("  P (K8s Target Provider): failed to remove namespace: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
				}
			}
		}
	case SERVICES, SERVICES_NS:
		updated := step.GetUpdatedComponents()
		if len(updated) > 0 {
			scope := dep.Instance.Scope
			if i.Config.DeploymentStrategy == SERVICES_NS {
				scope = dep.Instance.Name
			}
			for _, component := range components {
				if dep.Instance.Metadata != nil {
					if v, ok := dep.Instance.Metadata[ENV_NAME]; ok && v != "" {
						if component.Metadata == nil {
							component.Metadata = make(map[string]string)
						}
						component.Metadata[ENV_NAME] = v
					}
				}
				err = i.deployComponents(ctx, span, scope, component.Name, component.Metadata, []model.ComponentSpec{component}, projector, dep.Instance.Name)
				if err != nil {
					log.Debugf("  P (K8s Target Provider): failed to apply components: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
					return ret, err
				}
			}
		}
		deleted := step.GetDeletedComponents()
		if len(deleted) > 0 {
			scope := dep.Instance.Scope
			if i.Config.DeploymentStrategy == SERVICES_NS {
				scope = dep.Instance.Name
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
					log.Debugf("P (K8s Target Provider): failed to remove service: %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
					return ret, err
				}
				err = i.removeDeployment(ctx, scope, component.Name)
				if err != nil {
					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.DeleteFailed,
						Message: err.Error(),
					}
					log.Debugf("P (K8s Target Provider): failed to remove deployment: %s, traceId: %", err.Error(), span.SpanContext().TraceID().String())
					return ret, err
				}
				if i.Config.DeleteEmptyNamespace {
					err = i.removeNamespace(ctx, dep.Instance.Scope, i.Config.RetryCount, i.Config.RetryIntervalInSec)
					if err != nil {
						log.Debugf("P (K8s Target Provider): failed to remove namespace: %s, traceId: %", err.Error(), span.SpanContext().TraceID().String())
					}
				}
			}

		}
	}
	err = nil
	return ret, nil
}
func deploymentToComponents(deployment v1.Deployment) ([]model.ComponentSpec, error) {
	components := make([]model.ComponentSpec, len(deployment.Spec.Template.Spec.Containers))
	for i, c := range deployment.Spec.Template.Spec.Containers {
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
		components[i] = component
	}
	return components, nil
}
func metadataToService(scope string, name string, metadata map[string]string) (*apiv1.Service, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	servicePorts := make([]apiv1.ServicePort, 0)

	if v, ok := metadata["service.ports"]; ok && v != "" {
		log.Debugf("  P (K8s Target Provider): metadataToService - service ports: %s", v)
		e := json.Unmarshal([]byte(v), &servicePorts)
		if e != nil {
			log.Errorf("  P (K8s Target Provider): metadataToService - unmarshal: %v", e)
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
			Namespace: scope,
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
func componentsToDeployment(scope string, name string, metadata map[string]string, components []model.ComponentSpec, instanceName string) (*v1.Deployment, error) {
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
		ports := make([]apiv1.ContainerPort, 0)
		if v, ok := c.Properties["container.ports"].(string); ok && v != "" {
			e := json.Unmarshal([]byte(v), &ports)
			if e != nil {
				return nil, e
			}
		}
		container := apiv1.Container{
			Name:            c.Name,
			Image:           c.Properties[model.ContainerImage].(string),
			Ports:           ports,
			ImagePullPolicy: apiv1.PullPolicy(utils.ReadStringFromMapCompat(c.Properties, "container.imagePullPolicy", "Always")),
		}
		if v, ok := c.Properties["container.args"]; ok && v != "" {
			args := make([]string, 0)
			e := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &args)
			if e != nil {
				return nil, e
			}
			container.Args = args
		}
		if v, ok := c.Properties["container.commands"]; ok && v != "" {
			cmds := make([]string, 0)
			e := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &cmds)
			if e != nil {
				return nil, e
			}
			container.Command = cmds
		}
		if v, ok := c.Properties["container.resources"]; ok && v != "" {
			res := apiv1.ResourceRequirements{}
			e := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &res)
			if e != nil {
				return nil, e
			}
			container.Resources = res
		}
		if v, ok := c.Properties["container.volumeMounts"]; ok && v != "" {
			mounts := make([]apiv1.VolumeMount, 0)
			e := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &mounts)
			if e != nil {
				return nil, e
			}
			container.VolumeMounts = mounts
		}
		for k, v := range c.Properties {
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
		deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, container)
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
	log.Debug(string(data))

	return &deployment, nil
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
