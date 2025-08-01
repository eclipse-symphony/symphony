package utils

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// TestConfig holds configuration for test setup
type TestConfig struct {
	ProjectRoot    string
	ConfigPath     string
	ClientCertPath string
	ClientKeyPath  string
	CACertPath     string
	TargetName     string
	Namespace      string
	TopologyPath   string
	Protocol       string
	BaseURL        string
	BinaryPath     string
}

// SetupTestDirectory creates a temporary directory for test files with proper permissions
func SetupTestDirectory(t *testing.T) string {
	testDir, err := ioutil.TempDir("", "symphony-e2e-test-")
	require.NoError(t, err)

	// Set full permissions for the test directory to avoid permission issues
	err = os.Chmod(testDir, 0777)
	require.NoError(t, err)

	t.Logf("Created test directory with full permissions (0777): %s", testDir)
	return testDir
}

// GetProjectRoot returns the project root directory by walking up from current working directory
func GetProjectRoot(t *testing.T) string {
	// Start from the current working directory (where the test is running)
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	t.Logf("GetProjectRoot: Starting from directory: %s", currentDir)

	// Keep going up directories until we find the project root
	for {
		t.Logf("GetProjectRoot: Checking directory: %s", currentDir)

		// Check if this directory contains the expected project structure
		expectedDirs := []string{"api", "coa", "remote-agent", "test"}
		isProjectRoot := true

		for _, dir := range expectedDirs {
			fullPath := filepath.Join(currentDir, dir)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Logf("GetProjectRoot: Directory %s not found at %s", dir, fullPath)
				isProjectRoot = false
				break
			} else {
				t.Logf("GetProjectRoot: Found directory %s at %s", dir, fullPath)
			}
		}

		if isProjectRoot {
			t.Logf("Project root detected: %s", currentDir)
			return currentDir
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)

		// Check if we've reached the filesystem root
		if parentDir == currentDir {
			t.Fatalf("Could not find Symphony project root. Started from: %s", func() string {
				wd, _ := os.Getwd()
				return wd
			}())
		}

		currentDir = parentDir
	}
}

// CreateHTTPConfig creates HTTP configuration file for remote agent
func CreateHTTPConfig(t *testing.T, testDir, baseURL string) string {
	config := map[string]interface{}{
		"requestEndpoint":  fmt.Sprintf("%s/solution/tasks", baseURL),
		"responseEndpoint": fmt.Sprintf("%s/solution/task/getResult", baseURL),
		"baseUrl":          baseURL,
	}

	configBytes, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)

	configPath := filepath.Join(testDir, "config-http.json")
	err = ioutil.WriteFile(configPath, configBytes, 0644)
	require.NoError(t, err)

	return configPath
}

// CreateMQTTConfig creates MQTT configuration file for remote agent
func CreateMQTTConfig(t *testing.T, testDir, brokerAddress string, brokerPort int, targetName, namespace string) string {
	// Ensure directory has proper permissions first
	err := os.Chmod(testDir, 0777)
	if err != nil {
		t.Logf("Warning: Failed to ensure directory permissions: %v", err)
	}

	config := map[string]interface{}{
		"mqttBroker": brokerAddress,
		"mqttPort":   brokerPort,
		"targetName": targetName,
		"namespace":  namespace,
	}

	configBytes, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err, "Failed to marshal MQTT config to JSON")

	configPath := filepath.Join(testDir, "config-mqtt.json")
	t.Logf("Creating MQTT config file at: %s", configPath)
	t.Logf("Config content: %s", string(configBytes))

	err = ioutil.WriteFile(configPath, configBytes, 0666)
	if err != nil {
		t.Logf("Failed to write MQTT config file: %v", err)
		t.Logf("Target directory: %s", testDir)
		if info, statErr := os.Stat(testDir); statErr == nil {
			t.Logf("Directory permissions: %v", info.Mode())
		} else {
			t.Logf("Failed to get directory permissions: %v", statErr)
		}
	}
	require.NoError(t, err, "Failed to write MQTT config file")

	t.Logf("Successfully created MQTT config file: %s", configPath)
	return configPath
}

// CreateTestTopology creates a test topology file
func CreateTestTopology(t *testing.T, testDir string) string {
	// Ensure directory has proper permissions first
	err := os.Chmod(testDir, 0777)
	if err != nil {
		t.Logf("Warning: Failed to ensure directory permissions: %v", err)
	}

	topology := map[string]interface{}{
		"bindings": []map[string]interface{}{
			{
				"provider": "providers.target.script",
				"role":     "script",
			},
			{
				"provider": "providers.target.remote-agent",
				"role":     "remote-agent",
			},
			{
				"provider": "providers.target.http",
				"role":     "http",
			},
			{
				"provider": "providers.target.docker",
				"role":     "docker",
			},
		},
	}

	t.Logf("Creating test topology with bindings: %+v", topology)
	topologyBytes, err := json.MarshalIndent(topology, "", "  ")
	require.NoError(t, err, "Failed to marshal topology to JSON")

	topologyPath := filepath.Join(testDir, "topology.json")
	t.Logf("Creating topology file at: %s", topologyPath)
	t.Logf("Topology content: %s", string(topologyBytes))

	err = ioutil.WriteFile(topologyPath, topologyBytes, 0666)
	if err != nil {
		t.Logf("Failed to write topology file: %v", err)
		t.Logf("Target directory: %s", testDir)
		if info, statErr := os.Stat(testDir); statErr == nil {
			t.Logf("Directory permissions: %v", info.Mode())
		} else {
			t.Logf("Failed to get directory permissions: %v", statErr)
		}
	}
	require.NoError(t, err, "Failed to write topology file")

	t.Logf("Successfully created topology file: %s", topologyPath)
	return topologyPath
}

// CreateTargetYAML creates a Target resource YAML file
func CreateTargetYAML(t *testing.T, testDir, targetName, namespace string) string {
	yamlContent := fmt.Sprintf(`
apiVersion: fabric.symphony/v1
kind: Target
metadata:
  name: %s
  namespace: %s
spec:
  displayName: %s
  components:
  - name: remote-agent
    type: remote-agent
    properties:
      description: E2E test remote agent
  topologies:
  - bindings:
    - provider: providers.target.script
      role: script
    - provider: providers.target.remote-agent
      role: remote-agent
`, targetName, namespace, targetName)

	yamlPath := filepath.Join(testDir, "target.yaml")
	err := ioutil.WriteFile(yamlPath, []byte(strings.TrimSpace(yamlContent)), 0644)
	require.NoError(t, err)

	return yamlPath
}

// ApplyKubernetesManifest applies a YAML manifest to the cluster
func ApplyKubernetesManifest(t *testing.T, manifestPath string) error {
	cmd := exec.Command("kubectl", "apply", "-f", manifestPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("kubectl apply failed: %s", string(output))
		return err
	}

	t.Logf("Applied manifest: %s", manifestPath)
	return nil
}

// DeleteKubernetesManifest deletes a YAML manifest from the cluster
func DeleteKubernetesManifest(t *testing.T, manifestPath string) error {
	cmd := exec.Command("kubectl", "delete", "-f", manifestPath, "--ignore-not-found=true")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("kubectl delete failed: %s", string(output))
		return err
	}

	t.Logf("Deleted manifest: %s", manifestPath)
	return nil
}

// DeleteKubernetesManifestWithTimeout deletes a YAML manifest with timeout and wait
func DeleteKubernetesManifestWithTimeout(t *testing.T, manifestPath string, timeout time.Duration) error {
	t.Logf("Deleting manifest with timeout %v: %s", timeout, manifestPath)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// First try normal delete
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "-f", manifestPath, "--ignore-not-found=true", "--wait=true", "--timeout=60s")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("kubectl delete failed: %s", string(output))
		// If normal delete fails, try force delete
		t.Logf("Attempting force delete for: %s", manifestPath)
		forceCmd := exec.CommandContext(ctx, "kubectl", "delete", "-f", manifestPath, "--ignore-not-found=true", "--force", "--grace-period=0")
		forceOutput, forceErr := forceCmd.CombinedOutput()
		if forceErr != nil {
			t.Logf("Force delete also failed: %s", string(forceOutput))
			return forceErr
		}
		t.Logf("Force deleted manifest: %s", manifestPath)
		return nil
	}

	t.Logf("Successfully deleted manifest: %s", manifestPath)
	return nil
}

// DeleteSolutionManifestWithTimeout deletes a solution manifest that may contain both Solution and SolutionContainer
// It handles the deletion order required by admission webhooks: Solution -> SolutionContainer
// Following the pattern from CleanUpSymphonyObjects function
func DeleteSolutionManifestWithTimeout(t *testing.T, manifestPath string, timeout time.Duration) error {
	t.Logf("Deleting solution manifest with timeout %v: %s", timeout, manifestPath)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Read the manifest file to check if it contains both Solution and SolutionContainer
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Logf("Failed to read manifest file: %v", err)
		return err
	}

	contentStr := string(content)
	hasSolution := strings.Contains(contentStr, "kind: Solution")
	hasSolutionContainer := strings.Contains(contentStr, "kind: SolutionContainer")

	if hasSolution && hasSolutionContainer {
		// Extract namespace and solution name for targeted deletion
		lines := strings.Split(contentStr, "\n")
		var namespace, solutionName, solutionContainerName string

		inSolution := false
		inSolutionContainer := false

		for _, line := range lines {
			line = strings.TrimSpace(line)

			if line == "kind: Solution" {
				inSolution = true
				inSolutionContainer = false
				continue
			}
			if line == "kind: SolutionContainer" {
				inSolutionContainer = true
				inSolution = false
				continue
			}

			if strings.HasPrefix(line, "name:") && (inSolution || inSolutionContainer) {
				name := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				if inSolution {
					solutionName = name
				} else if inSolutionContainer {
					solutionContainerName = name
				}
			}

			if strings.HasPrefix(line, "namespace:") && (inSolution || inSolutionContainer) {
				namespace = strings.TrimSpace(strings.TrimPrefix(line, "namespace:"))
			}
		}

		// Delete Solution first (using the same pattern as CleanUpSymphonyObjects)
		if solutionName != "" {
			t.Logf("Deleting Solution: %s in namespace: %s", solutionName, namespace)
			var solutionCmd *exec.Cmd
			if namespace != "" {
				solutionCmd = exec.CommandContext(ctx, "kubectl", "delete", "solutions.solution.symphony", solutionName, "-n", namespace, "--ignore-not-found=true", "--timeout=60s")
			} else {
				solutionCmd = exec.CommandContext(ctx, "kubectl", "delete", "solutions.solution.symphony", solutionName, "--ignore-not-found=true", "--timeout=60s")
			}

			solutionOutput, solutionErr := solutionCmd.CombinedOutput()
			if solutionErr != nil {
				t.Logf("Failed to delete Solution: %s", string(solutionOutput))
				// Don't return error immediately, try to delete SolutionContainer anyway
			} else {
				t.Logf("Successfully deleted Solution: %s", solutionName)
			}
		}

		// Then delete SolutionContainer
		if solutionContainerName != "" {
			t.Logf("Deleting SolutionContainer: %s in namespace: %s", solutionContainerName, namespace)
			var containerCmd *exec.Cmd
			if namespace != "" {
				containerCmd = exec.CommandContext(ctx, "kubectl", "delete", "solutioncontainers.solution.symphony", solutionContainerName, "-n", namespace, "--ignore-not-found=true", "--timeout=60s")
			} else {
				containerCmd = exec.CommandContext(ctx, "kubectl", "delete", "solutioncontainers.solution.symphony", solutionContainerName, "--ignore-not-found=true", "--timeout=60s")
			}

			containerOutput, containerErr := containerCmd.CombinedOutput()
			if containerErr != nil {
				t.Logf("Failed to delete SolutionContainer: %s", string(containerOutput))
				return containerErr
			} else {
				t.Logf("Successfully deleted SolutionContainer: %s", solutionContainerName)
			}
		}

		t.Logf("Successfully deleted solution manifest: %s", manifestPath)
		return nil
	} else {
		// Fallback to normal deletion if it's not a combined manifest
		return DeleteKubernetesManifestWithTimeout(t, manifestPath, timeout)
	}
}

