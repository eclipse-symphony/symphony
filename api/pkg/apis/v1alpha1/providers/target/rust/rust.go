/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rust

/*
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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

const (
	providerName = "P (Rust Target)"
)

type RustTargetProviderConfig struct {
	Name    string `json:"name"`
	LibFile string `json:"libFile"`
	LibHash string `json:"libHash"`
}

type RustTargetProvider struct {
	provider *C.ProviderHandle
	Context  *contexts.ManagerContext
}

func RustTargetProviderConfiggFromMap(properties map[string]string) (RustTargetProviderConfig, error) {
	ret := RustTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["libFile"]; ok {
		ret.LibFile = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "'libFile' is missing in Rust provider config", v1alpha2.BadConfig)
	}
	if v, ok := properties["libHash"]; ok {
		ret.LibHash = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "'libHash' is missing in Rust provider config", v1alpha2.BadConfig)
	}
	return ret, nil
}

func toRustTargetProviderConfig(config providers.IProviderConfig) (RustTargetProviderConfig, error) {
	ret := RustTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *RustTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := RustTargetProviderConfiggFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *RustTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (r *RustTargetProvider) Init(config providers.IProviderConfig) error {
	rustConfig, err := toRustTargetProviderConfig(config)
	if err != nil {
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: expected HttpTargetProviderConfig", providerName), v1alpha2.InitFailed)
		return err
	}

	// Create Rust provider from file
	cProviderPath := C.CString(rustConfig.LibFile)
	cExpectedHash := C.CString(rustConfig.LibHash)
	defer C.free(unsafe.Pointer(cProviderPath))
	defer C.free(unsafe.Pointer(cExpectedHash))

	provider := C.create_provider_instance(cProviderPath, cExpectedHash)
	if provider == nil {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to create Rust provider from library file", providerName), v1alpha2.InitFailed)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to serialize Rust provider configuration", providerName), v1alpha2.InitFailed)
	}

	r.provider = provider

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

// func NewRustTargetProvider(providerLibPath string) (*RustTargetProvider, error) {

// 	cProviderPath := C.CString(providerLibPath)
// 	defer C.free(unsafe.Pointer(cProviderPath))

// 	provider := C.create_provider_instance(cProviderPath)
// 	if provider == nil {
// 		return nil, fmt.Errorf("failed to create provider instance")
// 	}

// 	return &RustTargetProvider{provider: provider}, nil
// }

func (r *RustTargetProvider) Close() {
	if r.provider != nil {
		C.destroy_provider_instance(r.provider)
	}
}
