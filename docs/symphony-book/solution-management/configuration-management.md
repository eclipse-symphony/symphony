# Configuration management

Symphony supports sophisticated configuration management capabilities for modeling and managing configurations across multiple applications in an industrial environment. Symphony can:

* Maintain a centralized configuration repository that can be synchronized to different physical sites.
* Create overrides at different levels while maintaining visibility of what is overridden and where.
* Use multiple configuration data sources at the same time with precedence.
* Compose configurations into one bigger configuration.
* Inject configurations into application as environment variables or as file mounts.
* Inject configurations into shared components when an application is deployed.
* Validate configuration schema.

Symphony separates the concern of configuration management and configuration consumption. On one hand, IT/OT can use the sophisticated configuration management capabilities to manage configurations at multiple levels and overrides. On the other hand, developers are shielded from all complexities and all they see are resolved configuration values injected as environment variables or as file mounts.

Configurations can be injected into solution components as environment variables or file mounts (on Kubernetes). An application can also query Symphony REST API to retrieve configurations.

## Related topics

* [Define configurations](./defining-configurations.md)
* [Serve configurations](./serving-configurations.md)