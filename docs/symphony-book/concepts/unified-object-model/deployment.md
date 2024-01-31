# Deployments

A `deployment` is a self-contained object that holds all the relevant information for Symphony to carry out a reconciliation operation. It packages the [solution](./solution.md), the [instance](./instance.md), and the impacted [target](./target.md)s into one JSON payload and submits it to Symphony API. An end user or external system should rarely use this API directly. Instead, they should manage Symphony objects, such as solutions and instances, through corresponding REST API routes.

The deployment object could become heavy, especially when lots of targets are involved. In future versions, Symphony may offer additional routes that optimize payload sizes for large-scale deployments. However, this should not impact on the end user if user keeps using REST API routes.

On the other hand, this object allows Symphony to be used as a fully stateless, imperative system. A user can trigger a deployment without creating any Symphony objects such as solutions or instances.
