---
type: docs
title: "Run Symphony on Kubernetes"
linkTitle: "Symphony on Kubernetes"
description: ""
weight: 10
---

This tutorial shows you how to deploy Symphony to a Kubernetes cluster and use `kubectl` to interact with Symphony API in addition to REST API.

## Part 1: Deploy Symphony to a Kubernetes cluster

1. Install Maestro and deploy Symphony to a Kubernetes cluster, to which your `kubectl` is currently pointing:

    {{< tabpane >}}
    {{< tab header="Bash (Linux/WSL/Mac)" lang="bash" >}}
    wget -q https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.sh -O - | /bin/bash
    maestro up
    {{< /tab >}}
    {{< tab header="PowerShell (Windows)" lang="powershell" >}}
    powershell -Command "iwr -useb https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.ps1 | iex"
    maestro up
    {{< /tab >}}
    {{< /tabpane >}}

    You should see outputs like:

    ```bash
    Checking kubectl ...found
    Checking kubectl context ...OK
    Checking Kubernetes connection ...OK
    Installing cert manager ...done
    Checking Helm ...found
    Checking Symphony API (Symphony) ...done
    Checking namespace existence ...done
    Deploying Symphony API (Symphony), chart: oci://ghcr.io/eclipse-symphony/helm/symphony, installServiceExt: true, with MQTT broker: false
    Deploying Symphony API (Symphony) ...done
    Checking public Symphony API address ...
    Symphony API: http://<Symphony REST API IP>:8080/v1alpha2/greetings

    Done!
    ```

2. Check Symphony's pod using `kubectl`:

    ```bash
    kubectl get pods
    ```

    You shoulde see a list of pods running:
    ```bash
    NAME                                   READY   STATUS  
    symphony-api-...                       1/1     Running 
    symphony-cert-manager-...              1/1     Running 
    symphony-cert-manager-cainjector-...   1/1     Running 
    symphony-cert-manager-webhook-...      1/1     Running 
    symphony-controller-manager-...        2/2     Running 
    symphony-redis-...                     1/1     Running 
    symphony-zipkin-...                    1/1     Running     
    ```
## Part 2: Deploying a simple service

1. Use maestro to deploy the hello-k8s sample, which deploys a Prometheus service on Kubernetes:
    ```bash
    maestro samples run hello-k8s
    ```
    You should see outputs like this:
    ```bash
    Creating target sample-k8s-target ... done
    Creating solution sample-prometheus-server ... done
    Creating solutionversion sample-prometheus-server-v-v1 ... done
    Creating instance sample-prometheus-instance ... done
    ... checking post-command condition
    ... checking post-command condition
    ⣯  NOTE: Navigate to http://<Prometheus service public IP>:9090/ to access the Promethus portal (it may take a few minutes for the LoadBalancer to be provisioned)
    ```
2. Navigae to the above URL to visit Prometheus portal.    

3. The sample creates a `sample-k8s-scope` namespace and deploys several resources. Use `kubectl` to examine them:

    ```bash
    kubectl get all -n sample-k8s-scope
    ```

    You should see something like:
    ```bash
    AME                                         READY   STATUS    RESTARTS   AGE
    pod/sample-prometheus-instance-...          1/1     Running   0          38s

    NAME                                        TYPE           CLUSTER-IP     EXTERNAL-IP       PORT(S)          AGE
    service/sample-prometheus-instance          LoadBalancer   10.0.214.131   <Public IP>       9090:32171/TCP   33s

    NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/sample-prometheus-instance   1/1     1            1           38s

    NAME                                         DESIRED   CURRENT   READY   AGE
    replicaset.apps/sample-prometheus-...        1         1         1       38s
    ```
    You should see something like:
    ```bash
    ... IMAGE           ... CREATED          STATUS             PORTS                       NAMES
    ... prom/prometheus ... 18 seconds ago   Up 17 seconds      0.0.0.0:9090->9090/tcp      sample-prometheus-server
    ```

## Part 3: Examine Symphony objects

The sample deploys several Symphony objects:
 * A `sample-k8s-target` Target object that represents a deployment target 
 * A `sample-prometheus-server` Solution object with a `sample-prometheus-server-v-v1` version that represent the software being managed
 * A `sample-prometheus-instance` Instance object that describes a deployment of the software to the target

 > NOTE: See [Orchestration Model](../../concepts/abstractions/orch_model/) for more details.                                         

 1. You can examine these objects using `kubectl`:

    ```bash
    kubectl get target
    kubectl get solution
    kubectl get solutionversion
    kubectl get instance
    ```

2. Alternatively, you can use `maestro` to query the same objects. `maestro` goes through the REST API, but you should see identical objects via either channel:

    ```bash
    maestro get target
    maestro get solution
    maestro get solutionversion
    maestro get instance
    ```

    `maestro` also support direct JSONPath queries, for example:
    ```bash
    maestro get instance --json-path "@.spec.target"
    ```
    This displays the Target name associated with the Instance:
    ```bash
    NAME
    sample-k8s-target
    ```

## Part 3: Clean up
1. Remove the Promethus service by removing the Instance object:

    ```bash
    kubectl delete instance sample-prometheus-instance
    ```

2. Once the Instance object is removed, Symphony state reconcilation removes what it represents. You can check with `kubecctl`:

    ```bash
    kubectl get all -n sample-k8s-scope
    ```
    This should return an empty resource list:
    ```bash
    No resources found in sample-k8s-scope namespace.
    ```
    > **NOTE:** Symphony Kubernetes provider doesn't remove namespaces

3. (Optional) Remove other Symphony objects:
    ```bash
    kubectl delete solutionversion sample-prometheus-server-v-v1
    kubectl delete solution sample-prometheus-server
    kubectl delete target sample-k8s-target
    ```
    Or, remove the namespece altogether:
    ```bash
    kubectl delete ns sample-k8s-scope
    ```