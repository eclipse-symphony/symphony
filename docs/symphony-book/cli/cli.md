# Symphony CLI (maestro)

_(last edit: 4/12/2023)_

Maestro is a cross-platform CLI for you to interact with your Symphony control planes. When running on Kubernetes, you can use maestro together with kubectl without conflicts.

## Download and install

### Linux/Mac

```bash
wget -q https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.sh -O - | /bin/bash
```

### Windows

```cmd
powershell -Command "iwr -useb https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.ps1 | iex"
```

## Install Symphony

Install all prerequisites and Symphony, including:

* Docker
* Kubernetes (Kind)
* Kubectl
* Helm 3
* Symphony API

```bash
./maestro up
```

## Check prerequisites

Check all Symphony dependencies

```bash
./maestro check
```
