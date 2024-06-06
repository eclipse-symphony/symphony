extern crate rust_binding;

use rust_binding::{ProviderConfig, ValidationRule, DeploymentSpec, ComponentStep, ComponentSpec, DeploymentStep, ComponentResultSpec, ITargetProvider};

pub struct MockProvider;

impl ITargetProvider for MockProvider {
    fn init(&self, _config: ProviderConfig) -> Result<(), String> {
        println!("MockProvider initialized");
        Ok(())
    }

    fn get_validation_rule(&self) -> ValidationRule {
        println!("Returning mock validation rule");
        ValidationRule {
            // Populate with mock data
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

#[no_mangle]
pub extern "C" fn create_mock_provider() -> *mut MockProvider {
    let provider = Box::new(MockProvider);
    Box::into_raw(provider)
}

#[no_mangle]
pub extern "C" fn destroy_mock_provider(provider: *mut MockProvider) {
    if !provider.is_null() {
        unsafe {
            drop(Box::from_raw(provider));
        }
    }
}