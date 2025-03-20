/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

 extern crate symphony;

 use symphony::models::{
     ProviderConfig, ValidationRule, DeploymentSpec, ComponentStep, ComponentSpec,
     DeploymentStep, ComponentResultSpec, State, ComponentAction,
 };
 use symphony::ITargetProvider;
 use symphony::ProviderWrapper;
 use std::collections::HashMap;
 
 use tokio::runtime::Runtime;
 use ankaios_sdk::{Ankaios, Workload};
 use tokio::time::Duration;
 use std::sync::{Arc};
 use tokio::sync::Mutex;
 
 pub struct AnkaiosProvider {
    runtime: Runtime,            // Tokio runtime for async execution
    ank: Arc<Mutex<Option<Ankaios>>>,
}

 #[no_mangle]
 pub extern "C" fn create_provider() -> *mut ProviderWrapper  {
    let provider: Box<dyn ITargetProvider> = Box::new(AnkaiosProvider {
        runtime: Runtime::new().unwrap(), // Initialize runtime
        ank: Arc::new(Mutex::new(None)),  // Lazy initialization
    });
     let wrapper = Box::new(ProviderWrapper { inner: provider });
     Box::into_raw(wrapper)
 }

 impl ITargetProvider for AnkaiosProvider {
    fn init(&self, _config: ProviderConfig) -> Result<(), String> {        
        let ank_clone = Arc::clone(&self.ank);
        let _  = self.runtime.block_on(async {
            let needs_init = {
                let ank_guard = ank_clone.lock().await;
                ank_guard.is_none() // Check if initialization is needed
            }; 
            if needs_init {
                match Ankaios::new().await {
                    Ok(ankaios_instance) => {
                        let mut ank_guard = ank_clone.lock().await;
                        *ank_guard = Some(ankaios_instance);
                        return Ok(());
                    }
                    Err(e) => {
                        return Err(format!("Failed to initialize Ankaios: {:?}", e));
                    }
                }                
            }
            Ok(())
        });
           
        Ok(())
    }
    fn get_validation_rule(&self) -> Result<ValidationRule, String> {
        Ok(ValidationRule::new())
    }

    fn get(
        &self,
        _deployment: DeploymentSpec,
        _references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String> {
        let runtime_handle = self.runtime.handle().clone();
        let ank_clone = Arc::clone(&self.ank);

        let result = runtime_handle.block_on(async {
            let mut ank_guard = ank_clone.lock().await;
            AnkaiosProvider::async_get(&mut ank_guard, _deployment, _references).await
        });

        result
    }
    fn apply(
        &self,
        deployment: DeploymentSpec,
        step: DeploymentStep,
        is_dry_run: bool,
    ) -> Result<HashMap<String, ComponentResultSpec>, String> {
        let runtime_handle = self.runtime.handle().clone();
        let ank_clone = Arc::clone(&self.ank);

        let result = runtime_handle.block_on(async {
            let mut ank_guard = ank_clone.lock().await;
            AnkaiosProvider::async_apply(&mut ank_guard, deployment, step, is_dry_run).await
        });

        result
    }
}

impl AnkaiosProvider {
    /// This is an internal async function, **not part of the trait**.
    async fn async_get(
        ank_guard: &mut Option<Ankaios>, 
        _deployment: DeploymentSpec,
        references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String> {
        
        let mut result_componentspecs: Vec<ComponentSpec> = vec![];

        if let Some(ank) = &mut *ank_guard { // Get a mutable reference
            if let Ok(complete_state) = ank
                .get_state(
                    Some(vec!["workloadStates".to_string(), "workloads".to_string()]),
                    Some(Duration::from_secs(5)),
                )
                .await
            {
                // Get the workload states present in the complete state
                let workload_states_dict = complete_state.get_workload_states().get_as_dict();                
                // Print the states of the workloads
                for (agent_name, workload_states) in workload_states_dict.iter() {
                    for (workload_name, workload_states) in workload_states.as_mapping().unwrap().iter() {
                        for (_workload_id, workload_state) in workload_states.as_mapping().unwrap().iter() {
                           
                            let _state = workload_state.get("state").unwrap().as_str().unwrap();
                            //if state == "running" {
                                for component in references.iter() {
                                    if component.component.name == workload_name.as_str().unwrap() {
                                        let mut ret_component = component.component.clone();
                                        let mut properties = ret_component.properties.clone().unwrap_or_default();
                                        // Ankaios agent name
                                        let agent_json_value: serde_json::Value = serde_json::to_value(agent_name).unwrap_or(serde_json::Value::Null);
                                        properties.insert("ankaios.agent".to_string(), agent_json_value);
                                        
                                        if let Some(workload_name_str) = workload_name.as_str() {                                         
                                            if let Ok(workload) = ank.get_workload(workload_name_str.to_owned(), Some(Duration::from_secs(5))).await {                                            
                                                let workload_properties = workload.to_dict();            
                                                // Ankaios runtime
                                                properties.insert("ankaios.runtime".to_string(), serde_json::Value::String(workload_properties["runtime"].as_str().unwrap().to_string()));
                                                // Ankasios restart policy
                                                properties.insert("ankaios.restartPolicy".to_string(), serde_json::Value::String(workload_properties["restartPolicy"].as_str().unwrap().to_string()));
                                                // runtimeConfig
                                                properties.insert("ankaios.runtimeConfig".to_string(), serde_json::Value::String(workload_properties["runtimeConfig"].as_str().unwrap().to_string()));
                                            }
                                        }                                         
                                        ret_component.properties = Some(properties);
                                        result_componentspecs.push(ret_component);
                                        break;
                                    }
                                }
                            //}
                        }
                    }
                }
            }
        } else {
            return Err("Failed to acquire lock".to_string());
        }
        Ok(result_componentspecs) // Simulated async operation
    }
    async fn async_apply(
        ank_guard: &mut Option<Ankaios>, 
        _deployment: DeploymentSpec,
        step: DeploymentStep,
        is_dry_run: bool,
    ) -> Result<HashMap<String, ComponentResultSpec>, String> {
        
        let mut result: HashMap<String, ComponentResultSpec> = HashMap::new();

        if let Some(ank) = &mut *ank_guard {
            if is_dry_run {
                println!("Dry run is enabled, skipping actual apply");
                return Ok(result);
            }

            for component in step.components.iter() {
                if component.action == ComponentAction::Delete {
                    // Simulate deletion
                    match ank.delete_workload(component.component.name.clone(), None).await {
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
                        .agent_name(component.component.properties.as_ref().and_then(|props| props.get("ankaios.agent")?.as_str()).unwrap_or("agent_A"))
                        .runtime(component.component.properties.as_ref().and_then(|props| props.get("ankaios.runtime")?.as_str()).unwrap_or("poaman"))
                        .restart_policy(component.component.properties.as_ref().and_then(|props| props.get("ankaios.restartPolicy")?.as_str()).unwrap_or("NEVER"))
                        .runtime_config(component.component.properties.as_ref().and_then(|props| props.get("ankaios.runtimeConfig")?.as_str()).unwrap_or(""))
                        .build()
                        .unwrap();
                    match ank.apply_workload(workload, None).await {
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
        }  else {
            return Err("Failed to acquire lock".to_string());
        }
        Ok(result) // Simulated async operation
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
