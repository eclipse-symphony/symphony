/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package iotedge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestInitWithNil(t *testing.T) {
	provider := IoTEdgeTargetProvider{}
	err := provider.Init(nil)
	assert.NotNil(t, err)
}

func TestInitWithEmptyAPIVersion(t *testing.T) {
	provider := IoTEdgeTargetProvider{}
	err := provider.Init(IoTEdgeTargetProviderConfig{})
	assert.Nil(t, err)
	assert.Equal(t, "2020-05-31-preview", provider.Config.APIVersion)
}

func TestInitWithEmptyKeyName(t *testing.T) {
	provider := IoTEdgeTargetProvider{}
	err := provider.Init(IoTEdgeTargetProviderConfig{})
	assert.Nil(t, err)
	assert.Equal(t, "iothubowner", provider.Config.KeyName)
}

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name": "name",
	}
	provider := IoTEdgeTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":             "name",
		"keyName":          "key",
		"key":              "value",
		"apiVersion":       "2020-05-31",
		"edgeAgentVersion": "1.2",
		"edgeHubVersion":   "1.2",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":       "name",
		"keyName":    "key",
		"key":        "value",
		"iotHub":     "iotHub",
		"deviceName": "device",
	}
	err = provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestGetNullDevice(t *testing.T) {
	ioTHubName := os.Getenv("S8CTEST_IOTHUB_NAME")
	if ioTHubName == "" {
		t.Skip("Skipping because S8CTEST_IOTHUB_NAME is not set")
	}

	provider := IoTEdgeTargetProvider{}
	err := provider.Init(IoTEdgeTargetProviderConfig{
		Name:       "iotedge",
		IoTHub:     os.Getenv("S8CTEST_IOTHUB_NAME"),
		DeviceName: os.Getenv("S8CTEST_IOTHUB_NULL_DEVICE_NAME"),
		APIVersion: os.Getenv("S8CTEST_IOTHUB_API_VERSION"),
		KeyName:    os.Getenv("S8CTEST_IOTHUB_KEY_NAME"),
		Key:        os.Getenv("S8CTEST_IOTHUB_KEY"),
	})
	assert.Nil(t, err)

	components, err := provider.Get(context.Background(), model.DeploymentSpec{}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, components)
	assert.Equal(t, 0, len(components)) //null device shouldn't have any modules
}

func TestGetVMDevice(t *testing.T) {
	ioTHubName := os.Getenv("S8CTEST_IOTHUB_NAME")
	if ioTHubName == "" {
		t.Skip("Skipping because S8CTEST_IOTHUB_NAME is not set")
	}

	provider := IoTEdgeTargetProvider{}
	err := provider.Init(IoTEdgeTargetProviderConfig{
		Name:       "iotedge",
		IoTHub:     os.Getenv("S8CTEST_IOTHUB_NAME"),
		DeviceName: os.Getenv("S8CTEST_IOTHUB_VM_DEVICE_NAME"),
		APIVersion: os.Getenv("S8CTEST_IOTHUB_API_VERSION"),
		KeyName:    os.Getenv("S8CTEST_IOTHUB_KEY_NAME"),
		Key:        os.Getenv("S8CTEST_IOTHUB_KEY"),
	})
	assert.Nil(t, err)

	components, err := provider.Get(context.Background(), model.DeploymentSpec{}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, components)
	assert.Equal(t, 1, len(components)) //test vm device should have 1 module
	assert.Equal(t, "SimulatedTemperatureSensor", components[0].Name)
}

func TestToMoudleEmptyVersion(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"container.version":  "",
			"container.type":     "docker",
			model.ContainerImage: "docker/hello-world",
		},
	}
	_, err := toModule(component, "test", "", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleNoVersions(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"container.type":     "docker",
			model.ContainerImage: "docker/hello-world",
		},
	}
	_, err := toModule(component, "test", "", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleEmptyType(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"container.version":  "1.0",
			"container.type":     "",
			model.ContainerImage: "docker/hello-world",
		},
	}
	_, err := toModule(component, "test", "", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleNoTypes(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"container.version":  "1.0",
			model.ContainerImage: "docker/hello-world",
		},
	}
	_, err := toModule(component, "test", "", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleEmptyImage(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"container.version":  "1.0",
			"container.type":     "docker",
			model.ContainerImage: "",
		},
	}
	_, err := toModule(component, "test", "", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleNoImage(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"container.version": "1.0",
			"container.type":    "docker",
		},
	}
	_, err := toModule(component, "test", "", "target")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}
