# Canary Development
In this scenario, you’ll deploy an application with a frontend and a backend. And as you roll out a new backend version, you want to do a canary deployment, in which you send a small portion of the backend traffic to the new backend version, validate its quality, and gradually shift all traffic to the new backend (and rollback if validation fails).

## Modeling the application
You can model the application using Symphony Solution in several different ways:

1. You can put the front end, the backend (v1), as well as an ingress into the same Solution definition as three components. When you need to roll out a new backend (v2), you patch the Solution object to add a new backend (v2) component, and then patch the ingress component to adjust traffic shape. During canary, you perform manual validations and adjust the traffic shape by modifying the ingress definition.

2. If you consider that ingress is an infrastructure component. You can move the ingress component to a Target object, or to a separate Solution object so that it can be managed separately. 
3. Further, if the front end and the backend are managed by different teams, you can split them into different Solutions as well.

## Generic flow
Regardless how you model your application, the canary process is the same:
1. Deploy a new version
2. Adjust traffic so that certain percentage of traffic is sent to the new version. If everything is shifted to the new version, then deployment is done. If everything is shifted back to the original version, then the deployment is rolled back.
3. Test the new version. Based on test result, either increment the weight of the new version, or reduce the weight of the new version and then return to step 2.
