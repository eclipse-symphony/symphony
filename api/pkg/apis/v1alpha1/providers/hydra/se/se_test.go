/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package se

import (
	"testing"

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
 "apps": [
  {
   "version": "1.0",
   "kind": "container",
   "metadata": {
    "name": "Softdpac_2",
    "labels": {
     "partnerName": "Softdpac_3",
     "preferredPrimary": "true",
     "partnerSoftdpacIP": "192.168.2.82",
     "pairUniqueId": "7b33ca87de1e45f5a1374b8c28635e4",
     "uuid": "516e3960-c306-4191-a888-d70b8a496ed1",
     "type": "softdpac"
    }
   },
   "spec": {
    "container": {
     "name": "Softdpac_2",
     "image": "softdpac-ha:v23.0.23264.08",
     "networks": [
      {
       "name": "softdpacDeviceNet",
       "address": "192.168.2.81"
      },
      {
       "name": "softdpacInterlinkNet",
       "address": "169.254.39.42"
      }
     ],
     "resources": {
      "limits": {
       "memory": "524288",
       "cpus" : "0,1"
      }
     }
    },
    "affinity": {
     "preferredHosts": ["ipc_left", "ipc_right", "ipc_spare"],
     "appAntiAffinity": ["Softdpac_3"]
    }
   }
  },
  {
   "kind": "container",
   "metadata": {
    "name": "Softdpac_3",
    "labels": {
     "partnerName": "Softdpac_2",
     "preferredPrimary": "false",
     "partnerSoftdpacIP": "192.168.2.81",
     "uuid": "2eb00770-e4dc-4947-8849-bf9c18b0b139",
     "pairUniqueId": "7b33ca87de1e45f5a1374b8c28635e4",
     "type": "softdpac"
    }
   },
   "spec": {
    "container": {
     "name": "Softdpac_3",
     "image": "softdpac-ha:v23.0.23264.08",
     "networks": [
      {
       "name": "softdpacDeviceNet",
       "address": "192.168.2.82"
      },
      {
       "name": "softdpacInterlinkNet",
       "address": "169.254.39.43"
      }
     ],
     "resources": {
      "limits": {
       "memory": "524288",
       "cpus" : "0,1"
      }
     }
    },
    "affinity": {
     "preferredHosts": ["ipc_right", "ipc_left", "ipc_spare"],
     "appAntiAffinity": ["Softdpac_2"]
    }
   },
            "status": {
                "status": "running",
                "timeStamp": "2023-10-01T12:00:00Z",
                "runningHost": "ipc_right"
            }
  }
 ],
 "devices": [
  {
   "kind": "ipc",
   "metadata": {
    "name": "ipc_left",
    "labels": {
     "uuid": "4eee4cf5-2265-4be6-ba41-ff51c992702b"
    }
   },
   "spec": {
    "addresses": ["192.168.2.79"],
                "networks": [
                    {
                        "nicList": ["eth0", "eth1"],
                        "netName": "deviceNet",
                        "nicName": "bond0",
                        "redundancyMode": "",
                        "ipv4": "",
                        "gateway": ""
                    }
                ],
    "containerNetworks": [
     {
      "subnet": "169.254.39.40/29",
      "gateway": "169.254.39.41",
      "networkId": "softdpacInterlinkNet",
      "nicName": "eninterlink",
      "type": "macvlan"
     },
     
     {
      "subnet": "192.168.2.0/24",
      "gateway": "192.168.2.1",
      "networkId": "softdpacDeviceNet",
      "nicName": "bond0",
      "type": "macvlan"
     }
    ]
   }
  },
  {
   "kind": "ipc",
   "metadata": {
    "name": "ipc_right",
    "labels": {
     "uuid": "2878dcaa-9c34-4837-baf1-812216daae36"
    }
   },
   "spec": {
    "addresses": ["192.168.2.80"],
    "containerNetworks": [
     {
      "subnet": "169.254.39.40/29",
      "gateway": "169.254.39.41",
      "networkId": "softdpacInterlinkNet",
      "nicName": "eninterlink",
      "type": "macvlan"
     },
     
     {
      "subnet": "192.168.2.0/24",
      "gateway": "192.168.2.1",
      "networkId": "softdpacDeviceNet",
      "nicName": "bond0",
      "type": "macvlan"
     }
    ]
   }
  },
  {
   "kind": "ipc",
   "metadata": {
    "name": "ipc_spare",
    "labels": {
     "uuid": "b9adbc50-ae14-435d-9970-9ee6b4a5c201"
    }
   },
   "spec": {
    "addresses": ["192.168.2.237"],
    "containerNetworks": [
     {
      "subnet": "169.254.39.40/29",
      "gateway": "169.254.39.41",
      "networkId": "softdpacInterlinkNet",
      "nicName": "eninterlink",
      "type": "macvlan"
     },
     
     {
      "subnet": "192.168.2.0/24",
      "gateway": "192.168.2.1",
      "networkId": "softdpacDeviceNet",
      "nicName": "bond0",
      "type": "macvlan"
     }
    ]
   }
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
	assert.Nil(t, err)
}
