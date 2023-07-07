# Managers

In Symphony's [HB-MVP pattern](https://www.linkedin.com/pulse/hb-mvp-design-pattern-extensible-systems-part-i-haishi-bai/), Managers implement protocol-agnostic, platform-agonistic business logic. They are the core Symphony business layer. These managers include:

* Jobs Manager

  Listens to ```job``` events and invokes Symphony reconcile API (offered by the [Solution Vednor](../vendors/solution.md)). It can also be configured with a timer that triggers peridoical reconcilations.

* Object managers

  For each object type in the [UOM](../uom/uom.md), there's a corresponding manager, named after the object type in plural form, including:
  * Devices manager
  * Instances manager
  * Solutions manager
  * Targets manager

  These managers implement CRUD operations on corresponding object types, and they are hosted by corresponding vendors such as Devices vendor and Target vendor. These vendors collectively offer Symphony REST API to manage Symphony objects.

  When hosted on Kubernetes, such object operations are delegated to Kubernetes API. In such a case, users interact with Symphony objects through native Kubernetes API instead of through these REST API routes. Please see [this diagram](../build_deployment/standalone.md) for more details.

* Solution Manager

  [Solution Manager](./solution-manager.md) implements the core Symphony state-seeking logic. 
  > **NOTE**: Solution Manager is not necessarily a good name. A more appropriate would be a Deployment Manager.

* Reference Manager

  A Reference manager allows object lookups. It also has special logic to resolve an [Azure Custom Vision](https://azure.microsoft.com/en-us/products/cognitive-services/custom-vision-service/) edge model.

* Target Manager

  A Target manager is mostly used by a Symphony agent (that runs on a target machine) to monitor associated [Devices](../uom/device.md).

* Users Manager

    A Users manager implements a simple user store for easy password-based authentication & authorization. This is mostly to facilitate testing. In a production environment, Symphony encourages claim-based architecture that delegate authentication to a trusted Identity Provider (IdP) such as AAD.   