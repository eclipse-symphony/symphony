/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package remoteAgent

import (
	"context"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/targets"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	k8ssecret "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/secret"
	k8sstate "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/states/k8s"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRemoteAgentSchedulerInit(t *testing.T) {
	vendorContext := vendorCtx.VendorContext{}
	managerConfig := managers.ManagerConfig{
		Name: "remoteagent-scheduler-manager",
		Type: "managers.symphony.remoteagentscheduler",
		Properties: map[string]string{
			"providers.persistentstate": "k8s-state",
			"singleton":                 "true",
			"RetentionDuration":         "30m",
		},
		Providers: map[string]managers.ProviderConfig{
			"k8s-state": managers.ProviderConfig{
				Type: "providers.state.k8s",
				Config: map[string]interface{}{
					"inCluster": true,
				},
			},
			"secret": managers.ProviderConfig{
				Type: "providers.secret.k8s",
				Config: map[string]interface{}{
					"inCluster": true,
				},
			},
		},
	}
	providers := map[string]providers.IProvider{
		"k8s-state": &k8sstate.K8sStateProvider{},
		"secret":    &k8ssecret.K8sSecretProvider{},
	}
	remoteAgentScheduler := RemoteTargetSchedulerManager{}
	err := remoteAgentScheduler.Init(&vendorContext, managerConfig, providers)
	assert.NotNil(t, remoteAgentScheduler)
	assert.Nil(t, err)
}

func TestRemoteAgentUpgrade(t *testing.T) {

	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	clientset := fake.NewSimpleClientset()
	os.Setenv("AGENT_VERSION", "0.0.0.1")

	cert := []byte(`-----BEGIN CERTIFICATE-----
MIIDZDCCAkygAwIBAgIRAM4UJK1WOsWMdO516YhKmc4wDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIc3ltcGhvbnkwHhcNMjQxMjMxMDU0MzQ0WhcNMjUwMzMxMDU0
MzQ0WjBPMRkwFwYDVQQKExBzeW1waG9ueS1zZXJ2aWNlMTIwMAYDVQQDEylDTj1k
ZWZhdWx0LXJlbW90ZS10YXJnZXQuc3ltcGhvbnktc2VydmljZTCCASIwDQYJKoZI
hvcNAQEBBQADggEPADCCAQoCggEBALbjcKK5KFMFl2C8A3jZv0zYbHjON6IrMCKH
WL/7R6xKEcYPlBwK5X+qMwD61G5DtGcJ0cOmyF1zszEMxT8Znsv1rvc2lgQarl+L
T8GKKEZXvWqOIriYhK0pFWF4P4F8oTOxQWqNfscpLKPxjvs7eXO1TX9jl5RFYMwH
eQtJWAq6uVDICLMG5jBkqR4FYnFdrRva/ArPOgpHw7M9t7YU/rub3Q82hYA5WIYq
syUhXTqV8ojjHbO3l5/SpKnbq3wv2O6Bi/dgVGTAB8bC/qARJcPd7MAo9hugcf9Q
Usnty2MkUagN9udUR8tj8xONeKlDgSgyueI8KZuYJVoQ7Yn7JmMCAwEAAaN3MHUw
DgYDVR0PAQH/BAQDAgWgMAwGA1UdEwEB/wQCMAAwHwYDVR0jBBgwFoAUaXDu5lg1
rcZX9Va5VhPWbMiQGyMwNAYDVR0RBC0wK4IpQ049ZGVmYXVsdC1yZW1vdGUtdGFy
Z2V0LnN5bXBob255LXNlcnZpY2UwDQYJKoZIhvcNAQELBQADggEBAHN7bVoJgsS4
06oFTy/H3wxFqjlNCZz386vMqAJxudBhwEuOeumISA33xHKI+RZhIDPe5Ck3IjiJ
yaf8vVuyWfTUakMab4oDUTYj4FnM946D96xN1uXQaYnNj4PnCXIJoZz4XPc3BWnH
399NyIWRyU+GNVGfuxXVLA0bPud/9jU3u2ER51UKxDh2TRnAtBtIIRhw617mpJ4O
PVE4yvNHUq8R9EA6jy8kCvXD1/RtPvLVmPaEwfrwpAaxflIF1dGRSeKQqYa+tUkT
IlblbTwENXKJXwwdB9Q97Ux57u3gk8ORDZXbY0QSDSkIaPgu4h635q7Yhej9maRl
OMpY47SLwvc=
-----END CERTIFICATE-----`)
	// Create a test secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tls",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"tls.crt": cert,
		},
	}

	// Create the secret using the fake client
	_, err := clientset.CoreV1().Secrets("default").Create(context.Background(), secret, metav1.CreateOptions{})
	assert.Nil(t, err)

	// Create a K8sSecretProvider
	secretProvider := &k8ssecret.K8sSecretProvider{
		Clientset: clientset,
		Config:    k8ssecret.K8sSecretProviderConfig{},
	}

	manager := targets.TargetsManager{
		StateProvider:  stateProvider,
		SecretProvider: secretProvider,
	}
	testTarget := model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &model.TargetSpec{
			Components: []model.ComponentSpec{
				{
					Type: "remote-agent",
					Name: "test",
				},
			},
		},
		Status: model.DeployableStatus{
			Properties: map[string]string{
				"targets.test.test": "{\"version\":\"0.0.0.1\", \"certificateExpiration\":\"2021-09-01T00:00:00Z\"}",
			},
		},
	}

	body := map[string]interface{}{
		"apiVersion": model.FabricGroup + "/v1",
		"kind":       "Target",
		"metadata":   testTarget.ObjectMeta,
		"spec":       testTarget.Spec,
		"status":     testTarget.Status,
	}
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   "test",
			Body: body,
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": testTarget.ObjectMeta.Namespace,
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	}

	_, err = stateProvider.Upsert(context.Background(), upsertRequest)
	assert.Nil(t, err)

	remoteAgentScheduler := RemoteTargetSchedulerManager{}
	remoteAgentScheduler.TargetsManager = &manager

	errs := remoteAgentScheduler.Poll()
	assert.Equal(t, 0, len(errs))

	target, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	component := target.Spec.Components[0]
	action := component.Properties["action"].(string)
	thumbprint := component.Properties["thumbprint"].(string)
	assert.Equal(t, "secretrotation", action)
	assert.Equal(t, "dff5df9b7bac5aa5e9a7ff5d78dd4b9ca4792ab6", thumbprint)

	testTarget.Status = model.DeployableStatus{
		Properties: map[string]string{
			"targets.test.test": "{\"version\":\"0.0.0.2\", \"certificateExpiration\":\"2100-09-01T00:00:00Z\"}",
		},
	}
	body = map[string]interface{}{
		"apiVersion": model.FabricGroup + "/v1",
		"kind":       "Target",
		"metadata":   testTarget.ObjectMeta,
		"spec":       testTarget.Spec,
		"status":     testTarget.Status,
	}
	upsertRequest = states.UpsertRequest{
		Value: states.StateEntry{
			ID:   "test",
			Body: body,
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": testTarget.ObjectMeta.Namespace,
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	}

	_, err = stateProvider.Upsert(context.Background(), upsertRequest)
	assert.Nil(t, err)

	errs = remoteAgentScheduler.Poll()
	assert.Equal(t, 0, len(errs))

	target, err = manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	component = target.Spec.Components[0]
	action = component.Properties["action"].(string)
	version := component.Properties["version"].(string)
	assert.Equal(t, "upgrade", action)
	assert.Equal(t, "0.0.0.1", version)
}