func TestToModuleDesiredProperties(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"image":              "docker/hello-world",
			"desired.A1":         "ABC",
			"desired.A2":         "DEF",
			"container.version":  "1.0",
			"container.type":     "docker",
			model.ContainerImage: "redis",
		},
	}
	module, err := toModule(component, "test", "", "target")
	assert.Nil(t, err)
	assert.Equal(t, "ABC", module.DesiredProperties["A1"].(string))
	assert.Equal(t, "DEF", module.DesiredProperties["A2"].(string))
}
func TestToModuleDesiredPropertiesWithAgent(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"image":              "docker/hello-world",
			"desired.A1":         "ABC",
			"desired.A2":         "DEF",
			"container.version":  "1.0",
			"container.type":     "docker",
			model.ContainerImage: "redis",
		},
	}
	module, err := toModule(component, "test", "symphony-agent", "target")
	assert.Nil(t, err)
	assert.Equal(t, "ABC", module.DesiredProperties["A1"].(string))
	assert.Equal(t, "DEF", module.DesiredProperties["A2"].(string))
	assert.Equal(t, "target-runtime-target-symphony-agent", module.Environments[ENV_NAME].Value)
}
func TestToModuleEnvVariables(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]interface{}{
			"image":              "docker/hello-world",
			"env.A1":             "ABC",
			"env.A2":             "DEF",
			"container.version":  "1.0",
			"container.type":     "docker",
			model.ContainerImage: "redis",
		},
	}
	module, err := toModule(component, "test", "", "target")
	assert.Nil(t, err)
	assert.Equal(t, "ABC", module.Environments["A1"].Value)
	assert.Equal(t, "DEF", module.Environments["A2"].Value)
}
func TestUpdateDeployment(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	updateDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
	}, ModuleTwin{}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 1, len(col))
	assert.Equal(t, "ABC", col["my-instance-1-my-module"].Settings["123"])
}
func TestUpdateDeploymentWithExistingModule(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	updateDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
	}, ModuleTwin{
		ModuleId: "king",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"modules": map[string]interface{}{
					"other-module": Module{
						Settings: map[string]string{
							"123": "DEF",
						},
					},
				},
			},
		},
	}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 2, len(col))
	assert.Equal(t, "ABC", col["my-instance-1-my-module"].Settings["123"])
	assert.Equal(t, "DEF", col["other-module"].Settings["123"])
}

