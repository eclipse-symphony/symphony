# Deploy Symphony
> **NOTE:** Jump to Step 4 if you want to deploy through Helm.

## 1. Deploy Symphony K8s binding
```bash
cd symphony-k8s
make deploy IMG=<Symphony K8s image tag, such as hbai/symphony-k8s:latest>
```
If you are using Kind, you also need to load the Docker image to Kind cluster:
```bash
kind load docker-image <image tag> --name <Kind cluster name>
```
## 2. Deploy Symphony API
```bash
cd symphony-api
# Symphony API pods
kubectl create -f deployment/symphony-api.yaml
# Symphony ClusterIP service
kubectl create -f deployment/symphony-service.yaml
# Optional LoadBalancer service for testing
kubectl create -f deployment/symphony-service-ext.yaml
```
## 3. Wait for pods to come up
It's highly recommended to wait for all pods to enter Running state before further actions. Use ```kubectl get pods --all-namespaces``` to check all pod states. The following is a sample of how the cluster should look like at this point (a Kind cluster):
```bash
NAMESPACE           NAME                                            READY   STATUS
cert-manager        cert-manager-...                                1/1     Running
cert-manager        cert-manager-cainjector-...                     1/1     Running
cert-manager        cert-manager-webhook-...                        1/1     Running
default             symphony-api-...                                1/1     Running
iotedge             edgeagent-...                                   2/2     Running
iotedge             edgehub-...                                     2/2     Running
iotedge             iotedged-...                                    1/1     Running
kube-system         coredns-...                                     1/1     Running
kube-system         coredns-...                                     1/1     Running
kube-system         etcd-symphony-control-plane                     1/1     Running
kube-system         kindnet-...                                     1/1     Running
kube-system         kube-apiserver-symphony-control-plane           1/1     Running
kube-system         kube-controller-manager-symphony-control-plane  1/1     Running
kube-system         kube-proxy-...                                  1/1     Running
kube-system         kube-scheduler-symphony-control-plane           1/1     Running
local-path-storage  local-path-provisioner-...                      1/1     Running
symphony-k8s-system symphony-k8s-controller-manager-...             2/2     Running
```
## 4. Deploy using Helm
If you have the Symphony repository cloned, you can find a Helm chart under the ```helm``` directory. And you can install Symphony using this Helm chart by:

```bash
cd symphony-k8s/helm
helm install symphony ./symphony
```
Or, you can download the Helm chart from Symphony container registry:

```bash
# enable experimental features
$env:HELM_EXPERIMENTAL_OCI=1 # powershell
SET HELM_EXPERIMENTAL_OCI=1 # cmd 
export HELM_EXPERIMENTAL_OCI=1 # bash
helm install symphony oci://symphonyk8s.azurecr.io/helm/symphony --version 0.1.0
```
If you use a Kubernetes cluster with both Windows nodes and Linux nodes, you need to set affinity of cert-manager and Symphony to use Linux nodes. You can use the values.aks-hci.yaml file as an example of expected value file. Then, deploy your chart using the value file:
```bash
helm install symphony ./symphony -f ./symphony/values.aks-hci.yaml
```
## 5. Uninstall
If you've used Helm to install Symphony, deleting it is easy:
```
helm delete symphony
```
If you've used manual deployment, you need to delete a few things:
```bash
kubectl delete ns symphony-k8s-system
kubectl delete deployment symphony-api
kubectl delete service symphony-service
kubectl delete crd solutions.solution.symphony
kubectl delete crd instances.solution.symphony
kubectl delete crd endpoints.solution.symphony
kubectl delete crd targets.solution.symphony
kubectl delete crd devices.fabric.symphony
kubectl delete crd lockers.security.symphony
kubectl delete crd skills.ai.symphony
```