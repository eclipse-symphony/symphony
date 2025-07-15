# Symphony Remote Agent Bootstrap Guide (MQTT Mode)

This guide describes how to deploy and manage the Symphony Remote Agent in MQTT mode on both Windows and Linux, including certificate preparation and authentication strategies.

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
  -mqtt_broker <broker IP> `
  -mqtt_port <mqtt_port> `
  -cert_path "/path/to/client.crt" `
  -key_path "/path/to/client.key" `
  -target_name <target name> `
  -namespace "default" `
  -topology "topologies.json" `
  -run_mode "schedule" `
  -ca_cert_path "/path/to/ca.crt" `
  -use_cert_subject $true
# The script will prompt you to input the remote agent binary path.
```

- No password prompt is needed; PEM files are used directly.

**About `run_mode`:**
- Use `schedule` if you need to install apps that require UI interaction.
- Use `service` if UI interaction is not needed.
- Service mode uses `remote-agent.exe install/start/stop/uninstall` for service management.

**Administrator Privileges:**  
Registering either a Windows service or a scheduled task requires running PowerShell as an administrator.

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
sudo ./bootstrap.sh mqtt <mqtt_broker> <mqtt_port> /path/to/client.crt /path/to/client.key <target_name> default topologies.json <user> <group> /path/to/remote-agent /path/to/ca.crt true
```

- You must provide the binary path and CA cert path.
- Use absolute paths for all file parameters to avoid issues with systemd.

**Root privileges are required to register the systemd service.**

---

## Certificate Preparation

### 1. Prepare CA certificate for MQTT server

```bash
kubectl create namespace cert-manager
kubectl create secret generic <mqtt_ca> --from-file=ca.crt=./ca.crt -n cert-manager
```

### 2. Prepare MQTT client certificate and key

```bash
kubectl create secret generic <client-secret-name> --from-file=client.crt=./client.crt --from-file=client.key=./client.key -n default
```

### 3. values.yaml and deployment command

No need to manually edit values.yaml, just pass parameters via `--set` when deploying:

```bash
mage cluster:deployWithSettings \
  "--set remoteAgent.used=true \
  --set remoteCert.remoteCAs.secretName=<mqtt_ca> \
  --set remoteCert.remoteCAs.secretKey=ca.crt \
  --set remoteCert.subjects=MyRootCA \
  --set mqtt.mqttClientCert.enabled=true \
  --set mqtt.mqttClientCert.secretName=<client-secret-name> \
  --set mqtt.mqttClientCert.crtKey=client.crt \
  --set mqtt.mqttClientCert.keyKey=client.key \
  --set mqtt.brokerAddress=tls://your-mqtt-broker:port \
  --set installServiceExt=true"
```

---

## Authentication Modes

Symphony supports three MQTT authentication modes:

1. **No certificate authentication**  
   - No certificate is used, all clients share the same topic. Least secure.
2. **Shared certificate authentication**  
   - Multiple clients use the same certificate (shared cert/key pair). `targetName` is used to distinguish clients.
3. **Strict certificate authentication (`use_cert_subject=true`)**  
   - Each client uses a unique certificate. The MQTT topic suffix is the client certificate subject (Common Name).

See the main documentation for ACL examples and security notes.

---

## Certificate Issuing Procedure

See the main documentation for detailed OpenSSL commands and best practices.  
When issuing the MQTT server certificate, include `extendedKeyUsage = serverAuth` in the certificate extensions.

---

## Notes

- Adjust parameters as needed for your environment.
- The topology.json file must conform to the Symphony remote agent topology specification.
- For HTTP mode, see [README-http.md](./README-http.md).
