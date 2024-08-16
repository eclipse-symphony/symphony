package main

/*
#cgo LDFLAGS: -L./target/release -lrust_binding -lm -ldl -lpthread
#cgo CFLAGS: -I./rust_binding/include

#include <stdlib.h>
#include "rust_target_provider.h"
*/
import "C"
import (
	"context"
	"fmt"
	"unsafe"
)

type ProviderConfig C.ProviderConfig

type ValidationRule C.ValidationRule

type DeploymentSpec C.DeploymentSpec

type ComponentStep C.ComponentStep

type ComponentSpec C.ComponentSpec

type DeploymentStep C.DeploymentStep

type ComponentResultSpec C.ComponentResultSpec

type ITargetProvider interface {
	Init(config ProviderConfig) error
	GetValidationRule(ctx context.Context) ValidationRule
	Get(ctx context.Context, deployment DeploymentSpec, references []ComponentStep) ([]ComponentSpec, error)
	Apply(ctx context.Context, deployment DeploymentSpec, step DeploymentStep, isDryRun bool) (map[string]ComponentResultSpec, error)
}

type RustTargetProvider struct {
	provider *C.ProviderHandle
}

func (r *RustTargetProvider) Init(config ProviderConfig) error {
	cConfig := C.ProviderConfig(config)
	res := C.init_provider(r.provider, &cConfig)
	if res != 0 {
		return fmt.Errorf("failed to initialize provider")
	}
	return nil
}

func (r *RustTargetProvider) GetValidationRule(ctx context.Context) ValidationRule {
	rule := C.get_validation_rule(r.provider)
	return ValidationRule(rule)
}

func (r *RustTargetProvider) Get(ctx context.Context, deployment DeploymentSpec, references []ComponentStep) ([]ComponentSpec, error) {
	cDeployment := C.DeploymentSpec(deployment)

	// Handle empty `references` slice
	var cReferences *C.ComponentStep
	if len(references) > 0 {
		cReferences = (*C.ComponentStep)(unsafe.Pointer(&references[0]))
	} else {
		// Create a dummy slice with a single zero-value element
		dummy := make([]C.ComponentStep, 1)
		cReferences = &dummy[0]
	}

	var count C.size_t

	cResult := C.get(r.provider, &cDeployment, cReferences, &count)
	if cResult == nil {
		return nil, fmt.Errorf("failed to get component specs")
	}

	// Convert C array to Go slice
	result := make([]ComponentSpec, count)
	for i := C.size_t(0); i < count; i++ {
		result[i] = ComponentSpec(C.ComponentSpec(*(*C.ComponentSpec)(unsafe.Pointer(uintptr(unsafe.Pointer(cResult)) + uintptr(i)*unsafe.Sizeof(*cResult)))))
	}

	return result, nil
}

func (r *RustTargetProvider) Apply(ctx context.Context, deployment DeploymentSpec, step DeploymentStep, isDryRun bool) (map[string]ComponentResultSpec, error) {
	cDeployment := C.DeploymentSpec(deployment)
	cStep := C.DeploymentStep(step)
	cIsDryRun := C.int(0)
	if isDryRun {
		cIsDryRun = 1
	}
	var count C.size_t

	cResult := C.apply(r.provider, &cDeployment, &cStep, cIsDryRun, &count)
	if cResult == nil {
		return nil, fmt.Errorf("failed to apply deployment step")
	}

	// Convert C array to Go map (if using a key, otherwise to a slice)
	result := make(map[string]ComponentResultSpec)
	for i := C.size_t(0); i < count; i++ {
		result[fmt.Sprintf("result_%d", i)] = ComponentResultSpec(C.ComponentResultSpec(*(*C.ComponentResultSpec)(unsafe.Pointer(uintptr(unsafe.Pointer(cResult)) + uintptr(i)*unsafe.Sizeof(*cResult)))))
	}

	return result, nil
}

func NewRustTargetProvider(providerType string, providerLibPath string) (*RustTargetProvider, error) {
	cProviderType := C.CString(providerType)
	defer C.free(unsafe.Pointer(cProviderType))

	cProviderPath := C.CString(providerLibPath)
	defer C.free(unsafe.Pointer(cProviderPath))

	provider := C.create_provider_instance(cProviderType, cProviderPath)
	if provider == nil {
		return nil, fmt.Errorf("failed to create provider instance")
	}

	return &RustTargetProvider{provider: provider}, nil
}

func (r *RustTargetProvider) Close() {
	if r.provider != nil {
		C.destroy_provider_instance(r.provider)
	}
}

func main() {
	config := ProviderConfig{}
	rustProvider, err := NewRustTargetProvider("mock", "./target/release/libmock.so")
	if err != nil {
		fmt.Println("Error loading provider:", err)
		return
	}
	defer rustProvider.Close()

	if err := rustProvider.Init(config); err != nil {
		fmt.Println("Error initializing provider:", err)
		return
	}
	fmt.Println("Provider initialized successfully")

	// Example usage of GetValidationRule
	rule := rustProvider.GetValidationRule(context.Background())
	fmt.Printf("Validation Rule: %+v\n", rule)

	// Example usage of Get
	deployment := DeploymentSpec{}
	references := []ComponentStep{}
	components, err := rustProvider.Get(context.Background(), deployment, references)
	if err != nil {
		fmt.Println("Error getting components:", err)
		return
	}
	fmt.Printf("Components: %+v\n", components)

	// Example usage of Apply
	step := DeploymentStep{}
	results, err := rustProvider.Apply(context.Background(), deployment, step, true)
	if err != nil {
		fmt.Println("Error applying step:", err)
		return
	}
	fmt.Printf("Results: %+v\n", results)
}
