/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"testing"

	configv1 "gopls-workspace/apis/config/v1"
	configutils "gopls-workspace/configutils"

	k8smodel "gopls-workspace/apis/model/v1"

	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplyValidationPoliciesNoItem(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"target": []configv1.ValidationPolicy{
			{
				SelectorType:   "topologies.bindings",
				SelectorKey:    "provider",
				SelectorValue:  "providers.target.azure.iotedge",
				SpecField:      "binding.config.deviceName",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := TargetList{
		Items: []Target{},
	}
	for _, p := range policies["target"] {
		pack := extractTargetValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", "my-device", p.ValidationType, pack)
		assert.Nil(t, err)
		assert.Equal(t, "", ret)
	}
}

func TestApplyValidationPoliciesSingleItem(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"target": []configv1.ValidationPolicy{
			{
				SelectorType:   "topologies.bindings",
				SelectorKey:    "provider",
				SelectorValue:  "providers.target.azure.iotedge",
				SpecField:      "binding.config.deviceName",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := TargetList{
		Items: []Target{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: k8smodel.TargetSpec{
					Topologies: []apimodel.TopologySpec{
						{
							Bindings: []apimodel.BindingSpec{
								{
									Provider: "providers.target.azure.iotedge",
									Config: map[string]string{
										"deviceName": "my-device",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, p := range policies["target"] {
		pack := extractTargetValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", "my-device", p.ValidationType, pack)
		assert.Nil(t, err)
		assert.Equal(t, "", ret)
	}
}

func TestApplyValidationPoliciesNoDuplicated(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"target": []configv1.ValidationPolicy{
			{
				SelectorType:   "topologies.bindings",
				SelectorKey:    "provider",
				SelectorValue:  "providers.target.azure.iotedge",
				SpecField:      "binding.config.deviceName",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := TargetList{
		Items: []Target{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: k8smodel.TargetSpec{
					Topologies: []apimodel.TopologySpec{
						{
							Bindings: []apimodel.BindingSpec{
								{
									Provider: "providers.target.azure.iotedge",
									Config: map[string]string{
										"deviceName": "my-device",
									},
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake2",
				},
				Spec: k8smodel.TargetSpec{
					Topologies: []apimodel.TopologySpec{
						{
							Bindings: []apimodel.BindingSpec{
								{
									Provider: "providers.target.azure.iotedge",
									Config: map[string]string{
										"deviceName": "my-device2",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, p := range policies["target"] {
		pack := extractTargetValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", "my-device", p.ValidationType, pack)
		assert.Nil(t, err)
		assert.Equal(t, "", ret)
	}
}

func TestApplyValidationPoliciesDuplicated(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"target": []configv1.ValidationPolicy{
			{
				SelectorType:   "topologies.bindings",
				SelectorKey:    "provider",
				SelectorValue:  "providers.target.azure.iotedge",
				SpecField:      "binding.config.deviceName",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := TargetList{
		Items: []Target{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: k8smodel.TargetSpec{
					Topologies: []apimodel.TopologySpec{
						{
							Bindings: []apimodel.BindingSpec{
								{
									Provider: "providers.target.azure.iotedge",
									Config: map[string]string{
										"deviceName": "my-device",
									},
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake2",
				},
				Spec: k8smodel.TargetSpec{
					Topologies: []apimodel.TopologySpec{
						{
							Bindings: []apimodel.BindingSpec{
								{
									Provider: "providers.target.azure.iotedge",
									Config: map[string]string{
										"deviceName": "my-device2",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, p := range policies["target"] {
		pack := extractTargetValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", "my-device2", p.ValidationType, pack)
		assert.Nil(t, err)
		assert.NotEqual(t, "", ret)
	}
}

func TestApplyValidationPoliciesNoConflict(t *testing.T) {
	policies := map[string][]configv1.ValidationPolicy{
		"target": []configv1.ValidationPolicy{
			{
				SelectorType:   "topologies.bindings",
				SelectorKey:    "provider",
				SelectorValue:  "providers.target.azure.iotedge",
				SpecField:      "binding.config.deviceName",
				ValidationType: "unique",
				Message:        "there's already a target associated with the IoT Edge device: %s",
			},
		},
	}
	list := TargetList{
		Items: []Target{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake",
				},
				Spec: k8smodel.TargetSpec{
					Topologies: []apimodel.TopologySpec{
						{
							Bindings: []apimodel.BindingSpec{
								{
									Provider: "providers.target.azure.iotedge",
									Config: map[string]string{
										"deviceName": "my-device",
									},
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quake2",
				},
				Spec: k8smodel.TargetSpec{
					Topologies: []apimodel.TopologySpec{
						{
							Bindings: []apimodel.BindingSpec{
								{
									Provider: "providers.target.azure.iotedge",
									Config: map[string]string{
										"deviceName": "my-device2",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, p := range policies["target"] {
		pack := extractTargetValidationPack(list, p)
		ret, err := configutils.CheckValidationPack("quake", "my-device3", p.ValidationType, pack)
		assert.Nil(t, err)
		assert.Equal(t, "", ret)
	}
}