// DeleteKubernetesResource deletes a single Kubernetes resource by type and name
// Following the pattern from CleanUpSymphonyObjects function
func DeleteKubernetesResource(t *testing.T, resourceType, resourceName, namespace string, timeout time.Duration) error {
	t.Logf("Deleting %s: %s in namespace: %s", resourceType, resourceName, namespace)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if namespace != "" {
		cmd = exec.CommandContext(ctx, "kubectl", "delete", resourceType, resourceName, "-n", namespace, "--ignore-not-found=true", "--timeout=60s")
	} else {
		cmd = exec.CommandContext(ctx, "kubectl", "delete", resourceType, resourceName, "--ignore-not-found=true", "--timeout=60s")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to delete %s %s: %s", resourceType, resourceName, string(output))
		return err
	} else {
		t.Logf("Successfully deleted %s: %s", resourceType, resourceName)
		return nil
	}
}

// WaitForResourceDeleted waits for a specific resource to be completely deleted
func WaitForResourceDeleted(t *testing.T, resourceType, resourceName, namespace string, timeout time.Duration) {
	t.Logf("Waiting for %s %s/%s to be deleted...", resourceType, namespace, resourceName)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Logf("Timeout waiting for %s %s/%s to be deleted", resourceType, namespace, resourceName)
			return // Don't fail the test, just log and continue
		case <-ticker.C:
			cmd := exec.Command("kubectl", "get", resourceType, resourceName, "-n", namespace)
			err := cmd.Run()
			if err != nil {
				// Resource not found, it's been deleted
				t.Logf("%s %s/%s has been deleted", resourceType, namespace, resourceName)
				return
			}
			t.Logf("Still waiting for %s %s/%s to be deleted...", resourceType, namespace, resourceName)
		}
	}
}

// GetRestConfig gets Kubernetes REST config
func GetRestConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// GetKubeClient gets Kubernetes clientset
func GetKubeClient() (kubernetes.Interface, error) {
	config, err := GetRestConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// GetDynamicClient gets Kubernetes dynamic client
func GetDynamicClient() (dynamic.Interface, error) {
	config, err := GetRestConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(config)
}

// WaitForTargetCreated waits for a Target resource to be created
func WaitForTargetCreated(t *testing.T, targetName, namespace string, timeout time.Duration) {
	dyn, err := GetDynamicClient()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for Target %s/%s to be created", namespace, targetName)
		case <-ticker.C:
			targets, err := dyn.Resource(schema.GroupVersionResource{
				Group:    "fabric.symphony",
				Version:  "v1",
				Resource: "targets",
			}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})

			if err == nil && len(targets.Items) > 0 {
				for _, item := range targets.Items {
					if item.GetName() == targetName {
						t.Logf("Target %s/%s created successfully", namespace, targetName)
						return
					}
				}
			}
		}
	}
}

// WaitForTargetReady waits for a Target to reach ready state
func WaitForTargetReady(t *testing.T, targetName, namespace string, timeout time.Duration) {
	dyn, err := GetDynamicClient()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Check immediately first
	target, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "fabric.symphony",
		Version:  "v1",
		Resource: "targets",
	}).Namespace(namespace).Get(context.Background(), targetName, metav1.GetOptions{})

	if err == nil {
		status, found, err := unstructured.NestedMap(target.Object, "status")
		if err == nil && found {
			provisioningStatus, found, err := unstructured.NestedMap(status, "provisioningStatus")
			if err == nil && found {
				statusStr, found, err := unstructured.NestedString(provisioningStatus, "status")
				if err == nil && found {
					t.Logf("Target %s/%s current status: %s", namespace, targetName, statusStr)
					if statusStr == "Succeeded" {
						t.Logf("Target %s/%s is already ready", namespace, targetName)
						return
					}
					if statusStr == "Failed" {
						t.Fatalf("Target %s/%s failed to deploy", namespace, targetName)
					}
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for Target %s/%s to be ready", namespace, targetName)
		case <-ticker.C:
			target, err := dyn.Resource(schema.GroupVersionResource{
				Group:    "fabric.symphony",
				Version:  "v1",
				Resource: "targets",
			}).Namespace(namespace).Get(context.Background(), targetName, metav1.GetOptions{})

			if err == nil {
				status, found, err := unstructured.NestedMap(target.Object, "status")
				if err == nil && found {
					provisioningStatus, found, err := unstructured.NestedMap(status, "provisioningStatus")
					if err == nil && found {
						statusStr, found, err := unstructured.NestedString(provisioningStatus, "status")
						if err == nil && found {
							t.Logf("Target %s/%s status: %s", namespace, targetName, statusStr)
							if statusStr == "Succeeded" {
								t.Logf("Target %s/%s is ready", namespace, targetName)
								return
							}
							if statusStr == "Failed" {
								t.Fatalf("Target %s/%s failed to deploy", namespace, targetName)
							}
						} else {
							t.Logf("Target %s/%s: provisioningStatus.status not found", namespace, targetName)
						}
					} else {
						t.Logf("Target %s/%s: provisioningStatus not found", namespace, targetName)
					}
				} else {
					t.Logf("Target %s/%s: status not found", namespace, targetName)
				}
			} else {
				t.Logf("Error getting Target %s/%s: %v", namespace, targetName, err)
			}
		}
	}
}

// WaitForInstanceReady waits for an Instance to reach ready state
func WaitForInstanceReady(t *testing.T, instanceName, namespace string, timeout time.Duration) {
	dyn, err := GetDynamicClient()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	t.Logf("Waiting for Instance %s/%s to be ready...", namespace, instanceName)

	for {
		select {
		case <-ctx.Done():
			t.Logf("Timeout waiting for Instance %s/%s to be ready", namespace, instanceName)
			// Don't fail the test, just continue - Instance deployment might take long
			return
		case <-ticker.C:
			instance, err := dyn.Resource(schema.GroupVersionResource{
				Group:    "solution.symphony",
				Version:  "v1",
				Resource: "instances",
			}).Namespace(namespace).Get(context.Background(), instanceName, metav1.GetOptions{})

			if err != nil {
				t.Logf("Error getting Instance %s/%s: %v", namespace, instanceName, err)
				continue
			}

			status, found, err := unstructured.NestedMap(instance.Object, "status")
			if err != nil || !found {
				t.Logf("Instance %s/%s: status not found", namespace, instanceName)
				continue
			}

			provisioningStatus, found, err := unstructured.NestedMap(status, "provisioningStatus")
			if err != nil || !found {
				t.Logf("Instance %s/%s: provisioningStatus not found", namespace, instanceName)
				continue
			}

			statusStr, found, err := unstructured.NestedString(provisioningStatus, "status")
			if err != nil || !found {
				t.Logf("Instance %s/%s: provisioningStatus.status not found", namespace, instanceName)
				continue
			}

			t.Logf("Instance %s/%s status: %s", namespace, instanceName, statusStr)

			if statusStr == "Succeeded" {
				t.Logf("Instance %s/%s is ready and deployed successfully", namespace, instanceName)
				return
			}
			if statusStr == "Failed" {
				t.Logf("Instance %s/%s failed to deploy, but continuing test", namespace, instanceName)
				return
			}

			// Check if there's deployment activity
			deployed, found, err := unstructured.NestedInt64(status, "deployed")
			if err == nil && found && deployed > 0 {
				t.Logf("Instance %s/%s has some deployments (%d), considering it ready", namespace, instanceName, deployed)
				return
			}

			t.Logf("Instance %s/%s still deploying, waiting...", namespace, instanceName)
		}
	}
}

// streamProcessLogs streams logs from a process reader to test output in real-time
func streamProcessLogs(t *testing.T, reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		t.Logf("[%s] %s", prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Logf("[%s] Error reading logs: %v", prefix, err)
	}
}

// BuildRemoteAgentBinary builds the remote agent binary
func BuildRemoteAgentBinary(t *testing.T, config TestConfig) string {
	binaryPath := filepath.Join(config.ProjectRoot, "remote-agent", "bootstrap", "remote-agent")

	t.Logf("Building remote agent binary at: %s", binaryPath)

	// Build the binary: GOOS=linux GOARCH=amd64 go build -o bootstrap/remote-agent
	buildCmd := exec.Command("go", "build", "-o", "bootstrap/remote-agent", ".")
	buildCmd.Dir = filepath.Join(config.ProjectRoot, "remote-agent")
	buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")

	var stdout, stderr bytes.Buffer
	buildCmd.Stdout = &stdout
	buildCmd.Stderr = &stderr

	err := buildCmd.Run()
	if err != nil {
		t.Logf("Build stdout: %s", stdout.String())
		t.Logf("Build stderr: %s", stderr.String())
	}
	require.NoError(t, err, "Failed to build remote agent binary")

	t.Logf("Successfully built remote agent binary")
	return binaryPath
}

// GetWorkingCertificates calls the getcert endpoint with bootstrap cert to obtain working certificates
func GetWorkingCertificates(t *testing.T, baseURL, targetName, namespace string, bootstrapCertPath, bootstrapKeyPath string, testDir string) (string, string) {
	t.Logf("Getting working certificates using bootstrap cert...")
	getCertEndpoint := fmt.Sprintf("%s/targets/bootstrap/%s?namespace=%s&osPlatform=linux", baseURL, targetName, namespace)
	t.Logf("Calling certificate endpoint: %s", getCertEndpoint)

	// Load bootstrap certificate
	cert, err := tls.LoadX509KeyPair(bootstrapCertPath, bootstrapKeyPath)
	require.NoError(t, err, "Failed to load bootstrap cert/key")

	// Create HTTP client with bootstrap certificate
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true, // Skip server cert verification for testing
	}
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	// Call getcert endpoint
	resp, err := client.Post(getCertEndpoint, "application/json", nil)
	require.NoError(t, err, "Failed to call certificate endpoint")
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		t.Logf("Certificate endpoint failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		require.Fail(t, "Certificate endpoint failed", "Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse JSON response
	var result struct {
		Public  string `json:"public"`
		Private string `json:"private"`
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	err = json.Unmarshal(bodyBytes, &result)
	require.NoError(t, err, "Failed to parse JSON response")

	t.Logf("Certificate endpoint response received")

	// Parse and format public certificate (same logic as bootstrap.sh)
	public := result.Public
	header := strings.Join(strings.Fields(public)[0:2], " ")
	footer := strings.Join(strings.Fields(public)[len(strings.Fields(public))-2:], " ")
	base64Content := strings.Join(strings.Fields(public)[2:len(strings.Fields(public))-2], "\n")
	correctedPublic := header + "\n" + base64Content + "\n" + footer

	// Parse and format private key
	private := result.Private
	headerPriv := strings.Join(strings.Fields(private)[0:4], " ")
	footerPriv := strings.Join(strings.Fields(private)[len(strings.Fields(private))-4:], " ")
	base64ContentPriv := strings.Join(strings.Fields(private)[4:len(strings.Fields(private))-4], "\n")
	correctedPrivate := headerPriv + "\n" + base64ContentPriv + "\n" + footerPriv

	// Save working certificates
	publicPath := filepath.Join(testDir, "working-public.pem")
	privatePath := filepath.Join(testDir, "working-private.pem")

	err = ioutil.WriteFile(publicPath, []byte(correctedPublic), 0644)
	require.NoError(t, err, "Failed to save working public certificate")

	err = ioutil.WriteFile(privatePath, []byte(correctedPrivate), 0644)
	require.NoError(t, err, "Failed to save working private key")

	t.Logf("Working certificates saved to %s and %s", publicPath, privatePath)
	return publicPath, privatePath
}

// StartRemoteAgentProcess starts the remote agent as a background process using binary with two-phase auth
func StartRemoteAgentProcess(t *testing.T, config TestConfig) *exec.Cmd {
	// First build the binary
	binaryPath := BuildRemoteAgentBinary(t, config)

	// Phase 1: Get working certificates using bootstrap cert (HTTP protocol only)
	var workingCertPath, workingKeyPath string
	if config.Protocol == "http" {
		fmt.Printf("Using HTTP protocol, obtaining working certificates...\n")
		workingCertPath, workingKeyPath = GetWorkingCertificates(t, config.BaseURL, config.TargetName, config.Namespace,
			config.ClientCertPath, config.ClientKeyPath, filepath.Dir(config.ConfigPath))
	} else {
		// For MQTT, use bootstrap certificates directly
		workingCertPath = config.ClientCertPath
		workingKeyPath = config.ClientKeyPath
	}

	// Phase 2: Start remote agent with working certificates
	args := []string{
		"-config", config.ConfigPath,
		"-client-cert", workingCertPath,
		"-client-key", workingKeyPath,
		"-target-name", config.TargetName,
		"-namespace", config.Namespace,
		"-topology", config.TopologyPath,
		"-protocol", config.Protocol,
	}

	if config.CACertPath != "" {
		args = append(args, "-ca-cert", config.CACertPath)
	}
	// Log the complete binary execution command to test output
	t.Logf("=== Remote Agent Binary Execution Command ===")
	t.Logf("Binary Path: %s", binaryPath)
	t.Logf("Working Directory: %s", filepath.Join(config.ProjectRoot, "remote-agent", "bootstrap"))
	t.Logf("Command Line: %s %s", binaryPath, strings.Join(args, " "))
	t.Logf("Full Arguments: %v", args)
	t.Logf("===============================================")

	fmt.Printf("Starting remote agent with arguments: %v\n", args)
	cmd := exec.Command(binaryPath, args...)
	// Set working directory to where the binary is located
	cmd.Dir = filepath.Join(config.ProjectRoot, "remote-agent", "bootstrap")

	// Create pipes for real-time log streaming
	stdoutPipe, err := cmd.StdoutPipe()
	require.NoError(t, err, "Failed to create stdout pipe")

	stderrPipe, err := cmd.StderrPipe()
	require.NoError(t, err, "Failed to create stderr pipe")

	// Also capture to buffers for final output
	var stdout, stderr bytes.Buffer
	stdoutTee := io.TeeReader(stdoutPipe, &stdout)
	stderrTee := io.TeeReader(stderrPipe, &stderr)

	err = cmd.Start()

	require.NoError(t, err)

	// Start real-time log streaming in background goroutines
	go streamProcessLogs(t, stdoutTee, "Remote Agent STDOUT")
	go streamProcessLogs(t, stderrTee, "Remote Agent STDERR")

	// Final output logging when process exits
	go func() {
		cmd.Wait()
		if stdout.Len() > 0 {
			t.Logf("Remote Agent final stdout: %s", stdout.String())
		}
		if stderr.Len() > 0 {
			t.Logf("Remote Agent final stderr: %s", stderr.String())
		}
	}()

	t.Cleanup(func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})

	t.Logf("Started remote agent process with PID: %d using working certificates", cmd.Process.Pid)
	t.Logf("Remote Agent logs will be shown in real-time with [Remote Agent STDOUT] and [Remote Agent STDERR] prefixes")
	return cmd
}

// WaitForProcessReady waits for a process to be ready by checking if it's still running
func WaitForProcessReady(t *testing.T, cmd *exec.Cmd, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for process to be ready")
		case <-ticker.C:
			// Check if process is still running
			if cmd.ProcessState == nil {
				t.Logf("Process is ready and running")
				return
			}
			if cmd.ProcessState.Exited() {
				t.Fatalf("Process exited unexpectedly: %s", cmd.ProcessState.String())
			}
		}
	}
}

