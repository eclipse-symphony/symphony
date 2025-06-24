# Symphony Remote Agent Bootstrap Script Guide

This directory provides scripts to automate the deployment of the Symphony Remote Agent on both Windows and Linux platforms.

## Script Overview

- **bootstrap.ps1**  
  For Windows. Automates certificate import, configuration file generation, remote agent binary download, and registers/starts the agent as a Windows service or scheduled task.

- **bootstrap.sh**  
  For Linux. Automates certificate and key handling, configuration file generation, binary download, and registers/starts the agent as a systemd service.

- **start-up-symphony.sh**  
  For local development. Orchestrates minikube startup, certificate generation, K8s resource deployment, DNS setup, and calls the bootstrap scripts.

---

### 1. bootstrap.ps1 (Windows)

#### Key Features (Windows)

- Imports client certificate (.pfx)
- Generates configuration files
- Downloads the remote agent binary
- Registers and starts the agent as a Windows service or scheduled task

#### Parameters (Windows)

| Name          | Description                                              | Required?  |
| ------------- | -------------------------------------------------------- | ---------- |
| endpoint      | Symphony server API address                              | Yes        |
| cert_path     | Path to client certificate (.pfx)                        | Yes        |
| target_name   | Target name (Remote Target Name)                         | Yes        |
| namespace     | K8s namespace, default is 'default'                      | Optional   |
| topology      | Path to topology file (.json)                            | Yes        |
| run_mode      | 'service' or 'schedule', default is 'schedule'           | Optional   |

#### Example (Windows)

```powershell
pwsh .\bootstrap.ps1 \
  -endpoint "https://symphony-service:8081/v1alpha2" \
  -cert_path ".\certfile\client.pfx" \
  -target_name "windows-target" \
  -namespace "default" \
  -topology "topologies.json" \
  -run_mode "service"
```

> The script will prompt for the certificate password interactively when needed.

> **About `run_mode`:**
>
> - Use `schedule` if you need to install apps that require UI interaction (e.g., UWP apps, retail demo scenarios). In this mode, remote-agent runs as a scheduled task and supports UI interaction.
> - Use `service` if UI interaction is not needed. The agent runs as a Windows service in the background, which is recommended for most production scenarios.

> **Administrator Privileges:**
>
> - Registering either a Windows service (`service` mode) or a scheduled task (`schedule` mode) requires running PowerShell as an administrator.
> - Please ensure you launch your terminal or PowerShell session with elevated (administrator) privileges before running the script.

---

### 2. bootstrap.sh (Linux)

#### Key Features (Linux)

- Handles client certificate (.crt) and key (.key)
- Generates configuration files
- Downloads the remote agent binary
- Registers and starts the agent as a systemd service

#### Parameters (Linux)

| Position | Name        | Description                                 |
| -------- | ----------- | ------------------------------------------- |
| $1       | endpoint    | Symphony server API address                 |
| $2       | cert_path   | Path to client certificate (.crt)           |
| $3       | key_path    | Path to client private key (.key)           |
| $4       | target_name | Target name (Remote Target Name)            |
| $5       | namespace   | K8s namespace                               |
| $6       | topology    | Path to topology file (.json)               |
| $7       | user        | Linux user to run remote-agent              |
| $8       | group       | Linux group to run remote-agent             |

#### Example (Linux)

```bash
sudo ./bootstrap.sh \
  https://symphony-service:8081/v1alpha2 \
  certfile/client.crt \
  certfile/client.key \
  remote-demo \
  default \
  topologies.json \
  <user> \
  <group>
```

> Root privileges are required to register the systemd service.

---

### 3. start-up-symphony.sh (Local End-to-End Startup)

#### Key Features (End-to-End)

- Starts minikube
- Installs openssl
- Generates local CA and client certificates
- Creates K8s Secret
- Deploys Symphony K8s resources
- Configures local hosts/DNS
- Waits for symphony-api-serving-cert and imports local CA
- Calls the bootstrap script to register and start the remote agent

#### Typical Usage

- **Linux:**

  ```bash
  sudo ./bootstrap.sh https://symphony-service:8081/v1alpha2 certfile/client.crt certfile/client.key <target_name> default topologies.json <user> <group>
  ```

- **Windows:**

  ```powershell
  pwsh .\bootstrap.ps1 -endpoint https://symphony-service:8081/v1alpha2 -cert_path .\certfile\client.pfx -target_name <target_name> -namespace default -topology topologies.json -run_mode <run mode>
  ```

---

### How It Works: Key Steps in `start-up-symphony.sh`

This script automates the setup of a local Symphony development environment and the remote agent. The main steps are:

1. **Start minikube** to provide a local Kubernetes cluster.
2. **Install OpenSSL** to enable certificate generation and management.
3. **Generate a local CA** and use it to sign a client certificate and key for secure communication.
4. **Create a Kubernetes secret** to store the client certificate in the cluster.
5. **Verify the secret** to ensure the stored certificate matches the generated one.
6. **Deploy Symphony services** to the cluster using `mage cluster:deployWithSettings`.
7. **Start a minikube tunnel** to expose services on localhost.
8. **(Optional) Remove the local CA cert** from the system trust store for cleanup.
9. **Wait for the Symphony API serving certificate** to be created, then extract and trust it locally.
10. **Configure DNS/hosts** to resolve `symphony-service` locally.
11. **Create the remote target** in Kubernetes.
12. **Stop any running remote-agent service** to avoid conflicts.
13. **Run the bootstrap script** (`bootstrap.sh` or `bootstrap.ps1`) to register and start the remote agent with the correct certificates and configuration.

This end-to-end process ensures a secure, reproducible, and automated setup for Symphony remote agent development and testing.

---

### How to Stop the Agent

- **Windows:**
  - If run_mode is `service` (Windows Service):
    - Stop and remove the Windows service:
      ```powershell
      Stop-Service -Name symphony-service -Force
      sc.exe delete symphony-service
      ```
  - If run_mode is `schedule` (Scheduled Task):
    - Stopping involves two steps:
      1. Unregister the scheduled task:
         ```powershell
         Unregister-ScheduledTask -TaskName "RemoteAgentTask" -Confirm:$false
         ```
      2. Stop the remote agent process (if running):
         ```powershell
         Stop-Process -Name remote-agent -Force
         ```
  - Run these commands in an elevated (administrator) PowerShell session.

- **Linux:**
  - Stop the remote agent systemd service:
    ```bash
    sudo systemctl stop remote-agent.service
    ```
  - To disable and remove the service from startup:
    ```bash
    sudo systemctl disable remote-agent.service
    sudo rm /etc/systemd/system/remote-agent.service
    sudo systemctl daemon-reload
    ```

---

### Notes

- Adjust parameters (endpoint, target_name, user, group, etc.) as needed for your environment.
- Ensure all dependencies are installed (e.g., jq, openssl, PowerShell 7+).
- On Linux, root privileges are required to register the systemd service.
- On Windows, running as a service requires administrator privileges.
- The topology.json file must conform to the Symphony remote agent topology specification.

---

For questions, please refer to the script comments or contact the project maintainer.
