package iotedge

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
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

	components, err := provider.Get(context.Background(), model.DeploymentSpec{})
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

	components, err := provider.Get(context.Background(), model.DeploymentSpec{})
	assert.Nil(t, err)
	assert.NotNil(t, components)
	assert.Equal(t, 1, len(components)) //test vm device should have 1 module
	assert.Equal(t, "SimulatedTemperatureSensor", components[0].Name)
}

func TestToMoudleEmptyVersion(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"container.version": "",
			"container.type":    "docker",
			"container.image":   "docker/hello-world",
		},
	}
	_, err := toModule(component, "test", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleNoVersions(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"container.type":  "docker",
			"container.image": "docker/hello-world",
		},
	}
	_, err := toModule(component, "test", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleEmptyType(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"container.version": "1.0",
			"container.type":    "",
			"container.image":   "docker/hello-world",
		},
	}
	_, err := toModule(component, "test", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleNoTypes(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"container.version": "1.0",
			"container.image":   "docker/hello-world",
		},
	}
	_, err := toModule(component, "test", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleEmptyImage(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"container.version": "1.0",
			"container.type":    "docker",
			"container.image":   "",
		},
	}
	_, err := toModule(component, "test", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}

func TestToMoudleNoImage(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"container.version": "1.0",
			"container.type":    "docker",
		},
	}
	_, err := toModule(component, "test", "")
	assert.NotNil(t, err)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
}
func TestToModuleDesiredProperties(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"image":             "docker/hello-world",
			"desired.A1":        "ABC",
			"desired.A2":        "DEF",
			"container.version": "1.0",
			"container.type":    "docker",
			"container.image":   "redis",
		},
	}
	module, err := toModule(component, "test", "")
	assert.Nil(t, err)
	assert.Equal(t, "ABC", module.DesiredProperties["A1"].(string))
	assert.Equal(t, "DEF", module.DesiredProperties["A2"].(string))
}
func TestToModuleDesiredPropertiesWithAgent(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"image":             "docker/hello-world",
			"desired.A1":        "ABC",
			"desired.A2":        "DEF",
			"container.version": "1.0",
			"container.type":    "docker",
			"container.image":   "redis",
		},
	}
	module, err := toModule(component, "test", "symphony-agent")
	assert.Nil(t, err)
	assert.Equal(t, "ABC", module.DesiredProperties["A1"].(string))
	assert.Equal(t, "DEF", module.DesiredProperties["A2"].(string))
	assert.Equal(t, "target-runtime-symphony-agent", module.Environments[ENV_NAME].Value)
}
func TestToModuleEnvVariables(t *testing.T) {
	component := model.ComponentSpec{
		Properties: map[string]string{
			"image":             "docker/hello-world",
			"env.A1":            "ABC",
			"env.A2":            "DEF",
			"container.version": "1.0",
			"container.type":    "docker",
			"container.image":   "redis",
		},
	}
	module, err := toModule(component, "test", "")
	assert.Nil(t, err)
	assert.Equal(t, "ABC", module.Environments["A1"].Value)
	assert.Equal(t, "DEF", module.Environments["A2"].Value)
}
func TestUpdateDeployment(t *testing.T) {
	deployment := makeDefaultDeployment(nil, "", "")

	updateDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	updateDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	updateDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	reduceDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	reduceDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	updateDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	updateDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	updateDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	reduceDeployment(&deployment, "my-instance-1", map[string]Module{
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
	deployment := makeDefaultDeployment(nil, "", "")

	reduceDeployment(&deployment, "my-instance-1", map[string]Module{
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
