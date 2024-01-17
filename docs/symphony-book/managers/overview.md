# Managers

In Symphony's [HB-MVP pattern](https://www.linkedin.com/pulse/hb-mvp-design-pattern-extensible-systems-part-i-haishi-bai/), managers implement protocol-agnostic, platform-agonistic business logic. They are the core Symphony business layer. These managers include:

* Jobs manager

  Listens to `job` events and invokes Symphony reconcile API (offered by the solution vendor). It can also be configured with a timer that triggers periodical reconciliations.

* Object managers

  For each object type in the [UOM](../concepts/unified-object-model/_overview.md), there's a corresponding manager, named after the object type in plural form, including:

  * Devices manager
  * Instances manager
  * Solutions manager
  * Targets manager

  These managers implement CRUD operations on corresponding object types, and they are hosted by corresponding vendors such as the devices vendor and target vendor. These vendors collectively offer Symphony REST API to manage Symphony objects.

  When hosted on Kubernetes, such object operations are delegated to Kubernetes API. In such a case, users interact with Symphony objects through native Kubernetes API instead of through these REST API routes. For an example, see [Run Symphony in standalone mode](../build_deployment/standalone.md).

* Solution manager

  [Solution manager](./solution-manager.md) implements the core Symphony state-seeking logic.

  > **NOTE**: Solution manager is not necessarily a good name. A more appropriate name would be deployment manager.

* Reference manager

  A reference manager allows object lookups. It also has special logic to resolve an [Azure Custom Vision](https://azure.microsoft.com/products/cognitive-services/custom-vision-service/) edge model.

* Target manager

  A target manager is mostly used by a Symphony agent (that runs on a target machine) to monitor associated [devices](../concepts/unified-object-model/device.md).

* Users manager

  A users manager implements a simple user store for easy password-based authentication and authorization. This is mostly to facilitate testing. In a production environment, Symphony encourages claim-based architecture that delegates authentication to a trusted identity provider (IdP) such as Microsoft Entra ID.
