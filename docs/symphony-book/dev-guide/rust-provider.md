# Implementing a Rust Target Provider
In this walkthrough, we’ll use implementing a Rust provider for [Eclipse Ankaios](https://eclipse-ankaios.github.io/ankaios/0.2/) as an example to walk you through the steps of creating a new Symphony [Target provider](../providers/target-providers/target_provider.md) using Rust. 

## 1. Deciding on the integration point 

Ankaios aims to bring cloud-native practices to automated HPCs while meeting the safety and real-time requirements of the automotive industry. It consists of an Ankaios server that manages multiple Ankaios agents. As a toolchain orchestrator, Symphony does not interfere with the internal workings of Ankaios components. Instead, Symphony treats the entire Ankaios system as a Target that provides in-vehicle orchestration. The Ankaios Target can be annotated with feature flags, enabling Symphony—acting as a fleet management layer in this case—to make fleet-level decisions based on these flags. For example, a Solution component can request installation on an Ankaios-enabled Target.

## 2. Setting up Ankaios test environment

When you develop a provider, it's important to have a high-fidelity test environment so that you can test your provider functionalities locally in isolation before conduting more complex integration tests. 

Install Ankaios following the [official instruction](https://eclipse-ankaios.github.io/ankaios/latest/usage/installation).
```bash
# Install without mTLS
curl -sfL https://github.com/eclipse-ankaios/ankaios/releases/latest/download/install.sh | bash -
# Start Ankaios server
sudo systemctl start ank-server
# Start an Ankaios agent
sudo systemctl start ank-agent
# Check server status
sudo systemctl status ank-server
# You sould see the server is active, and has received a AgentHello message from 'agent-A'
```
## 3. Preparing your provider project
>**NOTE:** We assume you've already have Rust and Cargo install.

**IMPORTANT:** If you are contributing your provider source code to Symphony, you should fork Symphony repository and put your provider project under the `api/pkg/apis/v1alpha1/providers/target/rust/rust_providers` folder. You should also modify the `api/pkg/apis/v1alpha1/providers/target/rust/Cargo.toml` file to include your project into the workspace. This will join your project into our automated build and release pipeline. For example:
```toml
[workspace]
members = [
    "symphony",
    "rust_providers/mock",
    "rust_providers/ankaios"
]
```
1. Create a new Rust project using Cargo:
    ```bash
    cargo new ankaios --lib
    cd ankaios
    ```

2. Modify your `Cargo.toml` file to add a reference to Symphony crate.
    ```toml
    [dependencies]
    symphony = "0.1.0"
    ```
3. Replace the content of your `lib.rs` with this code:
    ```rust
     extern crate symphony;

    use symphony::models::{
        ProviderConfig, ValidationRule, DeploymentSpec, ComponentStep, ComponentSpec,
        DeploymentStep, ComponentResultSpec,
        ComponentValidationRule
    };
    use symphony::ITargetProvider;
    use symphony::ProviderWrapper;
    use std::collections::HashMap;
    
    pub struct AnkaiosProvider;

    #[no_mangle]
    pub extern "C" fn create_provider() -> *mut ProviderWrapper  {
        let provider: Box<dyn ITargetProvider> = Box::new(AnkaiosProvider {});
        let wrapper = Box::new(ProviderWrapper { inner: provider });
        Box::into_raw(wrapper)
    }

    impl ITargetProvider for AnkaiosProvider {
        fn init(&self, _config: ProviderConfig) -> Result<(), String> {
            Ok(())
        }
        fn get_validation_rule(&self) -> Result<ValidationRule, String> {
            let validation_rule = ValidationRule {
                required_component_type: "".to_string(),
                component_validation_rule: ComponentValidationRule {
                    required_component_type: "".to_string(),
                    change_detection_properties: vec![],
                    change_detection_metadata: vec![],
                    required_properties: vec![],
                    optional_properties: vec![],
                    required_metadata: vec![],
                    optional_metadata: vec![],
                },
                sidecar_validation_rule: ComponentValidationRule {
                    required_component_type: "".to_string(),
                    change_detection_properties: vec![],
                    change_detection_metadata: vec![],
                    required_properties: vec![],
                    optional_properties: vec![],
                    required_metadata: vec![],
                    optional_metadata: vec![],
                },
                allow_sidecar: true,
                scope_isolation: true,
                instance_isolation: true,
            };
        
            Ok(validation_rule)
        }
        fn get(&self, _deployment: DeploymentSpec, _references: Vec<ComponentStep>) -> Result<Vec<ComponentSpec>, String> {
            Ok(vec![])
        }
        fn apply(
            &self,
            _deployment: DeploymentSpec,
            _step: DeploymentStep,
            _is_dry_run: bool,
        ) -> Result<HashMap<String, ComponentResultSpec>, String> {
            Ok(HashMap::new())
        }
    }
    ```
4. Build your project to make sure everything is in place:
    ```bash
    cargo build
    ```
👍 Great! Now you are ready to implement your provider!

## 4. Implementing the `get()` method
Symphony periodically calls the `get()` method to retrieve the current system state. Since Symphony does not require a provider to maintain any state, it specifies the relevant deployment (via the `deployment: DeploymentSpec` parameter) and components (via the references: `Vec<ComponentStep> parameter`) it is interested in. Typically, you should iterate over the components in the references parameter and construct a `Vec<ComponentSpec>` array as the return value.
### 4.1 Decide on what constitutes a `Component`
A Symphony `Solution` consists of one or more `Components`. When a system integrates with Symphony, it can choose to represent its entire (relevant) system state as a single Symphony `Component` or expose a more granular construct.

For Ankaios, you can either treat the entire Ankaios system state as a single `Component` or represent each Ankaios workload as a separate `Component`.

The choice of granularity depends on your specific use case. However, opting for finer granularity provides opportunities to leverage more Symphony features. For example, by treating each Ankaios workload as a separate `Component`, you can use Symphony's component dependency feature to ensure workloads are provisioned in the correct order.

In this walkthrough, we'll treat an Ankaios workload as a `Component`.

A Symphony `Component` consists of a name, a type, and a key-value property bag.

* Component **name**: A unique identifier within a Solution.
* Component **type**: An arbitrary string. However, for a specific system, it's best to use a consistent type string, such as `ankaios-workload`. Symphony uses this type string to identify the corresponding `TargetProvider` that claims to handle that component type.
* Component **properties**: A collection of key-value pairs that can store any relevant information. However, you must ensure that these properties can be reliably reconstructed when requested by Symphony. Symphony uses these properties—along with validation rules (covered in the next section)—to determine whether an update is required. If the property values are unstable, you may trigger constant reconciliations.

### 4.2 Deploying an Ankaios workload
We'll manually deploy an Ankaios workload for testing purposes. In this walkthrough, we'll run a Ngix server on Podman:
```bash
ank -k run workload \
nginx \
--runtime podman \
--agent agent_A \
--config 'image: docker.io/nginx:latest
commandOptions: ["-p", "8087:80"]'
```
### 4.3 Add Ankaios references
At the time of writing, Ankaios Rust SDK crate hasn't been published. You'll need to add a reference through their Git repository. Please consult Ankaios documents for updates.
Modify your `Cargo.toml` file to include a reference to `ankaios-sdk`:
```bash
[dependencies]
symphony = "0.1.0"
ankaios-sdk = { version = "0.5.0-rc1", git = "https://github.com/GabyUnalaq/ank-sdk-rust.git", branch = "first-version" }
```