func TestUpdateDeploymentWithCurrentModule(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	updateDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
	}, ModuleTwin{
		ModuleId: "king",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"modules": map[string]interface{}{
					"my-instance-1-my-module": Module{
						Settings: map[string]string{
							"123": "DEF",
						},
					},
				},
			},
		},
	}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 1, len(col))
	assert.Equal(t, "ABC", col["my-instance-1-my-module"].Settings["123"])
}
func TestReduceDeploymentWithCurrentModule(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	reduceDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
	}, ModuleTwin{
		ModuleId: "king",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"modules": map[string]interface{}{
					"my-instance-1-my-module": Module{
						Settings: map[string]string{
							"123": "DEF",
						},
					},
				},
			},
		},
	}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 0, len(col))
}
func TestReduceDeploymentWithMoreModule(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	reduceDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
	}, ModuleTwin{
		ModuleId: "king",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"modules": map[string]interface{}{
					"my-instance-1-my-module": Module{
						Settings: map[string]string{
							"123": "DEF",
						},
					},
					"my-instance-2-my-module": Module{
						Settings: map[string]string{
							"123": "HIJ",
						},
					},
				},
			},
		},
	}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 1, len(col))
	assert.Equal(t, "HIJ", col["my-instance-2-my-module"].Settings["123"])
}
func TestModifyRoutes(t *testing.T) {
	route := "FROM /messages/modules/SimulatedTemperatureSensor/outputs/temperatureOutput INTO BrokeredEndpoint(\"/modules/filtermodule/inputs/input1\")"
	route = modifyRoutes(route, "my-instance-1", map[string]Module{
		"SimulatedTemperatureSensor": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
		"filtermodule": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
	}, map[string]bool{
		"other": true,
	})
	assert.Equal(t, "FROM /messages/modules/my-instance-1-SimulatedTemperatureSensor/outputs/temperatureOutput INTO BrokeredEndpoint(\"/modules/my-instance-1-filtermodule/inputs/input1\")", route)
}
func TestModifyRoutesHitOther(t *testing.T) {
	route := "FROM /messages/modules/SimulatedTemperatureSensor/outputs/temperatureOutput INTO BrokeredEndpoint(\"/modules/filtermodule/inputs/input1\")"
	route = modifyRoutes(route, "my-instance-1", map[string]Module{
		"SimulatedTemperatureSensor": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
		"filtermodule": {
			Settings: map[string]string{
				"123": "ABC",
			},
		},
	}, map[string]bool{
		"filtermodule": true,
	})
	assert.Equal(t, "FROM /messages/modules/my-instance-1-SimulatedTemperatureSensor/outputs/temperatureOutput INTO BrokeredEndpoint(\"/modules/filtermodule/inputs/input1\")", route)
}
func TestUpdateDeploymentWithRoutes(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	updateDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
			IotHubRoutes: map[string]string{
				"route-1": "FROM messagees/modules/my-module INTO messages/modules/other",
			},
		},
	}, ModuleTwin{}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 1, len(col))
	col2 := deployment.ModulesContent["$edgeHub"].DesiredProperties["routes"].(map[string]string)
	assert.Equal(t, "FROM messagees/modules/my-instance-1-my-module INTO messages/modules/other", col2["my-instance-1-route-1"])
}
func TestUpdateDeploymentWithExistingModuleAndRoute(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	updateDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
			IotHubRoutes: map[string]string{
				"route-1": "FROM messagees/modules/my-module INTO messages/modules/other",
			},
		},
	}, ModuleTwin{
		ModuleId: "king",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"modules": map[string]interface{}{
					"other-module": Module{
						Settings: map[string]string{
							"123": "DEF",
						},
						IotHubRoutes: map[string]string{
							"route-1": "FROM messagees/modules/other-module INTO messages/modules/another",
						},
					},
				},
				"routes": map[string]string{
					"route-1": "FROM messagees/modules/other-module INTO messages/modules/another",
				},
			},
		},
	}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 2, len(col))
	col2 := deployment.ModulesContent["$edgeHub"].DesiredProperties["routes"].(map[string]string)
	assert.Equal(t, "FROM messagees/modules/my-instance-1-my-module INTO messages/modules/other", col2["my-instance-1-route-1"])
	assert.Equal(t, "FROM messagees/modules/other-module INTO messages/modules/another", col2["route-1"])
}
func TestUpdateDeploymentWithCurrentModuleAndRoutes(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	updateDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
			IotHubRoutes: map[string]string{
				"route-1": "FROM messagees/modules/my-module INTO messages/modules/cool",
			},
		},
	}, ModuleTwin{
		ModuleId: "king",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"modules": map[string]interface{}{
					"my-instance-1-my-module": Module{
						Settings: map[string]string{
							"123": "DEF",
						},
						IotHubRoutes: map[string]string{
							"my-instance-1-route-1": "FROM messagees/my-instance-1-my-module/my-module INTO messages/modules/other",
						},
					},
				},
				"routes": map[string]string{
					"my-instance-1-route-1": "FROM messagees/my-instance-1-my-module/my-module INTO messages/modules/other",
				},
			},
		},
	}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 1, len(col))
	col2 := deployment.ModulesContent["$edgeHub"].DesiredProperties["routes"].(map[string]string)
	assert.Equal(t, "FROM messagees/modules/my-instance-1-my-module INTO messages/modules/cool", col2["my-instance-1-route-1"])
}
func TestReduceDeploymentWithCurrentModuleRoute(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	reduceDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
			IotHubRoutes: map[string]string{
				"route-1": "FROM messagees/modules/my-module INTO messages/modules/cool",
			},
		},
	}, ModuleTwin{
		ModuleId: "king",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"modules": map[string]interface{}{
					"my-instance-1-my-module": Module{
						Settings: map[string]string{
							"123": "DEF",
						},
						IotHubRoutes: map[string]string{
							"my-instance-1-route-1": "FROM messagees/my-instance-1-my-module/my-module INTO messages/modules/other",
						},
					},
				},
				"routes": map[string]string{
					"my-instance-1-route-1": "FROM messagees/my-instance-1-my-module/my-module INTO messages/modules/other",
				},
			},
		},
	}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 0, len(col))
	col2 := deployment.ModulesContent["$edgeHub"].DesiredProperties["routes"].(map[string]string)
	assert.Equal(t, 0, len(col2))
}
func TestReduceDeploymentWithMoreModuleAndRoutes(t *testing.T) {
	deployment := makeDefaultDeployment(context.Background(), nil, "", "")

	reduceDeployment(context.Background(), &deployment, "my-instance-1", map[string]Module{
		"my-module": {
			Settings: map[string]string{
				"123": "ABC",
			},
			IotHubRoutes: map[string]string{
				"route-1": "FROM messagees/modules/my-module INTO messages/modules/cool",
			},
		},
	}, ModuleTwin{
		ModuleId: "king",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"modules": map[string]interface{}{
					"my-instance-1-my-module": Module{
						Settings: map[string]string{
							"123": "DEF",
						},
						IotHubRoutes: map[string]string{
							"my-instance-1-route-1": "FROM messagees/modules/my-instance-1-my-module INTO messages/modules/cool",
						},
					},
					"my-instance-2-my-module": Module{
						Settings: map[string]string{
							"123": "HIJ",
						},
						IotHubRoutes: map[string]string{
							"my-instance-2-route-1": "FROM messagees/modules/my-instance-2-my-module INTO messages/modules/cool",
						},
					},
				},
				"routes": map[string]string{
					"my-instance-1-route-1": "FROM messagees/my-instance-1-my-module/my-module INTO messages/modules/other",
					"my-instance-2-route-1": "FROM messagees/modules/my-instance-2-my-module INTO messages/modules/cool",
				},
			},
		},
	}, ModuleTwin{})

	col := deployment.ModulesContent["$edgeAgent"].DesiredProperties["modules"].(map[string]Module)
	assert.Equal(t, 1, len(col))
	assert.Equal(t, "HIJ", col["my-instance-2-my-module"].Settings["123"])
	col2 := deployment.ModulesContent["$edgeHub"].DesiredProperties["routes"].(map[string]string)
	assert.Equal(t, 1, len(col2))
	assert.Equal(t, "FROM messagees/modules/my-instance-2-my-module INTO messages/modules/cool", col2["my-instance-2-route-1"])
}

func TestGetFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	assert.Nil(t, err)

	provider := IoTEdgeTargetProvider{}
	err = provider.Init(IoTEdgeTargetProviderConfig{
		Name:       "name",
		IoTHub:     u.Host,
		DeviceName: "device",
		KeyName:    "key",
		Key:        "value",
	})
	assert.Nil(t, err)

	_, err = provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)
	assert.NotNil(t, err)
}

func TestApplyFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	assert.Nil(t, err)

	provider := IoTEdgeTargetProvider{}
	err = provider.Init(IoTEdgeTargetProviderConfig{
		Name:       "name",
		IoTHub:     u.Host,
		DeviceName: "device",
		KeyName:    "key",
		Key:        "value",
	})
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "MyApp",
		Properties: map[string]interface{}{
			"container.image":   "image",
			"container.version": "1.0.0",
			"container.type":    "type",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}

func TestToComponent(t *testing.T) {
	hubTwin := ModuleTwin{
		DeviceId: "id",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"routes": map[string]interface{}{
					"rt1": "modules/mid/abc",
				},
			},
		},
	}
	twin := ModuleTwin{
		DeviceId: "id",
		ModuleId: "mid",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"a": 1,
				"b": 1.22,
				"c": true,
				"d": "abc",
				"e": uint(123),
				"f": int8(123),
				"g": int16(123),
				"h": int32(123),
				"i": int64(123),
				"j": uint8(123),
				"k": uint16(123),
				"l": uint32(123),
				"m": uint64(123),
				"n": uint32(123),
				"o": uint32(123),
				"p": float32(123.12),
				"q": float64(123.12),
				"z": []interface{}{"hello", 123, 4.56, true},
			},
		},
	}
	module := Module{
		RestartPolicy: "policy",
		Version:       "1.0.0",
		Type:          "type",
		Settings: map[string]string{
			"createOptions": "opt",
			"image":         "img",
		},
	}
	component, err := toComponent(hubTwin, twin, "default", module)
	assert.Nil(t, err)
	assert.NotNil(t, component)
}

