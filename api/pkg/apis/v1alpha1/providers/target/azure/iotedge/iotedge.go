/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package iotedge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	azureutils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/cloudutils/azure"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/google/uuid"
)

var sLog = logger.NewLogger("coa.runtime")

const (
	ENV_NAME string = "SYMPHONY_AGENT_ADDRESS"
	ENV_SALT string = "SYMPHONY_VERSION_SALT"
)

// Provider config and type
type IoTEdgeTargetProviderConfig struct {
	Name             string `json:"name"`
	KeyName          string `json:"keyName"`
	Key              string `json:"key"`
	IoTHub           string `json:"iotHub"`
	APIVersion       string `json:"apiVersion"`
	DeviceName       string `json:"deviceName"`
	EdgeAgentVersion string `json:"edgeAgentVersion,omitempty"`
	EdgeHubVersion   string `json:"edgeHubVersion,omitempty"`
}
type IoTEdgeTargetProvider struct {
	Config  IoTEdgeTargetProviderConfig
	Context *contexts.ManagerContext
}

// Azure IoT Edge objects
type IoTEdgeDeployment struct {
	ModulesContent map[string]ModuleState `json:"modulesContent"`
}
type ModuleState struct {
	DesiredProperties map[string]interface{} `json:"properties.desired"`
}
type DesiredProperties struct {
	SchemaVersion string            `json:"schemaVersion"`
	Runtime       Runtime           `json:"runtime"`
	SystemModules map[string]Module `json:"systemModules"`
	Modules       map[string]Module `json:"modules"`
	Version       int               `json:"$version,omitempty"`
	Metadata      interface{}       `json:"$metadata,omitempty"`
}
type Runtime struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}
type RegistryCredential struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Address  string `json:"address"`
}
type Module struct {
	Type              string                 `json:"type"`
	Settings          map[string]string      `json:"settings"`
	Status            string                 `json:"status,omitempty"`
	RestartPolicy     string                 `json:"restartPolicy,omitempty"`
	Version           interface{}            `json:"version,omitempty"`
	DesiredProperties map[string]interface{} `json:"metadata,omitempty"`
	Graph             map[string]interface{} `json:"graph,omitempty"`
	GraphFlavor       string                 `json:"graphFlavor,omitempty"`
	IotHubRoutes      map[string]string      `json:"routes,omitempty"`
	Environments      map[string]EnvValue    `json:"env,omitempty"`
}
type EnvValue struct {
	Value string `json:"value"`
}
type ModuleID struct {
	ModuleId string `json:"moduleId"`
}
type ModuleTwin struct {
	DeviceId   string               `json:"deviceId"`
	ModuleId   string               `json:"moduleId"`
	Properties ModuleTwinProperties `json:"properties"`
	Version    interface{}          `json:"version"`
}
type ModuleTwinProperties struct {
	Desired  map[string]interface{} `json:"desired"`
	Reported map[string]interface{} `json:"reported"`
}

func IoTEdgeTargetProviderConfigFromMap(properties map[string]string) (IoTEdgeTargetProviderConfig, error) {
	ret := IoTEdgeTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["keyName"]; ok {
		ret.KeyName = v
	} else {
		ret.KeyName = "iothubowner"
	}
	if v, ok := properties["key"]; ok {
		ret.Key = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "IoT Edge update provider key is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["iotHub"]; ok {
		ret.IoTHub = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "IoT Edge update provider IoT Hub name is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["apiVersion"]; ok {
		ret.APIVersion = v
	} else {
		ret.APIVersion = "2020-05-31-preview"
	}
	if v, ok := properties["edgeAgentVersion"]; ok {
		ret.EdgeAgentVersion = v
	} else {
		ret.EdgeAgentVersion = "1.3"
	}
	if v, ok := properties["edgeHubVersion"]; ok {
		ret.EdgeHubVersion = v
	} else {
		ret.EdgeHubVersion = "1.3"
	}
	if v, ok := properties["deviceName"]; ok {
		ret.DeviceName = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "IoT Edge update provider device name is not set", v1alpha2.BadConfig)
	}
	return ret, nil
}

