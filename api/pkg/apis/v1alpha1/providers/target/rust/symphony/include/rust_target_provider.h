/*
 * Copyright (c) Microsoft Corporation and others.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

#ifndef RUST_TARGET_PROVIDER_H
#define RUST_TARGET_PROVIDER_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>
#include <stddef.h>

// Handle to the provider instance
typedef struct {
    void* provider;  // Pointer to the provider instance
    void* lib;       // Pointer to the dynamically loaded library (to keep it in memory)
} ProviderHandle;

// Creates a new Rust target provider instance based on a dynamically loaded shared library.
ProviderHandle *create_provider_instance(const char *provider_path, const char *expected_hash, const char *config_json);

// Destroys a provider instance.
void destroy_provider_instance(ProviderHandle* handle);

// Gets a target provider's validation rule.
const char* get_validation_rule(ProviderHandle* handle);

// Gets component specifications from a target provider.
const char* get(ProviderHandle* handle, const char* deployment_json, const char* references_json);

// Applies a deployment step to a target provider.
const char* apply(ProviderHandle* handle, const char* deployment_json, const char* step_json, int is_dry_run);

#ifdef __cplusplus
}
#endif

#endif // RUST_TARGET_PROVIDER_H