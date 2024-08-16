extern crate rust_binding;

use rust_binding::{ProviderConfig, ValidationRule, DeploymentSpec, ComponentStep, ComponentSpec, DeploymentStep, ComponentResultSpec, ITargetProvider, PropertyDesc, ComponentValidationRule, FFIArray, PersistentStrings};

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

    fn get_validation_rule(&self, persistent: &mut PersistentStrings) -> ValidationRule {
        println!("Returning mock validation rule");
    
        // Add strings to the persistent storage
        let name_ptr = persistent.add("example_property");
        let type_ptr = persistent.add("example_type");
        let required_property_name_ptr = persistent.add("required_property_name");
    
        let required_property = PropertyDesc {
            name: name_ptr,
            ignore_case: false,
            skip_if_missing: false,
            prefix_match: false,
            is_component_name: false,
        };
    
        let required_property_desc = PropertyDesc {
            name: required_property_name_ptr,
            ignore_case: true,
            skip_if_missing: false,
            prefix_match: true,
            is_component_name: false,
        };
    
        let change_detection_array = FFIArray::new(vec![required_property]);
        let required_properties_array = FFIArray::new(vec![required_property_desc.name]);
    
        let component_validation_rule = ComponentValidationRule {
            required_component_type: type_ptr,
            change_detection: change_detection_array.clone(),
            change_detection_metadata: FFIArray::new(Vec::new()), // Empty array
            required_properties: required_properties_array,        // Sample required property
            optional_properties: FFIArray::new(Vec::new()),        // Empty array
            required_metadata: FFIArray::new(Vec::new()),          // Empty array
            optional_metadata: FFIArray::new(Vec::new()),          // Empty array
        };
    
        ValidationRule {
            required_component_type: type_ptr,
            component_validation_rule: component_validation_rule.clone(),
            sidecar_validation_rule: component_validation_rule.clone(),
            allow_sidecar: true,
            scope_isolation: true,
            instance_isolation: true,
        }
    }
    

    fn get(&self, _deployment: DeploymentSpec, _references: Vec<ComponentStep>) -> Result<Vec<ComponentSpec>, String> {
        println!("Returning mock component specs");
        Ok(vec![
            ComponentSpec {
                // Populate with mock data
            },
        ])
    }

    fn apply(&self, _deployment: DeploymentSpec, _step: DeploymentStep, _is_dry_run: bool) -> Result<Vec<ComponentResultSpec>, String> {
        println!("Applying mock deployment step");
        Ok(vec![
            ComponentResultSpec {
                // Populate with mock data
            },
        ])
    }
}