func (i *IoTEdgeTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := IoTEdgeTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *IoTEdgeTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *IoTEdgeTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("IoT Edge Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (IoT Edge Target): Init()")

	updateConfig, err := toIoTEdgeTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (IoT Edge Target): expected IoTEdgeTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig

	return nil
}
func (i *IoTEdgeTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("IoT Edge Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (IoT Edge Target): getting components: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	hubTwin, err := i.getIoTEdgeModuleTwin(ctx, "$edgeHub")
	if err != nil {
		sLog.Errorf("  P (IoT Edge Target): failed to get edgeHub module twin: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}

	modules, err := i.getIoTEdgeModules(ctx)
	if err != nil {
		sLog.Errorf("  P (IoT Edge Target): failed to get modules: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	components := make([]model.ComponentSpec, 0)
	for k, m := range modules {
		if k != "$edgeAgent" && k != "$edgeHub" {
			var twin ModuleTwin
			twin, err = i.getIoTEdgeModuleTwin(ctx, k)
			if err != nil {
				sLog.Errorf("  P (IoT Edge Target): failed to get %s module: %+v, traceId: %s", k, err, span.SpanContext().TraceID().String())
				return nil, err
			}
			var component model.ComponentSpec
			component, err = toComponent(hubTwin, twin, deployment.Instance.Name, m)
			if err != nil {
				sLog.Errorf("  P (IoT Edge Target):failed to parse %s twin to component %+v, traceId: %s", k, err, span.SpanContext().TraceID().String())
				return nil, err
			}
			components = append(components, component)
		}
	}

	return components, nil
}

func (i *IoTEdgeTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("IoT Edge Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (IoT Edge Target): applying components, traceId: %s", span.SpanContext().TraceID().String())

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()

	edgeAgent, err := i.getIoTEdgeModuleTwin(ctx, "$edgeAgent")
	if err != nil {
		sLog.Errorf("  P (IoT Edge Target): failed to get edgeAgent moduel twin: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return ret, err
	}

	edgeHub, err := i.getIoTEdgeModuleTwin(ctx, "$edgeHub")
	if err != nil {
		sLog.Errorf("  P (IoT Edge Target): failed to get edgeHub module twin: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return ret, err
	}

	//updated
	modules := make(map[string]Module)
	for _, a := range components {
		module, e := toModule(a, deployment.Instance.Name, deployment.Instance.Metadata[ENV_NAME], step.Target)
		if e != nil {
			ret[a.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.UpdateFailed,
				Message: e.Error(),
			}
			err = e
			sLog.Errorf("  P (IoT Edge Target): failed to parse %s component to module: %+v, traceId: %s", a.Name, err, span.SpanContext().TraceID().String())
			return ret, err
		}
		modules[a.Name] = module
	}
	if len(modules) > 0 {
		err = i.deployToIoTEdge(ctx, deployment.Instance.Name, deployment.Instance.Metadata, modules, edgeAgent, edgeHub)
		if err != nil {
			sLog.Errorf("  P (IoT Edge Target): failed to deploy to IoT edge: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return ret, err
		}
	}

	//delete
	modules = make(map[string]Module)
	for _, a := range components {
		module, e := toModule(a, deployment.Instance.Name, deployment.Instance.Metadata[ENV_NAME], step.Target)
		if e != nil {
			ret[a.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.DeleteFailed,
				Message: e.Error(),
			}
			sLog.Errorf("  P (IoT Edge Target): failed to parse %s component to module: %+v, traceId: %s", a.Name, err, span.SpanContext().TraceID().String())
			return ret, err
		}
		modules[a.Name] = module
	}
	if len(modules) > 0 {
		err = i.remvoefromIoTEdge(ctx, deployment.Instance.Name, deployment.Instance.Metadata, modules, edgeAgent, edgeHub)
		if err != nil {
			sLog.Errorf("  P (IoT Edge Target): failed to remove from IoT edge: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return ret, err
		}
	}
	//TODO: Should we raise events to remove AVA graphs?
	return ret, nil
}
func (*IoTEdgeTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{model.ContainerImage, "container.version", "container.type"},
		OptionalProperties:    []string{"container.restartPolicy", "container.createOptions", "env.*"},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
		ChangeDetectionProperties: []model.PropertyDesc{
			{Name: "container.restartPolicy", IgnoreCase: false, SkipIfMissing: true},
			{Name: "container.createOptions", IgnoreCase: false, SkipIfMissing: true},
			{Name: "container.version", IgnoreCase: false, SkipIfMissing: true},
			{Name: "container.type", IgnoreCase: false, SkipIfMissing: true},
			{Name: model.ContainerImage, IgnoreCase: false, SkipIfMissing: true},
			{Name: "desired.*", IgnoreCase: false, SkipIfMissing: true},
			{Name: "env.*", IgnoreCase: false, SkipIfMissing: true},
		},
	}
}

func toIoTEdgeTargetProviderConfig(config providers.IProviderConfig) (IoTEdgeTargetProviderConfig, error) {
	ret := IoTEdgeTargetProviderConfig{}
	if config == nil {
		return ret, errors.New("IoTEdgeTargetProviderConfig is null")
	}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	// ret.IoTHub = providers.LoadEnv(ret.IoTHub)
	// ret.DeviceName = providers.LoadEnv(ret.DeviceName)
	// ret.APIVersion = providers.LoadEnv(ret.APIVersion)
	// ret.KeyName = providers.LoadEnv(ret.KeyName)
	// ret.Key = providers.LoadEnv(ret.Key)

	if ret.APIVersion == "" {
		ret.APIVersion = "2020-05-31-preview"
	}
	if ret.KeyName == "" {
		ret.KeyName = "iothubowner"
	}
	if ret.EdgeAgentVersion == "" {
		ret.EdgeAgentVersion = "1.3"
	}
	if ret.EdgeHubVersion == "" {
		ret.EdgeHubVersion = "1.3"
	}
	return ret, nil
}

func toComponent(hubTwin ModuleTwin, twin ModuleTwin, name string, module Module) (model.ComponentSpec, error) {
	moduleId, _ := reduceKey(twin.ModuleId, name)
	component := model.ComponentSpec{
		Name:       moduleId,
		Properties: make(map[string]interface{}),
		Routes:     make([]model.RouteSpec, 0),
	}
	for k, v := range module.Environments {
		if k != ENV_NAME && k != ENV_SALT {
			component.Properties["env."+k] = v.Value
		}
	}

	if v, ok := hubTwin.Properties.Desired["routes"]; ok {
		routes := v.(map[string]interface{})
		for k, iv := range routes {
			def := iv.(string)
			if strings.Contains(def, "modules/"+twin.ModuleId+"/") { //TODO: this check is not necessarily safe
				reducedRoute, _ := reduceKey(k, name)
				reducedDef, _ := replaceKey(def, name)
				component.Routes = append(component.Routes, model.RouteSpec{
					Route: reducedRoute,
					Type:  "iothub",
					Properties: map[string]string{
						"definition": reducedDef,
					},
				})
			}
		}
	}

	component.Properties["container.restartPolicy"] = module.RestartPolicy
	if module.Version != nil {
		component.Properties["container.version"] = module.Version.(string)
	}
	component.Properties["container.type"] = module.Type
	if v, ok := module.Settings["createOptions"]; ok {
		component.Properties["container.createOptions"] = v
	}
	if v, ok := module.Settings["image"]; ok {
		component.Properties[model.ContainerImage] = v
	}
	//TODO: We are extracting only keys starting with a lower-case letter here.
	interestedKey := regexp.MustCompile(`^[a-zA-Z]+`)
	for k, v := range twin.Properties.Desired { //We are reading desired instead of reported, as we leave IoT Edge state seeking to IoT Edge itself
		if interestedKey.MatchString(k) {
			switch v.(type) {
			case int:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case int8:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case int16:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case int32:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case int64:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case uint:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case uint8:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case uint16:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case uint32:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case uint64:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case float32:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case float64:
				component.Properties["desired."+k] = fmt.Sprintf("#%v", v)
			case string:
				component.Properties["desired."+k] = fmt.Sprintf("%s", v)
			case bool:
				component.Properties["desired."+k] = fmt.Sprintf("$%v", v)
			case []interface{}:
				data, err := json.Marshal(v)
				if err == nil {
					component.Properties["desired."+k] = string(data)
				} else {
					component.Properties["desired."+k] = fmt.Sprintf("%v", v) //The "desired." prefix is added to match with what's generated during Apply
				}
			default:
				data, err := json.Marshal(v)
				if err == nil {
					component.Properties["desired."+k] = string(data)
				} else {
					component.Properties["desired."+k] = fmt.Sprintf("%v", v) //The "desired." prefix is added to match with what's generated during Apply
				}
			}
		}
	}
	return component, nil
}
func readProperty(properties map[string]interface{}, key string, defaultVal string, required bool) (string, error) {
	if v, ok := properties[key]; ok && v != "" {
		return fmt.Sprintf("%v", v), nil
	}
	if required && defaultVal == "" {
		return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("required property '%s' is missng", key), v1alpha2.BadRequest)
	}
	return defaultVal, nil
}
func toModule(component model.ComponentSpec, name string, agentName string, targetName string) (Module, error) {
	policy, err := readProperty(component.Properties, "container.restartPolicy", "always", false)
	if err != nil {
		return Module{}, err
	}
	createOptions, err := readProperty(component.Properties, "container.createOptions", "", false)
	if err != nil {
		return Module{}, err
	}
	version, err := readProperty(component.Properties, "container.version", "", true)
	if err != nil {
		return Module{}, err
	}
	componentType, err := readProperty(component.Properties, "container.type", "", true)
	if err != nil {
		return Module{}, err
	}
	image, err := readProperty(component.Properties, model.ContainerImage, "", true)
	if err != nil {
		return Module{}, err
	}
	module := Module{
		Version:       version,
		Type:          componentType,
		RestartPolicy: policy,
		Status:        "running",
		Settings: map[string]string{
			"image":         image,
			"createOptions": createOptions,
		},
	}
	module.DesiredProperties = make(map[string]interface{})
	module.Graph = make(map[string]interface{})
	module.GraphFlavor = "ava"
	module.IotHubRoutes = make(map[string]string)
	module.Environments = make(map[string]EnvValue)
	for k, v := range component.Properties {
		// TODO: Transition from map[string]string to map[string]interface{}
		// for now we would only do this for string properties
		if sv, ok := v.(string); ok {
			tv := utils.ProjectValue(sv, name)
			if strings.HasPrefix(k, "desired.") {
				module.DesiredProperties[k[8:]] = tv
				// } else if strings.HasPrefix(k, "graph.") {
				// 	if k == "graph.methodFlavor" {
				// 		module.GraphFlavor = v
				// 	} else {
				// 		module.Graph[k[6:]] = v
				// 	}
			} else if strings.HasPrefix(k, "env.") {
				module.Environments[k[4:]] = EnvValue{Value: tv}
			}
		}
	}

	module.Environments[ENV_SALT] = EnvValue{Value: uuid.New().String()}

	if agentName != "" {
		module.Environments[ENV_NAME] = EnvValue{Value: fmt.Sprintf("%s-%s-%s", "target-runtime", targetName, agentName)}
	}
	for _, v := range component.Routes {
		if v.Type == "iothub" {
			module.IotHubRoutes[v.Route] = v.Properties["definition"]
		}
	}

	return module, nil
}
func (i *IoTEdgeTargetProvider) getIoTEdgeModuleTwin(ctx context.Context, id string) (ModuleTwin, error) {
	url := fmt.Sprintf("https://%s/twins/%s/modules/%s?api-version=%s", i.Config.IoTHub, i.Config.DeviceName, id, i.Config.APIVersion)
	_, span := observability.StartSpan("IoT Edge REST API", ctx, &map[string]string{
		"method": "getIoTEdgeModuleTwin",
		"url":    url,
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	module := ModuleTwin{}
	sasToken := azureutils.CreateSASToken(fmt.Sprintf("%s/devices/%s", i.Config.IoTHub, i.Config.DeviceName), i.Config.KeyName, i.Config.Key)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		sLog.Errorf("failed to get IoT Edge modules: %v", err)
		return module, v1alpha2.NewCOAError(err, "failed to get IoT Edge modules", v1alpha2.InternalError)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sasToken)
	resp, err := client.Do(req)
	if err != nil {
		sLog.Errorf("failed to get IoT Edge modules: %v", err)
		return module, v1alpha2.NewCOAError(err, "failed to get IoT Edge modules", v1alpha2.InternalError)
	}
	if resp.StatusCode != http.StatusOK {
		sLog.Errorf("failed to get IoT Edge modules: %v", resp)
		//return module, v1alpha1.NewCOAError(nil, "failed to get IoT Edge modules", v1alpha1.InternalError) //TODO: carry over HTTP status code
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sLog.Errorf("failed to get IoT Edge modules: %v", err)
		return module, v1alpha2.NewCOAError(err, "failed to get IoT Edge modules", v1alpha2.InternalError)
	}
	err = json.Unmarshal(bodyBytes, &module)
	if err != nil {
		sLog.Errorf("failed to get IoT Edge modules: %v", err)
		return module, v1alpha2.NewCOAError(err, "failed to get IoT Edge modules", v1alpha2.InternalError)
	}
	return module, nil
}
func (i *IoTEdgeTargetProvider) getIoTEdgeModules(ctx context.Context) (map[string]Module, error) {
	ret := make(map[string]Module)
	agentTwin, err := i.getIoTEdgeModuleTwin(ctx, "$edgeAgent")
	if err != nil {
		return ret, err
	}
	data, err := json.Marshal(agentTwin.Properties.Desired["modules"])
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (i *IoTEdgeTargetProvider) remvoefromIoTEdge(ctx context.Context, name string, metadata map[string]string, modules map[string]Module, agentRef ModuleTwin, hubRef ModuleTwin) error {
	deployment := makeDefaultDeployment(metadata, i.Config.EdgeAgentVersion, i.Config.EdgeHubVersion)
	err := reduceDeployment(&deployment, name, modules, agentRef, hubRef)
	if err != nil {
		return err
	}
	return i.applyIoTEdgeDeployment(ctx, deployment)
}

func (i *IoTEdgeTargetProvider) deployToIoTEdge(ctx context.Context, name string, metadata map[string]string, modules map[string]Module, agentRef ModuleTwin, hubRef ModuleTwin) error {

	deployment := makeDefaultDeployment(metadata, i.Config.EdgeAgentVersion, i.Config.EdgeHubVersion)

	err := updateDeployment(&deployment, name, modules, agentRef, hubRef)
	if err != nil {
		return err
	}
	return i.applyIoTEdgeDeployment(ctx, deployment)
}

func (i *IoTEdgeTargetProvider) applyIoTEdgeDeployment(ctx context.Context, deployment IoTEdgeDeployment) error {
	url := fmt.Sprintf("https://%s/devices/%s/applyConfigurationContent?api-version=%s", i.Config.IoTHub, i.Config.DeviceName, i.Config.APIVersion)
	_, span := observability.StartSpan("IoT Edge REST API", ctx, &map[string]string{
		"method": "applyIoTEdgeDeployment",
		"url":    url,
	})

	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sasToken := azureutils.CreateSASToken(fmt.Sprintf("%s/devices/%s", i.Config.IoTHub, i.Config.DeviceName), i.Config.KeyName, i.Config.Key)
	client := &http.Client{}
	payload, err := json.Marshal(deployment)
	if err != nil {
		return v1alpha2.NewCOAError(err, "failed to serialize IoT Edge deployemnt", v1alpha2.SerializationError)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		sLog.Errorf("failed to post IoT Edge deployment: %v", err)
		return v1alpha2.NewCOAError(err, "failed to post IoT Edge deployment", v1alpha2.InternalError)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sasToken)
	resp, err := client.Do(req)
	if err != nil {
		sLog.Errorf("failed to post IoT Edge deployment: %v", err)
		return v1alpha2.NewCOAError(err, "failed to post IoT Edge deployment", v1alpha2.InternalError)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		sLog.Errorf("failed to post IoT Edge deployment: %v", resp)
		return v1alpha2.NewCOAError(nil, "failed to post IoT Edge deployment", v1alpha2.InternalError) //TODO: carry over HTTP status code
	}
	return nil
}

func replaceKey(key string, name string) (string, bool) {
	if name != "" && strings.Contains(key, name+"-") {
		return strings.ReplaceAll(key, name+"-", ""), true
	}
	return key, false
}

func reduceKey(key string, name string) (string, bool) {
	if name != "" && strings.HasPrefix(key, name+"-") {
		return key[len(name)+1:], true
	}
	return key, false
}
func expandKey(key string, name string) string {
	if name != "" {
		return name + "-" + key
	}
	return key
}

func carryOverRoutes(deployment *IoTEdgeDeployment, ref ModuleTwin) {
	if ref.ModuleId != "" {
		if v, ok := ref.Properties.Desired["routes"]; ok {
			if vc, ok := v.(map[string]string); ok {
				m := deployment.ModulesContent["$edgeHub"].DesiredProperties["routes"].(map[string]string)
				for k, iv := range vc {
					m[k] = iv
				}
			}
		}
	}
}

func updateDeployment(deployment *IoTEdgeDeployment, name string, modules map[string]Module, agentRef ModuleTwin, hubRef ModuleTwin) error {

	// add all other modules that are not in the current module list so that we can write them back
	otherModules := map[string]bool{}
	if agentRef.ModuleId != "" {
		carryOverRoutes(deployment, agentRef)
		im, ok := agentRef.Properties.Desired["modules"].(map[string]interface{})
		if ok {
			for k, _ := range im {
				rk, reduced := reduceKey(k, name)
				if !reduced {
					strContent, _ := json.Marshal(im[k])
					mRef := Module{}
					err := json.Unmarshal(strContent, &mRef)
					if err != nil {
						return err
					}
					modules[rk] = mRef
					otherModules[rk] = true
				}
			}
		}
	}

	// create a new module collection
	deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"] = make(map[string]Module)

	rd := deployment.ModulesContent["$edgeHub"].DesiredProperties["routes"].(map[string]string)

	if v, ok := hubRef.Properties.Desired["routes"]; ok {
		routes := v.(map[string]interface{})
		for ik, iv := range routes {
			rd[ik] = iv.(string)
		}
	}

	// add all modules, wich include modules from current deployment as well as other modules
	for k, m := range modules {
		d := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
		ek := k
		if _, ok := otherModules[k]; !ok {
			ek = expandKey(k, name)
		}
		d[ek] = m
		if len(m.DesiredProperties) > 0 {
			deployment.ModulesContent[ek] = ModuleState{
				DesiredProperties: map[string]interface{}{},
			}
			for ik, iv := range m.DesiredProperties {
				deployment.ModulesContent[ek].DesiredProperties[ik] = iv
			}
		}
		if len(m.IotHubRoutes) > 0 {
			if _, ok := otherModules[k]; !ok {
				for rk, rv := range m.IotHubRoutes {
					rek := expandKey(rk, name)
					mrv := modifyRoutes(rv, name, modules, otherModules)
					rd[rek] = mrv
				}
			}
		}
	}
	return nil
}
func modifyRoutes(route string, name string, modules map[string]Module, otherModules map[string]bool) string {
	for k, _ := range modules {
		if _, ok := otherModules[k]; !ok {
			route = strings.ReplaceAll(route, "modules/"+k, "modules/"+name+"-"+k)
		}
	}
	return route
}

func reduceDeployment(deployment *IoTEdgeDeployment, name string, modules map[string]Module, ref ModuleTwin, hubRef ModuleTwin) error {

	otherModules := map[string]bool{}

	rd := deployment.ModulesContent["$edgeHub"].DesiredProperties["routes"].(map[string]string)

	if v, ok := hubRef.Properties.Desired["routes"]; ok {
		routes := v.(map[string]interface{})
		for ik, iv := range routes {
			rd[ik] = iv.(string)
		}
	}

	if ref.ModuleId != "" {
		carryOverRoutes(deployment, ref)
		im, ok := ref.Properties.Desired["modules"].(map[string]interface{})
		if ok {
			for k, _ := range im {
				rk, reduced := reduceKey(k, name)
				if !reduced {
					strContent, _ := json.Marshal(im[k])
					mRef := Module{}
					err := json.Unmarshal(strContent, &mRef)
					if err != nil {
						return err
					}
					modules[rk] = mRef
					otherModules[rk] = true
				} else {
					if len(modules[rk].IotHubRoutes) > 0 {
						for ik, _ := range modules[rk].IotHubRoutes {
							delete(rd, expandKey(ik, name))
						}
					}
					delete(modules, rk)
				}
			}
		}
	}

	deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"] = make(map[string]Module)
	for k, m := range modules {
		d := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
		ek := k
		if _, ok := otherModules[k]; !ok {
			ek = expandKey(k, name)
		}
		d[ek] = m
		if len(m.DesiredProperties) > 0 {
			deployment.ModulesContent[ek] = ModuleState{
				DesiredProperties: map[string]interface{}{},
			}
			for ik, iv := range m.DesiredProperties {
				deployment.ModulesContent[ek].DesiredProperties[ik] = iv
			}
		}
	}
	return nil
}

func makeDefaultDeployment(metadata map[string]string, edgeAgentVersion string, edgeHubVersion string) IoTEdgeDeployment {

	deployment := IoTEdgeDeployment{
		ModulesContent: map[string]ModuleState{
			"$edgeAgent": {
				DesiredProperties: map[string]interface{}{
					"schemaVersion": "1.0",
					"runtime": Runtime{
						Type: "docker",
						Settings: map[string]interface{}{
							"minDockerVersion": "v1.25",
							"loggingOption":    "",
						},
					},
					"systemModules": map[string]Module{
						"edgeAgent": Module{
							Type: "docker",
							Settings: map[string]string{
								"image":         "mcr.microsoft.com/azureiotedge-agent:" + edgeAgentVersion,
								"createOptions": "",
							},
						},
						"edgeHub": {
							Type:          "docker",
							RestartPolicy: "always",
							Status:        "running",
							Settings: map[string]string{
								"image":         "mcr.microsoft.com/azureiotedge-hub:" + edgeHubVersion,
								"createOptions": "{\"HostConfig\":{\"PortBindings\":{\"5671/tcp\":[{\"HostPort\":\"5671\"}],\"8883/tcp\":[{\"HostPort\":\"8883\"}],\"443/tcp\":[{\"HostPort\":\"443\"}]}}}",
							},
						},
					},
				},
			},
			"$edgeHub": {
				DesiredProperties: map[string]interface{}{
					"schemaVersion": "1.0",
					"routes":        map[string]string{},
					"storeAndForwardConfiguration": map[string]int{ //TODO: this is also a hack
						"timeToLiveSecs": 7200,
					},
				},
			},
		},
	}
	if v, ok := metadata["$edgeAgent.registryCredentials"]; ok && strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		credentials := make(map[string]RegistryCredential)
		data := []byte(v)
		err := json.Unmarshal(data, &credentials)
		if err == nil {
			(deployment.ModulesContent["$edgeAgent"].DesiredProperties["runtime"].(Runtime)).Settings["registryCredentials"] = credentials
		}
	}
	return deployment
}
