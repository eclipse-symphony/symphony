# Symphony Remote Agent End-to-End Tests

This directory contains comprehensive end-to-end tests for the Symphony remote agent integration, testing both HTTP and MQTT communication protocols.

## Overview

The test suite validates the complete workflow:
1. Symphony K8s server startup
2. Target resource creation
3. Remote agent connection and communication
4. Topology updates
5. Data interaction between server and agent

## Test Structure

```
test/e2e/remote-agent-integration/
├── scenarios/
│   ├── http-communication/
│   │   └── http_test.go           # HTTP protocol tests
│   └── mqtt-communication/
│       └── mqtt_test.go           # MQTT protocol tests
├── utils/
│   ├── cert_utils.go              # Certificate generation utilities
│   ├── test_helpers.go            # Kubernetes and general test helpers
│   └── mqtt_utils.go              # MQTT broker management utilities
├── go.mod                         # Go module definition
└── README.md                      # This file
```

## Prerequisites

### Required Software
- Go 1.21 or later
- Docker (for MQTT broker containers)
- kubectl (configured with access to a Kubernetes cluster)
- Kubernetes cluster with Symphony CRDs installed

### Optional Software
- mosquitto-clients (for MQTT connectivity verification)

### Environment Setup

1. **Kubernetes Cluster**: Ensure you have access to a Kubernetes cluster with Symphony CRDs installed.

2. **Symphony Server**: The tests assume a Symphony K8s server is running. You can either:
   - Run it manually before the tests
   - Modify the tests to start it automatically (see TODO comments)

3. **Docker Access**: For MQTT tests, Docker must be available to run the Eclipse Mosquitto broker.

## Running the Tests

### Initialize Dependencies

```bash
cd test/e2e/remote-agent-integration
go mod tidy
```

### Run HTTP Communication Tests

```bash
# Run HTTP tests only
go test -v ./scenarios/http-communication/

# Run with verbose output
go test -v -run TestE2EHttpCommunication ./scenarios/http-communication/
```

### Run MQTT Communication Tests

```bash
# Run MQTT tests only
go test -v ./scenarios/mqtt-communication/

# Run specific MQTT test
go test -v -run TestE2EMqttCommunication ./scenarios/mqtt-communication/

# Run MQTT certificate authentication test
go test -v -run TestMQTTCertificateAuthentication ./scenarios/mqtt-communication/

# Run MQTT reconnection test
go test -v -run TestMQTTReconnection ./scenarios/mqtt-communication/
```

### Run All Tests

```bash
# Run all E2E tests
go test -v ./...

# Run with timeout (recommended for E2E tests)
go test -v -timeout 10m ./...
```

## Test Configuration

### HTTP Tests
- **Default URL**: `https://localhost:8443/v1alpha2`
- **Target Name**: `test-http-target`
- **Namespace**: `test-http-ns`

### MQTT Tests
- **MQTT Broker**: Automatically started in Docker container
- **Ports**: 1883 (plain), 8883 (TLS)
- **Target Name**: `test-mqtt-target`
- **Namespace**: `test-mqtt-ns`
- **Topics**: `symphony/request/{target}`, `symphony/response/{target}`

## Test Features

### Automatic Test Environment
- **Certificate Generation**: Self-signed certificates for TLS testing
- **Temporary Directories**: Automatic cleanup after tests
- **Resource Management**: Kubernetes resources are automatically cleaned up
- **MQTT Broker**: Containerized broker with TLS configuration

### HTTP Communication Tests
- Target resource creation and management
- Remote agent startup with HTTP configuration
- Topology update verification
- Data interaction testing
- Certificate-based authentication

### MQTT Communication Tests
- MQTT broker setup with TLS support
- Certificate-based MQTT authentication
- Topic subscription verification
- Bi-directional MQTT communication
- Connection resilience testing
- Invalid certificate handling

## Troubleshooting

### Common Issues

1. **Docker Not Available**
   ```
   Error: Failed to start MQTT broker
   Solution: Ensure Docker is running and accessible
   ```

2. **Kubernetes Access Issues**
   ```
   Error: kubectl apply failed
   Solution: Check kubectl configuration and cluster access
   ```

3. **Symphony CRDs Missing**
   ```
   Error: no matches for kind "Target"
   Solution: Install Symphony CRDs in your cluster
   ```

4. **Certificate Issues**
   ```
   Error: x509: certificate signed by unknown authority
   Solution: Check certificate generation and paths
   ```

### Debug Information

Enable verbose logging:
```bash
go test -v -args -test.v
```

Check Docker containers:
```bash
docker ps | grep test-mqtt-broker
docker logs <container-id>
```

Check Kubernetes resources:
```bash
kubectl get targets,instances -A
kubectl describe target <target-name> -n <namespace>
```

## Customization

### Modifying Test Configuration

Edit the test files to change:
- **Timeouts**: Increase timeout values for slower environments
- **Namespaces**: Use different namespace prefixes
- **Ports**: Configure different MQTT ports
- **URLs**: Point to different Symphony server endpoints

### Adding New Test Scenarios

1. Create new test functions in existing files
2. Add new test files following the naming convention
3. Update this README with new test descriptions

### Environment Variables

The tests support these environment variables:
- `KUBECONFIG`: Path to kubeconfig file
- `SYMPHONY_SERVER_URL`: Override default Symphony server URL

## Integration with CI/CD

### Example GitHub Actions

```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Start minikube
      uses: medyagh/setup-minikube@master
    
    - name: Install Symphony CRDs
      run: kubectl apply -f k8s/config/
    
    - name: Run E2E Tests
      run: |
        cd test/e2e/remote-agent-integration
        go test -v -timeout 15m ./...
```

## Contributing

When adding new tests:
1. Follow existing patterns for setup and cleanup
2. Use descriptive test names and logging
3. Handle both success and failure scenarios
4. Update documentation

## Test Metrics

Typical test execution times:
- HTTP communication test: ~2-3 minutes
- MQTT communication test: ~3-5 minutes (includes Docker container startup)
- Certificate authentication test: ~1-2 minutes
- Reconnection test: ~2-3 minutes

## Security Considerations

- Tests generate self-signed certificates for testing only
- MQTT broker containers are automatically cleaned up
- Temporary files are removed after test completion
- No production credentials are required or used