// CreateYAMLFile creates a YAML file with the given content
func CreateYAMLFile(t *testing.T, filePath, content string) error {
	err := ioutil.WriteFile(filePath, []byte(strings.TrimSpace(content)), 0644)
	if err != nil {
		t.Logf("Failed to create YAML file %s: %v", filePath, err)
		return err
	}
	t.Logf("Created YAML file: %s", filePath)
	return nil
}

// CleanupNamespace deletes all Symphony resources in a namespace
func CleanupNamespace(t *testing.T, namespace string) {
	dyn, err := GetDynamicClient()
	if err != nil {
		t.Logf("Failed to get dynamic client for cleanup: %v", err)
		return
	}

	// Clean up Targets
	targets, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "fabric.symphony",
		Version:  "v1",
		Resource: "targets",
	}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})

	if err == nil {
		for _, target := range targets.Items {
			dyn.Resource(schema.GroupVersionResource{
				Group:    "fabric.symphony",
				Version:  "v1",
				Resource: "targets",
			}).Namespace(namespace).Delete(context.Background(), target.GetName(), metav1.DeleteOptions{})
		}
	}

	// Clean up Instances
	instances, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "solution.symphony",
		Version:  "v1",
		Resource: "instances",
	}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})

	if err == nil {
		for _, instance := range instances.Items {
			dyn.Resource(schema.GroupVersionResource{
				Group:    "solution.symphony",
				Version:  "v1",
				Resource: "instances",
			}).Namespace(namespace).Delete(context.Background(), instance.GetName(), metav1.DeleteOptions{})
		}
	}

	t.Logf("Cleaned up Symphony resources in namespace: %s", namespace)
}

// CreateCASecret creates CA secret in cert-manager namespace for trust bundle
func CreateCASecret(t *testing.T, certs CertificatePaths) string {
	secretName := "client-cert-secret"

	// Ensure cert-manager namespace exists
	cmd := exec.Command("kubectl", "create", "namespace", "cert-manager")
	cmd.Run() // Ignore error if namespace already exists

	// Create CA secret in cert-manager namespace with correct key name
	cmd = exec.Command("kubectl", "create", "secret", "generic", secretName,
		"--from-file=ca.crt="+certs.CACert,
		"-n", "cert-manager")

	err := cmd.Run()
	require.NoError(t, err)

	t.Logf("Created CA secret %s in cert-manager namespace", secretName)
	return secretName
}

// CreateClientCertSecret creates client certificate secret in test namespace
func CreateClientCertSecret(t *testing.T, namespace string, certs CertificatePaths) string {
	secretName := "remote-agent-client-secret"

	cmd := exec.Command("kubectl", "create", "secret", "generic", secretName,
		"--from-file=client.crt="+certs.ClientCert,
		"--from-file=client.key="+certs.ClientKey,
		"-n", namespace)

	err := cmd.Run()
	require.NoError(t, err)

	t.Logf("Created client cert secret %s in namespace %s", secretName, namespace)
	return secretName
}

// StartSymphonyWithRemoteAgentConfig starts Symphony with remote agent configuration
func StartSymphonyWithRemoteAgentConfig(t *testing.T, protocol string) {
	var helmValues string

	if protocol == "http" {
		helmValues = "--set remoteAgent.remoteCert.used=true " +
			"--set remoteAgent.remoteCert.trustCAs.secretName=client-cert-secret " +
			"--set remoteAgent.remoteCert.trustCAs.secretKey=ca.crt " +
			"--set remoteAgent.remoteCert.subjects=remote-agent-client " +
			"--set certManager.enabled=true " +
			"--set api.env.ISSUER_NAME=symphony-ca-issuer " +
			"--set api.env.SYMPHONY_SERVICE_NAME=symphony-service"
	} else if protocol == "mqtt" {
		helmValues = "--set remoteAgent.remoteCert.used=true " +
			"--set remoteAgent.remoteCert.trustCAs.secretName=client-cert-secret " +
			"--set remoteAgent.remoteCert.trustCAs.secretKey=ca.crt " +
			"--set remoteAgent.remoteCert.subjects=remote-agent-client " +
			"--set mqtt.mqttClientCert.enabled=true " +
			"--set mqtt.mqttClientCert.secretName=remote-agent-client-secret " +
			"--set mqtt.mqttClientCert.crt=client.crt " +
			"--set mqtt.mqttClientCert.key=client.key " +
			"--set mqtt.brokerAddress=tls://localhost:8883 " +
			"--set mqtt.enabled=true --set mqtt.useTLS=true " +
			"--set certManager.enabled=true " +
			"--set api.env.ISSUER_NAME=symphony-ca-issuer " +
			"--set api.env.SYMPHONY_SERVICE_NAME=symphony-service"
	}

	// Execute mage command from localenv directory
	projectRoot := GetProjectRoot(t)
	localenvDir := filepath.Join(projectRoot, "test", "localenv")

	t.Logf("StartSymphonyWithRemoteAgentConfig: Project root: %s", projectRoot)
	t.Logf("StartSymphonyWithRemoteAgentConfig: Localenv dir: %s", localenvDir)

	// Check if localenv directory exists
	if _, err := os.Stat(localenvDir); os.IsNotExist(err) {
		t.Fatalf("Localenv directory does not exist: %s", localenvDir)
	}

	cmd := exec.Command("mage", "cluster:deploywithsettings", helmValues)
	cmd.Dir = localenvDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Symphony deployment stdout: %s", stdout.String())
		t.Logf("Symphony deployment stderr: %s", stderr.String())

		// Check if the error is related to cert-manager webhook
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "cert-manager-webhook") &&
			strings.Contains(stderrStr, "x509: certificate signed by unknown authority") {
			t.Logf("Detected cert-manager webhook certificate issue, attempting to fix...")
			FixCertManagerWebhook(t)

			// Retry the deployment after fixing cert-manager
			t.Logf("Retrying Symphony deployment after cert-manager fix...")
			var retryStdout, retryStderr bytes.Buffer
			cmd.Stdout = &retryStdout
			cmd.Stderr = &retryStderr

			retryErr := cmd.Run()
			if retryErr != nil {
				t.Logf("Retry deployment stdout: %s", retryStdout.String())
				t.Logf("Retry deployment stderr: %s", retryStderr.String())
				require.NoError(t, retryErr)
			} else {
				t.Logf("Symphony deployment succeeded after cert-manager fix")
				err = nil // Clear the original error since retry succeeded
			}
		}
	}
	require.NoError(t, err)

	t.Logf("Started Symphony with remote agent configuration for %s protocol", protocol)
}

