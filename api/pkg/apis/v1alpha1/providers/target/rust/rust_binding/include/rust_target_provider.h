#ifndef RUST_TARGET_PROVIDER_H
#define RUST_TARGET_PROVIDER_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>
#include <stddef.h>

typedef struct {
    void* provider;  // Pointer to the provider instance
    void* lib;       // Pointer to the dynamically loaded library (to keep it in memory)
} ProviderHandle;

typedef struct {
    const void* ptr;  // Pointer to the array
    size_t len;       // Length of the array
} FFIArray;

typedef struct {
    const char* name;
    bool ignore_case;
    bool skip_if_missing;
    bool prefix_match;
    bool is_component_name;
} PropertyDesc;

typedef struct {
    const char* required_component_type;
    FFIArray change_detection;             // FFIArray of PropertyDesc
    FFIArray change_detection_metadata;    // FFIArray of PropertyDesc
    FFIArray required_properties;          // FFIArray of const char*
    FFIArray optional_properties;          // FFIArray of const char*
    FFIArray required_metadata;            // FFIArray of const char*
    FFIArray optional_metadata;            // FFIArray of const char*
} ComponentValidationRule;

typedef struct {
    // Fields for provider configuration
} ProviderConfig;

typedef struct {
    const char* required_component_type;
    ComponentValidationRule component_validation_rule;
    ComponentValidationRule sidecar_validation_rule;
    bool allow_sidecar;
    bool scope_isolation;
    bool instance_isolation;
} ValidationRule;

typedef struct {
    // Fields for deployment specification
} DeploymentSpec;

typedef struct {
    // Fields for component step
} ComponentStep;

typedef struct {
    // Fields for component specification
} ComponentSpec;

typedef struct {
    // Fields for deployment step
} DeploymentStep;

typedef struct {
    // Fields for component result specification
} ComponentResultSpec;

// Function to create a provider instance based on the provider type and path
ProviderHandle* create_provider_instance(const char* provider_type, const char* provider_path);

// Function to destroy a provider instance
void destroy_provider_instance(ProviderHandle* handle);

// Initialize the provider with the given configuration
int init_provider(ProviderHandle* handle, const ProviderConfig* config);

// Get the validation rule from the provider
ValidationRule get_validation_rule(ProviderHandle* handle);

// Get component specifications from the provider
ComponentSpec* get(ProviderHandle* handle, const DeploymentSpec* deployment, const ComponentStep* references, size_t* count);

// Apply a deployment step
ComponentResultSpec* apply(ProviderHandle* handle, const DeploymentSpec* deployment, const DeploymentStep* step, int is_dry_run, size_t* count);

#ifdef __cplusplus
}
#endif

#endif // RUST_TARGET_PROVIDER_H
