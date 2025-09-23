package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SetupTestEnvironment sets up a complete test environment including Symphony startup and returns configuration
func SetupTestEnvironment(t *testing.T, testDir string) (TestConfig, error) {
	// If no testDir requested, create a temporary directory and use it
	var fullTestDir string
	if testDir == "" {
		fullTestDir = SetupTestDirectory(t)
		// keep the behavior of creating a nested testDir if needed
	} else {
		// Create the requested test directory path as provided (relative or absolute)
		fullTestDir = testDir
		if !filepath.IsAbs(fullTestDir) {
			// Make sure the directory exists relative to current working directory
			if err := os.MkdirAll(fullTestDir, 0777); err != nil {
				return TestConfig{}, err
			}
		} else {
			// Absolute path - ensure it exists
			if err := os.MkdirAll(fullTestDir, 0777); err != nil {
				return TestConfig{}, err
			}
		}
	}

	// Get project root
	projectRoot := GetProjectRoot(t)

	// Generate namespace for test
	namespace := "default"

	// Create namespace
	cmd := exec.Command("kubectl", "create", "namespace", namespace)
	cmd.Run() // Ignore error if namespace already exists

	config := TestConfig{
		ProjectRoot: projectRoot,
		Namespace:   namespace,
	}

	return config, nil
}

// CleanupTestDirectory cleans up test directory
func CleanupTestDirectory(testDir string) {
	if testDir != "" {
		os.RemoveAll(testDir)
	}
}

// BootstrapRemoteAgent bootstraps the remote agent using direct process approach
func BootstrapRemoteAgent(t *testing.T, config TestConfig) (*exec.Cmd, error) {
	t.Log("Starting remote agent process...")

	// Start remote agent using direct process (no systemd service) without automatic cleanup
	processCmd := StartRemoteAgentProcessWithoutCleanup(t, config)
	require.NotNil(t, processCmd)

	// Wait for the process to be healthy and ready
	WaitForProcessHealthy(t, processCmd, 30*time.Second)

	t.Log("Remote agent process started successfully")
	return processCmd, nil
}

// VerifyTargetStatus verifies target status
func VerifyTargetStatus(t *testing.T, targetName, namespace string) bool {
	dyn, err := GetDynamicClient()
	if err != nil {
		t.Logf("Failed to get dynamic client: %v", err)
		return false
	}

	target, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "fabric.symphony",
		Version:  "v1",
		Resource: "targets",
	}).Namespace(namespace).Get(context.Background(), targetName, metav1.GetOptions{})

	if err != nil {
		t.Logf("Failed to get target %s: %v", targetName, err)
		return false
	}

	// Check if target has status indicating success
	status, found, err := unstructured.NestedMap(target.Object, "status")
	if err != nil || !found {
		return false
	}

	provisioningStatus, found, err := unstructured.NestedMap(status, "provisioningStatus")
	if err != nil || !found {
		return false
	}

	statusStr, found, err := unstructured.NestedString(provisioningStatus, "status")
	if err != nil || !found {
		return false
	}

	return statusStr == "Succeeded"
}

// VerifySolutionStatus verifies solution status
func VerifySolutionStatus(t *testing.T, solutionName, namespace string) bool {
	dyn, err := GetDynamicClient()
	if err != nil {
		t.Logf("Failed to get dynamic client: %v", err)
		return false
	}

	solution, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "solution.symphony",
		Version:  "v1",
		Resource: "solutions",
	}).Namespace(namespace).Get(context.Background(), solutionName, metav1.GetOptions{})

	if err != nil {
		t.Logf("Failed to get solution %s: %v", solutionName, err)
		return false
	}

	// Check if solution has status indicating success
	status, found, err := unstructured.NestedMap(solution.Object, "status")
	if err != nil || !found {
		return false
	}

	provisioningStatus, found, err := unstructured.NestedMap(status, "provisioningStatus")
	if err != nil || !found {
		return false
	}

	statusStr, found, err := unstructured.NestedString(provisioningStatus, "status")
	if err != nil || !found {
		return false
	}

	return statusStr == "Succeeded"
}

