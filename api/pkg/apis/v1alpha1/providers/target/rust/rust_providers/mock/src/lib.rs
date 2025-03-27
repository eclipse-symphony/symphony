/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

 extern crate symphony;

 use symphony::models::{
     ProviderConfig, ValidationRule, DeploymentSpec, ComponentStep, ComponentSpec,
     DeploymentStep, ComponentResultSpec,
     ComponentValidationRule, RouteSpec, SidecarSpec, State
 };
 use symphony::ITargetProvider;
 use symphony::ProviderWrapper;
 use std::collections::HashMap;
 
 pub struct MockProvider;
 
 #[no_mangle]
 pub extern "C" fn create_provider() -> *mut ProviderWrapper  {
     let provider: Box<dyn ITargetProvider> = Box::new(MockProvider {});
     let wrapper = Box::new(ProviderWrapper { inner: provider });
     Box::into_raw(wrapper)
 }
 
 impl ITargetProvider for MockProvider {
     fn init(&self, _config: ProviderConfig) -> Result<(), String> {
         println!("MOCK RUST PROVIDER: ------ init()");
         Ok(())
     }
 
     fn get_validation_rule(&self) -> Result<ValidationRule, String> {
         println!("MOCK RUST PROVIDER: ------ get_validation_rule()");
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
         println!("MOCK RUST PROVIDER: ------ get()");
         let component_spec = ComponentSpec {
             name: "example_component".to_string(),
             component_type: Some("example_type".to_string()),
             metadata: Some(HashMap::from([
                 ("example_metadata_key".to_string(), "example_metadata_value".to_string())
             ])),
             properties: Some(HashMap::from([
                 ("example_property_key".to_string(), serde_json::json!("example_property_value"))
             ])),
             parameters: Some(HashMap::from([
                 ("example_parameter_key".to_string(), "example_parameter_value".to_string())
             ])),
             routes: Some(vec![
                 RouteSpec {
                     route: "example_route".to_string(),
                     route_type: "example_type".to_string(),
                     properties: Some(HashMap::from([
                         ("example_route_property_key".to_string(), "example_route_property_value".to_string())
                     ])),
                     filters: Some(Vec::new()),
                 }
             ]),
             constraints: Some("example_constraint".to_string()),
             dependencies: Some(Vec::new()),
             skills: Some(Vec::new()),
             sidecars: Some(vec![
                 SidecarSpec {
                     name: Some("example_sidecar".to_string()),
                     sidecar_type: Some("example_type".to_string()),
                     properties: Some(HashMap::from([
                         ("example_sidecar_property_key".to_string(), serde_json::json!("example_sidecar_property_value"))
                     ])),
                 }
             ]),
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
                 message: format!("Component {} applied successfully", component_step.component.name),
             };
             result_map.insert(component_step.component.name.clone(), result_spec);
         }
 
         Ok(result_map)
     }
 }