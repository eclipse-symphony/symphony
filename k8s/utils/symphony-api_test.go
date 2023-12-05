/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	symphonyv1 "gopls-workspace/apis/symphony.microsoft.com/v1"
	"testing"

	k8smodel "github.com/azure/symphony/k8s/apis/model/v1"

	apimodel "github.com/azure/symphony/api/pkg/apis/v1alpha1/model"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSimpleStringMatch(t *testing.T) {
	assert.True(t, matchString("CPU", "CPU"))
}
func TestSimpleStringMismatch(t *testing.T) {
	assert.False(t, matchString("CPU", "cpu"))
}
func TestWildcardMatch(t *testing.T) {
	assert.True(t, matchString("CP*", "CPU-x64"))
}
func TestWildcardMatchShort(t *testing.T) {
	assert.True(t, matchString("CP*", "CP"))
}
func TestWildcardMatchShortMiddle(t *testing.T) {
	assert.True(t, matchString("C*U", "CU"))
}
func TestWildcardMatchMultiple(t *testing.T) {
	assert.True(t, matchString("C*P*", "CPU"))
}
func TestSingleWildcardMatch(t *testing.T) {
	assert.True(t, matchString("C%U", "CPU"))
}
func TestSingleWildcardMatchMultiple(t *testing.T) {
	assert.True(t, matchString("C%U%", "CPU1"))
}
func TestSingleWildcardMismatch(t *testing.T) {
	assert.False(t, matchString("C%U", "CPP"))
}

func TestDirectTargetMatch(t *testing.T) {
	instance := symphonyv1.Instance{
		Spec: apimodel.InstanceSpec{
			Target: apimodel.TargetSelector{
				Name: "gateway-1",
			},
		},
	}
	targets := symphonyv1.TargetList{
		Items: []symphonyv1.Target{
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-1",
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-2",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 1, len(ts))
	assert.Equal(t, "gateway-1", ts[0].ObjectMeta.Name)
}
func TestDirectTargetMismatch(t *testing.T) {
	instance := symphonyv1.Instance{
		Spec: apimodel.InstanceSpec{
			Target: apimodel.TargetSelector{
				Name: "gateway-1",
			},
		},
	}
	targets := symphonyv1.TargetList{
		Items: []symphonyv1.Target{
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-2",
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-3",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-3",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 0, len(ts))
}
func TestTargetWildcardMatch(t *testing.T) {
	instance := symphonyv1.Instance{
		Spec: apimodel.InstanceSpec{
			Target: apimodel.TargetSelector{
				Name: "gateway*",
			},
		},
	}
	targets := symphonyv1.TargetList{
		Items: []symphonyv1.Target{
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-1",
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-2",
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "file-server",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "file-server",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 2, len(ts))
}
func TestTargetSincleWildcardMatch(t *testing.T) {
	instance := symphonyv1.Instance{
		Spec: apimodel.InstanceSpec{
			Target: apimodel.TargetSelector{
				Name: "gateway%1",
			},
		},
	}
	targets := symphonyv1.TargetList{
		Items: []symphonyv1.Target{
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-1",
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway_1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway_1",
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway1",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 2, len(ts))
}
func TestTargetSelectorMatch(t *testing.T) {
	instance := symphonyv1.Instance{
		Spec: apimodel.InstanceSpec{
			Target: apimodel.TargetSelector{
				Selector: map[string]string{
					"OS": "windows",
				},
			},
		},
	}
	targets := symphonyv1.TargetList{
		Items: []symphonyv1.Target{
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-1",
					Properties: map[string]string{
						"OS": "windows",
					},
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-2",
					Properties: map[string]string{
						"OS": "linux",
					},
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-1",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 1, len(ts))
	assert.Equal(t, "gateway-1", ts[0].ObjectMeta.Name)
}
func TestTargetSelectorMismatch(t *testing.T) {
	instance := symphonyv1.Instance{
		Spec: apimodel.InstanceSpec{
			Target: apimodel.TargetSelector{
				Selector: map[string]string{
					"OS": "windows",
				},
			},
		},
	}
	targets := symphonyv1.TargetList{
		Items: []symphonyv1.Target{
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-1",
					Properties: map[string]string{
						"OS": "linux",
					},
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-2",
					Properties: map[string]string{
						"OS": "mac",
					},
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-3",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 0, len(ts))
}
func TestTargetSelectorWildcardMatch(t *testing.T) {
	instance := symphonyv1.Instance{
		Spec: apimodel.InstanceSpec{
			Target: apimodel.TargetSelector{
				Selector: map[string]string{
					"tag": "floor-*",
				},
			},
		},
	}
	targets := symphonyv1.TargetList{
		Items: []symphonyv1.Target{
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-1",
					Properties: map[string]string{
						"tag": "floor-23",
					},
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-2",
					Properties: map[string]string{
						"tag": "floor-34",
					},
				},
			},
			symphonyv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-3",
				},
				Spec: k8smodel.TargetSpec{
					DisplayName: "gateway-3",
					Properties: map[string]string{
						"tag": "floor-88",
					},
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 3, len(ts))
}

func TestCreateSymphonyDeployment(t *testing.T) {
	instance := symphonyv1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-instance",
		},
		Spec: apimodel.InstanceSpec{
			Target: apimodel.TargetSelector{
				Selector: map[string]string{
					"tag": "floor-*",
				},
			},
		},
	}
	targets := []symphonyv1.Target{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gateway-1",
			},
			Spec: k8smodel.TargetSpec{
				DisplayName: "gateway-1",
				Properties: map[string]string{
					"tag": "floor-23",
				},
				Components: []k8smodel.ComponentSpec{
					{
						Properties: runtime.RawExtension{
							Raw: []byte(`{"targetKey": "targetValue", "nested": {"targetKey2": "targetValue2"}}`),
						},
					},
				},
			},
		},
	}

	solution := symphonyv1.Solution{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-solution",
		},
		Spec: k8smodel.SolutionSpec{
			Components: []k8smodel.ComponentSpec{
				{
					Properties: runtime.RawExtension{
						Raw: []byte(`{"key": "value", "nested": {"key": "value"}}`),
					},
				},
			},
		},
	}
	deployment, err := CreateSymphonyDeployment(instance, solution, targets)
	assert.NoError(t, err)
	assert.IsType(t, map[string]interface{}{}, deployment.Solution.Components[0].Properties)
	assert.IsType(t, map[string]interface{}{}, deployment.Targets["gateway-1"].Components[0].Properties)
	propertiesMap := deployment.Solution.Components[0].Properties
	assert.Equal(t, "value", propertiesMap["key"])
	assert.Equal(t, "value", propertiesMap["nested"].(map[string]interface{})["key"])
	targetPropertiesMap := deployment.Targets["gateway-1"].Components[0].Properties
	assert.Equal(t, "targetValue", targetPropertiesMap["targetKey"])
	assert.Equal(t, "targetValue2", targetPropertiesMap["nested"].(map[string]interface{})["targetKey2"])
}