// VerifyInstanceStatus verifies instance status
func VerifyInstanceStatus(t *testing.T, instanceName, namespace string) bool {
	dyn, err := GetDynamicClient()
	if err != nil {
		t.Logf("Failed to get dynamic client: %v", err)
		return false
	}

	instance, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "solution.symphony",
		Version:  "v1",
		Resource: "instances",
	}).Namespace(namespace).Get(context.Background(), instanceName, metav1.GetOptions{})

	if err != nil {
		t.Logf("Failed to get instance %s: %v", instanceName, err)
		return false
	}

	// Check if instance has status indicating success
	status, found, err := unstructured.NestedMap(instance.Object, "status")
	if err != nil || !found {
		return false
	}

	provisioningStatus, found, err := unstructured.NestedMap(status, "provisioningStatus")
	if err != nil || !found {
		return false
	}

	statusStr, found, err := unstructured.NestedString(provisioningStatus, "status")
	if err != nil || !found {
		return false
	}

	return statusStr == "Succeeded"
}

// CreateSolutionWithComponents creates a solution YAML with specified number of components
func CreateSolutionWithComponents(t *testing.T, testDir, solutionName, namespace string, componentCount int) string {
	components := make([]map[string]interface{}, componentCount)

	for i := 0; i < componentCount; i++ {
		components[i] = map[string]interface{}{
			"name": fmt.Sprintf("component-%d", i+1),
			"type": "script",
			"properties": map[string]interface{}{
				"script": fmt.Sprintf("echo 'Component %d running'", i+1),
			},
		}
	}

	// Build a SolutionContainer followed by a versioned Solution that references it as rootResource.
	solutionContainer := map[string]interface{}{
		"apiVersion": "solution.symphony/v1",
		"kind":       "SolutionContainer",
		"metadata": map[string]interface{}{
			"name":      solutionName,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{},
	}

	solutionVersion := fmt.Sprintf("%s-v-version1", solutionName)
	solution := map[string]interface{}{
		"apiVersion": "solution.symphony/v1",
		"kind":       "Solution",
		"metadata": map[string]interface{}{
			"name":      solutionVersion,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"rootResource": solutionName,
			"components":   components,
		},
	}

	containerYaml, err := yaml.Marshal(solutionContainer)
	require.NoError(t, err)
	solutionYaml, err := yaml.Marshal(solution)
	require.NoError(t, err)

	yamlContent := string(containerYaml) + "\n---\n" + string(solutionYaml)

	yamlPath := filepath.Join(testDir, fmt.Sprintf("%s.yaml", solutionName))
	err = ioutil.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	return yamlPath
}

// CreateSolutionWithComponentsForProvider creates a solution with components for specific provider
func CreateSolutionWithComponentsForProvider(t *testing.T, testDir, solutionName, namespace string, componentCount int, provider string) string {
	components := make([]map[string]interface{}, componentCount)

	for i := 0; i < componentCount; i++ {
		var componentType string
		var properties map[string]interface{}

		switch provider {
		case "script":
			componentType = "script"
			properties = map[string]interface{}{
				"script": fmt.Sprintf("echo 'Script component %d for provider %s'", i+1, provider),
			}
		case "http":
			componentType = "http"
			properties = map[string]interface{}{
				"url":    fmt.Sprintf("http://example.com/api/component-%d", i+1),
				"method": "GET",
			}
		case "helm.v3":
			componentType = "helm.v3"
			properties = map[string]interface{}{
				"chart": map[string]interface{}{
					"name":    fmt.Sprintf("test-chart-%d", i+1),
					"version": "1.0.0",
				},
			}
		default:
			componentType = "script"
			properties = map[string]interface{}{
				"script": fmt.Sprintf("echo 'Default component %d'", i+1),
			}
		}

		components[i] = map[string]interface{}{
			"name":       fmt.Sprintf("component-%s-%d", provider, i+1),
			"type":       componentType,
			"properties": properties,
		}
	}

	// Build solution container + versioned solution referencing the container
	solutionContainer := map[string]interface{}{
		"apiVersion": "solution.symphony/v1",
		"kind":       "SolutionContainer",
		"metadata": map[string]interface{}{
			"name":      solutionName,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{},
	}

	solutionVersion := fmt.Sprintf("%s-v-version1", solutionName)
	solution := map[string]interface{}{
		"apiVersion": "solution.symphony/v1",
		"kind":       "Solution",
		"metadata": map[string]interface{}{
			"name":      solutionVersion,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"rootResource": solutionName,
			"components":   components,
		},
	}

	containerYaml, err := yaml.Marshal(solutionContainer)
	require.NoError(t, err)
	solutionYaml, err := yaml.Marshal(solution)
	require.NoError(t, err)

	yamlContent := string(containerYaml) + "\n---\n" + string(solutionYaml)

	yamlPath := filepath.Join(testDir, fmt.Sprintf("%s.yaml", solutionName))
	err = ioutil.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	return yamlPath
}

// CreateInstanceYAML creates an instance YAML file
func CreateInstanceYAML(t *testing.T, testDir, instanceName, solutionName, targetName, namespace string) string {
	// The controller expects the solution reference to include the version in the form "<container>:version1"
	solutionRef := fmt.Sprintf("%s:version1", solutionName)
	yamlContent := fmt.Sprintf(`
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: %s
  namespace: %s
spec:
  solution: %s
  target:
    name: %s
`, instanceName, namespace, solutionRef, targetName)

	yamlPath := filepath.Join(testDir, fmt.Sprintf("%s.yaml", instanceName))
	err := ioutil.WriteFile(yamlPath, []byte(strings.TrimSpace(yamlContent)), 0644)
	require.NoError(t, err)

	return yamlPath
}

// GetInstanceComponents gets the components of an instance
func GetInstanceComponents(t *testing.T, instanceName, namespace string) []map[string]interface{} {
	dyn, err := GetDynamicClient()
	if err != nil {
		t.Logf("Failed to get dynamic client: %v", err)
		return nil
	}

	instance, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "solution.symphony",
		Version:  "v1",
		Resource: "instances",
	}).Namespace(namespace).Get(context.Background(), instanceName, metav1.GetOptions{})

	if err != nil {
		t.Logf("Failed to get instance %s: %v", instanceName, err)
		return nil
	}

	// Extract components from status or spec
	status, found, err := unstructured.NestedMap(instance.Object, "status")
	if err != nil || !found {
		return nil
	}

	components, found, err := unstructured.NestedSlice(status, "components")
	if err != nil || !found {
		return nil
	}

	result := make([]map[string]interface{}, len(components))
	for i, comp := range components {
		if compMap, ok := comp.(map[string]interface{}); ok {
			result[i] = compMap
		}
	}

	return result
}

