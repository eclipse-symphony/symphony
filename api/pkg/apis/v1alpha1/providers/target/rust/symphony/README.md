# Eclipse Symphony Target Provider Rust Binding

This crate provides the [Eclipse Symphony](https://github.com/eclipse-symphony/symphony) Target provider Rust binding, which allows a custom Target provider to be written in Rust.

Symphony is a toolchain orchestrator that aims to provide consistent management experience across multiple toolchains. A key capability of Symphony is state seeking, where a system's current is brought towards a new desired state. Symphony allows different toolchains to join the state seeking process through a Target Provider trait.

```rust
pub trait ITargetProvider: Send + Sync {
    fn get_validation_rule(&self) -> Result<ValidationRule, String>;
    fn get(&self, deployment: DeploymentSpec, references: Vec<ComponentStep>) -> Result<Vec<ComponentSpec>, String>;
    fn apply(&self, deployment: DeploymentSpec, step: DeploymentStep, is_dry_run: bool) -> Result<HashMap<String, ComponentResultSpec>, String>; 
 }
 ```
 * The `get()` method returns the current state of a system.
 * The `apply()` method applies the new desired state.
 * The `get_validation_rule()` allows a provider to define what properties are expected in the incoming state spec, and what properties to be used for change detection.

 ## Current Rust based Target Providers

 | Provider  | Info |
 |-----------|------|
 | ankaios   | An [Eclipse Ankaios](https://github.com/eclipse-ankaios/ankaios) provider |
 | mock      | A mock provider for testing purposes |
 | uprotocol | A provider that allows deployment Targets to be implemented using [Eclipse uProtocol](https://uprotocol.org) |
