# Quick Start Guide - Remote Agent E2E Testing

## 🚀 Get Started in 5 Minutes

This guide shows you how to quickly run the Remote Agent E2E tests using both approaches.

## 📋 Prerequisites Check

### 1. Verify Environment
```bash
# Check required tools
docker --version          # For minikube
kubectl version --client  # Kubernetes CLI
go version                # Go 1.19+
minikube version          # Local Kubernetes

# Check if running in WSL/Linux environment
uname -a
```

### 2. Quick Setup
```bash
# Navigate to test directory
cd test/e2e/remote-agent-integration

# Verify Go modules
GOWORK=off go mod tidy
```

## 🎯 Method 1: Binary Testing (Recommended for Development)

### Run HTTP Test
```bash
# Simple one-command execution
GOWORK=off go test -v -timeout 25m ./scenarios/http-communication/ -run TestE2EHttpCommunication

# Expected output:
=== RUN   TestE2EHttpCommunication
=== RUN   TestE2EHttpCommunication/SetupFreshMinikubeCluster
    http_test.go:25: Creating fresh minikube cluster for isolated testing...
=== RUN   TestE2EHttpCommunication/CreateCertificateSecrets
    http_test.go:44: Created CA secret client-cert-secret in cert-manager namespace
=== RUN   TestE2EHttpCommunication/StartSymphonyServer
    http_test.go:52: Started Symphony with remote agent configuration for http protocol
# ... more detailed logs ...
--- PASS: TestE2EHttpCommunication (15m23s)
```

### What This Test Does
1. ✅ Starts fresh minikube cluster
2. ✅ Generates client certificates with `Subject=MyRootCA`
3. ✅ Deploys Symphony with remote agent configuration
4. ✅ Downloads Symphony server CA certificate
5. ✅ Builds Remote Agent binary (`GOOS=linux GOARCH=amd64`)
6. ✅ Starts Remote Agent with TLS validation
7. ✅ Verifies end-to-end communication
8. ✅ Cleans up everything

## 🔧 Method 2: Bootstrap Testing (Production-like)

### Setup Sudo Access
```bash
# Required for systemd service management
echo "$USER ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/test-automation

# Verify sudo works without password
sudo -n true && echo "Sudo configured correctly" || echo "Sudo configuration needed"
```

### Run Bootstrap Test
```bash
# Execute bootstrap-based test
GOWORK=off go test -v -timeout 25m ./scenarios/http-communication/ -run TestE2EHttpCommunicationWithBootstrap

# Expected output:
=== RUN   TestE2EHttpCommunicationWithBootstrap
=== RUN   TestE2EHttpCommunicationWithBootstrap/SetupFreshMinikubeCluster
    http_bootstrap_test.go:25: Creating fresh minikube cluster for isolated testing...
=== RUN   TestE2EHttpCommunicationWithBootstrap/StartRemoteAgentWithBootstrap
    test_helpers.go:831: Starting bootstrap.sh with args: [http https://localhost:8081/v1alpha2 ...]
    test_helpers.go:847: Bootstrap.sh started, systemd service should be created
    test_helpers.go:915: Waiting for systemd service remote-agent.service to be active...
    test_helpers.go:924: Systemd service remote-agent.service is active
# ... more detailed logs ...
--- PASS: TestE2EHttpCommunicationWithBootstrap (18m45s)
```

### What This Test Does
1. ✅ Everything from Method 1, PLUS:
2. ✅ Executes your actual `bootstrap.sh` script
3. ✅ Creates systemd service `/etc/systemd/system/remote-agent.service`
4. ✅ Manages service lifecycle with `systemctl`
5. ✅ Tests production deployment process
6. ✅ Cleans up systemd service and files

## 🐛 Troubleshooting

### Common Issues & Solutions

#### 1. Docker Not Running
```bash
# Error: Cannot connect to the Docker daemon
# Solution: Start Docker service
sudo systemctl start docker
sudo systemctl enable docker
docker version  # Verify it works
```

#### 2. Minikube Start Failed
```bash
# Error: minikube start failed
# Solution: Clean up and retry
minikube delete
docker system prune -f
# Then run test again
```

#### 3. Go Module Issues  
```bash
# Error: module not found
# Solution: 
cd test/e2e/remote-agent-integration
GOWORK=off go mod download
GOWORK=off go mod tidy
```

#### 4. Sudo Permission Required (Bootstrap Tests)
```bash
# Error: sudo: a password is required
# Solution: Configure passwordless sudo
echo "$USER ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/test-automation
```

