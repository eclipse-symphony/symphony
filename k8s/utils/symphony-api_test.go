package utils

import (
	fabricv1 "gopls-workspace/apis/fabric/v1"
	solutionv1 "gopls-workspace/apis/solution/v1"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	instance := solutionv1.Instance{
		Spec: solutionv1.InstanceSpec{
			Stages: []solutionv1.StageSpec{
				{
					Target: solutionv1.TargetSpec{
						Name: "gateway-1",
					},
				},
			},
		},
	}
	targets := fabricv1.TargetList{
		Items: []fabricv1.Target{
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-1",
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: fabricv1.TargetSpec{
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
	instance := solutionv1.Instance{
		Spec: solutionv1.InstanceSpec{
			Stages: []solutionv1.StageSpec{
				{
					Target: solutionv1.TargetSpec{
						Name: "gateway-1",
					},
				},
			},
		},
	}
	targets := fabricv1.TargetList{
		Items: []fabricv1.Target{
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-2",
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-3",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-3",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 0, len(ts))
}
func TestTargetWildcardMatch(t *testing.T) {
	instance := solutionv1.Instance{
		Spec: solutionv1.InstanceSpec{
			Stages: []solutionv1.StageSpec{
				{
					Target: solutionv1.TargetSpec{
						Name: "gateway*",
					},
				},
			},
		},
	}
	targets := fabricv1.TargetList{
		Items: []fabricv1.Target{
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-1",
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-2",
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "file-server",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "file-server",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 2, len(ts))
}
func TestTargetSincleWildcardMatch(t *testing.T) {
	instance := solutionv1.Instance{
		Spec: solutionv1.InstanceSpec{
			Stages: []solutionv1.StageSpec{
				{
					Target: solutionv1.TargetSpec{
						Name: "gateway%1",
					},
				},
			},
		},
	}
	targets := fabricv1.TargetList{
		Items: []fabricv1.Target{
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-1",
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway_1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway_1",
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway1",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 2, len(ts))
}
func TestTargetSelectorMatch(t *testing.T) {
	instance := solutionv1.Instance{
		Spec: solutionv1.InstanceSpec{
			Stages: []solutionv1.StageSpec{
				{
					Target: solutionv1.TargetSpec{
						Selector: map[string]string{
							"OS": "windows",
						},
					},
				},
			},
		},
	}
	targets := fabricv1.TargetList{
		Items: []fabricv1.Target{
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-1",
					Properties: map[string]string{
						"OS": "windows",
					},
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-2",
					Properties: map[string]string{
						"OS": "linux",
					},
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: fabricv1.TargetSpec{
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
	instance := solutionv1.Instance{
		Spec: solutionv1.InstanceSpec{
			Stages: []solutionv1.StageSpec{
				{
					Target: solutionv1.TargetSpec{
						Selector: map[string]string{
							"OS": "windows",
						},
					},
				},
			},
		},
	}
	targets := fabricv1.TargetList{
		Items: []fabricv1.Target{
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-1",
					Properties: map[string]string{
						"OS": "linux",
					},
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-2",
					Properties: map[string]string{
						"OS": "mac",
					},
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-3",
				},
			},
		},
	}
	ts := MatchTargets(instance, targets)
	assert.Equal(t, 0, len(ts))
}
func TestTargetSelectorWildcardMatch(t *testing.T) {
	instance := solutionv1.Instance{
		Spec: solutionv1.InstanceSpec{
			Stages: []solutionv1.StageSpec{
				{
					Target: solutionv1.TargetSpec{
						Selector: map[string]string{
							"tag": "floor-*",
						},
					},
				},
			},
		},
	}
	targets := fabricv1.TargetList{
		Items: []fabricv1.Target{
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-1",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-1",
					Properties: map[string]string{
						"tag": "floor-23",
					},
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-2",
				},
				Spec: fabricv1.TargetSpec{
					DisplayName: "gateway-2",
					Properties: map[string]string{
						"tag": "floor-34",
					},
				},
			},
			fabricv1.Target{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-3",
				},
				Spec: fabricv1.TargetSpec{
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
