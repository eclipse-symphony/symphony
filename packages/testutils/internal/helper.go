package internal

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Pod(name, namespace string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"management.azure.com/operationId": "test-operation-id",
				"test-annotation":                  "test-annotation-value",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  name,
					Image: fmt.Sprintf("image/%s", name),
				},
			},
		},
	}
}

func Resource(name, namespace string, gvk schema.GroupVersionKind) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"annotations": map[string]interface{}{
					"management.azure.com/operationId": "test-operation-id",
					"test-annotation":                  "test-annotation-value",
				},
			},
			"apiVersion": gvk.GroupVersion().String(),
			"kind":       gvk.Kind,
		},
	}
}

func Target(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"annotations": map[string]interface{}{
					"management.azure.com/operationId": "test-operation-id",
					"test-annotation":                  "test-annotation-value",
				},
			},
			"apiVersion": "fabric.symphony/v1",
			"kind":       "Target",
			"status": map[string]interface{}{
				"provisioningStatus": map[string]interface{}{
					"status":      "Succeeded",
					"operationId": "test-operation-id",
				},
			},
		},
	}
}

func Namespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
	}
}

func OutOfSyncResource(name, namespace string, gvk schema.GroupVersionKind) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"annotations": map[string]interface{}{
					"management.azure.com/operationId": "test-operation-new",
					"test-annotation":                  "test-annotation-value",
				},
			},
			"apiVersion": gvk.GroupVersion().String(),
			"kind":       gvk.Kind,
			"status": map[string]interface{}{
				"provisioningStatus": map[string]interface{}{
					"status":      "Succeeded",
					"operationId": "test-operation-id-old", // this is the old operation id
				},
			},
		},
	}
}

func GenerateTestApiResourceList() []*metav1.APIResourceList {
	// we want to test the following resources:
	// - core/v1: Pods, ConfigMaps, Namespaces
	// - orchestrator.iotoperations.azure.com/v1: Targets

	return []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{
					Name:         "pods",
					Namespaced:   true,
					Kind:         "Pod",
					SingularName: "pod",
					Group:        "",
					Version:      "v1",
				},
				{
					Name:         "configmaps",
					Namespaced:   true,
					Kind:         "ConfigMap",
					SingularName: "configmap",
					Group:        "",
					Version:      "v1",
				},
				{
					Name:         "namespaces",
					Namespaced:   false,
					Kind:         "Namespace",
					SingularName: "namespace",
					Group:        "",
					Version:      "v1",
				},
			},
		},
		{
			GroupVersion: "fabric.symphony/v1",
			APIResources: []metav1.APIResource{
				{
					Name:         "targets",
					Namespaced:   true,
					Kind:         "Target",
					SingularName: "target",
					Group:        "fabric.symphony",
					Version:      "v1",
				},
			},
		},
	}
}