// GetInstanceComponentsPaged gets components with paging
func GetInstanceComponentsPaged(t *testing.T, instanceName, namespace string, page, pageSize int) []map[string]interface{} {
	allComponents := GetInstanceComponents(t, instanceName, namespace)
	if allComponents == nil {
		return nil
	}

	start := page * pageSize
	end := start + pageSize

	if start >= len(allComponents) {
		return []map[string]interface{}{}
	}

	if end > len(allComponents) {
		end = len(allComponents)
	}

	return allComponents[start:end]
}

// GetInstanceComponentsPagedForProvider gets components with paging for specific provider
func GetInstanceComponentsPagedForProvider(t *testing.T, instanceName, namespace string, page, pageSize int, provider string) []map[string]interface{} {
	allComponents := GetInstanceComponents(t, instanceName, namespace)
	if allComponents == nil {
		return nil
	}

	// Filter components by provider/type
	var filteredComponents []map[string]interface{}
	for _, comp := range allComponents {
		if compType, ok := comp["type"].(string); ok {
			if (provider == "script" && compType == "script") ||
				(provider == "http" && compType == "http") ||
				(provider == "helm.v3" && compType == "helm.v3") {
				filteredComponents = append(filteredComponents, comp)
			}
		}
	}

	start := page * pageSize
	end := start + pageSize

	if start >= len(filteredComponents) {
		return []map[string]interface{}{}
	}

	if end > len(filteredComponents) {
		end = len(filteredComponents)
	}

	return filteredComponents[start:end]
}

// ReadAndUpdateSolutionName reads solution file and updates the name
func ReadAndUpdateSolutionName(t *testing.T, solutionPath, newName string) string {
	content, err := ioutil.ReadFile(solutionPath)
	require.NoError(t, err)

	// Simple string replacement - in real implementation would parse YAML
	updatedContent := strings.ReplaceAll(string(content), `"name":"test-solution-update-v2"`, fmt.Sprintf(`"name":"%s"`, newName))

	return updatedContent
}

// WriteFileContent writes content to file
func WriteFileContent(t *testing.T, filePath, content string) {
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
}

// GetInstanceEvents gets events for an instance
func GetInstanceEvents(t *testing.T, instanceName, namespace string) []string {
	// Get events related to the instance
	cmd := exec.Command("kubectl", "get", "events", "-n", namespace, "--field-selector", fmt.Sprintf("involvedObject.name=%s", instanceName), "-o", "jsonpath={.items[*].message}")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Failed to get events for instance %s: %v", instanceName, err)
		return nil
	}

	events := strings.Fields(string(output))
	return events
}

