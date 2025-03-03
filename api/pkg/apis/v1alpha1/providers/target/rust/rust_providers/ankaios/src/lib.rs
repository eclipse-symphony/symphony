/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

 extern crate symphony;

 use symphony::models::{
     ProviderConfig, ValidationRule, DeploymentSpec, ComponentStep, ComponentSpec,
     DeploymentStep, ComponentResultSpec,
     ComponentValidationRule
 };
 use symphony::ITargetProvider;
 use symphony::ProviderWrapper;
 use std::collections::HashMap;
 
 use tokio::runtime::Runtime;
 use ankaios_sdk::{Ankaios};
 use tokio::time::Duration;
 use std::sync::{Arc, Mutex};

 pub struct AnkaiosProvider {
    runtime: Runtime,            // Tokio runtime for async execution
    ank: Arc<Mutex<Option<Ankaios>>>, // Ankaios instance (Mutex for safe shared access)
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
        let mut ank = self.ank.lock().unwrap();
        if ank.is_none() {
            // Initialize Ankaios inside runtime
            *ank = Some(self.runtime.block_on(Ankaios::new()).unwrap());
        }
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

    fn get(
        &self,
        _deployment: DeploymentSpec,
        _references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String> {
        self.runtime.block_on(self.async_get(_deployment, _references))
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

impl AnkaiosProvider {
    /// This is an internal async function, **not part of the trait**.
    async fn async_get(
        &self,
        _deployment: DeploymentSpec,
        _references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String> {
        let mut ank_guard = self.ank.lock().unwrap(); // Acquire a mutable lock

        if let Some(ank) = &mut *ank_guard { // Get a mutable reference
            if let Ok(complete_state) = ank
                .get_state(
                    Some(vec!["workloadStates".to_string()]),
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
                            println!(
                                "Workload {} on agent {} has the state {:?}",
                                workload_name.as_str().unwrap(),
                                agent_name.as_str().unwrap(),
                                workload_state.get("state").unwrap().as_str().unwrap().to_string()
                            );
                        }
                    }
                }
            }
        } else {
            return Err("Ankaios is not initialized".to_string());
        }
        Ok(vec![]) // Simulated async operation
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use symphony::models::{DeploymentSpec};

    #[test]
    fn test_get() {
        let provider = AnkaiosProvider {
            runtime: tokio::runtime::Runtime::new().unwrap(),
            ank: std::sync::Arc::new(std::sync::Mutex::new(None)),
        };

        // Initialize provider
        provider.init(Default::default()).expect("Failed to initialize provider");

        let deployment = DeploymentSpec::empty();
        let references = vec![];

        let result = provider.get(deployment, references);
        assert!(result.is_ok(), "Expected Ok result, but got {:?}", result);
    }
}