#ifndef RUST_TARGET_PROVIDER_H
#define RUST_TARGET_PROVIDER_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>
#include <stddef.h>

typedef struct {
    // Fields for provider configuration
} ProviderConfig;

typedef struct {
    // Fields for validation rule
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

// Function to create a provider instance based on the provider type and path to the provider file
void* create_provider_instance(const char* provider_type, const char* path);

// Function to destroy a provider instance
void destroy_provider_instance(void* provider);

// Initialize the provider with the given configuration
int init_provider(void* provider, const ProviderConfig* config);

// Get the validation rule from the provider
ValidationRule get_validation_rule(void* provider);

// Get component specifications from the provider
ComponentSpec* get(void* provider, const DeploymentSpec* deployment, const ComponentStep* references, size_t* count);

// Apply a deployment step
ComponentResultSpec* apply(void* provider, const DeploymentSpec* deployment, const DeploymentStep* step, int is_dry_run, size_t* count);

#ifdef __cplusplus
}
#endif

#endif // RUST_TARGET_PROVIDER_H
