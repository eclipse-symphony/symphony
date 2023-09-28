# Configuration Management
Symphony supports sophisticated configuration management capabilities for modeling and managing configurations across multiple applications in an industrial environment. These features include:

* Maintains a centralized configuration repository that can be synchronized to different physical sites.
* Creates overrides at different levels while maintaining visibility of what is overridden and where.
* Uses multiple configuration data sources at the same time with precedence. 
* Composes configurations into one bigger configuration.
* Injects configurations to application as environment variables or as file mounts.
* Injects configurations into shared components when an application is deployed.
* Configuration schema validations.

Symphony separates the concern of configuration management and configuration consumption. On one hand, IT/OT can use the sophisticated configuration management capabilities to manage configurations at multiple levels and overrides. On the other hand, developers are shielded from all complexities and  all they see are resolved configuration values injected as environment variables or as file mounts.

## Topics

* [Defining Configurations](./defining-configurations.md)
* [Serving Configurations](./serving-configurations.md)