// CleanupCASecret cleans up CA secret from cert-manager namespace
func CleanupCASecret(t *testing.T, secretName string) {
	cmd := exec.Command("kubectl", "delete", "secret", secretName, "-n", "cert-manager", "--ignore-not-found=true")
	cmd.Run()
	t.Logf("Cleaned up CA secret %s from cert-manager namespace", secretName)
}

// CleanupClientSecret cleans up client certificate secret from namespace
func CleanupClientSecret(t *testing.T, namespace, secretName string) {
	cmd := exec.Command("kubectl", "delete", "secret", secretName, "-n", namespace, "--ignore-not-found=true")
	cmd.Run()
	t.Logf("Cleaned up client secret %s from namespace %s", secretName, namespace)
}

// CleanupSymphony cleans up Symphony deployment
func CleanupSymphony(t *testing.T, testName string) {
	// Dump logs first
	cmd := exec.Command("mage", "dumpSymphonyLogsForTest", fmt.Sprintf("'%s'", testName))
	cmd.Dir = "../../../localenv"
	cmd.Run()

	// Destroy symphony
	cmd = exec.Command("mage", "destroy", "all,nowait")
	cmd.Dir = "../../../localenv"
	cmd.Run()
	CleanupSystemdService(t)
	t.Logf("Cleaned up Symphony for test %s", testName)
}

// StartFreshMinikube always creates a brand new minikube cluster
func StartFreshMinikube(t *testing.T) {
	t.Logf("Creating fresh minikube cluster for isolated testing...")

	// Step 1: Always delete any existing cluster first
	t.Logf("Deleting any existing minikube cluster...")
	cmd := exec.Command("minikube", "delete")
	cmd.Run() // Ignore errors - cluster might not exist

	// Wait a moment for cleanup to complete
	time.Sleep(5 * time.Second)

	// Step 2: Start new cluster with optimal settings for testing
	t.Logf("Starting new minikube cluster...")
	cmd = exec.Command("minikube", "start")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Minikube start stdout: %s", stdout.String())
		t.Logf("Minikube start stderr: %s", stderr.String())
		t.Fatalf("Failed to start minikube: %v", err)
	}

	// Step 3: Wait for cluster to be fully ready
	WaitForMinikubeReady(t, 5*time.Minute)

	t.Logf("Fresh minikube cluster is ready for testing")
}

// WaitForMinikubeReady waits for the cluster to be fully operational
func WaitForMinikubeReady(t *testing.T, timeout time.Duration) {
	t.Logf("Waiting for minikube cluster to be ready...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for minikube to be ready after %v", timeout)
		case <-ticker.C:
			// Check 1: Can we get nodes?
			cmd := exec.Command("kubectl", "get", "nodes")
			if cmd.Run() != nil {
				t.Logf("Still waiting for kubectl to connect...")
				continue
			}

			// Check 2: Can we create secrets?
			cmd = exec.Command("kubectl", "auth", "can-i", "create", "secrets")
			if cmd.Run() != nil {
				t.Logf("Still waiting for RBAC permissions...")
				continue
			}

			// Check 3: Are system pods running?
			cmd = exec.Command("kubectl", "get", "pods", "-n", "kube-system", "--field-selector=status.phase=Running")
			output, err := cmd.Output()
			if err != nil || len(strings.TrimSpace(string(output))) == 0 {
				t.Logf("Still waiting for system pods to be running...")
				continue
			}

			t.Logf("Minikube cluster is fully ready!")
			return
		}
	}
}

// StartFreshMinikubeWithRetry starts minikube with retry mechanism
func StartFreshMinikubeWithRetry(t *testing.T, maxRetries int) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		t.Logf("Attempt %d/%d: Starting fresh minikube cluster...", attempt, maxRetries)

		// Delete any existing cluster
		exec.Command("minikube", "delete").Run()
		time.Sleep(5 * time.Second)

		// Try to start
		cmd := exec.Command("minikube", "start")

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		lastErr = cmd.Run()
		if lastErr == nil {
			// Success! Wait for readiness
			WaitForMinikubeReady(t, 5*time.Minute)
			t.Logf("Minikube started successfully on attempt %d", attempt)
			return
		}

		t.Logf("Attempt %d failed: %v", attempt, lastErr)
		t.Logf("Stdout: %s", stdout.String())
		t.Logf("Stderr: %s", stderr.String())

		if attempt < maxRetries {
			t.Logf("Retrying in 10 seconds...")
			time.Sleep(10 * time.Second)
		}
	}

	t.Fatalf("Failed to start minikube after %d attempts. Last error: %v", maxRetries, lastErr)
}

// CleanupMinikube ensures cluster is deleted after testing
func CleanupMinikube(t *testing.T) {
	t.Logf("Cleaning up minikube cluster...")

	cmd := exec.Command("minikube", "delete")
	err := cmd.Run()
	if err != nil {
		t.Logf("Warning: Failed to delete minikube cluster: %v", err)
	} else {
		t.Logf("Minikube cluster deleted successfully")
	}
}

// FixCertManagerWebhook fixes cert-manager webhook certificate issues
func FixCertManagerWebhook(t *testing.T) {
	t.Logf("Fixing cert-manager webhook certificate issues...")

	// Delete webhook configurations to force recreation
	webhookConfigs := []string{
		"cert-manager-webhook",
		"cert-manager-cainjector",
	}

	for _, config := range webhookConfigs {
		t.Logf("Deleting validating webhook configuration: %s", config)
		cmd := exec.Command("kubectl", "delete", "validatingwebhookconfiguration", config, "--ignore-not-found=true")
		cmd.Run() // Ignore errors as the webhook might not exist

		t.Logf("Deleting mutating webhook configuration: %s", config)
		cmd = exec.Command("kubectl", "delete", "mutatingwebhookconfiguration", config, "--ignore-not-found=true")
		cmd.Run() // Ignore errors as the webhook might not exist
	}

	// Restart cert-manager pods to regenerate certificates
	t.Logf("Restarting cert-manager deployments...")
	deployments := []string{
		"cert-manager",
		"cert-manager-webhook",
		"cert-manager-cainjector",
	}

	for _, deployment := range deployments {
		cmd := exec.Command("kubectl", "rollout", "restart", "deployment", deployment, "-n", "cert-manager")
		if err := cmd.Run(); err != nil {
			t.Logf("Warning: Failed to restart deployment %s: %v", deployment, err)
		}
	}

	// Wait for cert-manager to be ready again
	t.Logf("Waiting for cert-manager to be ready after restart...")
	time.Sleep(10 * time.Second)

	// Wait for deployments to be ready
	for _, deployment := range deployments {
		cmd := exec.Command("kubectl", "rollout", "status", "deployment", deployment, "-n", "cert-manager", "--timeout=120s")
		if err := cmd.Run(); err != nil {
			t.Logf("Warning: Deployment %s may not be ready: %v", deployment, err)
		}
	}

	t.Logf("Cert-manager webhook fix completed")
}

// WaitForCertManagerReady waits for cert-manager and CA issuer to be ready
func WaitForCertManagerReady(t *testing.T, timeout time.Duration) {
	t.Logf("Waiting for cert-manager and CA issuer to be ready...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	issuerFixed := false

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for cert-manager to be ready after %v", timeout)
		case <-ticker.C:
			// Step 1: Check if cert-manager pods are running
			cmd := exec.Command("kubectl", "get", "pods", "-n", "cert-manager", "--field-selector=status.phase=Running")
			output, err := cmd.Output()
			if err != nil || len(strings.TrimSpace(string(output))) == 0 {
				t.Logf("Still waiting for cert-manager pods to be running...")
				continue
			}

			// Step 2: Wait for Symphony API server cert to exist
			cmd = exec.Command("kubectl", "get", "secret", "symphony-api-serving-cert", "-n", "default")
			if cmd.Run() != nil {
				t.Logf("Still waiting for Symphony API server certificate...")
				continue
			}

			// Step 3: Check if CA issuer exists
			cmd = exec.Command("kubectl", "get", "issuer", "symphony-ca-issuer", "-n", "default")
			if cmd.Run() != nil {
				t.Logf("Still waiting for CA issuer symphony-ca-issuer...")
				continue
			}

			// Step 4: Check if CA issuer is ready
			cmd = exec.Command("kubectl", "get", "issuer", "symphony-ca-issuer", "-n", "default", "-o", "jsonpath={.status.conditions[0].status}")
			output, err = cmd.Output()
			if err != nil {
				t.Logf("Failed to check issuer status: %v", err)
				continue
			}

			status := strings.TrimSpace(string(output))
			if status != "True" {
				if !issuerFixed {
					t.Logf("CA issuer is not ready (status: %s), attempting to fix timing issue...", status)
					// Fix the timing issue by recreating the issuer
					err := fixIssuerTimingIssue(t)
					if err != nil {
						t.Logf("Failed to fix issuer: %v", err)
						continue
					}
					issuerFixed = true
					t.Logf("Issuer recreation completed, waiting for it to become ready...")
				}
				continue
			}

			t.Logf("Cert-manager and CA issuer are ready")
			return
		}
	}
}

// fixIssuerTimingIssue recreates the CA issuer to fix timing issues
func fixIssuerTimingIssue(t *testing.T) error {
	t.Logf("Fixing CA issuer timing issue...")

	// Delete the existing issuer
	cmd := exec.Command("kubectl", "delete", "issuer", "symphony-ca-issuer", "-n", "default", "--ignore-not-found=true")
	err := cmd.Run()
	if err != nil {
		t.Logf("Warning: Failed to delete issuer: %v", err)
	}

	// Wait a moment for deletion to complete
	time.Sleep(2 * time.Second)

	// Create the issuer with correct configuration
	issuerYAML := `
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: symphony-ca-issuer
  namespace: default
spec:
  ca:
    secretName: symphony-api-serving-cert
`

	// Apply the issuer
	cmd = exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(strings.TrimSpace(issuerYAML))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Logf("Failed to create issuer - stdout: %s, stderr: %s", stdout.String(), stderr.String())
		return err
	}

	t.Logf("CA issuer recreated successfully")
	return nil
}

// WaitForSymphonyServerCert waits for Symphony API server certificate to be created
func WaitForSymphonyServerCert(t *testing.T, timeout time.Duration) {
	t.Logf("Waiting for Symphony API server certificate to be created...")

	// First wait for cert-manager to be ready
	WaitForCertManagerReady(t, timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for Symphony server certificate after %v", timeout)
		case <-ticker.C:
			cmd := exec.Command("kubectl", "get", "secret", "symphony-api-serving-cert", "-n", "default")
			if cmd.Run() == nil {
				t.Logf("Symphony API server certificate is ready")
				return
			}
			t.Logf("Still waiting for Symphony API server certificate...")
		}
	}
}

