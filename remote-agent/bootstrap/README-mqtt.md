# Symphony Remote Agent Bootstrap Guide (MQTT Mode)

This guide describes how to deploy and manage the Symphony Remote Agent in MQTT mode on both Windows and Linux.

> **Note:** For detailed MQTT broker setup, TLS configuration, and certificate management, see the [MQTT Binding Guide](../../docs/symphony-book/bindings/mqtt-binding.md).

---

## 1. Windows (`bootstrap.ps1`)

### Key Features

- Uses client certificate (.crt) and key (.key)
- Generates configuration files
- **You must manually download the remote agent binary and provide its path when prompted**
- Registers and starts the agent as a Windows service or scheduled task

### Parameters

| Name            | Description                                         | Required?  |
| --------------- | --------------------------------------------------- | ---------- |
| protocol        | Must be 'mqtt'                                      | Yes        |
| mqtt_broker     | MQTT broker address                                 | Yes        |
| mqtt_port       | MQTT broker port                                    | Yes        |
| cert_path       | Path to client certificate (.crt)                   | Yes        |
| key_path        | Path to client private key (.key)                   | Yes        |
| target_name     | Target name (Remote Target Name)                    | Yes        |
| namespace       | K8s namespace, default is 'default'                 | Optional   |
| topology        | Path to topology file (.json)                       | Yes        |
| run_mode        | 'service' or 'schedule', default is 'schedule'      | Optional   |
| ca_cert_path    | Path to CA certificate for MQTT                     | Yes        |
| use_cert_subject| Use cert subject as MQTT topic suffix (true/false)  | Optional   |

### Example

```powershell
.\bootstrap.ps1 `
  -protocol mqtt `
  -mqtt_broker <mqtt_broker_ip> `
  -mqtt_port <mqtt_port> `
  -cert_path "/path/to/client.crt" `
  -key_path "/path/to/client.key" `
  -target_name <target_name> `
  -namespace "default" `
  -topology "topologies.json" `
  -run_mode "schedule" `
  -ca_cert_path "/path/to/ca.crt" `
  -use_cert_subject $true
# The script will prompt you to input the remote agent binary path.
```

**About `run_mode`:**
- Use `schedule` if you need to install apps that require UI interaction
- Use `service` if UI interaction is not needed
- Service mode uses `remote-agent.exe install/start/stop/uninstall` for service management

**Administrator Privileges:** Required for registering Windows service or scheduled task.

---

## 2. Linux (`bootstrap.sh`)

### Key Features

- Uses client certificate (.crt) and key (.key)
- Generates configuration files
- **You must manually download the remote agent binary and provide its path as a parameter**
- Registers and starts the agent as a systemd service

### Parameters

| Position | Name            | Description                                         |
| -------- | --------------- | --------------------------------------------------- |
| $1       | protocol        | Must be 'mqtt'                                      |
| $2       | mqtt_broker     | MQTT broker address                                 |
| $3       | mqtt_port       | MQTT broker port                                    |
| $4       | cert_path       | Path to client certificate (.crt)                   |
| $5       | key_path        | Path to client private key (.key)                   |
| $6       | target_name     | Target name (Remote Target Name)                    |
| $7       | namespace       | K8s namespace                                       |
| $8       | topology        | Path to topology file (.json)                       |
| $9       | user            | Linux user to run remote-agent                      |
| $10      | group           | Linux group to run remote-agent                     |
| $11      | binary_path     | Path to remote-agent binary                         |
| $12      | ca_cert_path    | Path to CA certificate for MQTT                     |
| $13      | use_cert_subject| Use cert subject as MQTT topic suffix (true/false)  |

### Example

```bash
sudo ./bootstrap.sh mqtt <mqtt_broker_ip> <mqtt_port> /path/to/client.crt /path/to/client.key <target_name> default topologies.json <user> <group> /path/to/remote-agent /path/to/ca.crt true
```

**Root privileges are required to register the systemd service.**

---

## Remote Agent Configuration

The bootstrap scripts generate configuration files that connect the remote agent to Symphony's MQTT broker. Key configuration elements include:

### MQTT Connection Settings
- Broker address and port
- Client certificates for TLS authentication
- Request/response topics for communication

### Target Configuration
- Target name mapping
- Namespace assignment
- Topology definitions

> **For complete MQTT setup including:**
> - MQTT broker deployment and configuration
> - Certificate authority setup and certificate generation
> - TLS security configuration
> - Access Control Lists (ACL) configuration
> 
> **See:** [MQTT Binding Documentation](../../docs/symphony-book/bindings/mqtt-binding.md)

---

## Symphony K8s Deployment with Remote Agent

### Configuration by Communication Mode

#### Option 1: MQTT with TLS

For Symphony deployment with MQTT TLS support:

```bash
# Create required secrets (see MQTT Binding guide for certificate creation)
# Important: mqtt-ca secret must contain the MQTT broker's CA certificate
kubectl create secret generic mqtt-ca --from-file=ca.crt=ca.crt
kubectl create secret generic mqtt-client-secret \
  --from-file=client.crt=client.crt \
  --from-file=client.key=client.key

