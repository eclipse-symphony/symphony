use std::ffi::CStr;
use std::os::raw::c_char;
use libloading::{Library, Symbol};

#[repr(C)]
#[derive(Clone)]
pub struct ProviderConfig {
    _private: [u8; 0], // Ensure the struct is not empty
}

#[repr(C)]
#[derive(Clone)]
pub struct ValidationRule {
    // Define fields for validation rule
}

#[repr(C)]
#[derive(Clone)]
pub struct DeploymentSpec {
    // Define fields for deployment specification
}

#[repr(C)]
#[derive(Clone)]
pub struct ComponentStep {
    // Define fields for component step
}

#[repr(C)]
#[derive(Clone)]
pub struct ComponentSpec {
    // Define fields for component specification
}

#[repr(C)]
#[derive(Clone)]
pub struct DeploymentStep {
    // Define fields for deployment step
}

#[repr(C)]
#[derive(Clone)]
pub struct ComponentResultSpec {
    // Define fields for component result specification
}

// Trait that all providers must implement
pub trait ITargetProvider: Send + Sync {
    fn init(&self, config: ProviderConfig) -> Result<(), String>;
    fn get_validation_rule(&self) -> ValidationRule;
    fn get(&self, deployment: DeploymentSpec, references: Vec<ComponentStep>) -> Result<Vec<ComponentSpec>, String>;
    fn apply(&self, deployment: DeploymentSpec, step: DeploymentStep, is_dry_run: bool) -> Result<Vec<ComponentResultSpec>, String>;
}

// Struct to hold the provider instance
#[repr(C)]
pub struct ProviderHandle {
    provider: Box<dyn ITargetProvider>,
    _lib: Library, // Keep the library loaded to ensure the provider's functions remain valid
}

#[no_mangle]
pub extern "C" fn create_provider_instance(provider_type: *const c_char, provider_path: *const c_char) -> *mut ProviderHandle {
    let provider_type = unsafe { CStr::from_ptr(provider_type) }
        .to_str()
        .expect("Invalid provider type");

    let provider_path = unsafe { CStr::from_ptr(provider_path) }
        .to_str()
        .expect("Invalid provider path");

    // Load the provider library dynamically
    let lib = unsafe { Library::new(provider_path).expect("Failed to load provider library") };

    // Define a type alias for the expected function signature
    type CreateProviderFn = unsafe fn() -> *mut dyn ITargetProvider;

    // Find the symbol in the library
    let create_provider: Symbol<CreateProviderFn> = unsafe {
        lib.get(b"create_provider\0").expect("Failed to load create_provider function")
    };

    // Call the create function from the provider
    let provider = unsafe { create_provider() };

    let handle = Box::new(ProviderHandle { provider: unsafe { Box::from_raw(provider) }, _lib: lib });
    Box::into_raw(handle)
}

#[no_mangle]
pub extern "C" fn destroy_provider_instance(handle: *mut ProviderHandle) {
    if !handle.is_null() {
        unsafe {
            drop(Box::from_raw(handle));
        }
    }
}

#[no_mangle]
pub extern "C" fn init_provider(handle: *mut ProviderHandle, config: ProviderConfig) -> i32 {
    let handle = unsafe { &*handle };
    match handle.provider.init(config) {
        Ok(_) => 0,
        Err(_) => -1,
    }
}

#[no_mangle]
pub extern "C" fn get_validation_rule(handle: *mut ProviderHandle) -> ValidationRule {
    let handle = unsafe { &*handle };
    handle.provider.get_validation_rule()
}

#[no_mangle]
pub extern "C" fn get(
    handle: *mut ProviderHandle,
    deployment: *const DeploymentSpec,
    references: *const Vec<ComponentStep>,
) -> *mut Vec<ComponentSpec> {
    let handle = unsafe { &*handle };
    let deployment = unsafe { &*deployment }.clone();
    let references = unsafe { &*references }.clone();
    match handle.provider.get(deployment, references) {
        Ok(result) => Box::into_raw(Box::new(result)),
        Err(_) => std::ptr::null_mut(),
    }
}

#[no_mangle]
pub extern "C" fn apply(
    handle: *mut ProviderHandle,
    deployment: *const DeploymentSpec,
    step: *const DeploymentStep,
    is_dry_run: i32,
) -> *mut Vec<ComponentResultSpec> {
    let handle = unsafe { &*handle };
    let deployment = unsafe { &*deployment }.clone();
    let step = unsafe { &*step }.clone();
    let is_dry_run = is_dry_run != 0;
    match handle.provider.apply(deployment, step, is_dry_run) {
        Ok(result) => Box::into_raw(Box::new(result)),
        Err(_) => std::ptr::null_mut(),
    }
}