// DownloadSymphonyCA downloads Symphony server CA certificate to a file
func DownloadSymphonyCA(t *testing.T, testDir string) string {
	caPath := filepath.Join(testDir, "symphony-server-ca.crt")

	t.Logf("Downloading Symphony server CA certificate...")
	cmd := exec.Command("kubectl", "get", "secret", "symphony-api-serving-cert", "-n", "default",
		"-o", "jsonpath={.data.ca\\.crt}")
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to get Symphony server CA certificate")

	// Decode base64
	caData, err := base64.StdEncoding.DecodeString(string(output))
	require.NoError(t, err, "Failed to decode Symphony server CA certificate")

	// Write to file
	err = ioutil.WriteFile(caPath, caData, 0644)
	require.NoError(t, err, "Failed to write Symphony server CA certificate")

	t.Logf("Symphony server CA certificate saved to: %s", caPath)
	return caPath
}

// WaitForPortForwardReady waits for port-forward to be ready by testing TCP connection
func WaitForPortForwardReady(t *testing.T, address string, timeout time.Duration) {
	t.Logf("Waiting for port-forward to be ready at %s...", address)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for port-forward to be ready at %s after %v", address, timeout)
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", address, 2*time.Second)
			if err == nil {
				conn.Close()
				t.Logf("Port-forward is ready and accepting connections at %s", address)
				return
			}
			t.Logf("Still waiting for port-forward at %s... (error: %v)", address, err)
		}
	}
}

// StartPortForward starts kubectl port-forward for Symphony service
func StartPortForward(t *testing.T) *exec.Cmd {
	t.Logf("Starting port-forward for Symphony service...")

	cmd := exec.Command("kubectl", "port-forward", "svc/symphony-service", "8081:8081", "-n", "default")
	err := cmd.Start()
	require.NoError(t, err, "Failed to start port-forward")

	// Wait for port-forward to be truly ready
	WaitForPortForwardReady(t, "127.0.0.1:8081", 30*time.Second)

	t.Cleanup(func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			t.Logf("Killed port-forward process with PID: %d", cmd.Process.Pid)
		}
	})

	t.Logf("Port-forward started with PID: %d and is ready for connections", cmd.Process.Pid)
	return cmd
}

// IsGitHubActions checks if we're running in GitHub Actions environment specifically
func IsGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") != ""
}

// setupGitHubActionsSudo sets up passwordless sudo specifically for GitHub Actions environment
func setupGitHubActionsSudo(t *testing.T) {
	currentUser := GetCurrentUser(t)

	// In GitHub Actions, we often need to add ourselves to sudoers or the user might already be root
	if currentUser == "root" {
		t.Logf("Running as root in GitHub Actions, sudo not needed")
		return
	}

	t.Logf("Setting up passwordless sudo for GitHub Actions environment (user: %s)", currentUser)

	// Create a more permissive sudo rule for GitHub Actions
	githubActionsSudoRule := fmt.Sprintf("%s ALL=(ALL) NOPASSWD: ALL\n", currentUser)
	tempSudoFile := "/etc/sudoers.d/github-actions-integration-test"

	// Write the sudoers rule directly (in GitHub Actions we often have write access)
	err := ioutil.WriteFile(tempSudoFile, []byte(githubActionsSudoRule), 0440)
	if err != nil {
		t.Logf("Failed to write sudo rule directly, trying with sudo...")

		// Fallback: try to use sudo to write the file
		tempFile := "/tmp/github-actions-sudo-rule"
		err = ioutil.WriteFile(tempFile, []byte(githubActionsSudoRule), 0644)
		if err != nil {
			t.Skip("Failed to create GitHub Actions sudo rule file")
		}

		// Copy with sudo
		cmd := exec.Command("sudo", "cp", tempFile, tempSudoFile)
		if err := cmd.Run(); err != nil {
			t.Skip("Failed to setup GitHub Actions sudo access")
		}

		// Set proper permissions
		cmd = exec.Command("sudo", "chmod", "440", tempSudoFile)
		cmd.Run()

		// Clean up temp file
		os.Remove(tempFile)
	}

	// Set up cleanup
	t.Cleanup(func() {
		cleanupCmd := exec.Command("sudo", "rm", "-f", tempSudoFile)
		cleanupCmd.Run()
		t.Logf("Cleaned up GitHub Actions sudo rule: %s", tempSudoFile)
	})

	// Give the system a moment to reload sudoers
	time.Sleep(1 * time.Second)

	// Verify the setup worked
	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		t.Logf("GitHub Actions sudo setup verification failed, but continuing...")
		PrintSudoSetupInstructions(t)
		// Don't skip in GitHub Actions, just warn and continue
	} else {
		t.Logf("GitHub Actions passwordless sudo configured successfully")
	}
}

// CheckSudoAccess checks if sudo access is available and sets up temporary passwordless sudo if needed
func CheckSudoAccess(t *testing.T) {
	// First check if we already have passwordless sudo
	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err == nil {
		t.Logf("Sudo access confirmed for automated testing")
		return
	}

	// Check if we're in GitHub Actions environment specifically
	if IsGitHubActions() {
		t.Logf("Detected GitHub Actions environment, attempting to setup passwordless sudo...")
		setupGitHubActionsSudo(t)
		return
	}

	// Check if we can at least use sudo with password (interactive)
	t.Logf("Checking if sudo access is available (may require password)...")
	cmd = exec.Command("sudo", "true")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		t.Skip("No sudo access available. Please ensure you have sudo privileges.")
	}

	// If not, try to set up temporary passwordless sudo
	t.Logf("Setting up temporary passwordless sudo for integration testing...")

	currentUser := GetCurrentUser(t)
	tempSudoFile := "/etc/sudoers.d/temp-integration-test"

	// Create comprehensive sudo rule for bootstrap.sh operations
	// This covers: systemctl commands, file operations for service creation, package management, and shell execution
	sudoRule := fmt.Sprintf("%s ALL=(ALL) NOPASSWD: /bin/systemctl *, /usr/bin/systemctl *, /bin/bash -c *, /usr/bin/bash -c *, /bin/apt-get *, /usr/bin/apt-get *, /usr/bin/apt *, /bin/apt *, /bin/chmod *, /usr/bin/chmod *, /bin/mkdir *, /usr/bin/mkdir *, /bin/cp *, /usr/bin/cp *, /bin/rm *, /usr/bin/rm *\n", currentUser)

	t.Logf("Creating temporary sudo rule for user '%s'...", currentUser)
	t.Logf("You may be prompted for your sudo password to set up passwordless access for this test.")

	// Write the sudoers rule to a temporary file first
	tempFile := "/tmp/temp-sudo-rule"
	err := ioutil.WriteFile(tempFile, []byte(sudoRule), 0644)
	if err != nil {
		t.Skip("Failed to create temporary sudo rule file.")
	}

	// Copy the file to the sudoers.d directory with proper permissions
	cmd = exec.Command("sudo", "cp", tempFile, tempSudoFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Skip("Failed to set up temporary sudo access. Please ensure you have sudo privileges or configure passwordless sudo manually.")
	}

	// Set proper permissions on the sudoers file
	cmd = exec.Command("sudo", "chmod", "440", tempSudoFile)
	err = cmd.Run()
	if err != nil {
		t.Logf("Warning: Failed to set proper permissions on sudoers file: %v", err)
	}

	// Clean up the temporary file
	os.Remove(tempFile)

	// Give the system a moment to reload sudoers
	time.Sleep(1 * time.Second)

	// Set up cleanup to remove the temporary sudo rule
	t.Cleanup(func() {
		cleanupCmd := exec.Command("sudo", "rm", "-f", tempSudoFile)
		cleanupCmd.Run() // Ignore errors during cleanup
		t.Logf("Cleaned up temporary sudo rule: %s", tempSudoFile)
	})

	// Verify the setup worked
	cmd = exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		// Try to debug the issue
		t.Logf("Sudo verification failed, checking sudoers file...")

		// Check if the file exists and has correct content
		checkCmd := exec.Command("sudo", "cat", tempSudoFile)
		if output, checkErr := checkCmd.Output(); checkErr == nil {
			t.Logf("Sudoers file content: %s", string(output))
		} else {
			t.Logf("Failed to read sudoers file: %v", checkErr)
		}

		// Check sudoers syntax
		syntaxCmd := exec.Command("sudo", "visudo", "-c", "-f", tempSudoFile)
		if syntaxOutput, syntaxErr := syntaxCmd.CombinedOutput(); syntaxErr != nil {
			t.Logf("Sudoers syntax check failed: %v, output: %s", syntaxErr, string(syntaxOutput))
		} else {
			t.Logf("Sudoers syntax is valid")
		}

		PrintSudoSetupInstructions(t)
		t.Skip("Failed to verify temporary sudo setup. The sudoers rule was created but sudo -n still requires password.")
	}

	t.Logf("Temporary passwordless sudo configured successfully for testing")
}

// CheckSudoAccessWithFallback checks sudo access and provides fallback options for testing
func CheckSudoAccessWithFallback(t *testing.T) bool {
	// First check if we already have passwordless sudo
	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err == nil {
		t.Logf("Passwordless sudo access confirmed for automated testing")
		return true
	}

	// Check if we can at least use sudo with password (interactive)
	t.Logf("Checking if interactive sudo access is available...")
	cmd = exec.Command("sudo", "true")
	if err := cmd.Run(); err != nil {
		t.Logf("No sudo access available. Some tests may be skipped.")
		return false
	}

	t.Logf("Interactive sudo access confirmed, but automated testing may require password input")
	return true
}

// PrintSudoSetupInstructions prints instructions for manual sudo setup
func PrintSudoSetupInstructions(t *testing.T) {
	currentUser := GetCurrentUser(t)
	t.Logf("=== Manual Sudo Setup Instructions ===")
	t.Logf("To enable passwordless sudo for testing, create a file:")
	t.Logf("  sudo visudo -f /etc/sudoers.d/symphony-testing")
	t.Logf("Add this line:")
	t.Logf("  %s ALL=(ALL) NOPASSWD: /bin/systemctl *, /usr/bin/systemctl *, /bin/bash -c *, /usr/bin/bash -c *, /bin/apt-get *, /usr/bin/apt-get *, /usr/bin/apt *, /bin/apt *, /bin/chmod *, /usr/bin/chmod *, /bin/mkdir *, /usr/bin/mkdir *, /bin/cp *, /usr/bin/cp *, /bin/rm *, /usr/bin/rm *", currentUser)
	t.Logf("Save and exit. Then re-run the test.")
	t.Logf("===========================================")
}

// GetCurrentUser gets the current user for systemd service
func GetCurrentUser(t *testing.T) string {
	user := os.Getenv("USER")
	if user == "" {
		// Try alternative environment variables
		if u := os.Getenv("USERNAME"); u != "" {
			return u
		}
		// Fallback for containers
		return "root"
	}
	return user
}

// GetCurrentGroup gets the current group for systemd service
func GetCurrentGroup(t *testing.T) string {
	// Usually group name is same as user name in most systems
	user := GetCurrentUser(t)

	// Could also try to get actual group with: id -gn
	cmd := exec.Command("id", "-gn")
	if output, err := cmd.Output(); err == nil {
		group := strings.TrimSpace(string(output))
		if group != "" {
			return group
		}
	}

	// Fallback to username
	return user
}

