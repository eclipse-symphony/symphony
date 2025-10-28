/*
 * Copyright (c) Microsoft Corporation and others.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

use std::collections::HashMap;
use std::ffi::{c_char, CStr};
use symphony::models::{
    ComponentResultSpec, ComponentSpec, ComponentStep, DeploymentSpec, DeploymentStep, RouteSpec,
    SidecarSpec, State, ValidationRule,
};
use symphony::{ITargetProvider, ProviderWrapper};

pub struct MockProvider;

/// Creates a new provider for configuration data.
///
/// # Safety
///
/// Client code must make sure that the provided pointer is valid.
#[no_mangle]
pub unsafe extern "C" fn create_provider(config_json: *const c_char) -> *mut ProviderWrapper {
    if !config_json.is_null() {
        unsafe {
            if let Ok(config_str) = CStr::from_ptr(config_json).to_str() {
                println!("creating provider for configuration: {}", config_str);
            }
        }
    }
    let provider = Box::new(MockProvider {});
    let wrapper = Box::new(ProviderWrapper { inner: provider });
    Box::into_raw(wrapper)
}

impl ITargetProvider for MockProvider {
    fn get_validation_rule(&self) -> Result<ValidationRule, String> {
        println!("MOCK RUST PROVIDER: ------ get_validation_rule()");
        Ok(ValidationRule::default())
    }

    fn get(
        &self,
        _deployment: DeploymentSpec,
        _references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String> {
        println!("MOCK RUST PROVIDER: ------ get()");
        let component_spec = ComponentSpec {
            name: "example_component".to_string(),
            component_type: Some("example_type".to_string()),
            metadata: Some(HashMap::from([(
                "example_metadata_key".to_string(),
                "example_metadata_value".to_string(),
            )])),
            properties: Some(HashMap::from([(
                "example_property_key".to_string(),
                serde_json::json!("example_property_value"),
            )])),
            parameters: Some(HashMap::from([(
                "example_parameter_key".to_string(),
                "example_parameter_value".to_string(),
            )])),
            routes: Some(vec![RouteSpec {
                route: "example_route".to_string(),
                route_type: "example_type".to_string(),
                properties: Some(HashMap::from([(
                    "example_route_property_key".to_string(),
                    "example_route_property_value".to_string(),
                )])),
                filters: Some(Vec::new()),
            }]),
            constraints: Some("example_constraint".to_string()),
            dependencies: Some(Vec::new()),
            skills: Some(Vec::new()),
            sidecars: Some(vec![SidecarSpec {
                name: Some("example_sidecar".to_string()),
                sidecar_type: Some("example_type".to_string()),
                properties: Some(HashMap::from([(
                    "example_sidecar_property_key".to_string(),
                    serde_json::json!("example_sidecar_property_value"),
                )])),
            }]),
        };

        Ok(vec![component_spec])
    }

    fn apply(
        &self,
        _deployment: DeploymentSpec,
        step: DeploymentStep,
        _is_dry_run: bool,
    ) -> Result<HashMap<String, ComponentResultSpec>, String> {
        println!("MOCK RUST PROVIDER: ------ apply()");
        let mut result_map = HashMap::new();

        for component_step in step.components {
            let result_spec = ComponentResultSpec {
                status: State::OK,
                message: format!(
                    "Component {} applied successfully",
                    component_step.component.name
                ),
            };
            result_map.insert(component_step.component.name.clone(), result_spec);
        }

        Ok(result_map)
    }
}
