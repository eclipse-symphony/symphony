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

package v1

import (
	"testing"

	configv1 "gopls-workspace/apis/config/v1"
	configutils "gopls-workspace/configutils"

	apimodel "github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	k8smodel "github.com/azure/symphony/k8s/apis/model/v1"
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
