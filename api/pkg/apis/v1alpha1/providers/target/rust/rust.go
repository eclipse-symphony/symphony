//go:build !azure

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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	log = logger.NewLogger("providers.target.rust")
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
	ctx, span := observability.StartSpan(
		"Rust Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfoCtx(ctx, "  P (Rust Target): Init()")

	rustConfig, err := toRustTargetProviderConfig(config)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): expected RustTargetProviderConfig - %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: expected RustTargetProviderConfig", providerName), v1alpha2.InitFailed)
	}

	// Create Rust provider from file
	cProviderPath := C.CString(rustConfig.LibFile)
	cExpectedHash := C.CString(rustConfig.LibHash)
	defer C.free(unsafe.Pointer(cProviderPath))
	defer C.free(unsafe.Pointer(cExpectedHash))

	r.provider = C.create_provider_instance(cProviderPath, cExpectedHash)
	if r.provider == nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to create Rust provider from library file - %+v", err)
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("%s: failed to create Rust provider from library file", providerName), v1alpha2.InitFailed)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to serialize Rust provider configuration - %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to serialize Rust provider configuration", providerName), v1alpha2.InitFailed)
	}

	cConfigJSON := C.CString(string(configJSON))
	defer C.free(unsafe.Pointer(cConfigJSON))

	res := C.init_provider(r.provider, cConfigJSON)
	if res != 0 {
		log.ErrorfCtx(ctx, "  P (Rust Target): ailed to initialize provider - %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: ailed to initialize provider", providerName), v1alpha2.InitFailed)
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
	ctx, span := observability.StartSpan("Rust Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (Rust Target Provider): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	deploymentJSON, err := json.Marshal(deployment)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to marshal deployment - %+v", err)
		err = v1alpha2.NewCOAError(err, "failed to marshal deployment", v1alpha2.BadRequest)
		return nil, err
	}

	referencesJSON, err := json.Marshal(references)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to marshal reference - %+v", err)
		err = v1alpha2.NewCOAError(err, "failed to marshal reference", v1alpha2.BadRequest)
		return nil, err
	}

	cDeploymentJSON := C.CString(string(deploymentJSON))
	defer C.free(unsafe.Pointer(cDeploymentJSON))

	cReferencesJSON := C.CString(string(referencesJSON))
	defer C.free(unsafe.Pointer(cReferencesJSON))

	cResult := C.get(r.provider, cDeploymentJSON, cReferencesJSON)
	if cResult == nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to marshal reference - %+v", err)
		err = v1alpha2.NewCOAError(err, "failed to marshal reference", v1alpha2.BadRequest)
		return nil, err
	}
	defer C.free(unsafe.Pointer(cResult))

	var components []model.ComponentSpec
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &components); err != nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to unmarshal component specs - %+v", err)
		err = v1alpha2.NewCOAError(err, "failed to unmarshal component specs", v1alpha2.GetComponentSpecFailed)
		return nil, err
	}

	return components, nil
}

func (r *RustTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Rust Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, "  P (Rust Target Provider): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	deploymentJSON, err := json.Marshal(deployment)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to marshal deployment - %+v", err)
		err = v1alpha2.NewCOAError(err, "failed to marshal deployment", v1alpha2.BadRequest)
		return nil, err
	}

	stepJSON, err := json.Marshal(step)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to marshal step - %+v", err)
		err = v1alpha2.NewCOAError(err, "failed to marshal step", v1alpha2.BadRequest)
		return nil, err
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
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to apply deployment ste - %+v", err)
		err = v1alpha2.NewCOAError(err, "failed to apply deployment ste", v1alpha2.ApplyResourceFailed)
		return nil, err
	}
	defer C.free(unsafe.Pointer(cResult))

	var result map[string]model.ComponentResultSpec
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		log.ErrorfCtx(ctx, "  P (Rust Target): failed to unmarshal apply result - %+v", err)
		err = v1alpha2.NewCOAError(err, "failed to unmarshal apply result", v1alpha2.ApplyResourceFailed)
		return nil, err
	}

	return result, nil
}

func (r *RustTargetProvider) Close() {
	if r.provider != nil {
		C.destroy_provider_instance(r.provider)
	}
}
