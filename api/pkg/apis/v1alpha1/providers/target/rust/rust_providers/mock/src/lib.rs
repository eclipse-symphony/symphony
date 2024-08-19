extern crate rust_binding;

use rust_binding::models::{
    ProviderConfig, ValidationRule, DeploymentSpec, ComponentStep, ComponentSpec,
    DeploymentStep, ComponentResultSpec, PropertyDesc,
    ComponentValidationRule, RouteSpec, SidecarSpec, State
};
use rust_binding::ITargetProvider;

use std::collections::HashMap;

pub struct MockProvider;

#[no_mangle]
pub extern "C" fn create_provider() -> *mut dyn ITargetProvider {
    let provider = Box::new(MockProvider {});
    Box::into_raw(provider)
}

impl ITargetProvider for MockProvider {
    fn init(&self, _config: ProviderConfig) -> Result<(), String> {
        println!("MockProvider initialized");
        Ok(())
    }

    fn get_validation_rule(&self) -> Result<ValidationRule, String> {
        println!("Returning mock validation rule");
    
        let validation_rule = ValidationRule {
            required_component_type: "example_type".to_string(),
            component_validation_rule: ComponentValidationRule {
                required_component_type: "example_type".to_string(),
                change_detection_properties: vec![
                    PropertyDesc {
                        name: "example_property".to_string(),
                        ignore_case: true,
                        skip_if_missing: true,
                        prefix_match: true,
                        is_component_name: true,
                    }
                ],
                change_detection_metadata: vec![
                    PropertyDesc {
                        name: "example_metadata_property".to_string(),
                        ignore_case: false,
                        skip_if_missing: false,
                        prefix_match: false,
                        is_component_name: false,
                    }
                ],
                required_properties: vec!["required_property_name".to_string()],
                optional_properties: vec!["optional_property_name".to_string()],
                required_metadata: vec!["required_metadata_name".to_string()],
                optional_metadata: vec!["optional_metadata_name".to_string()],
            },
            sidecar_validation_rule: ComponentValidationRule {
                required_component_type: "sidecar_example_type".to_string(),
                change_detection_properties: vec![
                    PropertyDesc {
                        name: "sidecar_example_property".to_string(),
                        ignore_case: false,
                        skip_if_missing: false,
                        prefix_match: true,
                        is_component_name: false,
                    }
                ],
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
        println!("Returning mock component specs");

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

    fn apply(&self, _deployment: DeploymentSpec, _step: DeploymentStep, _is_dry_run: bool) -> Result<Vec<ComponentResultSpec>, String> {
        println!("Applying mock deployment step");

        let result_spec = ComponentResultSpec {
            status: State::OK,
            message: "Operation successful".to_string(),
        };

        Ok(vec![result_spec])
    }
}