// StartRemoteAgentWithBootstrap starts remote agent using bootstrap.sh script
func StartRemoteAgentWithBootstrap(t *testing.T, config TestConfig) *exec.Cmd {
	// Check sudo access first with improved command list
	CheckSudoAccess(t)
	hasSudo := CheckSudoAccessWithFallback(t)
	if !hasSudo {
		t.Skip("Sudo access is required for bootstrap testing but is not available")
	}

	// Build the binary first
	if config.Protocol == "mqtt" {
		binaryPath := BuildRemoteAgentBinary(t, config)
		config.BinaryPath = binaryPath
	}

	// Get current user and group
	currentUser := GetCurrentUser(t)
	currentGroup := GetCurrentGroup(t)

	t.Logf("Using user: %s, group: %s for systemd service", currentUser, currentGroup)

	// Prepare bootstrap.sh arguments
	var args []string

	if config.Protocol == "http" {
		// HTTP mode arguments
		args = []string{
			"http",                // protocol
			config.BaseURL,        // endpoint
			config.ClientCertPath, // cert_path
			config.ClientKeyPath,  // key_path
			config.TargetName,     // target_name
			config.Namespace,      // namespace
			config.TopologyPath,   // topology
			currentUser,           // user
			currentGroup,          // group
		}

		// Add Symphony CA certificate if available
		if config.CACertPath != "" {
			args = append(args, config.CACertPath)
			t.Logf("Adding Symphony CA certificate to bootstrap.sh: %s", config.CACertPath)
		}
	} else if config.Protocol == "mqtt" {
		// MQTT mode arguments
		args = []string{
			"mqtt",                // protocol
			"localhost",           // broker_address (will be from config)
			"8883",                // broker_port (will be from config)
			config.ClientCertPath, // cert_path
			config.ClientKeyPath,  // key_path
			config.TargetName,     // target_name
			config.Namespace,      // namespace
			config.TopologyPath,   // topology
			currentUser,           // user
			currentGroup,          // group
			config.BinaryPath,     // binary_path
			config.CACertPath,     // ca_cert_path
			"false",               // use_cert_subject
		}
	} else {
		t.Fatalf("Unsupported protocol: %s", config.Protocol)
	}

	// Start bootstrap.sh
	cmd := exec.Command("./bootstrap.sh", args...)
	cmd.Dir = filepath.Join(config.ProjectRoot, "remote-agent", "bootstrap")

	// Set environment to avoid interactive prompts
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	t.Logf("Starting bootstrap.sh with args: %v", args)
	err := cmd.Start()
	require.NoError(t, err, "Failed to start bootstrap.sh")

	t.Logf("Bootstrap.sh started with PID: %d", cmd.Process.Pid)

	// Wait for bootstrap.sh to complete - increased timeout for GitHub Actions
	go func() {
		err := cmd.Wait()
		if err != nil {
			t.Logf("Bootstrap.sh exited with error: %v", err)
		} else {
			t.Logf("Bootstrap.sh completed successfully")
		}
		t.Logf("Bootstrap.sh stdout: %s", stdout.String())
		if stderr.Len() > 0 {
			t.Logf("Bootstrap.sh stderr: %s", stderr.String())
		}
	}()

	t.Logf("Bootstrap.sh started, systemd service should be created")
	return cmd
}

// CleanupSystemdService cleans up the systemd service created by bootstrap.sh
func CleanupSystemdService(t *testing.T) {
	t.Logf("Cleaning up systemd remote-agent service...")

	// Stop the service
	cmd := exec.Command("sudo", "systemctl", "stop", "remote-agent.service")
	err := cmd.Run()
	if err != nil {
		t.Logf("Warning: Failed to stop service: %v", err)
	}

	// Disable the service
	cmd = exec.Command("sudo", "systemctl", "disable", "remote-agent.service")
	err = cmd.Run()
	if err != nil {
		t.Logf("Warning: Failed to disable service: %v", err)
	}

	// Remove service file
	cmd = exec.Command("sudo", "rm", "-f", "/etc/systemd/system/remote-agent.service")
	err = cmd.Run()
	if err != nil {
		t.Logf("Warning: Failed to remove service file: %v", err)
	}

	// Reload systemd daemon
	cmd = exec.Command("sudo", "systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		t.Logf("Warning: Failed to reload systemd daemon: %v", err)
	}

	t.Logf("Systemd service cleanup completed")
}

// WaitForSystemdService waits for systemd service to be active
func WaitForSystemdService(t *testing.T, serviceName string, timeout time.Duration) {
	t.Logf("Waiting for systemd service %s to be active...", serviceName)

	// First check current status immediately
	CheckSystemdServiceStatus(t, serviceName)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Logf("Timeout waiting for systemd service %s to be active", serviceName)
			// Before failing, check the final status
			CheckSystemdServiceStatus(t, serviceName)
			// Also check if the process is actually running
			CheckServiceProcess(t, serviceName)
			t.Fatalf("Timeout waiting for systemd service %s to be active after %v", serviceName, timeout)
		case <-ticker.C:
			// Check with detailed output
			cmd := exec.Command("sudo", "systemctl", "is-active", serviceName)
			output, err := cmd.CombinedOutput()
			activeStatus := strings.TrimSpace(string(output))

			if err == nil && activeStatus == "active" {
				t.Logf("Systemd service %s is active", serviceName)
				return
			}

			// Log detailed status
			t.Logf("Still waiting for systemd service %s... (current status: %s)", serviceName, activeStatus)
			if activeStatus == "failed" || activeStatus == "inactive" {
				t.Logf("Service %s is in %s state, checking details...", serviceName, activeStatus)
				CheckSystemdServiceStatus(t, serviceName)
				// If service failed, we should fail fast instead of waiting
				if activeStatus == "failed" {
					t.Fatalf("Systemd service %s failed to start", serviceName)
				}
			}
		}
	}
}

// CheckSystemdServiceStatus checks the status of systemd service
func CheckSystemdServiceStatus(t *testing.T, serviceName string) {
	cmd := exec.Command("sudo", "systemctl", "status", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Service %s status check failed: %v", serviceName, err)
	} else {
		t.Logf("Service %s status: %s", serviceName, string(output))
	}
}

// CheckServiceProcess checks if the service process is actually running
func CheckServiceProcess(t *testing.T, serviceName string) {
	t.Logf("Checking if %s process is running...", serviceName)

	// Get the main PID of the service
	cmd := exec.Command("sudo", "systemctl", "show", serviceName, "--property=MainPID")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Failed to get MainPID for %s: %v", serviceName, err)
		return
	}

	pidLine := strings.TrimSpace(string(output))
	if !strings.HasPrefix(pidLine, "MainPID=") {
		t.Logf("Invalid MainPID output for %s: %s", serviceName, pidLine)
		return
	}

	pidStr := strings.TrimPrefix(pidLine, "MainPID=")
	if pidStr == "0" {
		t.Logf("Service %s has no main process (MainPID=0)", serviceName)
		return
	}

	t.Logf("Service %s MainPID: %s", serviceName, pidStr)

	// Check if the process is actually running
	cmd = exec.Command("ps", "-p", pidStr, "-o", "pid,cmd")
	output, err = cmd.Output()
	if err != nil {
		t.Logf("Process %s for service %s is not running: %v", pidStr, serviceName, err)
	} else {
		t.Logf("Process info for %s: %s", serviceName, string(output))
	}
}

// AddHostsEntry adds an entry to /etc/hosts file
func AddHostsEntry(t *testing.T, hostname, ip string) {
	t.Logf("Adding hosts entry: %s %s", ip, hostname)

	// Add entry to /etc/hosts
	entry := fmt.Sprintf("%s %s", ip, hostname)
	cmd := exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/hosts", entry))
	err := cmd.Run()
	require.NoError(t, err, "Failed to add hosts entry")

	// Setup cleanup to remove the entry
	t.Cleanup(func() {
		RemoveHostsEntry(t, hostname)
	})

	t.Logf("Added hosts entry: %s -> %s", hostname, ip)
}

// RemoveHostsEntry removes an entry from /etc/hosts file
func RemoveHostsEntry(t *testing.T, hostname string) {
	t.Logf("Removing hosts entry for: %s", hostname)

	// Remove entry from /etc/hosts
	cmd := exec.Command("sudo", "sed", "-i", fmt.Sprintf("/127.0.0.1 %s/d", hostname), "/etc/hosts")
	err := cmd.Run()
	if err != nil {
		t.Logf("Warning: Failed to remove hosts entry for %s: %v", hostname, err)
	} else {
		t.Logf("Removed hosts entry for: %s", hostname)
	}
}

// SetupSymphonyHosts configures hosts file for Symphony service access
func SetupSymphonyHosts(t *testing.T) {
	// Add symphony-service -> 127.0.0.1 mapping
	AddHostsEntry(t, "symphony-service", "127.0.0.1")
}

// SetupSymphonyHostsForMainTest configures hosts file with main test cleanup
func SetupSymphonyHostsForMainTest(t *testing.T) {
	t.Logf("Adding hosts entry: 127.0.0.1 symphony-service")

	// Add entry to /etc/hosts
	entry := "127.0.0.1 symphony-service"
	cmd := exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/hosts", entry))
	err := cmd.Run()
	require.NoError(t, err, "Failed to add hosts entry")

	// Setup cleanup at main test level
	t.Cleanup(func() {
		t.Logf("Removing hosts entry for: symphony-service")
		cmd := exec.Command("sudo", "sed", "-i", "/127.0.0.1 symphony-service/d", "/etc/hosts")
		if err := cmd.Run(); err != nil {
			t.Logf("Warning: Failed to remove hosts entry for symphony-service: %v", err)
		} else {
			t.Logf("Removed hosts entry for: symphony-service")
		}
	})

	t.Logf("Added hosts entry: symphony-service -> 127.0.0.1")
}

// StartPortForwardForMainTest starts port-forward with main test cleanup
func StartPortForwardForMainTest(t *testing.T) *exec.Cmd {
	t.Logf("Starting port-forward for Symphony service...")

	cmd := exec.Command("kubectl", "port-forward", "svc/symphony-service", "8081:8081", "-n", "default")
	err := cmd.Start()
	require.NoError(t, err, "Failed to start port-forward")

	// Wait for port-forward to be truly ready
	WaitForPortForwardReady(t, "127.0.0.1:8081", 30*time.Second)

	// Setup cleanup at main test level
	t.Cleanup(func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			t.Logf("Killed port-forward process with PID: %d", cmd.Process.Pid)
		}
	})

	t.Logf("Port-forward started with PID: %d and is ready for connections", cmd.Process.Pid)
	return cmd
}

// MQTT-specific helper functions

// CreateMQTTCASecret creates CA secret in cert-manager namespace for MQTT trust bundle
func CreateMQTTCASecret(t *testing.T, certs MQTTCertificatePaths) string {
	secretName := "mqtt-ca"

	// Ensure cert-manager namespace exists
	t.Logf("Creating cert-manager namespace...")
	cmd := exec.Command("kubectl", "create", "namespace", "cert-manager")
	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "already exists") {
		t.Logf("Failed to create cert-manager namespace: %s", string(output))
	}

	// Create CA secret in cert-manager namespace
	t.Logf("Creating CA secret: kubectl create secret generic %s --from-file=ca.crt=%s -n cert-manager", secretName, certs.CACert)
	cmd = exec.Command("kubectl", "create", "secret", "generic", secretName,
		"--from-file=ca.crt="+certs.CACert,
		"-n", "cert-manager")

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to create CA secret: %s", string(output))
	}
	require.NoError(t, err)

	t.Logf("Created CA secret %s in cert-manager namespace", secretName)
	return secretName
}

