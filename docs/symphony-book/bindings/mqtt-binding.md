# MQTT Binding

MQTT binding allows you to access Symphony API through [MQTT](https://mqtt.org/). This enables Symphony to work in IoT environments and bypass firewall restrictions common in enterprise networks.

## Overview

The MQTT binding provides:
- **Asynchronous Communication** - Request/response pattern via MQTT topics
- **TLS Security** - Full encryption and certificate-based authentication
- **Flexible Authentication** - Anonymous, username/password, or certificate-based
- **Remote Agent Support** - Enables MQTT-based remote agent deployment

## Quick Start

| Configuration | Description | Setup Guide |
|---------------|-------------|-------------|
| **No-TLS Setup** | Setup without TLS encryption | [No-TLS Setup](#no-tls-setup) |
| **TLS Setup** | Secure setup with TLS and certificates | [TLS Setup](#tls-setup) |

> **💡 With MQTT configured, you can deploy [Remote Agents](../../remote-agent/bootstrap/README-mqtt.md) to manage remote targets via MQTT.**

## Architecture

```
[Remote Agent] ─── MQTT ───┐
[IoT Device]   ─── MQTT ───┤
[Mobile App]   ─── MQTT ───┤── [MQTT Broker] ── TLS ── [Symphony API]
[Edge Gateway] ─── MQTT ───┤
[Legacy System]─── MQTT ───┘
```

## Key Features

### Request/Response Pattern
- Symphony subscribes to `requestTopic`
- Processes incoming API requests via MQTT
- Publishes responses to `responseTopic`
- Supports correlation ID for request tracking

### Security Options
1. **Anonymous** - No authentication (development only)
2. **Username/Password** - Basic authentication
3. **TLS + Client Certificates** - Mutual TLS authentication (production recommended)

### Topic Patterns
- **Fixed Topics** - `coa-request` / `coa-response`
- **Dynamic Topics** - `symphony/request/{targetName}` / `symphony/response/{targetName}`
- **Custom Topics** - User-defined topic patterns

---

## No-TLS Setup

### Prerequisites
- MQTT broker (Mosquitto recommended)
- Symphony development environment

### 1. Start MQTT Broker
```bash
# Quick start with Docker
docker run --name mosquitto -d -p 1883:1883 eclipse-mosquitto
```

### 2. Configure Symphony

#### Option A: Helm Deployment
```yaml
# values-dev.yaml
mqtt:
  enabled: true
  useTLS: false
  brokerAddress: "tcp://localhost:1883"
  # clientID: "symphony-client"  # Optional, defaults to "symphony-mqtt-client"
  requestTopic: "coa-request"
  responseTopic: "coa-response"
```

Deploy:
```bash
mage cluster:deployWithSettings -f values-dev.yaml
```

#### Option B: Direct Configuration
```json
{
  "bindings": [{
    "type": "bindings.mqtt",
    "config": {
      "brokerAddress": "tcp://localhost:1883",
      "clientID": "symphony-client",
      "requestTopic": "coa-request",
      "responseTopic": "coa-response",
      "useTLS": "false"
    }
  }]
}
```

### 3. Test Connection
```bash
# Subscribe to responses
mosquitto_sub -h localhost -p 1883 -t coa-response

# Send test request
mosquitto_pub -h localhost -p 1883 -t coa-request \
  -m '{"route":"greetings","method":"GET"}'
```

---

## TLS Setup

### Prerequisites
- Secure MQTT broker with TLS
- TLS certificates (CA, client cert, client key)
- Kubernetes cluster

### 1. Generate Certificates

#### Create CA Certificate
```bash
# Generate CA private key
openssl genrsa -out ca.key 4096

# Create CA certificate
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=MyRootCA"
```

#### Create Client Certificate
```bash
# Generate client private key
openssl genrsa -out client.key 2048

# Create client CSR
openssl req -new -key client.key -out client.csr \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=symphony-client"

# Sign client certificate
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out client.crt -days 365
```

### 2. Create Kubernetes Secrets
```bash
# Create CA certificate secret
kubectl create secret generic mqtt-ca --from-file=ca.crt=ca.crt

# Create client certificate secret
kubectl create secret generic mqtt-client-secret \
  --from-file=client.crt=client.crt \
  --from-file=client.key=client.key
```

### 3. Configure Symphony

#### Certificate Configuration Explained
For TLS connections, Symphony uses a dual certificate configuration:

1. **MQTT CA Certificate** (`trustCAs`): Used to verify the MQTT broker's server certificate
2. **Client Certificate** (`mqttClientCert`): Used for client authentication to the MQTT broker

```yaml
# values-prod.yaml
installServiceExt: true

# MQTT CA certificate configuration
# This is where you place the MQTT broker's CA certificate for TLS verification
remoteAgent:
  used: true
  remoteCert:
    trustCAs: 
      secretName: "mqtt-ca"          # Secret containing MQTT broker's CA certificate
      secretKey: "ca.crt"            # Key in the secret (must be MQTT CA cert)
    subjects: "system:serviceaccount:default:symphony-api-sp;target.symphony.microsoft.com"

# MQTT binding configuration
mqtt:
  enabled: true
  useTLS: true
  brokerAddress: "tls://<mqtt-broker-ip>:8883"
  clientID: "symphony-client"
  requestTopic: "coa-request"
  responseTopic: "coa-response"
  # Client certificate for mutual TLS authentication
  mqttClientCert: 
    enabled: true
    secretName: "mqtt-client-secret"  # Secret containing client cert and key
    crt: "client.crt"                # Client certificate file key
    key: "client.key"                # Client private key file key
    mountPath: "/etc/mqtt-client"
```

> **Important**: The `trustCAs` section must contain the MQTT broker's CA certificate, not Symphony's internal CA. This certificate is used to validate the MQTT broker's TLS certificate during connection.

### 4. Deploy
```bash
mage cluster:deployWithSettings -f values-prod.yaml
```

### 5. Test TLS Connection
```bash
# Test with mosquitto client
mosquitto_sub \
  --cafile ca.crt \
  --cert client.crt \
  --key client.key \
  -h <mqtt-broker-ip> -p 8883 \
  -t coa-response
```

---

## Configuration Reference

### Required Parameters
| Parameter | Description | Example |
|-----------|-------------|---------|
| `brokerAddress` | MQTT broker URL | `tcp://localhost:1883` or `tls://broker:8883` |
| `requestTopic` | Topic for incoming requests | `coa-request` |
| `responseTopic` | Topic for outgoing responses | `coa-response` |

### Optional Parameters (Core)
| Parameter | Description | Default | Example |
|-----------|-------------|---------|---------|
| `clientID` | Unique client identifier | `symphony-mqtt-client` | `symphony-client` |
| `useTLS` | Enable TLS encryption | `false` | `true` |

### Optional Parameters (Connection)
| Parameter | Description | Default |
|-----------|-------------|---------|
| `username` | MQTT username | - |
| `password` | MQTT password | - |
| `timeoutSeconds` | Connection timeout | 100 |
| `keepAliveSeconds` | Keep-alive interval | 200 |
| `pingTimeoutSeconds` | Ping timeout | 100 |

### TLS Parameters
| Parameter | Description | Notes |
|-----------|-------------|-------|
| `caCertPath` | CA certificate path | Managed via `remoteAgent.remoteCert.trustCAs` |
| `clientCertPath` | Client certificate path | Managed via `mqtt.mqttClientCert` |
| `clientKeyPath` | Client private key path | Managed via `mqtt.mqttClientCert` |
| `insecureSkipVerify` | Skip TLS verification | For testing only |

### Helm TLS Configuration

| Helm Parameter | Description | Purpose |
|----------------|-------------|---------|
| `remoteAgent.remoteCert.trustCAs.secretName` | Kubernetes secret containing MQTT broker's CA certificate | Validates MQTT broker's TLS certificate |
| `remoteAgent.remoteCert.trustCAs.secretKey` | Key in the secret (e.g., "ca.crt") | Points to the CA certificate file |
| `mqtt.mqttClientCert.secretName` | Kubernetes secret containing client certificate and key | Client authentication to MQTT broker |
| `mqtt.mqttClientCert.crt` | Client certificate file key in secret | TLS client certificate |
| `mqtt.mqttClientCert.key` | Client private key file key in secret | TLS client private key |

---

## Authentication Modes

### 1. Anonymous Access (Development Only)
```yaml
mqtt:
  enabled: true
  useTLS: false
  brokerAddress: "tcp://localhost:1883"
```

### 2. Username/Password
```yaml
mqtt:
  enabled: true
  useTLS: false
  brokerAddress: "tcp://<broker-ip>:1883"
  username: "symphony-user"
  password: "your-password"
```

### 3. TLS with Client Certificates (Production)
```yaml
mqtt:
  enabled: true
  useTLS: true
  brokerAddress: "tls://<broker-ip>:8883"
  mqttClientCert: 
    enabled: true
    secretName: "mqtt-client-secret"
```

---

## Common Configurations

### Multi-Environment Setup
```yaml
# Development
mqtt:
  brokerAddress: "tcp://localhost:1883"
  requestTopic: "dev/request"
  responseTopic: "dev/response"

# Staging
mqtt:
  brokerAddress: "tls://mqtt-staging.company.com:8883"
  requestTopic: "staging/request"
  responseTopic: "staging/response"

# Production
mqtt:
  brokerAddress: "tls://mqtt-prod.company.com:8883"
  requestTopic: "prod/request"
  responseTopic: "prod/response"
```

### High Availability Configuration
```yaml
mqtt:
  enabled: true
  brokerAddress: "tls://mqtt-ha.company.com:8883"
  timeoutSeconds: 30
  keepAliveSeconds: 60
  pingTimeoutSeconds: 10
```

---

## Troubleshooting

### Common Issues

#### Connection Failed
- **Check broker connectivity**: `telnet <broker-ip> <port>`
- **Verify broker is running**: Check broker logs
- **Network issues**: Check firewall rules

#### TLS Certificate Errors
- **"certificate signed by unknown authority"**
  - Solution: Set correct `caCertPath` or use `insecureSkipVerify: "true"` for testing
- **"tls: bad certificate"**
  - Solution: Verify client certificate and key paths
  - Check certificate validity: `openssl verify -CAfile ca.crt client.crt`

#### Authentication Failed
- **Verify credentials**: Test with mosquitto client tools
- **Check broker ACL**: Review broker authentication configuration

### Debug Commands
```bash
# Test broker connectivity
telnet <broker-ip> <port>

# Verify certificates
openssl verify -CAfile ca.crt client.crt
openssl x509 -in client.crt -text -noout

# Test MQTT connection
mosquitto_pub -h <broker-ip> -p <port> \
  --cafile ca.crt --cert client.crt --key client.key \
  -t test -m "hello"

# Check Symphony logs
kubectl logs deployment/symphony-api | grep -i mqtt
```

---

## Security Best Practices

1. **Always use TLS in production**
2. **Use strong certificates** (RSA 2048+ bits)
3. **Implement certificate rotation**
4. **Configure proper broker ACLs**
5. **Store certificates in Kubernetes secrets**
6. **Monitor certificate expiration**
7. **Use unique client IDs per deployment**

---

## Advanced Topics

- **Remote Agent Setup**: [Remote Agent MQTT Guide](../../remote-agent/bootstrap/README-mqtt.md)
- **MQTT Broker Configuration**: See broker-specific documentation
- **Certificate Automation**: Use cert-manager for automatic certificate rotation
- **Monitoring**: Integrate with Prometheus for MQTT metrics

## Limitations

- MQTT binding doesn't support middleware at this time
- No support for MQTT QoS levels other than 0
- Limited to request/response pattern (no pub/sub broadcasting)

## Related Documentation

- [Remote Agent Bootstrap](../../remote-agent/bootstrap/README-mqtt.md)
- [Security Configuration](../security/)
