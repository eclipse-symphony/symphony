package kube

import (
	"context"
	"fmt"

	"github.com/eclipse-symphony/symphony/packages/testutils/conditions"
	"github.com/eclipse-symphony/symphony/packages/testutils/conditions/jq"
	"github.com/eclipse-symphony/symphony/packages/testutils/conditions/jsonpath"
	"github.com/eclipse-symphony/symphony/packages/testutils/helpers"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	statusDescription = jq.WithDescription("Provisioning Status")
)

var (
	/**
	 * These are some common kubernetes resource conditions
	 */

	PodReadyCondition types.Condition = conditions.All(
		NewKubernetesStatusCondition("Ready", true),
		NewKubernetesStatusCondition("Initialized", true),
		NewKubernetesStatusCondition("ContainersReady", true),
	) // can be used for pods and certificates

	DeploymentCompleteCondition types.Condition = conditions.All(
		NewKubernetesStatusCondition("Available", true),
		NewKubernetesStatusCondition("Progressing", true),
	) // can be used for deployments and statefulsets

	/**
	 *  These are some common conditions for azure iot orchestration resources
	 */

	// AioManagerLabelCondition is a condition that checks if the resource is managed by the aio orc api
	AioManagerLabelCondition types.Condition = NewLabelMatchCondition("iotoperations.azure.com/managed-by", "symphony-api")

	// ProvisioningSucceededCondition is a condition that checks if the resource has succeeded provisioning
	ProvisioningSucceededCondition types.Condition = jq.Equality(".status.provisioningStatus.status", "Succeeded", statusDescription)

	// ProvisioningFailedCondition is a condition that checks if the resource has failed provisioning
	ProvisioningFailedCondition types.Condition = jq.Equality(".status.provisioningStatus.status", "Failed", statusDescription)
	// OperationIdMatchCondition is a condition that checks if the resource has the operation id  annotation and
	// ensures that it matches the operationId in the status of the resource
	OperationIdMatchCondition types.Condition = jq.MustNew(
		fmt.Sprintf(`.metadata.annotations["%s"]`, "management.azure.com/operationId"),
		jq.WithCustomMatcher(operationJqMatcher),
		jq.WithDescription("Operation Id"),
	)
)

func operationJqMatcher(ctx context.Context, value, resource interface{}, log logger.Logger) error {
	operationId, err := getProvisioningOperationIdFromStatus(resource.(map[string]interface{}))
	if err != nil {
		return err
	}
	switch value := value.(type) {
	case string:
		log("Comparing %s with %s", operationId, value)
		if operationId == value {
			return nil
		}
		return fmt.Errorf("expected %s, got %s", operationId, value)
	default:
		return fmt.Errorf("expected operationId to be string, got %T", value)
	}
}

func getProvisioningOperationIdFromStatus(resource map[string]interface{}) (string, error) {
	status, ok := resource["status"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("status field not found")
	}
	provisioningStatus, ok := status["provisioningStatus"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("provisioningStatus field not found")
	}
	operationId, ok := provisioningStatus["operationId"].(string)
	if !ok {
		return "", fmt.Errorf("operationId field not found")
	}
	return operationId, nil
}

// NewKubernetesStatusCondition returns a condition that checks the status of a kubernetes resource's condition
func NewKubernetesStatusCondition(conditionType string, status bool) types.Condition {
	statusString := "False"
	if status {
		statusString = "True"
	}
	return jsonpath.MustNew(
		fmt.Sprintf("$.status.conditions[?(@.type == '%s')].status", conditionType),
		jsonpath.WithValue([]interface{}{statusString}),
		jsonpath.WithDescription(fmt.Sprintf("Condition %s", conditionType)),
	)
}

// NewLabelMatchCondition returns a condition that checks if the resource has the label and value
func NewLabelMatchCondition(label string, value string) types.Condition {
	return jq.Equality(
		fmt.Sprintf(`.metadata.labels["%s"]`, label),
		jq.WithValue(value),
		jq.WithDescription(fmt.Sprintf("Label %s", label)),
	)
}

func ProvisioningStatusComponentOutput(componentKey string, value interface{}) types.Condition {
	return jq.Equality(fmt.Sprintf(`.status.provisioningStatus.output["%s"]`, componentKey), value)
}

// AbsentResource returns an expectation for the resources is/are absent from the cluster
func AbsentResource(name, namespace string, gvk schema.GroupVersionKind, opts ...Option) (*KubeExpectation, error) {
	opts = append(opts, IsAbsent())
	return Resource(name, namespace, gvk, opts...)
}

// NewAnnotationMatchCondition returns a condition that checks if the resource has the annotation and value
func NewAnnotationMatchCondition(annotation string, value string) types.Condition {
	return jq.MustNew(
		fmt.Sprintf(`.metadata.annotations["%s"]`, annotation),
		jq.WithValue(value),
		jq.WithDescription(fmt.Sprintf("Annotation %s", annotation)),
	)
}

// Pod returns an expectation expectation for a pod(s) in the cluster
func Pod(name, namespace string, opts ...Option) (*KubeExpectation, error) {
	return Resource(name, namespace, helpers.PodGVK, opts...)
}

// AbsentPod returns an expectation that the pod(s) is/are absent from the cluster
func AbsentPod(name, namespace string, opts ...Option) (*KubeExpectation, error) {
	return AbsentResource(name, namespace, helpers.PodGVK, opts...)
}

// Target returns an expectation for a target(s) in the cluster
func Target(name, namespace string, opts ...Option) (*KubeExpectation, error) {
	return Resource(name, namespace, helpers.TargetGVK, opts...)
}

// AbsentTarget returns an expectation that the target(s) is/are absent from the cluster
func AbsentTarget(name, namespace string, opts ...Option) (*KubeExpectation, error) {
	return AbsentResource(name, namespace, helpers.TargetGVK, opts...)
}

// Instance returns an expectation for a instance(s) in the cluster
func Instance(name, namespace string, opts ...Option) (*KubeExpectation, error) {
	return Resource(name, namespace, helpers.InstanceGVK, opts...)
}

// AbsentInstance returns an expectation that the instance(s) is/are absent from the cluster
func AbsentInstance(name, namespace string, opts ...Option) (*KubeExpectation, error) {
	return AbsentResource(name, namespace, helpers.InstanceGVK, opts...)
}

// Solution returns an expectation for a solution(s) in the cluster
func Solution(name, namespace string, opts ...Option) (*KubeExpectation, error) {
	return Resource(name, namespace, helpers.SolutionGVK, opts...)
}

// AbsentSolution returns an expectation that the solution(s) is/are absent from the cluster
func AbsentSolution(name, namespace string, opts ...Option) (*KubeExpectation, error) {
	return AbsentResource(name, namespace, helpers.SolutionGVK, opts...)
}

// Must returns a resource expectation or panics if there is an error
func Must(resource *KubeExpectation, err error) *KubeExpectation {
	if err != nil {
		panic(err)
	}
	return resource
}