# Deploy with MQTT TLS configuration
mage cluster:deployWithSettings \
  "--set remoteAgent.remoteCert.used=true \
  --set remoteAgent.remoteCert.used=true \
  --set remoteAgent.remoteCert.trustCAs.secretName=mqtt-ca \
  --set remoteAgent.remoteCert.trustCAs.secretKey=ca.crt \
  --set mqtt.enabled=true \
  --set mqtt.useTLS=true \
  --set mqtt.brokerAddress=tls://<mqtt_broker_ip>:8883 \
  --set mqtt.mqttClientCert.enabled=true \
  --set mqtt.mqttClientCert.secretName=mqtt-client-secret"
```

#### Option 2: MQTT without TLS

For Symphony deployment with MQTT No-TLS (development/testing):

```bash
# No secrets required for MQTT without TLS
# Deploy with MQTT No-TLS configuration
mage cluster:deployWithSettings \
  "--set mqtt.enabled=true \
  --set mqtt.useTLS=false \
  --set mqtt.brokerAddress=tcp://<mqtt_broker_ip>:1883"
```

#### Option 3: HTTP Protocol

For Symphony deployment with HTTP protocol (always requires certificates):

```bash
# Create required secrets for HTTP (certificates always needed)
kubectl create secret generic remote-agent-ca --from-file=ca.crt=ca.crt
kubectl create secret generic remote-agent-client-secret \
  --from-file=client.crt=client.crt \
  --from-file=client.key=client.key

# Deploy with HTTP configuration
mage cluster:deployWithSettings \
  "--set remoteAgent.remoteCert.used=true \
  --set remoteAgent.remoteCert.used=true \
  --set remoteAgent.remoteCert.trustCAs.secretName=remote-agent-ca \
  --set remoteAgent.remoteCert.trustCAs.secretKey=ca.crt"
```

#### Certificate Configuration for Remote Agent

The MQTT configuration requires proper certificate setup:

1. **MQTT CA Certificate** (`mqtt-ca` secret): Contains the MQTT broker's CA certificate used to verify the broker's TLS certificate
2. **Client Certificate** (`mqtt-client-secret` secret): Contains client certificate and key for authentication

**Complete values.yaml example:**
```yaml
remoteAgent:
  remoteCert:
    used: true
    trustCAs: 
      secretName: "mqtt-ca"              # MQTT broker's CA certificate
      secretKey: "ca.crt"
    subjects: "system:serviceaccount:default:symphony-api-sp;target.symphony.microsoft.com"

mqtt:
  enabled: true
  useTLS: true
  brokerAddress: "tls://<mqtt_broker_ip>:8883"
  requestTopic: "coa-request"
  responseTopic: "coa-response"
  mqttClientCert: 
    enabled: true
    secretName: "mqtt-client-secret"    # Client certificate for authentication
    crt: "client.crt"
    key: "client.key"
    mountPath: "/etc/mqtt-client"
