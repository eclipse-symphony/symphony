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

package docker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var sLog = logger.NewLogger("coa.runtime")

type DockerTargetProviderConfig struct {
	Name string `json:"name"`
}

type DockerTargetProvider struct {
	Config DockerTargetProviderConfig
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
func (d *DockerTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Docker Target Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("  P (Docker Target): Init()")

	// convert config to DockerTargetProviderConfig type
	dockerConfig, err := toDockerTargetProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
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

func (i *DockerTargetProvider) Get(ctx context.Context, dep model.DeploymentSpec) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Docker Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Infof("  P (Docker Target): getting artifacts: %s - %s", dep.Instance.Scope, dep.Instance.Name)

	components := make([]model.ComponentSpec, 0)
	slice := dep.GetComponentSlice()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Docker Target): failed to create docker client: %+v", err)
		return nil, err
	}

	for _, component := range slice {
		info, err := cli.ContainerInspect(ctx, component.Name)
		if err == nil {
			name := info.Name
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
			component := model.ComponentSpec{
				Name:       name,
				Properties: make(map[string]string),
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
			components = append(components, component)
		}
	}

	observ_utils.CloseSpanWithError(span, nil)
	return components, nil
}

func (d *DockerTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, isDryRun bool) error {
	_, span := observability.StartSpan("Docker Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("  P (Docker Target): applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	injections := &model.ValueInjections{
		InstanceId: deployment.Instance.Name,
		SolutionId: deployment.Instance.Solution,
		TargetId:   deployment.ActiveTarget,
	}

	components := deployment.GetComponentSlice()

	err := d.GetValidationRule(ctx).Validate(components)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return err
	}
	if isDryRun {
		observ_utils.CloseSpanWithError(span, nil)
		return nil
	}

	for _, component := range components {
		if component.Type == "container" {
			image := model.ReadProperty(component.Properties, model.ContainerImage, injections)
			resources := model.ReadProperty(component.Properties, "container.resources", injections)
			if image == "" {
				err := errors.New("component doesn't have container.image property")
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Helm Target): component doesn't have container.image property")
				return err
			}

			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Docker Target): failed to create docker client: %+v", err)
				return err
			}

			isNew := true
			containerInfo, err := cli.ContainerInspect(ctx, component.Name)
			if err == nil {
				isNew = false
			}

			reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (Docker Target): failed to pull docker image: %+v", err)
				return err
			}

			defer reader.Close()
			io.Copy(os.Stdout, reader)

			if !isNew && containerInfo.Image != image {
				err = cli.ContainerRemove(context.Background(), component.Name, types.ContainerRemoveOptions{})
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Docker Target): failed to remove existing container: %+v", err)
					return err
				}
				isNew = true
			}

			if isNew {
				containerConfig := container.Config{
					Image: image,
				}
				var hostConfig *container.HostConfig
				if resources != "" {
					var resourceSpec container.Resources
					err := json.Unmarshal([]byte(resources), &resourceSpec)
					if err != nil {
						observ_utils.CloseSpanWithError(span, err)
						sLog.Errorf("  P (Docker Target): failed to read container resource settings: %+v", err)
						return err
					}
					hostConfig = &container.HostConfig{
						Resources: resourceSpec,
					}
				}
				container, err := cli.ContainerCreate(context.Background(), &containerConfig, hostConfig, nil, nil, component.Name)
				if err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Docker Target): failed to create container: %+v", err)
					return err
				}

				if err := cli.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{}); err != nil {
					observ_utils.CloseSpanWithError(span, err)
					sLog.Errorf("  P (Docker Target): failed to start container: %+v", err)
					return err
				}
			} else {
				if resources != "" {
					var resourceObj container.Resources
					err = json.Unmarshal([]byte(resources), &resourceObj)
					if err != nil {
						observ_utils.CloseSpanWithError(span, err)
						sLog.Errorf("  P (Docker Target): failed to unmarshal container resources spec: %+v", err)
						return err
					}
					_, err = cli.ContainerUpdate(context.Background(), component.Name, container.UpdateConfig{
						Resources: resourceObj,
					})
					if err != nil {
						observ_utils.CloseSpanWithError(span, err)
						sLog.Errorf("  P (Docker Target): failed to update container resources: %+v", err)
						return err
					}
				}

			}

		}
	}
	return nil
}

func (d *DockerTargetProvider) Remove(ctx context.Context, dep model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan("Docker Target Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	sLog.Infof("  P (Docker Target): deleting artifacts: %s - %s", dep.Instance.Scope, dep.Instance.Name)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Docker Target): failed to create docker client: %+v", err)
		return err
	}

	components := dep.GetComponentSlice()
	for _, component := range components {
		//if component.Type == "container" {
		err = cli.ContainerStop(context.Background(), component.Name, nil)
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Docker Target): failed to stop a running container: %+v", err)
			return err
		}
		err = cli.ContainerRemove(context.Background(), component.Name, types.ContainerRemoveOptions{})
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Errorf("  P (Docker Target): failed to remove existing container: %+v", err)
			return err
		}
		//}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func (d *DockerTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	_, span := observability.StartSpan("Docker Target Provider", ctx, &map[string]string{
		"method": "NeedsRemove",
	})
	sLog.Infof("  P (Docker Target Provider): NeedsRemove: %d - %d", len(desired), len(current))
	for _, dc := range desired {
		for _, cc := range current {
			if cc.Name == dc.Name && cc.Properties[model.ContainerImage] == dc.Properties[model.ContainerImage] {
				observ_utils.CloseSpanWithError(span, nil)
				sLog.Info("  P (Docker Target Provider): NeedsRemove: returning true")
				return true
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	sLog.Info("  P (Docker Target Provider): NeedsRemove: returning false")
	return false
}

func (d *DockerTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	ctx, span := observability.StartSpan("Docker Target Provider", ctx, &map[string]string{
		"method": "NeedsUpdate",
	})
	sLog.Infof(" P (Docker Target): NeedsUpdate: %d - %d", len(desired), len(current))

	for _, dc := range desired {
		needsUpdate := true
		for _, cc := range current {
			if cc.Name == dc.Name {
				// compare container image
				if cc.Properties[model.ContainerImage] != dc.Properties[model.ContainerImage] {
					needsUpdate = true
					break
				}
				// compare container port
				if cc.Properties["container.ports"] != "" && dc.Properties["container.ports"] != "" && cc.Properties["container.ports"] != dc.Properties["container.ports"] {
					needsUpdate = true
					break
				}
				//compare container resources
				if cc.Properties["container.resources"] != "" && dc.Properties["container.resources"] != "" && cc.Properties["container.resources"] != dc.Properties["container.resources"] {
					needsUpdate = true
					break
				}
				needsUpdate = false
				break
			}
		}
		if needsUpdate {
			// container needs an update
			sLog.Info(" P (Docker Target): NeedsUpdate: returning true")
			observ_utils.CloseSpanWithError(span, nil)
			return true
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	sLog.Info(" P (Docker Target): NeedsUpdate: returning false")
	return false
}
func (*DockerTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{model.ContainerImage},
		OptionalProperties:    []string{"container.resources"},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}