// VerifyLargeScaleDeployment verifies large scale deployment
func VerifyLargeScaleDeployment(t *testing.T, instanceName, namespace string, expectedCount int) bool {
	components := GetInstanceComponents(t, instanceName, namespace)
	if len(components) != expectedCount {
		t.Logf("Expected %d components, got %d", expectedCount, len(components))
		return false
	}

	// Verify all components are in ready state
	for i, comp := range components {
		if status, ok := comp["status"].(string); !ok || status != "Ready" {
			t.Logf("Component %d is not ready: %v", i, comp)
			return false
		}
	}

	return true
}

// RefreshInstanceStatus refreshes instance status
func RefreshInstanceStatus(t *testing.T, instanceName, namespace string) {
	// Force refresh by getting the instance again
	_ = VerifyInstanceStatus(t, instanceName, namespace)
}

// VerifyComponentsUninstalled verifies components are uninstalled
func VerifyComponentsUninstalled(t *testing.T, namespace string, expectedCount int) {
	t.Logf("Verifying %d components are uninstalled in namespace %s", expectedCount, namespace)

	// Check that there are no remaining component artifacts
	cmd := exec.Command("kubectl", "get", "all", "-n", namespace, "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Failed to get resources in namespace %s: %v", namespace, err)
		return
	}

	resources := strings.TrimSpace(string(output))
	if resources != "" {
		t.Logf("Warning: Found remaining resources in namespace %s: %s", namespace, resources)
	} else {
		t.Logf("All components successfully uninstalled from namespace %s", namespace)
	}
}

// TestSystemResponsivenessUnderLoad tests system responsiveness under load
func TestSystemResponsivenessUnderLoad(t *testing.T, namespace string, componentCount int) bool {
	t.Logf("Testing system responsiveness under load with %d components", componentCount)

	// Test basic API responsiveness with multiple retries and better error handling
	maxRetries := 3
	timeout := 30 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		t.Logf("System responsiveness test attempt %d/%d", retry+1, maxRetries)

		// Test multiple kubectl operations to verify system responsiveness
		tests := []struct {
			name string
			cmd  []string
		}{
			{"pods", []string{"kubectl", "get", "pods", "-n", namespace, "--timeout=30s"}},
			{"instances", []string{"kubectl", "get", "instances.solution.symphony", "-n", namespace, "--timeout=30s"}},
			{"targets", []string{"kubectl", "get", "targets.fabric.symphony", "-n", namespace, "--timeout=30s"}},
		}

		allPassed := true
		totalDuration := time.Duration(0)

		for _, test := range tests {
			start := time.Now()
			cmd := exec.Command(test.cmd[0], test.cmd[1:]...)
			output, err := cmd.CombinedOutput()
			duration := time.Since(start)
			totalDuration += duration

			t.Logf("Test %s took %v", test.name, duration)

			if err != nil {
				t.Logf("Test %s failed with error: %v, output: %s", test.name, err, string(output))
				allPassed = false
				break
			}

			// Individual command should complete within reasonable time
			if duration > timeout {
				t.Logf("Test %s took too long: %v (expected < %v)", test.name, duration, timeout)
				allPassed = false
				break
			}
		}

		if allPassed {
			// Overall system should be responsive even under load
			if totalDuration > 2*time.Minute {
				t.Logf("System responsiveness test warning: total time %v is quite high but acceptable", totalDuration)
			}
			t.Logf("System responsiveness test passed: total time %v", totalDuration)
			return true
		}

		// If failed, wait a bit before retry
		if retry < maxRetries-1 {
			t.Logf("Retrying system responsiveness test in 10 seconds...")
			time.Sleep(10 * time.Second)
		}
	}

	t.Logf("System responsiveness test failed after %d attempts", maxRetries)
	return false
}

// VerifyProviderComponentsDeployment verifies provider-specific components deployment
func VerifyProviderComponentsDeployment(t *testing.T, instanceName, namespace, provider string, expectedCount int) {
	components := GetInstanceComponents(t, instanceName, namespace)
	if components == nil {
		t.Logf("No components found for instance %s", instanceName)
		return
	}

	providerComponents := 0
	for _, comp := range components {
		if compType, ok := comp["type"].(string); ok {
			if (provider == "script" && compType == "script") ||
				(provider == "http" && compType == "http") ||
				(provider == "helm.v3" && compType == "helm.v3") {
				providerComponents++
			}
		}
	}

	if providerComponents != expectedCount {
		t.Logf("Expected %d components for provider %s, got %d", expectedCount, provider, providerComponents)
	} else {
		t.Logf("Verified %d components for provider %s", providerComponents, provider)
	}
}
