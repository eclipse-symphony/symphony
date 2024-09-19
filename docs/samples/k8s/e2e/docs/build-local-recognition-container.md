# Building Local Face Recognition Container
The local recognition container is based on [DeepStack](hhttps://deepstack.readthedocs.io/en/latest/index.html). This document explains how the local face recognition container is built. 
> **NOTE**: To set up demo environment, you can use a pre-built Docker image: ```hbai/face-detection```.

## Setting up build environment
The following instructions are recorded on a Ubuntu WSL 2 environment on Windows. They should apply to a regular Ubuntu machine as well.
### Prerequisites
* [Docker](https://www.docker.com/)
* [WSL 2](https://docs.microsoft.com/en-us/windows/wsl/install)

### Steps
1. Install DeepStack:
   ```bash
    wget https://deepquest.sfo2.digitaloceanspaces.com/deepstack/install-deepstack.sh
    chmod +x install-deepstack.sh
    ./install-deepstack.sh
   ```
2. 