// CreateMQTTClientCertSecret creates Symphony MQTT client certificate secret in specified namespace
func CreateMQTTClientCertSecret(t *testing.T, namespace string, certs MQTTCertificatePaths) string {
	secretName := "mqtt-client-secret"

	t.Logf("Creating MQTT client secret: kubectl create secret generic %s --from-file=client.crt=%s --from-file=client.key=%s -n %s",
		secretName, certs.SymphonyServerCert, certs.SymphonyServerKey, namespace)
	cmd := exec.Command("kubectl", "create", "secret", "generic", secretName,
		"--from-file=client.crt="+certs.SymphonyServerCert,
		"--from-file=client.key="+certs.SymphonyServerKey,
		"-n", namespace)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to create MQTT client secret: %s", string(output))
	}
	require.NoError(t, err)

	t.Logf("Created MQTT client cert secret %s in namespace %s", secretName, namespace)
	return secretName
}

// SetupExternalMQTTBroker sets up MQTT broker on host machine using Docker
func SetupExternalMQTTBroker(t *testing.T, certs MQTTCertificatePaths, brokerPort int) {
	t.Logf("Setting up external MQTT broker on host machine using Docker on port %d", brokerPort)

	// Create mosquitto configuration file using actual certificate file names
	configContent := fmt.Sprintf(`
port %d
cafile /mqtt/certs/%s
certfile /mqtt/certs/%s
keyfile /mqtt/certs/%s
require_certificate true
use_identity_as_username false
allow_anonymous true
log_dest stdout
log_type all
`, brokerPort, filepath.Base(certs.CACert), filepath.Base(certs.MQTTServerCert), filepath.Base(certs.MQTTServerKey))

	configPath := filepath.Join(filepath.Dir(certs.CACert), "mosquitto.conf")
	err := ioutil.WriteFile(configPath, []byte(strings.TrimSpace(configContent)), 0644)
	require.NoError(t, err)

	// Stop any existing mosquitto container
	t.Logf("Stopping any existing mosquitto container...")
	exec.Command("docker", "stop", "mqtt-broker").Run()
	exec.Command("docker", "rm", "mqtt-broker").Run()

	// Start mosquitto broker with Docker
	certsDir := filepath.Dir(certs.CACert)
	t.Logf("Starting MQTT broker with Docker...")
	t.Logf("Using certificates directly:")
	t.Logf("  CA Cert: %s -> /mqtt/certs/%s", certs.CACert, filepath.Base(certs.CACert))
	t.Logf("  Server Cert: %s -> /mqtt/certs/%s", certs.MQTTServerCert, filepath.Base(certs.MQTTServerCert))
	t.Logf("  Server Key: %s -> /mqtt/certs/%s", certs.MQTTServerKey, filepath.Base(certs.MQTTServerKey))

	t.Logf("Command: docker run -d --name mqtt-broker -p %d:%d -v %s:/mqtt/certs -v %s:/mosquitto/config eclipse-mosquitto:2.0 mosquitto -c /mosquitto/config/mosquitto.conf",
		brokerPort, brokerPort, certsDir, certsDir)

	cmd := exec.Command("docker", "run", "-d",
		"--name", "mqtt-broker",
		"-p", fmt.Sprintf("0.0.0.0:%d:%d", brokerPort, brokerPort),
		"-v", fmt.Sprintf("%s:/mqtt/certs", certsDir),
		"-v", fmt.Sprintf("%s:/mosquitto/config", certsDir),
		"eclipse-mosquitto:2.0",
		"mosquitto", "-c", "/mosquitto/config/mosquitto.conf")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Logf("Docker run stdout: %s", stdout.String())
		t.Logf("Docker run stderr: %s", stderr.String())
	}
	require.NoError(t, err, "Failed to start MQTT broker with Docker")

	t.Logf("MQTT broker started with Docker container ID: %s", strings.TrimSpace(stdout.String()))

	// Wait for broker to be ready
	t.Logf("Waiting for MQTT broker to be ready...")
	time.Sleep(10 * time.Second) // Give Docker time to start

	// // Setup cleanup
	// t.Cleanup(func() {
	// CleanupExternalMQTTBroker(t)
	// })

	t.Logf("External MQTT broker deployed and ready on host:%d", brokerPort)
}