func TestGetIoTEdgeModules(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	assert.Nil(t, err)

	provider := IoTEdgeTargetProvider{}
	err = provider.Init(IoTEdgeTargetProviderConfig{
		Name:       "name",
		IoTHub:     u.Host,
		DeviceName: "device",
		KeyName:    "key",
		Key:        "value",
	})
	assert.Nil(t, err)

	twin := ModuleTwin{
		DeviceId: "id",
		ModuleId: "mid",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"routes": map[string]interface{}{
					"rt1": "modules/mid/abc",
				},
				"modules": map[string]interface{}{
					"m1": "value1",
				},
			},
		},
	}

	_, err = provider.getIoTEdgeModules(context.Background())
	assert.NotNil(t, err)

	err = provider.deployToIoTEdge(context.Background(), "default", map[string]string{}, nil, twin, twin)
	assert.NotNil(t, err)

	err = provider.remvoefromIoTEdge(context.Background(), "default", map[string]string{}, nil, twin, twin)
	assert.NotNil(t, err)
}

func TestDeployRemove(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	assert.Nil(t, err)

	provider := IoTEdgeTargetProvider{}
	err = provider.Init(IoTEdgeTargetProviderConfig{
		Name:       "name",
		IoTHub:     u.Host,
		DeviceName: "device",
		KeyName:    "key",
		Key:        "value",
	})
	assert.Nil(t, err)

	twin := ModuleTwin{
		DeviceId: "id",
		ModuleId: "mid",
		Properties: ModuleTwinProperties{
			Desired: map[string]interface{}{
				"routes": map[string]interface{}{
					"rt1": "modules/mid/abc",
				},
				"modules": map[string]interface{}{
					"m1": "value1",
				},
			},
		},
	}

	err = provider.deployToIoTEdge(context.Background(), "default", map[string]string{}, nil, twin, twin)
	assert.NotNil(t, err)

	err = provider.remvoefromIoTEdge(context.Background(), "default", map[string]string{}, nil, twin, twin)
	assert.NotNil(t, err)
}

func TestApplyIoTEdgeDeployment(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	assert.Nil(t, err)

	provider := IoTEdgeTargetProvider{}
	err = provider.Init(IoTEdgeTargetProviderConfig{
		Name:       "name",
		IoTHub:     u.Host,
		DeviceName: "device",
		KeyName:    "key",
		Key:        "value",
	})
	assert.Nil(t, err)

	deployment := IoTEdgeDeployment{
		ModulesContent: map[string]ModuleState{
			"desire1": {
				map[string]interface{}{
					"m1": "value1",
				},
			},
		},
	}
	err = provider.applyIoTEdgeDeployment(context.Background(), deployment)
	assert.NotNil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &IoTEdgeTargetProvider{}
	_ = provider.Init(IoTEdgeTargetProvider{})
	// assert.Nil(t, err) it's okay if provider is not fully initialized
	conformance.ConformanceSuite(t, provider)
}