```

> **For detailed Symphony deployment configuration with MQTT:** See [MQTT Binding](../../docs/symphony-book/bindings/mqtt-binding.md)

---

## Authentication Modes

Symphony remote agent supports three MQTT authentication modes with different security levels and certificate requirements:

### Mode Comparison

| Mode | `use_cert_subject` | `remoteCert.used` | Security Level | Certificate Requirements | Client ID | Use Case |
|------|-------------------|--------------------|---------------|-------------------------|-----------|----------|
| **No Authentication** | `false` | `false` | Low | None | Shared/Fixed | Development/Testing |
| **Shared Certificate** | `false` | `true` | Medium | Few shared certificates | Target Name | Medium-scale deployment |
| **One-to-One Authentication** | `true` | `true` | High | One certificate per agent | Certificate Subject | High-security production |

### 1. No Authentication Mode

**Configuration:**
- `use_cert_subject=false`
- No client certificates required
- No TLS encryption (plain TCP connection)
- Uses shared MQTT topics

**Security:** Lowest - No encryption, no client authentication
**Certificate Management:** None required
**Suitable for:** Development and testing environments

**Remote Agent Example:**
```bash
./bootstrap.sh mqtt <broker_ip> 1883 "" "" <target_name> default topologies.json <user> <group> /path/to/agent "" false
```

**Symphony Configuration:**
```bash
# Deploy Symphony with No-TLS MQTT configuration
mage cluster:deployWithSettings \
  "--set mqtt.enabled=true \
  --set mqtt.useTLS=false \
  --set mqtt.brokerAddress=tcp://<mqtt_broker_ip>:1883 \
  --set mqtt.requestTopic=coa-request \
  --set mqtt.responseTopic=coa-response"
```

**Note:** Uses plain TCP connection on port 1883 without TLS encryption.

### 2. Shared Certificate Authentication

**Configuration:**
- `use_cert_subject=false`
- Uses shared client certificates
- Target name used for client identification
- Multiple agents can share the same certificate

**Security:** Medium - Certificate-based authentication with shared certs
**Certificate Management:** Few certificates can serve multiple agents
**Suitable for:** Medium-scale deployments where certificate management overhead needs to be minimized

**Example:**
```bash
./bootstrap.sh mqtt <broker_ip> 8883 /path/to/shared-client.crt /path/to/shared-client.key <target_name> default topologies.json <user> <group> /path/to/agent /path/to/ca.crt false
```

### 3. One-to-One Authentication (Most Secure)

**Configuration:**
- `use_cert_subject=true`
- Each remote agent requires a unique certificate
- **Critical Requirement:** Certificate subject **MUST** match the target name exactly
- Certificate subject becomes the MQTT client ID

**Security:** Highest - Unique certificate per agent with subject-based identification
**Certificate Management:** One certificate per remote agent required
**Suitable for:** High-security production environments

**Key Requirements:**
- Certificate subject = Target name (exact match)
- Each agent needs its own unique certificate
- Certificate subject is used as MQTT client ID

**Certificate Generation Example:**
```bash
# For target "edge-device-01", the certificate subject MUST be "edge-device-01"
openssl req -new -key client.key -out client.csr \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=edge-device-01"

# Sign the certificate
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out edge-device-01.crt -days 365
```

**Bootstrap Example:**
```bash
# target_name and certificate subject must match
./bootstrap.sh mqtt <broker_ip> 8883 /path/to/edge-device-01.crt /path/to/edge-device-01.key edge-device-01 default topologies.json <user> <group> /path/to/agent /path/to/ca.crt true
```

### Security Considerations

**Certificate Subject Matching (Mode 3):**
- When `use_cert_subject=true`, the system extracts the CN (Common Name) from the certificate subject
- This CN becomes the MQTT client ID and must match the target name exactly
- Mismatched subject and target name will prevent proper task execution
- Each remote agent requires a unique certificate with its own subject

**Trade-offs:**
- **Mode 1:** Easy setup, no security
- **Mode 2:** Balanced security and certificate management overhead
- **Mode 3:** Maximum security, highest certificate management overhead

**Recommendation:**
- Use Mode 1 for development and testing
- Use Mode 2 for production environments with moderate security requirements
- Use Mode 3 for high-security production environments where each agent needs unique identification

> **For ACL configuration and security best practices:** See [MQTT Binding Authentication](../../docs/symphony-book/bindings/mqtt-binding.md#authentication-modes)

---

## Notes

- Adjust parameters as needed for your environment
- Replace `<mqtt_broker_ip>` with your actual MQTT broker IP address
- The topology.json file must conform to the Symphony remote agent topology specification
- For HTTP mode, see [README-http.md](./README-http.md)
- For comprehensive MQTT configuration, see [MQTT Binding Guide](../../docs/symphony-book/bindings/mqtt-binding.md)
