# Remote Agent Communication Scenario

This scenario tests the communication between Symphony and remote agents using both HTTP and MQTT protocols.

## Overview

This integration test scenario validates:
1. **HTTP Communication with Bootstrap**: Remote agent communicates with Symphony API server using HTTPS with mutual TLS authentication via bootstrap script
2. **MQTT Communication with Bootstrap**: Remote agent communicates with Symphony through an external MQTT broker using TLS via bootstrap script
3. **HTTP Communication with Process**: Direct process-based remote agent communication with Symphony API server using HTTPS
4. **MQTT Communication with Process**: Direct process-based remote agent communication with Symphony through MQTT broker using TLS

## Test Structure

The tests are designed to run **sequentially** to avoid file conflicts during bootstrap script execution:

1. `TestE2EHttpCommunicationWithBootstrap` - HTTP-based communication test with bootstrap script
2. `TestE2EMQTTCommunicationWithBootstrap` - MQTT-based communication test with bootstrap script
3. `TestE2EHttpCommunicationWithProcess` - HTTP-based communication test with direct process
4. `TestE2EMQTTCommunicationWithProcess` - MQTT-based communication test with direct process

## Test Components

### HTTP Bootstrap Test (`verify/http_bootstrap_test.go`)

- Sets up a fresh Minikube cluster
- Generates test certificates for mutual TLS
- Deploys Symphony with HTTP configuration
- Uses `bootstrap.sh` to download and configure remote agent
- Creates Symphony resources (Target, Solution, Instance)
- Validates end-to-end communication

### MQTT Bootstrap Test (`verify/mqtt_test.go`)

- Sets up a fresh Minikube cluster
- Generates MQTT-specific certificates
- Deploys external MQTT broker with TLS support
- Configures Symphony to use MQTT broker
- Uses `bootstrap.sh` with pre-built agent binary
- Creates Symphony resources and validates MQTT communication

### HTTP Process Test (`verify/http_process_test.go`)

- Sets up a fresh Minikube cluster
- Generates test certificates for mutual TLS
- Deploys Symphony with HTTP configuration
- Starts remote agent as a direct process (no systemd service)
- Creates Symphony resources (Target, Solution, Instance)
- Validates end-to-end communication through direct process

### MQTT Process Test (`verify/mqtt_process_test.go`)

- Sets up a fresh Minikube cluster
- Generates MQTT-specific certificates
- Deploys external MQTT broker with TLS support
- Configures Symphony to use MQTT broker
- Starts remote agent as a direct process (no systemd service)
- Creates Symphony resources and validates MQTT communication through direct process

## Running Tests

### Using Mage (Recommended)

```bash
# Run all tests sequentially
mage test

# Run only verification tests
mage verify

# Setup test environment
mage setup

# Cleanup resources
mage cleanup
```

### Using Go Test Directly

```bash
# Run HTTP bootstrap test only
go test -v ./verify -run TestE2EHttpCommunicationWithBootstrap -timeout 30m

# Run MQTT bootstrap test only
go test -v ./verify -run TestE2EMQTTCommunicationWithBootstrap -timeout 30m

# Run HTTP process test only
go test -v ./verify -run TestE2EHttpCommunicationWithProcess -timeout 30m

# Run MQTT process test only
go test -v ./verify -run TestE2EMQTTCommunicationWithProcess -timeout 30m

# Run all tests (may cause conflicts due to parallel execution)
go test -v ./verify -timeout 30m
```

## Prerequisites

- Docker (for MQTT broker)
- Minikube
- kubectl
- Go 1.21+
- Sudo access (for systemd service management)

## Key Features

### Certificate Management

- HTTP: Uses Symphony-generated certificates with CA trust
- MQTT: Uses separate certificate hierarchy for broker and client authentication

### Bootstrap Script Integration

- Bootstrap tests use `bootstrap.sh` for agent setup
- HTTP: Downloads agent binary from Symphony API
- MQTT: Uses pre-built binary with custom configuration

### Process Integration

- Process tests start remote agent as direct process (no systemd service)
- HTTP: Direct HTTP communication with Symphony API
- MQTT: Direct MQTT communication through broker

### Sequential Execution

The tests are configured to run sequentially in the mage file to prevent:

- File conflicts during binary downloads
- Systemd service naming conflicts
- Port binding conflicts

### Cleanup Strategy

- Proper resource cleanup order: Instance → Solution → Target
- Systemd service cleanup (bootstrap tests)
- Process cleanup (process tests)
- Minikube cluster cleanup
- Certificate and secret cleanup

## Troubleshooting

### Common Issues

1. **File Conflicts**: Ensure tests run sequentially, not in parallel
2. **Sudo Permissions**: Tests require passwordless sudo for systemd operations
3. **Port Conflicts**: MQTT broker uses port 8883, ensure it's available
4. **Certificate Issues**: Check certificate generation and trust chain setup

### Debug Commands

```bash
# Check systemd service status
sudo systemctl status remote-agent.service

# View service logs
sudo journalctl -u remote-agent.service -f

# Check MQTT broker
docker ps | grep mqtt

# Verify certificates
openssl x509 -in cert.pem -text -noout
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Remote Agent  │    │   Symphony API  │    │  MQTT Broker    │
│   (Host/WSL)    │    │   (Minikube)    │    │   (Docker)      │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │  HTTP/TLS (Test 1)   │                      │
          ├──────────────────────┤                      │
          │                      │                      │
          │        MQTT/TLS (Test 2)                    │
          └──────────────────────────────────────────────┘
```

## Notes

- Tests create fresh Minikube clusters for isolation
- Each test manages its own certificate hierarchy
- Bootstrap script handles binary management and systemd configuration
- Cleanup is handled automatically via Go test cleanup functions
