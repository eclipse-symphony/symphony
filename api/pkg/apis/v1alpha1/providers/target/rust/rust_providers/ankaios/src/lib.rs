/*
 * Copyright (c) Microsoft Corporation and others.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

 #![cfg(all(unix))]

use std::collections::HashMap;
use std::ffi::{c_char, CStr};
use std::ptr;
use std::time::Duration;
use symphony::models::{
    ComponentAction, ComponentResultSpec, ComponentSpec, ComponentStep, DeploymentSpec,
    DeploymentStep, ProviderConfig, State, ValidationRule,
};
use symphony::ITargetProvider;
use symphony::ProviderWrapper;

use ankaios_sdk::{Ankaios, Workload};
use tracing::{debug, error};

/// Creates a new Ankaios target provider instance.
///
/// # Safety
///
/// Client code needs to make sure that the passed in pointer is valid.
#[no_mangle]
pub unsafe extern "C" fn create_provider(config_json: *const c_char) -> *mut ProviderWrapper {
    // try to configure tracing to output Events to stdout
    if tracing_subscriber::fmt::try_init().is_err() {
        // This will fail on subsequent invocations of the target provider's functions
        // because the provider is stateless and thus newly created for each and every invocation.
        // Consequently, we do not need to clutter stdout with corresponding warning
        // but can assume that the subscriber has been initialized one way or the other ...
    }
    if config_json.is_null() {
        error!("Pointer to configuration JSON string is null");
        return ptr::null_mut();
    }

    let config_str = match unsafe { CStr::from_ptr(config_json).to_str() } {
        Ok(str) => str,
        Err(err) => {
            error!("Error converting pointer to JSON string: {:?}", err);
            return ptr::null_mut();
        }
    };
    debug!("creating AnkaiosProvider using config: {}", config_str);

    let config: ProviderConfig = match serde_json::from_str(config_str) {
        Ok(cfg) => cfg,
        Err(err) => {
            error!("Error deserializing configuration JSON string: {:?}", err);
            return ptr::null_mut();
        }
    };

    let provider = match AnkaiosProvider::new(&config) {
        Ok(provider) => Box::new(provider),
        Err(e) => {
            error!("Error creating AnkaiosTargetProvider: {:?}", e);
            return ptr::null_mut();
        }
    };
    let wrapper = Box::new(ProviderWrapper { inner: provider });
    Box::into_raw(wrapper)
}

pub struct AnkaiosProvider {
    // Tokio runtime for async execution
    runtime: tokio::runtime::Runtime,
    ank: tokio::sync::Mutex<Ankaios>,
    validation_rule: ValidationRule,
}

impl AnkaiosProvider {
    fn new(_config: &ProviderConfig) -> Result<Self, Box<dyn core::error::Error>> {
        let tokio_runtime = tokio::runtime::Runtime::new()?;

        let ankaios_client = tokio_runtime
            .block_on(Ankaios::new_with_timeout(Duration::from_secs(5)))
            .inspect_err(|err| {
                error!("Failed to initialize Ankaios: {err}");
            })
            .map(tokio::sync::Mutex::new)?;

        // this might be adapted to determine the validation rule from config properties
        let validation_rule = ValidationRule::default();

        Ok(Self {
            runtime: tokio_runtime,
            ank: ankaios_client,
            validation_rule,
        })
    }

    async fn async_get(
        &self,
        _deployment: DeploymentSpec,
        references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String> {
        let complete_state = {
            // drop the guard as soon as we have retrieved the state from Ankaios
            let mut ankaios_client = self.ank.lock().await;
            ankaios_client
                .get_state(vec!["workloadStates".to_string(), "workloads".to_string()])
                .await
                .map_err(|err| {
                    error!("Failed to retrieve component state from Ankaios: {err}");
                    err.to_string()
                })?
        };

        let result_componentspecs = references
            .iter()
            .map(|step| step.component.clone())
            .filter_map(|mut spec| {
                // determine corresponding Ankaios workflow state
                if let Some(workload_state) = complete_state
                    .get_workload(&spec.name)
                    .map(|wl| wl.to_dict())
                {
                    let mut props = spec.properties.unwrap_or_default();
                    if let Some(v) = workload_state["agent"].as_str() {
                        props.insert(
                            "ankaios.agent".to_string(),
                            serde_json::Value::String(v.to_string()),
                        );
                    }
                    if let Some(v) = workload_state["runtime"].as_str() {
                        props.insert(
                            "ankaios.runtime".to_string(),
                            serde_json::Value::String(v.to_string()),
                        );
                    }
                    if let Some(v) = workload_state["restartPolicy"].as_str() {
                        props.insert(
                            "ankaios.restartPolicy".to_string(),
                            serde_json::Value::String(v.to_string()),
                        );
                    }
                    if let Some(v) = workload_state["runtimeConfig"].as_str() {
                        props.insert(
                            "ankaios.runtimeConfig".to_string(),
                            serde_json::Value::String(v.to_string()),
                        );
                    }
                    spec.properties = Some(props);
                    Some(spec)
                } else {
                    // no workload with the given component name found
                    None
                }
            })
            .collect();

        // let mut result_componentspecs = vec![];

        // // Get the workload states present in the complete state
        // let workload_states_dict = complete_state.get_workload_states().as_dict();
        // // Print the states of the workloads
        // for (agent_name, workload_states) in workload_states_dict.iter() {
        //     for (workload_name, workload_execution_states) in workload_states.iter() {
        //         for (_workload_id, workload_execution_state) in workload_execution_states.iter() {
        //             let _state = workload_execution_state.state;
        //             //if state == WorkloadStateEnum::Running {
        //             for component in references.iter() {
        //                 if component.component.name == workload_name.as_str() {
        //                     let mut ret_component = component.component.clone();
        //                     let mut properties =
        //                         ret_component.properties.clone().unwrap_or_default();
        //                     // Ankaios agent name
        //                     let agent_json_value: serde_json::Value =
        //                         serde_json::to_value(agent_name).unwrap_or(serde_json::Value::Null);
        //                     properties.insert("ankaios.agent".to_string(), agent_json_value);

        //                     if let Ok(workload) =
        //                         self.ank.get_workload(workload_name.to_owned()).await
        //                     {
        //                         let workload_properties = workload.to_dict();
        //                         // Ankaios runtime
        //                         properties.insert(
        //                             "ankaios.runtime".to_string(),
        //                             serde_json::Value::String(
        //                                 workload_properties["runtime"]
        //                                     .as_str()
        //                                     .unwrap()
        //                                     .to_string(),
        //                             ),
        //                         );
        //                         // Ankaios restart policy
        //                         properties.insert(
        //                             "ankaios.restartPolicy".to_string(),
        //                             serde_json::Value::String(
        //                                 workload_properties["restartPolicy"]
        //                                     .as_str()
        //                                     .unwrap()
        //                                     .to_string(),
        //                             ),
        //                         );
        //                         // runtimeConfig
        //                         properties.insert(
        //                             "ankaios.runtimeConfig".to_string(),
        //                             serde_json::Value::String(
        //                                 workload_properties["runtimeConfig"]
        //                                     .as_str()
        //                                     .unwrap()
        //                                     .to_string(),
        //                             ),
        //                         );
        //                     }
        //                     ret_component.properties = Some(properties);
        //                     result_componentspecs.push(ret_component);
        //                     break;
        //                 }
        //             }
        //             //}
        //         }
        //     }
        // }
        Ok(result_componentspecs)
    }

    async fn async_apply(
        &self,
        _deployment: DeploymentSpec,
        step: DeploymentStep,
    ) -> Result<HashMap<String, ComponentResultSpec>, String> {
        let mut result: HashMap<String, ComponentResultSpec> = HashMap::new();
        let mut ankaios_client = self.ank.lock().await;

        for component in step.components.iter() {
            if component.action == ComponentAction::Delete {
                // Simulate deletion
                match ankaios_client.delete_workload(component.component.name.clone()).await {
                    Ok(_) => {
                        let component_result = ComponentResultSpec {
                            status: State::OK,
                            message: "Component deleted successfully".to_string(),
                        };
                        result.insert(component.component.name.clone(), component_result);
                    }
                    Err(e) => {
                        let component_result = ComponentResultSpec {
                            status: State::InternalError,
                            message: format!("Failed to delete workload: {:?}", e),
                        };
                        result.insert(component.component.name.clone(), component_result);
                    }
                }
            } else if component.action == ComponentAction::Update {
                let workload = Workload::builder()
                    .workload_name(component.component.name.clone())
                    .agent_name(
                        component
                            .component
                            .properties
                            .as_ref()
                            .and_then(|props| props.get("ankaios.agent")?.as_str())
                            .unwrap_or("agent_A"),
                    )
                    .runtime(
                        component
                            .component
                            .properties
                            .as_ref()
                            .and_then(|props| props.get("ankaios.runtime")?.as_str())
                            .unwrap_or("poaman"),
                    )
                    .restart_policy(
                        component
                            .component
                            .properties
                            .as_ref()
                            .and_then(|props| props.get("ankaios.restartPolicy")?.as_str())
                            .unwrap_or("NEVER"),
                    )
                    .runtime_config(
                        component
                            .component
                            .properties
                            .as_ref()
                            .and_then(|props| props.get("ankaios.runtimeConfig")?.as_str())
                            .unwrap_or(""),
                    )
                    .build()
                    .unwrap();
                match ankaios_client.apply_workload(workload).await {
                    Ok(_) => {
                        let component_result = ComponentResultSpec {
                            status: State::OK,
                            message: "Component applied successfully".to_string(),
                        };
                        result.insert(component.component.name.clone(), component_result);
                    }
                    Err(e) => {
                        let component_result = ComponentResultSpec {
                            status: State::InternalError,
                            message: format!("Failed to apply workload: {:?}", e),
                        };
                        result.insert(component.component.name.clone(), component_result);
                    }
                }
            }
        }
        Ok(result) // Simulated async operation
    }
}

impl ITargetProvider for AnkaiosProvider {
    fn get_validation_rule(&self) -> Result<ValidationRule, String> {
        Ok(self.validation_rule.clone())
    }

    fn get(
        &self,
        deployment: DeploymentSpec,
        references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String> {
        self.runtime
            .block_on(self.async_get(deployment, references))
    }

    fn apply(
        &self,
        deployment: DeploymentSpec,
        step: DeploymentStep,
        is_dry_run: bool,
    ) -> Result<HashMap<String, ComponentResultSpec>, String> {
        if is_dry_run {
            println!("Dry run is enabled, skipping actual apply");
            return Ok(HashMap::new());
        }

        self.runtime.block_on(self.async_apply(deployment, step))
    }
}

// fn main() {
//     // Create provider using `create_provider()`
//     let provider_ptr = unsafe { create_provider() };

//     if provider_ptr.is_null() {
//         eprintln!("Failed to create provider.");
//         return;
//     }

//     // Convert raw pointer back into a `Box` and extract the provider reference
//     let provider_wrapper = unsafe { Box::from_raw(provider_ptr) };
//     let provider = &provider_wrapper.inner;

//     // Initialize provider
//     if let Err(e) = provider.init(ProviderConfig::default()) {
//         eprintln!("Initialization failed: {}", e);
//         return;
//     }

//     eprintln!("Provider initialized successfully.");

//     let deployment = DeploymentSpec::empty();

//     // Create a mock `references` list for `get()`
//     let references = vec![ComponentStep {
//         action: ComponentAction::Update,
//         component: ComponentSpec {
//             name: "symphony".to_string(),
//             component_type: Some("workload".to_string()),
//             properties: Some(HashMap::new()),
//             metadata: Some(HashMap::new()),
//             dependencies: Some(vec![]),
//             constraints: None,
//             parameters: Some(HashMap::new()),
//             routes: Some(vec![]),
//             sidecars: Some(vec![]),
//             skills: Some(vec![]),
//         },
//     }];

//     match provider.get(deployment.clone(), references) {
//         Ok(components) => {
//             println!("Get result: {:?}", components);
//             if components.is_empty() {
//                 println!("No components retrieved.");
//             }
//         }
//         Err(e) => eprintln!("Failed to get components: {}", e),
//     }

//     let step = DeploymentStep {
//         is_first: false,
//         role: "ankaios_workload".to_string(),
//         target: Some("my-target".to_string()),
//         components: vec![ComponentStep {
//             action: ComponentAction::Delete,
//             component: ComponentSpec {
//                 name: "yyds".to_string(),
//                 component_type: Some("ankaios_workload".to_string()),
//                 properties: Some(HashMap::from([
//                     ("ankaios.agent".to_string(), Value::String("agent_A".to_string())),
//                     ("ankaios.runtime".to_string(), Value::String("podman".to_string())),
//                     ("ankaios.restartPolicy".to_string(), Value::String("NEVER".to_string())),
//                     ("ankaios.runtimeConfig".to_string(), Value::String(
//                         "image: localhost/latest\ncommandOptions: [\"-p\", \"8082:8082\", \"-e\", \"CONFIG=/symphony-api-no-k8s.json\", \"-e\", \"USE_SERVICE_ACCOUNT_TOKENS=false\", \"-e\", \"SYMPHONY_API_URL=http://localhost:8082/v1alpha2/\"]".to_string()
//                     )),
//                 ])),
//                 metadata: Some(HashMap::new()),
//                 dependencies: Some(vec![]),
//                 constraints: None,
//                 parameters: Some(HashMap::new()),
//                 routes: Some(vec![]),
//                 sidecars: Some(vec![]),
//                 skills: Some(vec![]),
//             },
//         }],
//     };

//     match provider.apply(deployment, step, false) {
//         Ok(components) => {
//             println!("Apply result: {:?}", components);
//         }
//         Err(e) => eprintln!("Failed to get components: {}", e),
//     }

//     // Prevent double free, manually leak the box to avoid deallocation issues
//     Box::leak(provider_wrapper);
// }
