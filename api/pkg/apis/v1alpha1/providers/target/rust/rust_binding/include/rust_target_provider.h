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

// Function to create a provider instance based on the provider type and path
ProviderHandle* create_provider_instance(const char* provider_path, const char* expected_hash);

// Function to destroy a provider instance
void destroy_provider_instance(ProviderHandle* handle);

// Initialize the provider with the given configuration
int init_provider(ProviderHandle* handle, const char* config_json);

// Get the validation rule from the provider
const char* get_validation_rule(ProviderHandle* handle);

// Get component specifications from the provider
const char* get(ProviderHandle* handle, const char* deployment_json, const char* references_json);

// Apply a deployment step
const char* apply(ProviderHandle* handle, const char* deployment_json, const char* step_json, int is_dry_run);

#ifdef __cplusplus
}
#endif

#endif // RUST_TARGET_PROVIDER_H