#### 5. Minikube Tunnel Issues
```bash
# Error: minikube tunnel failed to start
# Solution: Check sudo access and clean up
sudo minikube tunnel --cleanup
# Or restart minikube
minikube delete && minikube start
```

#### 6. Certificate Issues
```bash
# Error: x509: certificate signed by unknown authority
# Solution: Tests handle this automatically, but if persistent:
rm -rf ~/.minikube/certs
minikube delete && minikube start
```

## 📊 Test Results Interpretation

### Successful Test Output
```bash
--- PASS: TestE2EHttpCommunication (15m23s)
    --- PASS: TestE2EHttpCommunication/SetupFreshMinikubeCluster (2m34s)
    --- PASS: TestE2EHttpCommunication/CreateCertificateSecrets (0m15s)
    --- PASS: TestE2EHttpCommunication/StartSymphonyServer (3m45s)
    --- PASS: TestE2EHttpCommunication/SetupSymphonyConnection (0m30s)
    --- PASS: TestE2EHttpCommunication/CreateTestConfigurations (0m10s)
    --- PASS: TestE2EHttpCommunication/StartRemoteAgent (1m20s)
    --- PASS: TestE2EHttpCommunication/VerifyTargetStatus (2m15s)
    --- PASS: TestE2EHttpCommunication/VerifyTopologyUpdate (0m05s)
    --- PASS: TestE2EHttpCommunication/TestDataInteraction (0m15s)
```

### What Success Means
- ✅ **Minikube Setup**: Fresh Kubernetes cluster ready
- ✅ **Certificate Management**: TLS certificates properly generated and deployed
- ✅ **Symphony Deployment**: Server running with remote agent configuration  
- ✅ **TLS Communication**: Mutual TLS authentication working
- ✅ **Remote Agent Connection**: Agent successfully connects to Symphony
- ✅ **Topology Update**: Agent can update topology via API
- ✅ **Data Interaction**: End-to-end workflow functional

## 💡 Pro Tips

### 1. Speed Up Development
```bash
# Skip minikube recreation for faster iteration
# (Manual setup - not in automated tests)
minikube start  # Keep running between tests
# Then modify test to skip minikube setup
```

### 2. Debug Failed Tests
```bash
# Check minikube status
minikube status

# View Symphony logs
kubectl logs -n default deployment/symphony-api -f

# Check systemd service (bootstrap tests)
sudo systemctl status remote-agent.service
sudo journalctl -u remote-agent.service -f
```

### 3. Test Specific Components
```bash
# Run only specific test phases
GOWORK=off go test -v ./scenarios/http-communication/ -run TestE2EHttpCommunication/StartRemoteAgent
```

### 4. Environment Variables
```bash
# Customize test behavior
export MINIKUBE_MEMORY=6144      # More memory for Symphony
export MINIKUBE_CPUS=4           # More CPUs
export TEST_TIMEOUT=30m          # Longer timeout

# Run with custom settings
GOWORK=off go test -v -timeout $TEST_TIMEOUT ./scenarios/http-communication/
```

## 🎯 Next Steps

### 1. Verify Your Setup
```bash
# Run a quick test to verify everything works
GOWORK=off go test -v -timeout 25m ./scenarios/http-communication/ -run TestE2EHttpCommunication
```

### 2. Integrate into Development Workflow
```bash
# Add to your development scripts
#!/bin/bash
echo "Running Remote Agent E2E tests..."
cd test/e2e/remote-agent-integration
GOWORK=off go test -v -timeout 25m ./scenarios/http-communication/
```

### 3. Explore Advanced Features
- Check `BOOTSTRAP_TESTING.md` for production testing
- Review `IMPLEMENTATION_SUMMARY.md` for technical details
- Customize test configurations in `test_helpers.go`

### 4. CI/CD Integration
```yaml
# Example GitHub Actions step
- name: Run Remote Agent E2E Tests
  run: |
    cd test/e2e/remote-agent-integration
    GOWORK=off go test -v -timeout 25m ./scenarios/http-communication/
```

## 🎉 You're Ready!

Your Remote Agent E2E testing framework is fully functional. You can now:

- ✅ Test development changes quickly with binary approach
- ✅ Validate production deployment with bootstrap approach  
- ✅ Run automated tests in CI/CD pipelines
- ✅ Debug issues with comprehensive logging
- ✅ Verify TLS security end-to-end

Happy testing! 🚀
