/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package docker

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type DockerTargetProviderConfig struct {
	Name string `json:"name"`
}

type DockerTargetProvider struct {
	Config  DockerTargetProviderConfig
	Context *contexts.ManagerContext
}

func DockerTargetProviderConfigFromMap(properties map[string]string) (DockerTargetProviderConfig, error) {
	ret := DockerTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	return ret, nil
}
func (d *DockerTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := DockerTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return d.Init(config)
}
func (s *DockerTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (d *DockerTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Docker Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (Docker Target): Init()")

	// convert config to DockerTargetProviderConfig type
	dockerConfig, err := toDockerTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Docker Target): expected DockerTargetProviderConfig: %+v", err)
		return err
	}

	d.Config = dockerConfig
	return nil
}
func toDockerTargetProviderConfig(config providers.IProviderConfig) (DockerTargetProviderConfig, error) {
	ret := DockerTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *DockerTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Docker Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (Docker Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		sLog.Errorf("  P (Docker Target): failed to create docker client: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}

	ret := make([]model.ComponentSpec, 0)
	for _, component := range references {
		var info types.ContainerJSON
		info, err = cli.ContainerInspect(ctx, component.Component.Name)
		if err == nil {
			name := info.Name
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
			component := model.ComponentSpec{
				Name:       name,
				Properties: make(map[string]interface{}),
			}
			// container.args
			if len(info.Args) > 0 {
				argsData, _ := json.Marshal(info.Args)
				component.Properties["container.args"] = string(argsData)
			}
			// container.image
			component.Properties[model.ContainerImage] = info.Config.Image
			if info.HostConfig != nil {
				resources, _ := json.Marshal(info.HostConfig.Resources)
				component.Properties["container.resources"] = string(resources)
			}
			// container.ports
			if info.NetworkSettings != nil && len(info.NetworkSettings.Ports) > 0 {
				ports, _ := json.Marshal(info.NetworkSettings.Ports)
				component.Properties["container.ports"] = string(ports)
			}
			// container.cmd
			if len(info.Config.Cmd) > 0 {
				cmdData, _ := json.Marshal(info.Config.Cmd)
				component.Properties["container.commands"] = string(cmdData)
			}
			// container.volumeMounts
			if len(info.Mounts) > 0 {
				volumeData, _ := json.Marshal(info.Mounts)
				component.Properties["container.volumeMounts"] = string(volumeData)
			}
			// get environment varibles that are passed in by the reference
			env := info.Config.Env
			if len(env) > 0 {
				for _, e := range env {
					pair := strings.Split(e, "=")
					if len(pair) == 2 {
						for _, s := range references {
							if s.Component.Name == component.Name {
								for k, _ := range s.Component.Properties {
									if k == "env."+pair[0] {
										component.Properties[k] = pair[1]
									}
								}
							}
						}
					}
				}
			}
			ret = append(ret, component)
		}
	}

	return ret, nil
}

func (i *DockerTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Docker Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (Docker Target): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	injections := &model.ValueInjections{
		InstanceId: deployment.Instance.Name,
		SolutionId: deployment.Instance.Solution,
		TargetId:   deployment.ActiveTarget,
	}

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.Errorf("  P (Docker Target): failed to validate components: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		sLog.Errorf("  P (Docker Target): failed to create docker client: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return ret, err
	}

	for _, component := range step.Components {
		if component.Action == "update" {
			image := model.ReadPropertyCompat(component.Component.Properties, model.ContainerImage, injections)
			resources := model.ReadPropertyCompat(component.Component.Properties, "container.resources", injections)
			if image == "" {
				err = errors.New("component doesn't have container.image property")
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.Errorf("  P (Helm Target): %+v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}

			alreadyRunning := true
			_, err = cli.ContainerInspect(ctx, component.Component.Name)
			if err != nil { //TODO: check if the error is ErrNotFound
				alreadyRunning = false
			}

			// TODO: I don't think we need to do an explict image pull here, as Docker will pull the image upon cache miss
			// reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
			// if err != nil {
			// 	observ_utils.CloseSpanWithError(span, &err)
			// 	sLog.Errorf("  P (Docker Target): failed to pull docker image: %+v", err)
			// 	return err
			// }

			// defer reader.Close()
			// io.Copy(os.Stdout, reader)

			if alreadyRunning {
				err = cli.ContainerStop(context.TODO(), component.Component.Name, nil)
				if err != nil {
					if !client.IsErrNotFound(err) {
						sLog.Errorf("  P (Docker Target): failed to stop a running container: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
						return ret, err
					}
				}
				err = cli.ContainerRemove(context.TODO(), component.Component.Name, types.ContainerRemoveOptions{})
				if err != nil {
					ret[component.Component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.UpdateFailed,
						Message: err.Error(),
					}
					sLog.Errorf("  P (Docker Target): failed to remove existing container: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
			}

			// prepare environment variables
			env := make([]string, 0)
			for k, v := range component.Component.Properties {
				if strings.HasPrefix(k, "env.") {
					env = append(env, strings.TrimPrefix(k, "env.")+"="+v.(string))
				}
			}

			containerConfig := container.Config{
				Image: image,
				Env:   env,
			}
			var hostConfig *container.HostConfig
			if resources != "" {
				var resourceSpec container.Resources
				err = json.Unmarshal([]byte(resources), &resourceSpec)
				if err != nil {
					ret[component.Component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.UpdateFailed,
						Message: err.Error(),
					}
					sLog.Errorf("  P (Docker Target): failed to read container resource settings: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
				hostConfig = &container.HostConfig{
					Resources: resourceSpec,
				}
			}
			var container container.ContainerCreateCreatedBody
			container, err = cli.ContainerCreate(context.TODO(), &containerConfig, hostConfig, nil, nil, component.Component.Name)
			if err != nil {
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.Errorf("  P (Docker Target): failed to create container: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}

			if err = cli.ContainerStart(context.TODO(), container.ID, types.ContainerStartOptions{}); err != nil {
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.Errorf("  P (Docker Target): failed to start container: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}
			ret[component.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Updated,
				Message: "",
			}
		} else {
			err = cli.ContainerStop(context.TODO(), component.Component.Name, nil)
			if err != nil {
				if !client.IsErrNotFound(err) {
					sLog.Errorf("  P (Docker Target): failed to stop a running container: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
			}
			err = cli.ContainerRemove(context.TODO(), component.Component.Name, types.ContainerRemoveOptions{})
			if err != nil {
				if !client.IsErrNotFound(err) {
					sLog.Errorf("  P (Docker Target): failed to remove existing container: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
					return ret, err
				}
			}
			ret[component.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Deleted,
				Message: "",
			}
		}
	}
	return ret, nil
}

func (*DockerTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{model.ContainerImage},
		OptionalProperties:    []string{"container.resources"},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
		ChangeDetectionProperties: []model.PropertyDesc{
			{Name: model.ContainerImage, IgnoreCase: false, SkipIfMissing: false},
			{Name: "container.ports", IgnoreCase: false, SkipIfMissing: true},
			{Name: "container.resources", IgnoreCase: false, SkipIfMissing: true},
		},
	}
}
