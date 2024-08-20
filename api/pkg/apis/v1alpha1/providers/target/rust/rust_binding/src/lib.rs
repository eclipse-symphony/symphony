/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use libloading::{Library, Symbol};
use std::ptr;
use std::collections::HashMap;

pub mod models;
use crate::models::*;

// Trait that all Rust-based Target providers must implement
pub trait ITargetProvider: Send + Sync {
    fn init(&self, config: ProviderConfig) -> Result<(), String>;
    fn get_validation_rule(&self) -> Result<ValidationRule, String>; // Return Rust native type
    fn get(&self, deployment: DeploymentSpec, references: Vec<ComponentStep>) -> Result<Vec<ComponentSpec>, String>; // Return Rust native types
    fn apply(&self, deployment: DeploymentSpec, step: DeploymentStep, is_dry_run: bool) -> Result<HashMap<String, ComponentResultSpec>, String>; 
}

// Struct to hold the provider instance
#[repr(C)]
pub struct ProviderHandle {
    provider: Box<dyn ITargetProvider>,
    _lib: Library, // Keep the library loaded to ensure the provider's functions remain valid
}

// External function to create the provider instance
#[no_mangle]
pub extern "C" fn create_provider_instance(provider_type: *const c_char, provider_path: *const c_char) -> *mut ProviderHandle {
    let _provider_type = unsafe { CStr::from_ptr(provider_type) }
        .to_str()
        .expect("Invalid provider type");

    let provider_path = unsafe { CStr::from_ptr(provider_path) }
        .to_str()
        .expect("Invalid provider path");

    let lib = unsafe { Library::new(provider_path).expect("Failed to load provider library") };

    type CreateProviderFn = unsafe fn() -> *mut dyn ITargetProvider;
    let create_provider: Symbol<CreateProviderFn> = unsafe {
        lib.get(b"create_provider\0").expect("Failed to load create_provider function")
    };

    let provider = unsafe { create_provider() };

    let handle = Box::new(ProviderHandle {
        provider: unsafe { Box::from_raw(provider) },
        _lib: lib,
    });

    Box::into_raw(handle)
}

// Destroy the provider instance
#[no_mangle]
pub extern "C" fn destroy_provider_instance(handle: *mut ProviderHandle) {
    if !handle.is_null() {
        unsafe {
            drop(Box::from_raw(handle));
        }
    }
}

// Initialize the provider with JSON configuration
#[no_mangle]
pub extern "C" fn init_provider(handle: *mut ProviderHandle, config_json: *const c_char) -> i32 {
    let config_str = unsafe { CStr::from_ptr(config_json).to_str().unwrap() };
    let config: ProviderConfig = serde_json::from_str(config_str).unwrap();

    let handle = unsafe { &*handle };
    match handle.provider.init(config) {
        Ok(_) => 0,
        Err(_) => -1,
    }
}

// Get validation rule as a JSON string
#[no_mangle]
pub extern "C" fn get_validation_rule(handle: *mut ProviderHandle) -> *mut c_char {
    let handle = unsafe { &*handle };
    match handle.provider.get_validation_rule() {
        Ok(validation_rule) => {
            let json = serde_json::to_string(&validation_rule).unwrap();
            CString::new(json).unwrap().into_raw()
        }
        Err(_) => ptr::null_mut(),
    }
}

// Get components as a JSON string
#[no_mangle]
pub extern "C" fn get(
    handle: *mut ProviderHandle,
    deployment_json: *const c_char,
    references_json: *const c_char,
) -> *mut c_char {
    let handle = unsafe { &*handle };
    let deployment_str = unsafe { CStr::from_ptr(deployment_json).to_str().unwrap() };
    let references_str = unsafe { CStr::from_ptr(references_json).to_str().unwrap() };
    let deployment: DeploymentSpec = serde_json::from_str(deployment_str).unwrap();
    let references: Vec<ComponentStep> = serde_json::from_str(references_str).unwrap();
    match handle.provider.get(deployment, references) {
        Ok(components) => {
            let json = serde_json::to_string(&components).unwrap();
            CString::new(json).unwrap().into_raw()
        }
        Err(_) => {
            println!("Error getting components");
            ptr::null_mut()
        } 
    }
}

// Apply deployment as a JSON string
#[no_mangle]
pub extern "C" fn apply(
    handle: *mut ProviderHandle,
    deployment_json: *const c_char,
    step_json: *const c_char,
    is_dry_run: i32,
) -> *mut c_char {
    let handle = unsafe { &*handle };
    let deployment_str = unsafe { CStr::from_ptr(deployment_json).to_str().unwrap() };
    let step_str = unsafe { CStr::from_ptr(step_json).to_str().unwrap() };

    let deployment: DeploymentSpec = serde_json::from_str(deployment_str).unwrap();
    let step: DeploymentStep = serde_json::from_str(step_str).unwrap();
    let is_dry_run = is_dry_run != 0;

    match handle.provider.apply(deployment, step, is_dry_run) {
        Ok(result_map) => {
            let json = serde_json::to_string(&result_map).unwrap();
            CString::new(json).unwrap().into_raw()
        }
        Err(_) => ptr::null_mut(),
    }
}