---
type: docs
title: "Run Symphony as a Single Process"
linkTitle: "Single Process"
description: ""
weight: 10
---

This tutorial shows you how to launch Symphony as a single process and call the Symphony APIs. Since all state is kept in memory, you can shut it down safely with no cleanup and no leftovers.

## Part 1: Launch Symphony in a Process

1. Install Maestro and launch Symphony as a process (enabled by the ```--no-k8s``` switch):

    {{< tabpane >}}
    {{< tab header="Bash (Linux/WSL/Mac)" lang="bash" >}}
    wget -q https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.sh -O - | /bin/bash
    maestro up --no-k8s
    {{< /tab >}}
    {{< tab header="PowerShell (Windows)" lang="powershell" >}}
    powershell -Command "iwr -useb https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.ps1 | iex"
    maestro up --no-k8s
    {{< /tab >}}
    {{< /tabpane >}}

2. In a separate Terminal Window, Invoke the ```greeting``` endpoint which doesn't require authentication:

    {{< tabpane >}}
    {{< tab header="Bash (Linux/WSL/Mac)" lang="bash" >}}
    curl -L http://localhost:8082/v1alpha2/greetings
    {{< /tab >}}
    {{< tab header="PowerShell (Windows)" lang="powershell" >}}
    (iwr http://localhost:8082/v1alpha2/greetings -UseBasicParsing).Content
    maestro up --no-k8s
    {{< /tab >}}
    {{< /tabpane >}}

    You shoulde see a string reponse:
    ```bash
    Hello from Symphony K8s control plane (S8C)
    ```
## Part 2: Deploying a Docker container

> **NOTE:** You need Docker for this part

1. Use maestro to deploy the hello-world sample, which deploys a Prometheus server Docker container to your machine:
    ```bash
    maestro samples run hello-world
    ```
    You should see outputs like this:
    ```bash
    Creating target sample-local-target ... done
    Creating solution-container sample-prometheus-server ... done
    Creating solution sample-prometheus-server-v-v1 ... done
    Creating instance sample-prometheus-instance ... done
    ⣟  NOTE: Navigate to http://localhost:9090/ to access the Promethus portal (it may take a few seconds for the Prometheus server to be ready)
    ```
2. After a few moments, you should be able to see the Docker container deployed:
    ```bash
    docker ps
    ```
    you should see a `sample-prometheus-server` container in the list.

3. Symphony does continuous state seeking. If you delete the container, Symphony brings it back because that's what the desired state is. To test this, manually shut down the container:
    ```bash
    docker rm -f sample-prometheus-server
    ```
    Then, after a few moments (during next reconcilation loop), the container should be brought back.
    ```bash
    docker ps
    ```
    You should see something like:
    ```bash
    ... IMAGE           ... CREATED          STATUS             PORTS                       NAMES
    ... prom/prometheus ... 18 seconds ago   Up 17 seconds      0.0.0.0:9090->9090/tcp      sample-prometheus-server

    ```
4. (Optional) Open a browser and navigate to `http://localhost:9090` to see the Prometheus interface.

## Part 3: Clean up
1. Kill the Symphony process with `Ctrl+C`.
3. (Optional) Remove the container:
    ```bash
    docker rm -f sample-prometheus-server
    ```