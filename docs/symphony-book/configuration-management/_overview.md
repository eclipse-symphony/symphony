# Configuration management

Symphony supports sophisticated configuration management capabilities for modeling and managing configurations across multiple applications in large-scale industrial environments. Symphony can:

* Maintain a centralized configuration repository that can be synchronized to different physical sites.
* Contextualize application configurations based on actual deployment environments.
* Create overrides at different levels while maintaining visibility of what is overridden and where.
* Use multiple configuration data sources at the same time with precedence.
* Dynamically assemble configurations into one unified configuration.
* Inject configurations into application as environment variables or as file mounts.
* Inject configurations into shared components when an application is deployed.
* Validate configuration schema.

Symphony follows several key principles in its configuration management design, including:

* **Separation of concerns**	

 Symphony looks at different personas and their roles in the configuration management and provides targeted features that enable them to work together while without being exposed to unnecessary details. For example, Symphony separates configuration management and configuration serving. IT/OT can use elaborated configuration models to handle large-scale management challenges; while developers get served an rendered configuration files without being exposed to any of the complexities. 

* **Breaking away from a static file-based mindset**

Configurations are often referred as “Configuration files”, which often leads to the wrong impression that configuration management problems are configuration FILE management problems. A file-based system has its inherited limitations in configuration modeling. This leads to lots of duplications and inconsistencies not only in static modeling but also during updates. Symphony provides a rich configuration modeling language that allows configurations to be modeled in a more object-oriented manner with constructs such as decomposition, inheritance, and overrides. 

* **Contextualization** 

Configuration objects in Symphony are assembled at the last moment during deployment. This allows Symphony to dynamically inject contextual information based on the actual deployment topology, drastically reducing the required number of configuration objects. Such contextualization is also critical for day-N operations where new machine configurations are managed by different personas other than who manage applications. With late-assembly, Symphony is able to pick up those changes and make sure the application’s configuration is up to date. 

* **Configuration is not an isolated problem**

Configuration management can’t be done in a vacuum. Instead, configurations are parts of the system desired state, and they should comply to all enterprise policies, workflows (like approval and maintenance windows), versioning and auditing practices. 

* **Meeting customers where they are**

Symphony doesn’t force customers to be modified how their applications behave just for the sake of Symphony. And Symphony aims to protect customer’s existing investments in configuration management as well. This is why Symphony is designed to allow configuration objects to refer to external data sources or services. 

## Fundamentals
 
* [Creating a configuration object](./define-configurations.md)
* [Serving a configuration object](./serve-configurations.md)
* [Late assembly](./late-assembly.md)
* [Secret management](./secret-management.md)

## Configuration Modeling
* [Componentization](./componentization.md)
* [Inheritance](./inheritance.md)
* [Contextualization and calculated configurations](./contextualization.md)

## At-scale management
* [Schema reinforcement](./schema-reinforcement.md)
* [Versioning](./versioning.md)
* [RBAC](./rbac.md)
* [Multi-site distribution](./multi-site-distribution.md)
* [External configuration sources](./external-sources.md)
* [Approvals, auditability and other workflows](./approval-and-workflows.md)

## Related topics

* [Define configurations](./define-configurations.md)
* [Serve configurations](./serve-configurations.md)

