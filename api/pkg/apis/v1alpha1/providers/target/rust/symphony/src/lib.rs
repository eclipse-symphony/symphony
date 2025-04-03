/*
 * Copyright (c) Microsoft Corporation and others.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

use core::ffi::c_char;
use libloading::Library;
use sha2::{Digest, Sha256};
use std::collections::HashMap;
use std::ffi::{CStr, CString};
use std::fs::File;
use std::io;
use std::ptr;
use tracing::{debug, error, warn};

pub mod models;
use crate::models::*;

/// Computes the SHA-256 hash of a file.
fn compute_sha256_hash(file_path: &str) -> Result<String, io::Error> {
    let mut file = File::open(file_path)?;
    let mut hasher = Sha256::new();
    io::copy(&mut file, &mut hasher)?;
    let hash = hasher.finalize();
    Ok(format!("{:x}", hash))
}

/// Validates the computed hash against the expected hash value.
fn validate_hash(file_path: &str, expected_hash: &str) -> Result<(), String> {
    debug!("expected hash value: {}", expected_hash);
    if "any".eq(expected_hash) {
        warn!("Validation of shared library's hash value has been disabled! Make sure that the shared library is what you expect it to be!");
        return Ok(());
    }
    let computed_hash =
        compute_sha256_hash(file_path).map_err(|e| format!("Error computing hash: {}", e))?;

    if computed_hash == expected_hash {
        Ok(())
    } else {
        Err(format!(
            "Hash mismatch! Expected: {}, Got: {}",
            expected_hash, computed_hash
        ))
    }
}

/// A Rust-based Symphony Target Provider.
pub trait ITargetProvider: Send + Sync {
    fn get_validation_rule(&self) -> Result<ValidationRule, String>;

    fn get(
        &self,
        deployment: DeploymentSpec,
        references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String>;

    fn apply(
        &self,
        deployment: DeploymentSpec,
        step: DeploymentStep,
        is_dry_run: bool,
    ) -> Result<HashMap<String, ComponentResultSpec>, String>;
}

/// A reference to the Target Provider instance that
/// has been created using a dynamically loaded library.
#[repr(C)]
pub struct ProviderHandle {
    provider: Box<dyn ITargetProvider>,
    // Keep the library loaded to ensure the provider's functions remain valid
    lib: Library,
}

pub struct ProviderWrapper {
    pub inner: Box<dyn ITargetProvider>,
}

/// Creates a new instance of the provider.
///
/// The Symphony runtime invokes this function to create a new Rust-based Target Provider based on
/// configuration properties specified for a Symphony _Target_'s binding.
///
/// This function tries to create the new instance by means of a dynamically loaded shared library
/// that it expects to find at the given path. The library must contain a function with the
/// following signature:
///
/// ```
/// use core::ffi::c_char;
/// use rust_binding::ProviderWrapper;
///
/// #[no_mangle]
/// pub unsafe extern "C" fn create_provider(config_json: *const c_char) -> *mut ProviderWrapper {
///   todo!()
/// }
/// ```
///
/// The library's provenance/integrity is verified by means of computing the SHA256 hash value for the library file
/// and comparing it to the expected hash value.
///
/// # Arguments
///
/// * `provider_path` - A UTF8 string containing the absolute path to the shared
///                     library that contains the provider code.
///                     The new instance will be created by looking up and
///                     invoking the shared library's `create_provider` function.
/// * `expected_hash` - A UTF8 string containing the hex representation of the hash
///                     code to use for verifying the integrity of the shared library.
///                     This value will be compared to the SHA256 hash value computed
///                     for the given library.
///                     A value of `any` disables the check. This is useful during
///                     development, when updating the expected hash value for each
///                     build cycle seems undesirable/unnecessary.
/// * `config_json` - The UTF8 representation of the JSON encoded provider configuration.
///
/// # Returns
///
/// A wrapper around the newly created provider instance or a `null` pointer
/// if the provider instance could not be created using the given library, e.g. because the
/// library file does not exist or its hash value does not match the expected value or the
/// library does not contain the required symbol.
///
/// # Safety
///
/// Client code needs to make sure that the provided pointers are valid.
#[no_mangle]
pub unsafe extern "C" fn create_provider_instance(
    provider_path: *const c_char,
    expected_hash: *const c_char,
    config_json: *const c_char,
) -> *mut ProviderHandle {
    // try to configure tracing to output Events to stdout
    if tracing_subscriber::fmt::try_init().is_err() {
        // This will fail on subsequent invocations of the target provider's functions
        // because the provider is stateless and thus newly created for each and every invocation.
        // Consequently, we do not need to clutter stdout with corresponding warning
        // but can assume that the subscriber has been initialized one way or the other ...
    }
    let Ok(provider_path) = unsafe { CStr::from_ptr(provider_path) }.to_str() else {
        error!("Path to shared library is not valid UTF-8");
        return ptr::null_mut();
    };

    let Ok(expected_hash) = unsafe { CStr::from_ptr(expected_hash) }.to_str() else {
        error!("Hash value is not valid UTF-8");
        return ptr::null_mut();
    };

    // Validate the hash before loading the provider
    if let Err(err) = validate_hash(provider_path, expected_hash) {
        error!("Hash validation failed: {}", err);
        return ptr::null_mut();
    }

    let Ok(lib) = (unsafe {
        Library::new(provider_path).inspect_err(|err| {
            error!("Failed to load provider library [{}]: {err}", provider_path);
        })
    }) else {
        return ptr::null_mut();
    };

    let Ok(newly_created_provider_wrapper) = (unsafe {
        type CreateProviderFn = unsafe extern "C" fn(*const c_char) -> *mut ProviderWrapper;
        lib.get::<CreateProviderFn>(b"create_provider\0")
            .map(|create_provider| create_provider(config_json))
    }) else {
        error!(
            "Shared library [{}] does not contain required symbol: create_provider",
            provider_path
        );
        return ptr::null_mut();
    };

    if newly_created_provider_wrapper.is_null() {
        error!("Error: create_provider returned null pointer");
        return ptr::null_mut();
    }

    // Take ownership of the `Box<dyn ITargetProvider>` from the wrapper
    let provider = unsafe { Box::from_raw(newly_created_provider_wrapper).inner };

    let handle = Box::new(ProviderHandle {
        provider, // Move the Box into ProviderHandle
        lib,
    });

    Box::into_raw(handle)
}

/// Destroys the provider instance.
///
/// # Safety
///
/// Client code needs to make sure that the passed in handle is a valid (raw) pointer.
#[no_mangle]
pub unsafe extern "C" fn destroy_provider_instance(handle: *mut ProviderHandle) {
    if !handle.is_null() {
        unsafe {
            drop(Box::from_raw(handle));
        }
    }
}

/// Gets validation rule as a JSON string.
///
/// # Safety
///
/// Client code needs to make sure that the passed in arguments are valid (raw) pointers.
#[no_mangle]
pub unsafe extern "C" fn get_validation_rule(handle: *mut ProviderHandle) -> *mut c_char {
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

/// Gets components as a JSON string.
///
/// # Safety
///
/// Client code needs to make sure that the passed in arguments are valid (raw) pointers.
#[no_mangle]
pub unsafe extern "C" fn get(
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
        Err(err) => {
            debug!("Error getting components: {err}");
            ptr::null_mut()
        }
    }
}

/// Applies deployment as a JSON string.
///
/// # Safety
///
/// Client code needs to make sure that the passed in arguments are valid (raw) pointers.
#[no_mangle]
pub unsafe extern "C" fn apply(
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
        Err(err) => {
            debug!("Error applying changes: {err}");
            ptr::null_mut()
        }
    }
}
