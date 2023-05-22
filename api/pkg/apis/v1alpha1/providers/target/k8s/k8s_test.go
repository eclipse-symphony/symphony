package k8s

import (
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestK8sTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := K8sTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestK8sTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := K8sTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestMetadataToServiceNil(t *testing.T) {
	s, e := metadataToService("", "", nil)
	assert.Nil(t, e)
	assert.Nil(t, s)
}
func TestInitWithBadConfigType(t *testing.T) {
	config := K8sTargetProviderConfig{
		ConfigType: "Bad",
	}
	provider := K8sTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyFile(t *testing.T) {
	config := K8sTargetProviderConfig{
		ConfigType: "path",
	}
	provider := K8sTargetProvider{}
	provider.Init(config)
	// assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}
func TestInitWithBadFile(t *testing.T) {
	config := K8sTargetProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := K8sTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyData(t *testing.T) {
	config := K8sTargetProviderConfig{
		ConfigType: "bytes",
	}
	provider := K8sTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithBadData(t *testing.T) {
	config := K8sTargetProviderConfig{
		ConfigType: "bytes",
		ConfigData: "bad data",
	}
	provider := K8sTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestComponentToServiceFull(t *testing.T) {
	s, e := metadataToService("default", "name", map[string]string{
		"service.ports": "[{\"name\":\"port8888\",\"port\":8888},{\"name\":\"port7788\",\"port\":7788}]",
		"service.annotation.service.beta.kubernetes.io/azure-load-balancer-resource-group": "MC_EVS_evsfoakssouth_southcentralus # change to the resource group of your public IP address",
		"service.annotation.service.beta.kubernetes.io/azure-dns-label-name":               "evsfoakssouth # change to the dns name associated with your public IP address",
		"service.type":           "LoadBalancer",
		"service.loadBalancerIP": "20.189.28.227",
	})
	assert.Nil(t, e)
	assert.Equal(t, apiv1.ServiceType("LoadBalancer"), s.Spec.Type)
	assert.Equal(t, "20.189.28.227", s.Spec.LoadBalancerIP)
	assert.Equal(t, 2, len(s.ObjectMeta.Annotations))
	assert.Equal(t, "evsfoakssouth # change to the dns name associated with your public IP address", s.ObjectMeta.Annotations["service.beta.kubernetes.io/azure-dns-label-name"])
	assert.Equal(t, 2, len(s.Spec.Ports))
	assert.Equal(t, "port8888", s.Spec.Ports[0].Name)
	assert.Equal(t, "name", s.ObjectMeta.Name)
	assert.Equal(t, "default", s.ObjectMeta.Namespace)
	assert.Equal(t, "name", s.ObjectMeta.Labels["app"])
	assert.Equal(t, "name", s.Spec.Selector["app"])
}
func TestDeploymentToComponents(t *testing.T) {
	cores := resource.MustParse("1")
	deployment := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "evs",
							Image:           "evaamscontreg.azurecr.io/evsclient:latest",
							ImagePullPolicy: "Always",
							Args:            []string{"endpointLocal=http://localhost:7788/api/ImageItems", "line=https://aka.ms/linesample"},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 8888,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"cpu": cores,
								},
								Requests: apiv1.ResourceList{
									"cpu": cores,
								},
							},
						},
						{
							Name:            "rocket",
							Image:           "evaamscontreg.azurecr.io/rocket:detection",
							ImagePullPolicy: "Always",
							Args:            []string{"pipeline=3", "line=https://aka.ms/lineeast960", "cat=car person"},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 7788,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"cpu": cores,
								},
								Requests: apiv1.ResourceList{
									"cpu": cores,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "azure-rocket",
									MountPath: "/app/output",
								},
							},
						},
					},
				},
			},
		},
	}
	components, err := deploymentToComponents(deployment)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(components))
	assert.Equal(t, "evs", components[0].Name)
	assert.Equal(t, "rocket", components[1].Name)
	assert.Equal(t, "evaamscontreg.azurecr.io/evsclient:latest", components[0].Properties[model.ContainerImage])
	assert.Equal(t, "evaamscontreg.azurecr.io/rocket:detection", components[1].Properties[model.ContainerImage])
	assert.Equal(t, "Always", components[0].Properties["container.imagePullPolicy"])
	assert.Equal(t, "Always", components[1].Properties["container.imagePullPolicy"])
	assert.Equal(t, "[\"endpointLocal=http://localhost:7788/api/ImageItems\",\"line=https://aka.ms/linesample\"]", components[0].Properties["container.args"])
	assert.Equal(t, "[\"pipeline=3\",\"line=https://aka.ms/lineeast960\",\"cat=car person\"]", components[1].Properties["container.args"])
	assert.Equal(t, "[{\"containerPort\":8888}]", components[0].Properties["container.ports"])
	assert.Equal(t, "[{\"containerPort\":7788}]", components[1].Properties["container.ports"])
	assert.Equal(t, "{\"limits\":{\"cpu\":\"1\"},\"requests\":{\"cpu\":\"1\"}}", components[0].Properties["container.resources"])
	assert.Equal(t, "{\"limits\":{\"cpu\":\"1\"},\"requests\":{\"cpu\":\"1\"}}", components[1].Properties["container.resources"])
	assert.Equal(t, "[{\"name\":\"azure-rocket\",\"mountPath\":\"/app/output\"}]", components[1].Properties["container.volumeMounts"])
}
func TestComponentsToDeploymentFull(t *testing.T) {
	d, e := componentsToDeployment("default", "name", map[string]string{
		"deployment.replicas":         "#3",
		"deployment.imagePullSecrets": "[{\"name\":\"acr-evaamscontreg-secret\"}]",
		"deployment.volumes":          "[{\"name\":\"azure-evs\", \"azureFile\": {\"secretName\":\"azure-fireshare-secret\",\"shareName\":\"evs/output\",\"readOnly\":false}},{\"name\":\"azure-rocket\",\"azureFile\":{\"secretName\":\"azure-fileshare-secret\",\"shareName\":\"rocket/heavy\",\"readOnly\":false}}]",
	}, []model.ComponentSpec{
		{
			Name: "evs",
			Properties: map[string]interface{}{
				"container.image":           "evaamscontreg.azurecr.io/evsclient:latest",
				"container.ports":           "[{\"containerPort\":8888}]",
				"container.args":            "[\"endpointLocal=http://localhost:7788/api/ImageItems\", \"line=https://aka.ms/linesample\"]",
				"container.imagePullPolicy": "Always",
				"container.resources":       "{\"requests\": {\"cpu\":1}, \"limits\": {\"cpu\": 1}}",
			},
		},
		{
			Name: "rocket",
			Properties: map[string]interface{}{
				"container.image":           "evaamscontreg.azurecr.io/rocket:detection",
				"container.ports":           "[{\"containerPort\":7788}]",
				"container.args":            "[\"pipeline=3\", \"line=https://aka.ms/lineeast960\", \"cat=car person\"]",
				"container.imagePullPolicy": "Always",
				"container.resources":       "{\"requests\": {\"cpu\":1}, \"limits\": {\"cpu\": 1, \"nvidia.com/gpu\":1}}",
				"container.volumeMounts":    "[{\"name\":\"azure-rocket\",\"mountPath\":\"/app/output\"}]",
			},
		},
	}, "instance-1")
	assert.Nil(t, e)
	assert.Equal(t, "name", d.ObjectMeta.Name)
	assert.Equal(t, "name", d.Spec.Selector.MatchLabels["app"])
	assert.Equal(t, "evs", d.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "rocket", d.Spec.Template.Spec.Containers[1].Name)
	assert.Equal(t, "evaamscontreg.azurecr.io/evsclient:latest", d.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "evaamscontreg.azurecr.io/rocket:detection", d.Spec.Template.Spec.Containers[1].Image)
	assert.Equal(t, "line=https://aka.ms/linesample", d.Spec.Template.Spec.Containers[0].Args[1])
	assert.Equal(t, "cat=car person", d.Spec.Template.Spec.Containers[1].Args[2])
	assert.Equal(t, apiv1.PullPolicy("Always"), d.Spec.Template.Spec.Containers[0].ImagePullPolicy)
	cores := resource.MustParse("1")
	actualCores := d.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"]
	assert.Equal(t, cores.Value(), actualCores.Value())
	actualCores = d.Spec.Template.Spec.Containers[1].Resources.Limits["nvidia.com/gpu"]
	assert.Equal(t, cores.Value(), actualCores.Value())
	assert.Equal(t, "/app/output", d.Spec.Template.Spec.Containers[1].VolumeMounts[0].MountPath)
	assert.Equal(t, int32Ptr(3), d.Spec.Replicas)
	assert.Equal(t, "acr-evaamscontreg-secret", d.Spec.Template.Spec.ImagePullSecrets[0].Name)
}
func TestNoOpProjection(t *testing.T) {
	cores := resource.MustParse("1")
	deployment := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "evs",
							Image:           "evaamscontreg.azurecr.io/evsclient:latest",
							ImagePullPolicy: "Always",
							Args:            []string{"endpointLocal=http://localhost:7788/api/ImageItems", "line=https://aka.ms/linesample"},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 8888,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"cpu": cores,
								},
								Requests: apiv1.ResourceList{
									"cpu": cores,
								},
							},
						},
						{
							Name:            "rocket",
							Image:           "evaamscontreg.azurecr.io/rocket:detection",
							ImagePullPolicy: "Always",
							Args:            []string{"pipeline=3", "line=https://aka.ms/lineeast960", "cat=car person"},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 7788,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"cpu": cores,
								},
								Requests: apiv1.ResourceList{
									"cpu": cores,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "azure-rocket",
									MountPath: "/app/output",
								},
							},
						},
					},
				},
			},
		},
	}
	projector, err := createProjector("noop")
	assert.Nil(t, err)
	projector.ProjectDeployment("default", "name", nil, nil, &deployment)
	assert.Equal(t, "evs", deployment.Spec.Template.Spec.Containers[0].Name)
}
func TestNullProjector(t *testing.T) {
	projector, err := createProjector("")
	assert.Nil(t, err)
	assert.Nil(t, projector)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &K8sTargetProvider{}
	_ = provider.Init(K8sTargetProviderConfig{})
	// assert.Nil(t, err) okay if provider is not fully initialized
	conformance.ConformanceSuite(t, provider)
}
