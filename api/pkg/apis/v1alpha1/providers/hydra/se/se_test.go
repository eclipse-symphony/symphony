/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package se

import (
	"encoding/json"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestInitWithNil(t *testing.T) {
	provider := SEProvider{}
	err := provider.Init(SEProviderConfig{})
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name": "name",
	}
	provider := SEProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestSetArtifact(t *testing.T) {
	jsonData := `{
    "Apps": [
      {
        "Version": "",
        "Kind": "container",
        "Metadata": {
          "Name": "softdpacc-1",
          "Uuid": "142292d7-dd0c-4a11-888e-3ad880ed4ce0",
          "Labels": {
            "type": "softdpac",
            "pairUniqueId": "16af79fa971b4a1bbce180999d178fd3",
            "originalAppName": "SoftdpacC_1",
            "partnerDeviceId": "1c014c0b-9f8f-4251-b138-4fde88492a9b",
            "partnerOwnerId": "4b0c5d10-fb9c-414b-a331-1859f778f1f4",
            "partnerName": "SoftdpacD_1",
            "partnerSoftdpacIP": "192.168.200.65",
            "preferredPrimary": "true"
          },
          "OwnerId": "4b0c5d10-fb9c-414b-a331-1859f778f1f4"
        },
        "Spec": {
          "Container": {
            "Name": "softdpacc-1",
            "Image": "",
            "Networks": [
              {
                "Ipv6": "",
                "Ipv4": "192.168.200.63",
                "NetworkId": "softdpacDeviceNet"
              }
            ],
            "Resources": {
              "Limits": {
                "Memory": "524288",
                "Cpus": "0,1,2,3"
              }
            }
          },
          "Blob": [],
          "HasBlob": false,
          "Affinity": {
            "PreferredHosts": [
              "22eb4ca1-3694-483d-a1c8-c188ec540377",
              "7613e804-a59d-4efc-a1d1-ada86954e3e0",
              "81d4ecbd-36f9-46aa-b13b-f096dde17bc7",
              "39ac2cd3-a6a4-446e-94e2-074f4083a3d5"
            ],
            "AppAntiAffinity": []
          },
          "DataCase": 1
        },
        "Status": {
          "Status": "Primary",
          "TimeStamp": "2025-08-05T22:37:48.8672836Z",
          "RunningHost": "",
          "InterlinkStatus": "Ok"
        },
        "Deleted": true
      },
      {
        "Version": "",
        "Kind": "container",
        "Metadata": {
          "Name": "softdpacd-1",
          "Uuid": "1c014c0b-9f8f-4251-b138-4fde88492a9b",
          "Labels": {
            "type": "softdpac",
            "pairUniqueId": "16af79fa971b4a1bbce180999d178fd3",
            "originalAppName": "SoftdpacD_1",
            "partnerDeviceId": "142292d7-dd0c-4a11-888e-3ad880ed4ce0",
            "partnerOwnerId": "4b0c5d10-fb9c-414b-a331-1859f778f1f4",
            "partnerName": "SoftdpacC_1",
            "partnerSoftdpacIP": "192.168.200.63",
            "preferredPrimary": "false"
          },
          "OwnerId": "4b0c5d10-fb9c-414b-a331-1859f778f1f4"
        },
        "Spec": {
          "Container": {
            "Name": "softdpacd-1",
            "Image": "",
            "Networks": [
              {
                "Ipv6": "",
                "Ipv4": "192.168.200.65",
                "NetworkId": "softdpacDeviceNet"
              }
            ],
            "Resources": {
              "Limits": {
                "Memory": "524288",
                "Cpus": "0,1,2,3"
              }
            }
          },
          "Blob": [],
          "HasBlob": false,
          "Affinity": {
            "PreferredHosts": [
              "22eb4ca1-3694-483d-a1c8-c188ec540377",
              "7613e804-a59d-4efc-a1d1-ada86954e3e0",
              "81d4ecbd-36f9-46aa-b13b-f096dde17bc7",
              "39ac2cd3-a6a4-446e-94e2-074f4083a3d5"
            ],
            "AppAntiAffinity": []
          },
          "DataCase": 1
        },
        "Status": {
          "Status": "Secondary",
          "TimeStamp": "2025-08-05T22:37:48.8677331Z",
          "RunningHost": "",
          "InterlinkStatus": "Ok"
        },
        "Deleted": true
      }
    ],
    "Devices": [
      {
        "Kind": "ipc",
        "Metadata": {
          "Name": "asrock-left",
          "Uuid": "22eb4ca1-3694-483d-a1c8-c188ec540377",
          "Labels": {
            "type": "iPCHADA",
            "originalDeviceName": "asrock_left"
          },
          "OwnerId": "4b0c5d10-fb9c-414b-a331-1859f778f1f4"
        },
        "Spec": {
          "Addresses": [
            "192.168.200.62"
          ],
          "Networks": [
            {
              "NetName": "softdpacDeviceNet",
              "NicName": "bond0",
              "RedundancyMode": "",
              "NicList": [],
              "Ipv4": "192.168.200.62",
              "Gateway": "192.168.1.1"
            },
            {
              "NetName": "softdpacInterlinkNet",
              "NicName": "eninterlink",
              "RedundancyMode": "",
              "NicList": [],
              "Ipv4": "10.10.1.64",
              "Gateway": "10.10.1.64"
            }
          ],
          "ContainerNetworks": [
            {
              "NetworkId": "softdpacDeviceNet",
              "Subnet": "192.168.0.0/16",
              "Gateway": "192.168.1.1",
              "NicName": "bond0",
              "Type": "macvlan"
            },
            {
              "NetworkId": "softdpacInterlinkNet",
              "Subnet": "10.10.1.0/24",
              "Gateway": "10.10.1.64",
              "NicName": "eninterlink",
              "Type": "macvlan"
            }
          ],
          "ReservedAppInterlinkIp": "10.10.1.65"
        },
        "TimeSetting": null,
        "Status": {
          "Status": "Reachable",
          "TimeStamp": "2025-08-05T22:37:48.8765753Z",
          "RunningAppInstances": [],
          "InterlinkStatus": "Ok",
          "NtpStatus": "Ok"
        },
        "Deleted": false
      },
      {
        "Kind": "ipc",
        "Metadata": {
          "Name": "asrock-right",
          "Uuid": "7613e804-a59d-4efc-a1d1-ada86954e3e0",
          "Labels": {
            "type": "iPCHADA",
            "originalDeviceName": "asrock_right"
          },
          "OwnerId": "4b0c5d10-fb9c-414b-a331-1859f778f1f4"
        },
        "Spec": {
          "Addresses": [
            "192.168.200.64"
          ],
          "Networks": [
            {
              "NetName": "softdpacDeviceNet",
              "NicName": "bond0",
              "RedundancyMode": "",
              "NicList": [],
              "Ipv4": "192.168.200.64",
              "Gateway": "192.168.1.1"
            },
            {
              "NetName": "softdpacInterlinkNet",
              "NicName": "eninterlink",
              "RedundancyMode": "",
              "NicList": [],
              "Ipv4": "10.10.1.66",
              "Gateway": "10.10.1.64"
            }
          ],
          "ContainerNetworks": [
            {
              "NetworkId": "softdpacDeviceNet",
              "Subnet": "192.168.0.0/16",
              "Gateway": "192.168.1.1",
              "NicName": "bond0",
              "Type": "macvlan"
            },
            {
              "NetworkId": "softdpacInterlinkNet",
              "Subnet": "10.10.1.0/24",
              "Gateway": "10.10.1.64",
              "NicName": "eninterlink",
              "Type": "macvlan"
            }
          ],
          "ReservedAppInterlinkIp": "10.10.1.67"
        },
        "TimeSetting": null,
        "Status": {
          "Status": "Reachable",
          "TimeStamp": "2025-08-05T22:37:48.8896309Z",
          "RunningAppInstances": [],
          "InterlinkStatus": "Ok",
          "NtpStatus": "Ok"
        },
        "Deleted": false
      },
      {
        "Kind": "ipc",
        "Metadata": {
          "Name": "asrock-spare",
          "Uuid": "81d4ecbd-36f9-46aa-b13b-f096dde17bc7",
          "Labels": {
            "type": "iPCHADA",
            "originalDeviceName": "asrock_spare"
          },
          "OwnerId": "4b0c5d10-fb9c-414b-a331-1859f778f1f4"
        },
        "Spec": {
          "Addresses": [
            "192.168.200.66"
          ],
          "Networks": [
            {
              "NetName": "softdpacDeviceNet",
              "NicName": "bond0",
              "RedundancyMode": "",
              "NicList": [],
              "Ipv4": "192.168.200.66",
              "Gateway": "192.168.1.1"
            },
            {
              "NetName": "softdpacInterlinkNet",
              "NicName": "eninterlink",
              "RedundancyMode": "",
              "NicList": [],
              "Ipv4": "10.10.1.68",
              "Gateway": "10.10.1.64"
            }
          ],
          "ContainerNetworks": [
            {
              "NetworkId": "softdpacDeviceNet",
              "Subnet": "192.168.0.0/16",
              "Gateway": "192.168.1.1",
              "NicName": "bond0",
              "Type": "macvlan"
            },
            {
              "NetworkId": "softdpacInterlinkNet",
              "Subnet": "10.10.1.0/24",
              "Gateway": "10.10.1.64",
              "NicName": "eninterlink",
              "Type": "macvlan"
            }
          ],
          "ReservedAppInterlinkIp": "10.10.1.69"
        },
        "TimeSetting": null,
        "Status": {
          "Status": "Reachable",
          "TimeStamp": "2025-08-05T22:37:48.8827865Z",
          "RunningAppInstances": [],
          "InterlinkStatus": "Ok",
          "NtpStatus": "Ok"
        },
        "Deleted": false
      },
      {
        "Kind": "ipc",
        "Metadata": {
          "Name": "ipcs3-vm-spare",
          "Uuid": "39ac2cd3-a6a4-446e-94e2-074f4083a3d5",
          "Labels": {
            "type": "iPCHADA",
            "originalDeviceName": "ipcs3_vm_spare"
          },
          "OwnerId": "4b0c5d10-fb9c-414b-a331-1859f778f1f4"
        },
        "Spec": {
          "Addresses": [
            "192.168.200.60"
          ],
          "Networks": [
            {
              "NetName": "softdpacDeviceNet",
              "NicName": "bond0",
              "RedundancyMode": "",
              "NicList": [],
              "Ipv4": "192.168.200.60",
              "Gateway": "192.168.1.1"
            },
            {
              "NetName": "softdpacInterlinkNet",
              "NicName": "eninterlink",
              "RedundancyMode": "",
              "NicList": [],
              "Ipv4": "10.10.1.60",
              "Gateway": "10.10.1.64"
            }
          ],
          "ContainerNetworks": [
            {
              "NetworkId": "softdpacDeviceNet",
              "Subnet": "192.168.0.0/16",
              "Gateway": "192.168.1.1",
              "NicName": "bond0",
              "Type": "macvlan"
            },
            {
              "NetworkId": "softdpacInterlinkNet",
              "Subnet": "10.10.1.0/24",
              "Gateway": "10.10.1.64",
              "NicName": "eninterlink",
              "Type": "macvlan"
            }
          ],
          "ReservedAppInterlinkIp": "10.10.1.61"
        },
        "TimeSetting": null,
        "Status": {
          "Status": "Reachable",
          "TimeStamp": "2025-08-05T22:37:48.8727949Z",
          "RunningAppInstances": [],
          "InterlinkStatus": "Ok",
          "NtpStatus": "Ok"
        },
        "Deleted": false
      }
    ]
  }`
	provider := SEProvider{}
	err := provider.Init(SEProviderConfig{
		Name:  "se",
		HASet: "ha-set1",
	})
	assert.Nil(t, err)
	artifactPack, err := provider.SetArtifact("state", []byte(jsonData))
	assert.NotNil(t, artifactPack)
	// check targets
	assert.Equal(t, 5, len(artifactPack.Targets))
	assert.Equal(t, "ha-set1", artifactPack.Targets[0].ObjectMeta.Labels["haSet"])
	assert.Equal(t, "ha-set1", artifactPack.Targets[1].ObjectMeta.Labels["haSet"])
	assert.Equal(t, "ha-set1", artifactPack.Targets[2].ObjectMeta.Labels["haSet"])
	assert.Equal(t, "ha-set1", artifactPack.Targets[3].ObjectMeta.Labels["haSet"])
	assert.Equal(t, "ipc", artifactPack.Targets[0].ObjectMeta.Labels["kind"])
	assert.Equal(t, "ipc", artifactPack.Targets[1].ObjectMeta.Labels["kind"])
	assert.Equal(t, "ipc", artifactPack.Targets[2].ObjectMeta.Labels["kind"])
	assert.Equal(t, "ipc", artifactPack.Targets[3].ObjectMeta.Labels["kind"])
	assert.Equal(t, "group", artifactPack.Targets[4].ObjectMeta.Labels["kind"])

	var selector model.TargetSelector
	err = json.Unmarshal([]byte(artifactPack.Targets[4].Spec.Topologies[0].Bindings[0].Config["targetSelector"]), &selector)
	assert.Equal(t, "instance", artifactPack.Targets[4].Spec.Topologies[0].Bindings[0].Role)
	assert.Equal(t, "providers.target.group", artifactPack.Targets[4].Spec.Topologies[0].Bindings[0].Provider)
	assert.Equal(t, "ha-set1", selector.LabelSelector["haSet"])
	assert.Equal(t, "ipc", selector.LabelSelector["kind"])

	// check solutioncontainers
	assert.Equal(t, 1, len(artifactPack.SolutionContainers))

	// check solutions
	assert.Equal(t, 1, len(artifactPack.Solutions))
	assert.Equal(t, 2, len(artifactPack.Solutions[0].Spec.Components))
	assert.Equal(t, "container", artifactPack.Solutions[0].Spec.Components[0].Type)
	assert.Equal(t, "container", artifactPack.Solutions[0].Spec.Components[1].Type)
	assert.Equal(t, artifactPack.SolutionContainers[0].ObjectMeta.Name+"-v-v1", artifactPack.Solutions[0].ObjectMeta.Name)
	assert.Equal(t, artifactPack.SolutionContainers[0].ObjectMeta.Name, artifactPack.Solutions[0].Spec.RootResource)

	// check instances
	assert.Equal(t, 1, len(artifactPack.Instances))
	assert.Equal(t, "ha-set1:v1", artifactPack.Instances[0].Spec.Solution)
	assert.Equal(t, "ha-set1", artifactPack.Instances[0].Spec.Target.LabelSelector["haSet"])
	assert.Equal(t, "group", artifactPack.Instances[0].Spec.Target.LabelSelector["kind"])

	assert.Nil(t, err)
}
