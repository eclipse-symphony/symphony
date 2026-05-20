# Managers

In Symphony's [HB-MVP pattern](https://www.linkedin.com/pulse/hb-mvp-design-pattern-extensible-systems-part-i-haishi-bai/), managers implement protocol-agnostic, platform-agonistic business logic. They are the core Symphony business layer. These managers include:

* Jobs manager

  Listens to `job` events and invokes Symphony reconcile API (offered by the solutionversion vendor). It can also be configured with a timer that triggers periodical reconciliations.

* Object managers

  For each object type in the [unified object model](../concepts/unified-object-model/_overview.md), there's a corresponding manager named after the object type in plural form, including:

  * Devices manager
  * Instances manager
  * SolutionVersions manager
  * Targets manager
  * CampaignVersions manager
  * Activations manager
  * CatalogVersions manager

  These managers implement CRUD operations on corresponding object types, and they are hosted by corresponding vendors such as the devices vendor and target vendor. These vendors collectively offer Symphony REST API to manage Symphony objects.

  When hosted on Kubernetes, such object operations are delegated to Kubernetes API. In such a case, users interact with Symphony objects through native Kubernetes API instead of through these REST API routes. For an example, see [Run Symphony in kubernetes mode](../build_deployment/symphony_mode.md).

* SolutionVersion manager

  [SolutionVersion manager](./solutionversion-manager.md) implements the core Symphony state-seeking logic.

  > **NOTE**: SolutionVersion manager is not necessarily a good name. A more appropriate name would be deployment manager.

* Reference manager

  A reference manager allows object lookups. It also has special logic to resolve an [Azure Custom Vision](https://azure.microsoft.com/products/cognitive-services/custom-vision-service/) edge model.

* Target manager

  A target manager is mostly used by a Symphony agent (that runs on a target machine) to monitor associated [devices](../concepts/unified-object-model/device.md).

* Users manager

  A users manager implements a simple user store for easy password-based authentication and authorization. This is mostly to facilitate testing. In a production environment, Symphony encourages claim-based architecture that delegates authentication to a trusted identity provider (IdP) such as Microsoft Entra ID.

* Stage manager

  A stage manager is used in symphony workflow. It triggers stage provider defined in each stage and report stage output to activation status.

* Staging manager

  A staging manager is used for catalogversion synchronization and remote job distribution between symphony parent and child sites.

## Choose appropriate state providers for managers

  Most of the managers need to load state provider to store, query or delete objects in the state store. Some managers store important information to state store like symphony unified objects while some others only store cache which is tolerant to process crash. Symphony explicitly defines whether a manager requires persistent state provider to help user choose appropriate data store in their scenario.
  | manager | state provider |
  |---|---|
  | jobs manager | persistent, volatile |
  | object manager <br> ( instances manager, solutionversions manager, <br> targets manager, device manager, <br> campaignversions manager, activations manager, <br> catalogversions manager)  | persistent |
  | solutionversion manager | persistent |
  | reference manager | volatile |
  | stage manager | volatile |
  | staging manager | volatile |
  | user manager | volatile |
