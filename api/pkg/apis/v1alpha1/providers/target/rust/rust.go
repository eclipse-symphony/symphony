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
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type RustTargetProviderConfig struct {
}

type RustTargetProvider struct {
	provider *C.ProviderHandle
}

func (r *RustTargetProvider) Init(config providers.IProviderConfig) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	cConfigJSON := C.CString(string(configJSON))
	defer C.free(unsafe.Pointer(cConfigJSON))

	res := C.init_provider(r.provider, cConfigJSON)
	if res != 0 {
		return fmt.Errorf("failed to initialize provider")
	}
	return nil
}

func (r *RustTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	cRule := C.get_validation_rule(r.provider)
	cRuleStr := C.GoString(cRule)
	C.free(unsafe.Pointer(cRule))

	var validationRule model.ValidationRule
	if err := json.Unmarshal([]byte(cRuleStr), &validationRule); err != nil {
		//TODO: Handle error
		return model.ValidationRule{}
	}

	return validationRule
}

func (r *RustTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	deploymentJSON, err := json.Marshal(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment: %v", err)
	}

	referencesJSON, err := json.Marshal(references)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal references: %v", err)
	}

	cDeploymentJSON := C.CString(string(deploymentJSON))
	defer C.free(unsafe.Pointer(cDeploymentJSON))

	cReferencesJSON := C.CString(string(referencesJSON))
	defer C.free(unsafe.Pointer(cReferencesJSON))

	cResult := C.get(r.provider, cDeploymentJSON, cReferencesJSON)
	if cResult == nil {
		return nil, fmt.Errorf("failed to get component specs")
	}
	defer C.free(unsafe.Pointer(cResult))

	var components []model.ComponentSpec
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &components); err != nil {
		return nil, fmt.Errorf("failed to unmarshal component specs: %v", err)
	}

	return components, nil
}

func (r *RustTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	deploymentJSON, err := json.Marshal(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment: %v", err)
	}

	stepJSON, err := json.Marshal(step)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal step: %v", err)
	}

	cDeploymentJSON := C.CString(string(deploymentJSON))
	defer C.free(unsafe.Pointer(cDeploymentJSON))

	cStepJSON := C.CString(string(stepJSON))
	defer C.free(unsafe.Pointer(cStepJSON))

	cIsDryRun := C.int(0)
	if isDryRun {
		cIsDryRun = 1
	}

	cResult := C.apply(r.provider, cDeploymentJSON, cStepJSON, cIsDryRun)
	if cResult == nil {
		return nil, fmt.Errorf("failed to apply deployment step")
	}
	defer C.free(unsafe.Pointer(cResult))

	var result map[string]model.ComponentResultSpec
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal apply result: %v", err)
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
