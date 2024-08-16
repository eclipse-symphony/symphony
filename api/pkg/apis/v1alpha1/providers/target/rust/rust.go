package rust

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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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

// Convert a C string to a Go string
func cStringToGoString(cStr *C.char) string {
	return C.GoString(cStr)
}

// Convert an FFIArray of C strings to a Go slice of strings
func cStringArrayToGoSlice(array C.FFIArray) []string {
	length := int(array.len)
	if length == 0 || array.ptr == nil {
		return nil
	}
	cArray := (*[1 << 28]*C.char)(unsafe.Pointer(array.ptr))[:length:length]
	goSlice := make([]string, length)
	for i, s := range cArray {
		goSlice[i] = cStringToGoString(s)
	}
	return goSlice
}

// Convert an FFIArray of PropertyDesc to a Go slice of PropertyDesc
func cPropertyDescArrayToGoSlice(array C.FFIArray) []model.PropertyDesc {
	length := int(array.len)
	if length == 0 || array.ptr == nil {
		return nil
	}
	cArray := (*[1 << 28]C.PropertyDesc)(unsafe.Pointer(array.ptr))[:length:length]
	goSlice := make([]model.PropertyDesc, length)
	for i, p := range cArray {
		goSlice[i] = model.PropertyDesc{
			Name:            cStringToGoString(p.name),
			IgnoreCase:      bool(p.ignore_case),
			SkipIfMissing:   bool(p.skip_if_missing),
			PrefixMatch:     bool(p.prefix_match),
			IsComponentName: bool(p.is_component_name),
		}
	}
	return goSlice
}

// Convert a C ValidationRule to a Go ValidationRule
func cValidationRuleToGoValidationRule(cRule C.ValidationRule) model.ValidationRule {
	return model.ValidationRule{
		RequiredComponentType: cStringToGoString(cRule.required_component_type),
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredComponentType:     cStringToGoString(cRule.component_validation_rule.required_component_type),
			ChangeDetectionProperties: cPropertyDescArrayToGoSlice(cRule.component_validation_rule.change_detection),
			ChangeDetectionMetadata:   cPropertyDescArrayToGoSlice(cRule.component_validation_rule.change_detection_metadata),
			RequiredProperties:        cStringArrayToGoSlice(cRule.component_validation_rule.required_properties),
			OptionalProperties:        cStringArrayToGoSlice(cRule.component_validation_rule.optional_properties),
			RequiredMetadata:          cStringArrayToGoSlice(cRule.component_validation_rule.required_metadata),
			OptionalMetadata:          cStringArrayToGoSlice(cRule.component_validation_rule.optional_metadata),
		},
		SidecarValidationRule: model.ComponentValidationRule{
			RequiredComponentType:     cStringToGoString(cRule.sidecar_validation_rule.required_component_type),
			ChangeDetectionProperties: cPropertyDescArrayToGoSlice(cRule.sidecar_validation_rule.change_detection),
			ChangeDetectionMetadata:   cPropertyDescArrayToGoSlice(cRule.sidecar_validation_rule.change_detection_metadata),
			RequiredProperties:        cStringArrayToGoSlice(cRule.sidecar_validation_rule.required_properties),
			OptionalProperties:        cStringArrayToGoSlice(cRule.sidecar_validation_rule.optional_properties),
			RequiredMetadata:          cStringArrayToGoSlice(cRule.sidecar_validation_rule.required_metadata),
			OptionalMetadata:          cStringArrayToGoSlice(cRule.sidecar_validation_rule.optional_metadata),
		},
		AllowSidecar:      bool(cRule.allow_sidecar),
		ScopeIsolation:    bool(cRule.scope_isolation),
		InstanceIsolation: bool(cRule.instance_isolation),
	}
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

func (r *RustTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	// Call the C function to get the validation rule
	cRule := C.get_validation_rule(r.provider)

	// Convert the C ValidationRule to the Go ValidationRule
	goRule := cValidationRuleToGoValidationRule(cRule)

	// Return the Go model ValidationRule
	return goRule
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
