# Configuring shared components
Many enterprises use services that are shared among multiple applications, such as database services, proxy/routing services, caching services and others. In some systems, these shared services need to be updated accordingly when applications are deployed or removed from the system. Symphony allows configurations of such services to be modeled and managed through coordinated desired state mutation via Symphony workflows.

## Modeling shared components

Symphony allows the shared components to be modeled in different ways depending on specific scenarios. Users can model shared components either as part of a Target stack, or a dedicated shared service Solution.

### As part of a Target stack

If users consider the shared components part of the enterprise platform, they can model these components as parts of a Target specification. For example, on a Target representing a database server, it’s natural to expect the database service is described as a component of that Target. 

### As a dedicated shared service Solution

Users can also capture shared services in a dedicated Solution object. This is a more flexible and preferred way of modeling as it decouples shared services from the infrastructure and allows flexible topologies of deployments, especially when multiple groups of shared services are scoped to support different sets of applications. 

## Updating configuration of shared components

Regardless of how the shared components are modeled, Symphony uses coordinated desired state mutation to manage shared component configurations. Coordinated desired mutation means when Symphony updates one artifact, it can use its workflow to automatically trigger mutation of other related artifacts. 

For example, when a new application is deployed (by creating a new Instance object), a Symphony workflow can update configurations of a shared database server configuration as part of the deployment flow. For this purposes, Symphony workflow (Campaign) supports a `patch` stage type that can patch up Symphony objects through declarative code. 

The following is an example of updating a shared Ingress controller to shift all traffic to a newly deployed service version:

```yaml
finalize:
      ...
      inputs:
        objectType: solution
        objectName: test-app
        patchSource: inline
        patchContent:
          name: ingress
          type: ingress
          metadata:
            annotations.nginx.ingress.kubernetes.io/rewrite-target: "'/$2'"
          properties:
            ingressClassName: nginx
            rules:
            - http:
                paths:
                - path: "'/api(/|$)(.*)'"
                  pathType: ImplementationSpecific
                  backend:
                    service:
                      name: backend-v2
                      port:
                        number: 3013
        patchAction: add
      stageSelector: "finalize-2"
```
 
The coordinated state mutation allows Symphony to support very flexible shared component configuration updates without interrupting the foundational state seeking paradigm. 