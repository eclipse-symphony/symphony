# Writing a Target Provider in Rust
## High-Level Workflow
To create a new provider:

1. Create a New Folder:

    Under the `api/pkg/apis/v1alpha1/providers/rust/rust_providers` directory, create a new library project using `cargo`:
    ```bash
    cargo new myprovider --lib
    ```
2. Update the `Cargo.toml` File:

    In your project folder, modify the `Cargo.toml` file to set the library type to a dynamic C library and to add the necessary dependencies:

    ```toml
    [lib]
    crate-type = ["cdylib"]

    [dependencies]
    rust_binding = { path = "../../rust_binding" }
    serde = "1.0"
    serde_json = "1.0"
    ```

3. Add the Provider to the Cargo Workspace:

    Update the `Cargo.toml` file in the `api/pkg/apis/v1alpha1/providers/rust` directory (two levels up) to include your new provider in the Cargo workspace:

    ```toml
    [workspace]
    members = [
        "rust_binding",        
        "rust_providers/myprovider"
    ]
    ```

4. Start Implementing the Provider:

    To begin implementing your provider, copy the source code from `api/pkg/apis/v1alpha1/providers/rust/rust_providers/mock/src/lib.rs` into your lib.rs file, and then rename MockProvider to your provider's name.

5. Build the Workspace:

    Build the workspace to ensure everything compiles correctly:

    ```bash
    # Navigate to the api/pkg/apis/v1alpha1/providers/rust folder
    cargo build --release
    ```
    You should see a myprovider.so library in the target/release folder. This file will be deployed to your Symphony container/pod/process.

6. Update the Provider Methods:
    Next, update the `get_validation_rule()`, `get()`, and `apply()` methods to interact with the toolchain you intend to support.

    > **Note:** You typically do not need to modify the `init()` method. However, if initialization work is required, you can access the provider configuration through the ProviderConfig parameter.

## Implement `get_validation_rule()` method
Symphony's component specification is an open schema that includes a properties collection and a metadata collection. Providers in Symphony have the flexibility to define validation rules for the key-value pairs within these collections. Specifically, a provider can specify:

* **Required Properties or Metadata:** Define which properties or metadata entries are mandatory.
* **Optional Properties or Metadata:** Identify which properties or metadata entries are optional.
* **Change Detection Properties or Metadata:** Specify which properties or metadata should be monitored for changes. Symphony uses this information to determine if a component requires an update, focusing only on the change detection collections while ignoring others.
* **Required Component Type:** Define a required component type, which is simply a string name. By specifying this, a provider ensures that a component is explicitly tagged with the expected type, preventing users from accidentally assigning a component to an incompatible provider.
* **Support for Sidecars:** Indicate whether the provider supports sidecar containers.
* **Instance Isolation:** Specify if the provider supports instance isolation, meaning multiple instances can be deployed to the same target without conflicts.
* **Scope Isolation:** Indicate if the provider supports scope isolation, which, in the context of Kubernetes, refers to namespace support.
* **Sidecar Validation Rules:** Define validation rules for sidecars, similar to those for components.

## Implement `get()` method
The `get()` method is responsible for returning the current state on the target, and it is expected to return an array of `ComponentSpec` objects. Since Symphony does not require providers to be stateful, it passes the current deployment and a list of components of interest to the get() method. These structures can be used to help identify which components should be probed.

## Implement `apply()` method
The `apply()` method is responsible for applying the new desired state. When this method is invoked, Symphony passes in the current deployment specification, the current `DeploymentStep` (as deployments may involve multiple steps), and a flag indicating whether it is a dry-run. The `apply()` method processes the components in the step and returns the operation results as a map of `ComponentResultSpec`.

A deployment step may include one or more `ComponentStep` structures. Each `ComponentStep` can represent either an `Update` or a `Delete` action, and the provider must update or remove the component accordingly.

## Building the Provider
To ensure that dependencies are correctly resolved and the build environment is consistent, we provide a Rust development container that you can use to build your Rust project. The Dockerfile for this development container is located in the `api/pkg/apis/v1alpha1/providers/rust/` folder.

To build the Cargo workspace that includes your provider project, follow these steps:

1. Run the Build Container:

    Use the following command to build the workspace:
    ```bash
    docker run --rm -v <path-to-cargo-workspace>:/workspace <your-dev-container>
    ```
    Replace <path-to-cargo-workspace> with the actual path to your Cargo workspace on your local machine.
    Replace <your-dev-container> with the name of the development container image.
2. Retrieve the Built Library:

    After the build process completes, the required .so file will be located in the target/release directory within your Cargo workspace folder.
    You can find it at:
    ```bash
    <path-to-cargo-workspace>/target/release/<provider library>.so
    ```
By following these steps, you can ensure that your Rust project is built in a controlled environment, reducing the risk of dependency issues and ensuring compatibility with the deployment environment.


## Deploying the Provider
Steps to upload your provider library (the `.so` file) to `symphony-api` pod on Kubernetes:
1. Build the above Cargo wokspace:
    ```bash
    # Navigate to the api/pkg/apis/v1alpha1/providers/rust folder
    cargo build --release
    ```
2. To avoid loading arbitary code, when you load a Rust provider, Symphony also requries you to provide a match hash of the provider. To generate the hash, you can use the `sha256sum` command:
    ```bash
    sha256sum ./target/release/<your library file>
    ```
    Copy the returned hashcode. You'll need to enter this to your provider configuration.
1. Get `symphony-api` pod:
    ```bash
    kubectl get pod
    ```
2. Copy your library file:
    ```bash
    kubectl cp ./target/release/<your so file> symphony-api-<...>:/etc/symphony-api/extensions/
    ```

    > **NOTE**: If you've insalled Symphony Helm chart with `extensions` enabled, you'll get this `/etc/symphony-api/extensions/` folder automatically mounted. 
## Test with the Provider

1. Define a Target object
    ```yaml
    apiVersion: fabric.symphony/v1
    kind: Target
    metadata:
        name: rust-test-target
    spec:  
    components:
    - name: mock
      properties:
        foo: bar
    forceRedeploy: true
    topologies:
    - bindings:
      - role: instance
        provider: providers.target.rust
        config:
          name: "rust-lib"
          libFile: "/etc/symphony-api/extensions/<library file name>"
          libHash: "<library hash code>"
      ```
2. Apply the target:
    ```bash
    kubectl apply -f <the above target.yaml file>
    ```
3. Observe the deployment status:
    ```bash
    kubectl get target -w
    ```
    