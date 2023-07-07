# Deployments

Deployment is a self-contained object that holds all relevant information for Symphony to carry out a reconciliation operation. It packages the [Solution](./solution.md), the [Instance](./instance.md), and the impacted [Target](./target.md)s into one JSON payload and submits to Symphony API. And end user or external system should rarely use this API directly. Instead, they should manage Symphony objects, such as Solutions and Instances, through corresponding REST API routes. 

The deployment object could become heavy, especially when lots of targets are involved. In future versions, Symphony may offer additional routes that optimize payload sizes for large-scale deployments. However, this should not impact on the end user if user keeps using REST API routes. 

On the other hand, this object allows Symphony to be used as a fully stateless, action-based system. A user can trigger a deployment without creating any Symphony objects such as Solutions or Instances.  