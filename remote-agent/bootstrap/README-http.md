# Symphony Remote Agent Bootstrap Guide (HTTP Mode)

This guide describes how to deploy and manage the Symphony Remote Agent in HTTP mode on both Windows and Linux.

---

## 1. Windows (`bootstrap.ps1`)

### Key Features

- Imports client certificate (.pfx)
- Generates configuration files
- Downloads the remote agent binary automatically
- Registers and starts the agent as a Windows service or scheduled task

### Parameters

| Name        | Description                                 | Required?  |
| ----------- | ------------------------------------------- | ---------- |
| endpoint    | Symphony server API address                 | Yes        |
| cert_path   | Path to client certificate (.pfx)           | Yes        |
| target_name | Target name (Remote Target Name)            | Yes        |
| namespace   | K8s namespace, default is 'default'         | Optional   |
| topology    | Path to topology file (.json)               | Yes        |
| run_mode    | 'service' or 'schedule', default is 'schedule' | Optional   |
| protocol    | Must be 'http'                              | Yes        |

### Example

```powershell
.\bootstrap.ps1 `
  -protocol http `
  -endpoint "https://symphony-service:8081/v1alpha2" `
  -cert_path "/path/to/client.pfx" `
  -target_name <target name> `
  -namespace "default" `
  -topology "topologies.json" `
  -run_mode "schedule"
```

- The script will prompt for the certificate password interactively.

**About `run_mode`:**
- Use `schedule` if you need to install apps that require UI interaction (e.g., UWP apps, retail demo scenarios).
- Use `service` if UI interaction is not needed (recommended for most production scenarios).
- Service mode uses `remote-agent.exe install/start/stop/uninstall` for service management.

**Administrator Privileges:**  
Registering either a Windows service or a scheduled task requires running PowerShell as an administrator.

---

## 2. Linux (`bootstrap.sh`)

### Key Features

- Handles client certificate (.crt) and key (.key)
- Generates configuration files
- Downloads the remote agent binary automatically
- Registers and starts the agent as a systemd service

### Parameters

| Position | Name        | Description                                 |
| -------- | ----------- | ------------------------------------------- |
| $1       | protocol    | Must be 'http'                              |
| $2       | endpoint    | Symphony server API address                 |
| $3       | cert_path   | Path to client certificate (.crt)           |
| $4       | key_path    | Path to client private key (.key)           |
| $5       | target_name | Target name (Remote Target Name)            |
| $6       | namespace   | K8s namespace                               |
| $7       | topology    | Path to topology file (.json)               |
| $8       | user        | Linux user to run remote-agent              |
| $9       | group       | Linux group to run remote-agent             |

### Example

```bash
sudo ./bootstrap.sh http https://symphony-service:8081/v1alpha2 /path/to/client.crt /path/to/client.key <target_name> default topologies.json <user> <group>
```

- The remote-agent binary will be downloaded automatically.
- Use absolute paths for all file parameters to avoid issues with systemd.

**Root privileges are required to register the systemd service.**

---

## Notes

- Adjust parameters as needed for your environment.
- The topology.json file must conform to the Symphony remote agent topology specification.
- For MQTT mode, see [README-mqtt.md](./README-mqtt.md).
