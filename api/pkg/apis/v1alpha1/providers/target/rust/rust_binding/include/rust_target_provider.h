#ifndef RUST_TARGET_PROVIDER_H
#define RUST_TARGET_PROVIDER_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>

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

int init_provider(void* provider, const ProviderConfig* config);
// Declarations for other functions

#ifdef __cplusplus
}
#endif

#endif // RUST_TARGET_PROVIDER_H
