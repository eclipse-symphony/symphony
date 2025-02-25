/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

 use std::ffi::{CStr, CString};
 use std::os::raw::c_char;
 use libloading::{Library, Symbol};
 use std::ptr;
 use std::fs::File;
 use std::io::{self, Read};
 use sha2::{Sha256, Digest};
 use std::collections::HashMap; 
 
 pub mod models;
 use crate::models::*;
 
 // Function to compute the SHA-256 hash of a file
 fn compute_sha256_hash(file_path: &str) -> Result<String, io::Error> {
     let mut file = File::open(file_path)?;
     let mut hasher = Sha256::new();
     let mut buffer = [0; 4096];
     
     loop {
         let bytes_read = file.read(&mut buffer)?;
         if bytes_read == 0 {
             break;
         }
         hasher.update(&buffer[..bytes_read]);
     }
     
     let hash = hasher.finalize();
     Ok(format!("{:x}", hash))
 }
 
 // Function to validate the computed hash against the expected hash value
 fn validate_hash(file_path: &str, expected_hash: &str) -> Result<(), String> {
     let computed_hash = compute_sha256_hash(file_path).map_err(|e| format!("Error computing hash: {}", e))?;
     
     if computed_hash == expected_hash {        
         Ok(())
     } else {
         Err(format!("Hash mismatch! Expected: {}, Got: {}", expected_hash, computed_hash))
     }
 }
 
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
     lib: Library, // Keep the library loaded to ensure the provider's functions remain valid
 }
 
 pub struct ProviderWrapper {
     pub inner: Box<dyn ITargetProvider>,
 }
 
 // External function to create the provider instance
 #[no_mangle]
 pub extern "C" fn create_provider_instance(
     provider_path: *const c_char, 
     expected_hash: *const c_char
 ) -> *mut ProviderHandle {
     let provider_path = unsafe { CStr::from_ptr(provider_path) }
         .to_str()
         .expect("Invalid provider path");
 
     let expected_hash = unsafe { CStr::from_ptr(expected_hash) }
         .to_str()
         .expect("Invalid hash value");
 
     // Validate the hash before loading the provider
     if let Err(err) = validate_hash(provider_path, expected_hash) {
         eprintln!("Hash validation failed: {}", err);
         return ptr::null_mut();
     }
 
     let lib = unsafe { Library::new(provider_path).expect("Failed to load provider library") };
 
     type CreateProviderFn = unsafe fn() -> *mut ProviderWrapper;
     let create_provider: Symbol<CreateProviderFn> = unsafe {
         lib.get(b"create_provider\0").expect("Failed to load create_provider function")
     };
 
     let wrapper = unsafe { create_provider() };
 
     if wrapper.is_null() {
         eprintln!("Error: create_provider returned null pointer");
         return ptr::null_mut();
     }
 
     // Take ownership of the `Box<dyn ITargetProvider>` from the wrapper
     let provider_box = unsafe { Box::from_raw(wrapper).inner };
 
     let handle = Box::new(ProviderHandle {
         provider: provider_box, // Move the Box into ProviderHandle
         lib: lib,
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
     if handle.is_null() {
         eprintln!("Error: handle pointer is null");
         return -1;
     }
     
     if config_json.is_null() {
         eprintln!("Error: config_json pointer is null");
         return -1;
     }
     
     let config_str = match unsafe { CStr::from_ptr(config_json).to_str() } {
         Ok(str) => str,
         Err(err) => {
             eprintln!("Error converting config_json to string: {:?}", err);
             return -1;
         },
     };
     let config: ProviderConfig = match serde_json::from_str(config_str) {
         Ok(cfg) => cfg,
         Err(err) => {
             eprintln!("Error deserializing config_json: {:?}", err);
             return -1;
         },
     };
     let handle_ref = unsafe { &*handle };
     match handle_ref.provider.init(config) {
         Ok(_) => {
             return 0;
         }
         Err(err) => {
             eprintln!("Error during provider initialization: {:?}", err);
             return -1;
         },
     }
 }
 
 // Get validation rule as a JSON string
 #[no_mangle]
 pub extern "C" fn get_validation_rule(handle: *mut ProviderHandle) -> *mut c_char {
     let handle = unsafe { &*handle };
     match handle.provider.get_validation_rule() {
         Ok(validation_rule) => {
             match CString::new(serde_json::to_string(&validation_rule).unwrap()) {
                 Ok(cstr) => cstr.into_raw(),
                 Err(_) => ptr::null_mut(),
             }
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