// SetupMQTTBroker deploys and configures MQTT broker with TLS (legacy function for backward compatibility)
func SetupMQTTBroker(t *testing.T, certs MQTTCertificatePaths, brokerPort int) {
	t.Logf("Setting up MQTT broker with TLS on port %d", brokerPort)

	// Create MQTT broker configuration
	brokerConfig := fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: mosquitto-config
  namespace: default
data:
  mosquitto.conf: |
    port %d
    cafile /mqtt/certs/ca.crt
    certfile /mqtt/certs/mqtt-server.crt
    keyfile /mqtt/certs/mqtt-server.key
    require_certificate true
    use_identity_as_username false
    allow_anonymous false
    log_dest stdout
    log_type all
---
apiVersion: v1
kind: Secret
metadata:
  name: mqtt-server-certs
  namespace: default
type: Opaque
data:
  ca.crt: %s
  mqtt-server.crt: %s
  mqtt-server.key: %s
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mosquitto-broker
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mosquitto-broker
  template:
    metadata:
      labels:
        app: mosquitto-broker
    spec:
      containers:
      - name: mosquitto
        image: eclipse-mosquitto:2.0
        ports:
        - containerPort: %d
        volumeMounts:
        - name: config
          mountPath: /mosquitto/config
        - name: certs
          mountPath: /mqtt/certs
        command: ["/usr/sbin/mosquitto", "-c", "/mosquitto/config/mosquitto.conf"]
      volumes:
      - name: config
        configMap:
          name: mosquitto-config
      - name: certs
        secret:
          secretName: mqtt-server-certs
---
apiVersion: v1
kind: Service
metadata:
  name: mosquitto-service
  namespace: default
spec:
  selector:
    app: mosquitto-broker
  ports:
  - port: %d
    targetPort: %d
  type: ClusterIP
`, brokerPort,
		base64.StdEncoding.EncodeToString(readFileBytes(t, certs.CACert)),
		base64.StdEncoding.EncodeToString(readFileBytes(t, certs.MQTTServerCert)),
		base64.StdEncoding.EncodeToString(readFileBytes(t, certs.MQTTServerKey)),
		brokerPort, brokerPort, brokerPort)

	// Save and apply broker configuration
	brokerPath := filepath.Join(filepath.Dir(certs.CACert), "mqtt-broker.yaml")
	err := ioutil.WriteFile(brokerPath, []byte(strings.TrimSpace(brokerConfig)), 0644)
	require.NoError(t, err)

	t.Logf("Applying MQTT broker configuration: kubectl apply -f %s", brokerPath)
	err = ApplyKubernetesManifest(t, brokerPath)
	require.NoError(t, err)

	// Wait for broker to be ready
	t.Logf("Waiting for MQTT broker to be ready...")
	WaitForDeploymentReady(t, "mosquitto-broker", "default", 60*time.Second)

	t.Logf("MQTT broker deployed and ready")
}

// readFileBytes reads file content as bytes for base64 encoding
func readFileBytes(t *testing.T, filePath string) []byte {
	data, err := ioutil.ReadFile(filePath)
	require.NoError(t, err)
	return data
}

// WaitForDeploymentReady waits for a deployment to be ready
func WaitForDeploymentReady(t *testing.T, deploymentName, namespace string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for deployment %s/%s to be ready", namespace, deploymentName)
		case <-ticker.C:
			cmd := exec.Command("kubectl", "get", "deployment", deploymentName, "-n", namespace, "-o", "jsonpath={.status.readyReplicas}")
			output, err := cmd.Output()
			if err == nil {
				readyReplicas := strings.TrimSpace(string(output))
				if readyReplicas == "1" {
					t.Logf("Deployment %s/%s is ready", namespace, deploymentName)
					return
				}
			}
			t.Logf("Still waiting for deployment %s/%s to be ready...", namespace, deploymentName)
		}
	}
}

// TestMQTTConnectivity tests MQTT broker connectivity before proceeding
func TestMQTTConnectivity(t *testing.T, brokerAddress string, brokerPort int, certs MQTTCertificatePaths) {
	t.Logf("Testing MQTT broker connectivity at %s:%d", brokerAddress, brokerPort)

	// Use kubectl port-forward to make MQTT broker accessible
	cmd := exec.Command("kubectl", "port-forward", "svc/mosquitto-service", fmt.Sprintf("%d:%d", brokerPort, brokerPort), "-n", "default")
	err := cmd.Start()
	require.NoError(t, err)

	// Wait for port-forward to be ready
	time.Sleep(5 * time.Second)

	// Cleanup port-forward
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Test basic connectivity (simplified - in real implementation you'd use MQTT client library)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", brokerPort), 10*time.Second)
	if err == nil {
		conn.Close()
		t.Logf("MQTT broker connectivity test passed")
	} else {
		t.Logf("MQTT broker connectivity test failed: %v", err)
		require.NoError(t, err)
	}
}

// StartSymphonyWithMQTTConfig starts Symphony with MQTT configuration
func StartSymphonyWithMQTTConfig(t *testing.T, brokerAddress string) {
	helmValues := fmt.Sprintf("--set remoteAgent.remoteCert.used=true "+
		"--set remoteAgent.remoteCert.trustCAs.secretName=mqtt-ca "+
		"--set remoteAgent.remoteCert.trustCAs.secretKey=ca.crt "+
		"--set remoteAgent.remoteCert.subjects=MyRootCA;localhost "+
		"--set http.enabled=true "+
		"--set mqtt.enabled=true "+
		"--set mqtt.useTLS=true "+
		"--set mqtt.mqttClientCert.enabled=true "+
		"--set mqtt.mqttClientCert.secretName=mqtt-client-secret "+
		"--set mqtt.brokerAddress=%s "+
		"--set certManager.enabled=true "+
		"--set api.env.ISSUER_NAME=symphony-ca-issuer "+
		"--set api.env.SYMPHONY_SERVICE_NAME=symphony-service", brokerAddress)

	t.Logf("Deploying Symphony with MQTT configuration...")
	t.Logf("Command: mage cluster:deployWithSettings \"%s\"", helmValues)

	// Execute mage command from localenv directory
	projectRoot := GetProjectRoot(t)
	localenvDir := filepath.Join(projectRoot, "test", "localenv")

	t.Logf("StartSymphonyWithMQTTConfig: Project root: %s", projectRoot)
	t.Logf("StartSymphonyWithMQTTConfig: Localenv dir: %s", localenvDir)

	// Check if localenv directory exists
	if _, err := os.Stat(localenvDir); os.IsNotExist(err) {
		t.Fatalf("Localenv directory does not exist: %s", localenvDir)
	}

	// Set a longer timeout for MQTT deployment as it's more complex
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "mage", "cluster:deploywithsettings", helmValues)
	cmd.Dir = localenvDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Symphony MQTT deployment stdout: %s", stdout.String())
		t.Logf("Symphony MQTT deployment stderr: %s", stderr.String())

		// Check if the error is related to cert-manager webhook
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "cert-manager-webhook") &&
			strings.Contains(stderrStr, "x509: certificate signed by unknown authority") {
			t.Logf("Detected cert-manager webhook certificate issue, attempting to fix...")
			FixCertManagerWebhook(t)

			// Retry the deployment after fixing cert-manager
			t.Logf("Retrying Symphony MQTT deployment after cert-manager fix...")
			var retryStdout, retryStderr bytes.Buffer
			cmd.Stdout = &retryStdout
			cmd.Stderr = &retryStderr

			retryErr := cmd.Run()
			if retryErr != nil {
				t.Logf("Retry MQTT deployment stdout: %s", retryStdout.String())
				t.Logf("Retry MQTT deployment stderr: %s", retryStderr.String())
				require.NoError(t, retryErr)
			} else {
				t.Logf("Symphony MQTT deployment succeeded after cert-manager fix")
				err = nil // Clear the original error since retry succeeded
			}
		}
	}
	require.NoError(t, err)

	t.Logf("Started Symphony with MQTT configuration")
}

// CleanupExternalMQTTBroker cleans up external MQTT broker Docker container
func CleanupExternalMQTTBroker(t *testing.T) {
	t.Logf("Cleaning up external MQTT broker Docker container...")

	// Stop and remove Docker container
	exec.Command("docker", "stop", "mqtt-broker").Run()
	exec.Command("docker", "rm", "mqtt-broker").Run()

	t.Logf("External MQTT broker cleanup completed")
}

// CleanupMQTTBroker cleans up MQTT broker deployment
func CleanupMQTTBroker(t *testing.T) {
	t.Logf("Cleaning up MQTT broker...")

	// Delete broker deployment and service
	exec.Command("kubectl", "delete", "deployment", "mosquitto-broker", "-n", "default", "--ignore-not-found=true").Run()
	exec.Command("kubectl", "delete", "service", "mosquitto-service", "-n", "default", "--ignore-not-found=true").Run()
	exec.Command("kubectl", "delete", "configmap", "mosquitto-config", "-n", "default", "--ignore-not-found=true").Run()
	exec.Command("kubectl", "delete", "secret", "mqtt-server-certs", "-n", "default", "--ignore-not-found=true").Run()

	t.Logf("MQTT broker cleanup completed")
}

// CleanupMQTTCASecret cleans up MQTT CA secret from cert-manager namespace
func CleanupMQTTCASecret(t *testing.T, secretName string) {
	cmd := exec.Command("kubectl", "delete", "secret", secretName, "-n", "cert-manager", "--ignore-not-found=true")
	cmd.Run()
	t.Logf("Cleaned up MQTT CA secret %s from cert-manager namespace", secretName)
}

// CleanupMQTTClientSecret cleans up MQTT client certificate secret from namespace
func CleanupMQTTClientSecret(t *testing.T, namespace, secretName string) {
	cmd := exec.Command("kubectl", "delete", "secret", secretName, "-n", namespace, "--ignore-not-found=true")
	cmd.Run()
	t.Logf("Cleaned up MQTT client secret %s from namespace %s", secretName, namespace)
}

// StartRemoteAgentProcessComplete starts remote agent as a complete process with full lifecycle management
func StartRemoteAgentProcessComplete(t *testing.T, config TestConfig) *exec.Cmd {
	// First build the binary
	binaryPath := BuildRemoteAgentBinary(t, config)

	// Phase 1: Get working certificates using bootstrap cert (HTTP protocol only)
	var workingCertPath, workingKeyPath string
	if config.Protocol == "http" {
		t.Logf("Using HTTP protocol, obtaining working certificates...")
		workingCertPath, workingKeyPath = GetWorkingCertificates(t, config.BaseURL, config.TargetName, config.Namespace,
			config.ClientCertPath, config.ClientKeyPath, filepath.Dir(config.ConfigPath))
	} else {
		// For MQTT, use bootstrap certificates directly
		workingCertPath = config.ClientCertPath
		workingKeyPath = config.ClientKeyPath
	}

	// Phase 2: Start remote agent with working certificates
	args := []string{
		"-config", config.ConfigPath,
		"-client-cert", workingCertPath,
		"-client-key", workingKeyPath,
		"-target-name", config.TargetName,
		"-namespace", config.Namespace,
		"-topology", config.TopologyPath,
		"-protocol", config.Protocol,
	}

	if config.CACertPath != "" {
		args = append(args, "-ca-cert", config.CACertPath)
	}

	// Log the complete binary execution command to test output
	t.Logf("=== Remote Agent Process Execution Command ===")
	t.Logf("Binary Path: %s", binaryPath)
	t.Logf("Working Directory: %s", filepath.Join(config.ProjectRoot, "remote-agent", "bootstrap"))
	t.Logf("Command Line: %s %s", binaryPath, strings.Join(args, " "))
	t.Logf("Full Arguments: %v", args)
	t.Logf("===============================================")

	t.Logf("Starting remote agent process with arguments: %v", args)
	cmd := exec.Command(binaryPath, args...)
	// Set working directory to where the binary is located
	cmd.Dir = filepath.Join(config.ProjectRoot, "remote-agent", "bootstrap")

	// Create pipes for real-time log streaming
	stdoutPipe, err := cmd.StdoutPipe()
	require.NoError(t, err, "Failed to create stdout pipe")

	stderrPipe, err := cmd.StderrPipe()
	require.NoError(t, err, "Failed to create stderr pipe")

	// Also capture to buffers for final output
	var stdout, stderr bytes.Buffer
	stdoutTee := io.TeeReader(stdoutPipe, &stdout)
	stderrTee := io.TeeReader(stderrPipe, &stderr)

	err = cmd.Start()
	require.NoError(t, err, "Failed to start remote agent process")

	// Start real-time log streaming in background goroutines
	go streamProcessLogs(t, stdoutTee, "Process STDOUT")
	go streamProcessLogs(t, stderrTee, "Process STDERR")

	// Final output logging when process exits
	go func() {
		cmd.Wait()
		if stdout.Len() > 0 {
			t.Logf("Remote agent process final stdout: %s", stdout.String())
		}
		if stderr.Len() > 0 {
			t.Logf("Remote agent process final stderr: %s", stderr.String())
		}
	}()

	// Setup automatic cleanup
	t.Cleanup(func() {
		CleanupRemoteAgentProcess(t, cmd)
	})

	t.Logf("Started remote agent process with PID: %d using working certificates", cmd.Process.Pid)
	t.Logf("Remote agent process logs will be shown in real-time with [Process STDOUT] and [Process STDERR] prefixes")
	return cmd
}

// StartRemoteAgentProcessWithoutCleanup starts remote agent as a complete process but doesn't set up automatic cleanup
// The caller is responsible for calling CleanupRemoteAgentProcess when needed
func StartRemoteAgentProcessWithoutCleanup(t *testing.T, config TestConfig) *exec.Cmd {
	// First build the binary
	binaryPath := BuildRemoteAgentBinary(t, config)

	// Phase 1: Get working certificates using bootstrap cert (HTTP protocol only)
	var workingCertPath, workingKeyPath string
	if config.Protocol == "http" {
		t.Logf("Using HTTP protocol, obtaining working certificates...")
		workingCertPath, workingKeyPath = GetWorkingCertificates(t, config.BaseURL, config.TargetName, config.Namespace,
			config.ClientCertPath, config.ClientKeyPath, filepath.Dir(config.ConfigPath))
	} else {
		// For MQTT, use bootstrap certificates directly
		workingCertPath = config.ClientCertPath
		workingKeyPath = config.ClientKeyPath
	}

	// Phase 2: Start remote agent with working certificates
	args := []string{
		"-config", config.ConfigPath,
		"-client-cert", workingCertPath,
		"-client-key", workingKeyPath,
		"-target-name", config.TargetName,
		"-namespace", config.Namespace,
		"-topology", config.TopologyPath,
		"-protocol", config.Protocol,
	}

	if config.CACertPath != "" {
		args = append(args, "-ca-cert", config.CACertPath)
	}

	// Log the complete binary execution command to test output
	t.Logf("=== Remote Agent Process Execution Command ===")
	t.Logf("Binary Path: %s", binaryPath)
	t.Logf("Working Directory: %s", filepath.Join(config.ProjectRoot, "remote-agent", "bootstrap"))
	t.Logf("Command Line: %s %s", binaryPath, strings.Join(args, " "))
	t.Logf("Full Arguments: %v", args)
	t.Logf("===============================================")

	t.Logf("Starting remote agent process with arguments: %v", args)
	cmd := exec.Command(binaryPath, args...)
	// Set working directory to where the binary is located
	cmd.Dir = filepath.Join(config.ProjectRoot, "remote-agent", "bootstrap")

	// Create pipes for real-time log streaming
	stdoutPipe, err := cmd.StdoutPipe()
	require.NoError(t, err, "Failed to create stdout pipe")

	stderrPipe, err := cmd.StderrPipe()
	require.NoError(t, err, "Failed to create stderr pipe")

	// Also capture to buffers for final output
	var stdout, stderr bytes.Buffer
	stdoutTee := io.TeeReader(stdoutPipe, &stdout)
	stderrTee := io.TeeReader(stderrPipe, &stderr)

	err = cmd.Start()
	require.NoError(t, err, "Failed to start remote agent process")

	// Start real-time log streaming in background goroutines
	go streamProcessLogs(t, stdoutTee, "Process STDOUT")
	go streamProcessLogs(t, stderrTee, "Process STDERR")

	// Final output logging when process exits
	go func() {
		cmd.Wait()
		if stdout.Len() > 0 {
			t.Logf("Remote agent process final stdout: %s", stdout.String())
		}
		if stderr.Len() > 0 {
			t.Logf("Remote agent process final stderr: %s", stderr.String())
		}
	}()

	// NOTE: No automatic cleanup - caller must call CleanupRemoteAgentProcess manually

	t.Logf("Started remote agent process with PID: %d using working certificates", cmd.Process.Pid)
	t.Logf("Remote agent process logs will be shown in real-time with [Process STDOUT] and [Process STDERR] prefixes")
	return cmd
}

// WaitForProcessHealthy waits for a process to be healthy and ready
func WaitForProcessHealthy(t *testing.T, cmd *exec.Cmd, timeout time.Duration) {
	t.Logf("Waiting for remote agent process to be healthy...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for process to be healthy after %v", timeout)
		case <-ticker.C:
			// Check if process is still running
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				t.Fatalf("Process exited unexpectedly: %s", cmd.ProcessState.String())
			}

			elapsed := time.Since(startTime)
			t.Logf("Process health check: PID %d running for %v", cmd.Process.Pid, elapsed)

			// Process is considered healthy if it's been running for at least 10 seconds
			// without exiting (indicating successful startup and connection)
			if elapsed >= 10*time.Second {
				t.Logf("Process is healthy and ready (running for %v)", elapsed)
				return
			}
		}
	}
}

// CleanupRemoteAgentProcess cleans up the remote agent process
func CleanupRemoteAgentProcess(t *testing.T, cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		t.Logf("No process to cleanup")
		return
	}

	t.Logf("Cleaning up remote agent process with PID: %d", cmd.Process.Pid)

	// First try graceful termination
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Logf("Failed to send interrupt signal: %v", err)
	}

	// Wait a moment for graceful shutdown
	time.Sleep(2 * time.Second)

	// Check if process has exited
	if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
		t.Logf("Process still running, force killing...")
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("Failed to kill process: %v", err)
		} else {
			t.Logf("Process force killed")
		}
	}

	// Wait for process to finish
	cmd.Wait()
	t.Logf("Remote agent process cleanup completed")
}
