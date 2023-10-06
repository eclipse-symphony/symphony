# Canary Development
In this scenario, you’ll deploy an application with a front end and a backend. And as you roll out a new backend version, you want to do a canary deployment, in which you send a small portion of the backend traffic to the new backend version, validate its quality, and gradually shift all traffic to the new backend (and rollback if validation fails).

## Modeling the application
You can model the application using Symphony Solution in several different ways:

1. You can put the front-end, the backend (v1), as well as an ingress into the same Solution definition as three components. When you need to roll out a new backend (v2), you patch the Solution object to add a new backend (v2) component, and then add a canary ingress component to adjust traffic shape. During canary, you perform manual or automated validations and adjust the traffic shape by modifying the canary ingress definition.

2. If you consider that ingress is an infrastructure component. You can move the ingress component to a Target object, or to a separate Solution object so that it can be managed separately. 

3. Further, if the front-end and the backend are managed by different teams, you can split them into different Solutions as well.

## Generic flow
Regardless of how you model your application, the canary process is the same:
1. Deploy a new version
2. Adjust traffic so that a certain percentage of traffic is sent to the new version. If everything is shifted to the new version, then deployment is done. If everything is shifted back to the original version, then the deployment is rolled back.
3. Test the new version. Based on test results, either increment the weight of the new version, or reduce the weight of the new version and then return to step 2.

## Sample Artifacts
You can find sample artifacts under ```docs/samples/canary```.

| Artifact | Purpose |
|--------|--------|
| ```activation.yaml``` | Activate the canary workflow |
| ```campaign.yaml``` | Canary workflow definition |
| ```instance.yaml``` | Initial applicaion deployment (front-end + backend (v1)) |
|```solution.yaml``` | Initial applicaion definition (front-end + backend (v1)) |
| ```target.yaml``` | Target definition (current K8s cluster) |

The following diagram illustrates how the stages in the canary workflow are defined, with corresponding stage names in ```campaign.yaml```.

![campaign](../images/canary-flow.png)

## Steps

1. Deploy the original application (front-end + backend v1):
   ```bash
   kubectl apply -f target.yaml
   kubectl apply -f solution.yaml
   kubectl apply -f instance.yaml
   ```
2. Wait for deployment to finish. You can monitor the progress with ```kubectl get instance -w```. Initially, the instance is likely to fail, because Nginx Ingress controller takes time to initialize, causing ```Ingress``` creation to fail. The instance should return to a healthy state after a few minutes (after the next round of state reconciliation happens).
3. In a separate Terminal Window, attach to the front-end pod and ping the backend every second:
   ```bash
   kubectl exec -it <front-end pod name> /bin/bash
   # inside container
   while true; do curl http://<ingress IP address>/api/env/APP_VERSION; sleep 1; done
   ```
   Keep this Terminal window open. You can observe how backend traffic is gradually shifted to v2 without interrupting the front-end.
4. Define and activate the campaign
   ```bash
   kubectl apply -f campaign.yaml
   kubectl apply -f activation.yaml
   ```
5. The campaign takes a few minutes to run. Eventually, you should see all traffic is shifted to v2 in the above Terminal window. Optionally, in a separate Terminal window, you can examine various objects:
   ```bash
   # check how the Solution is patched
   kubectl get solution test-app -o yaml
   # check how the canary ingress is configured (such as weight assignment)
   kubectl get ingress canary-ingress -o yaml
   ```
6. Once everything is done, check the final state of objects:
   ```bash
   # check services, backend-v1 should have been removed
   kubectl get svc
   # check ingresses, canary-ingress should have been removed
   kubectl get ingress
   # check ingress, it should have been reconfigured to route all traffic to v2
   kubectl get ingress ingress -o yaml
   ```



