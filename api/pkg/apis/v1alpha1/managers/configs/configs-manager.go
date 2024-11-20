/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package configs

import (
	"context"
	"fmt"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type ConfigsManager struct {
	managers.Manager
	ConfigProviders map[string]config.IConfigProvider
	Precedence      []string
}

func (s *ConfigsManager) Init(context *contexts.VendorContext, cfg managers.ManagerConfig, providers map[string]providers.IProvider) error {
	log.Debug(" M (Config): Init")
	err := s.Manager.Init(context, cfg, providers)
	if err != nil {
		return err
	}
	s.ConfigProviders = make(map[string]config.IConfigProvider)
	for key, provider := range providers {
		if cProvider, ok := provider.(config.IConfigProvider); ok {
			s.ConfigProviders[key] = cProvider
		}
	}
	if val, ok := cfg.Properties["precedence"]; ok {
		s.Precedence = strings.Split(val, ",")
	}
	if len(s.ConfigProviders) == 0 {
		log.Error(" M (Config): No config providers found")
		return v1alpha2.NewCOAError(nil, "No config providers found", v1alpha2.BadConfig)
	}
	if len(s.Precedence) < len(s.ConfigProviders) && len(s.ConfigProviders) > 1 {
		log.Error(" M (Config): Not enough precedence values")
		return v1alpha2.NewCOAError(nil, "Not enough precedence values", v1alpha2.BadConfig)
	}
	if len(s.ConfigProviders) > 1 {
		var provderKeys []string
		for key := range s.ConfigProviders {
			provderKeys = append(provderKeys, key)
		}
		if !utils.AreSlicesEqual(provderKeys, s.Precedence) {
			log.Error(" M (Config): Precedence does not match with config providers")
			return v1alpha2.NewCOAError(nil, "Precedence does not match with config providers", v1alpha2.BadConfig)
		}
	}
	return nil
}

func (s *ConfigsManager) Get(ctx context.Context, object string, field string, overlays []string, localContext interface{}) (interface{}, error) {
	ctx, span := observability.StartSpan("Config Manager", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.DebugfCtx(ctx, " M (Config): Get %v, config provider size %d", object, len(s.ConfigProviders))
	if strings.Index(object, "::") > 0 {
		parts := strings.Split(object, "::")
		if len(parts) != 2 {
			log.ErrorfCtx(ctx, " M (Config): Invalid object: %s", object)
			return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			if field == "" {
				configObj, err := s.getObjectWithOverlay(ctx, provider, parts[1], overlays, localContext)
				if err != nil {
					return "", err
				}
				return configObj, nil
			} else {
				return s.getWithOverlay(ctx, provider, parts[1], field, overlays, localContext)
			}
		}
		log.ErrorfCtx(ctx, " M (Config): Invalid provider: %s", parts[0])
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
		return "", err
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			if field == "" {
				configObj, err := s.getObjectWithOverlay(ctx, provider, object, overlays, localContext)
				if err != nil {
					return "", err
				}
				return configObj, nil
			} else {
				if value, err := s.getWithOverlay(ctx, provider, object, field, overlays, localContext); err == nil {
					return value, nil
				} else {
					return "", err
				}
			}
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if field == "" {
				configObj, err := s.getObjectWithOverlay(ctx, provider, object, overlays, localContext)
				if err != nil {
					return "", err
				}
				return configObj, nil
			} else {
				if value, err := s.getWithOverlay(ctx, provider, object, field, overlays, localContext); err == nil {
					return value, nil
				}
			}
		}
	}

	log.ErrorfCtx(ctx, " M (Config): Invalid config object or key: %s, %s", object, field)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object or key: %s, %s", object, field), v1alpha2.BadRequest)
	return "", err
}

func (s *ConfigsManager) getWithOverlay(ctx context.Context, provider config.IConfigProvider, object string, field string, overlays []string, localContext interface{}) (interface{}, error) {
	if len(overlays) > 0 {
		for _, overlay := range overlays {
			if overlayObject, err := provider.Read(ctx, overlay, field, localContext); err == nil {
				return overlayObject, nil
			}
		}
	}
	return provider.Read(ctx, object, field, localContext)
}

func (s *ConfigsManager) GetObject(ctx context.Context, object string, overlays []string, localContext interface{}) (map[string]interface{}, error) {
	ctx, span := observability.StartSpan("Config Manager", ctx, &map[string]string{
		"method": "GetObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.DebugfCtx(ctx, " M (Config): GetObject %v, config provider size %d", object, len(s.ConfigProviders))
	if strings.Index(object, "::") > 0 {
		parts := strings.Split(object, "::")
		if len(parts) != 2 {
			log.ErrorfCtx(ctx, " M (Config): Invalid object: %s", object)
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return s.getObjectWithOverlay(ctx, provider, parts[1], overlays, localContext)
		}
		log.ErrorfCtx(ctx, " M (Config): Invalid provider: %s", parts[0])
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
		return nil, err
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return s.getObjectWithOverlay(ctx, provider, object, overlays, localContext)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			return s.getObjectWithOverlay(ctx, provider, object, overlays, localContext)
		}
	}

	log.ErrorfCtx(ctx, " M (Config): Invalid config object: %s", object)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object: %s", object), v1alpha2.BadRequest)
	return nil, err
}

func (s *ConfigsManager) getObjectWithOverlay(ctx context.Context, provider config.IConfigProvider, object string, overlays []string, localContext interface{}) (map[string]interface{}, error) {
	if len(overlays) > 0 {
		for _, overlay := range overlays {
			overlayObject, err := provider.ReadObject(ctx, overlay, localContext)
			return overlayObject, err
		}
	}
	return provider.ReadObject(ctx, object, localContext)
}

func (s *ConfigsManager) Set(ctx context.Context, object string, field string, value interface{}) error {
	ctx, span := observability.StartSpan("Config Manager", ctx, &map[string]string{
		"method": "Set",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.DebugfCtx(ctx, " M (Config): Set %v, config provider size %d", object, len(s.ConfigProviders))
	if strings.Index(object, "::") > 0 {
		parts := strings.Split(object, "::")
		if len(parts) != 2 {
			log.ErrorfCtx(ctx, " M (Config): Invalid object: %s", object)
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
			return err
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return provider.Set(ctx, parts[1], field, value)
		}
		log.ErrorfCtx(ctx, " M (Config): Invalid provider: %s", parts[0])
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
		return err
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return provider.Set(ctx, object, field, value)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if err := provider.Set(ctx, object, field, value); err == nil {
				return nil
			}
		}
	}

	log.ErrorfCtx(ctx, " M (Config): Invalid config object or key: %s, %s", object, field)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object or key: %s, %s", object, field), v1alpha2.BadRequest)
	return err
}

func (s *ConfigsManager) SetObject(ctx context.Context, object string, values map[string]interface{}) error {
	ctx, span := observability.StartSpan("Config Manager", ctx, &map[string]string{
		"method": "SetObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.DebugfCtx(ctx, " M (Config): SetObject %v, config provider size %d", object, len(s.ConfigProviders))
	if strings.Index(object, "::") > 0 {
		parts := strings.Split(object, "::")
		if len(parts) != 2 {
			log.ErrorfCtx(ctx, " M (Config): Invalid object: %s", object)
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
			return err
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return provider.SetObject(ctx, parts[1], values)
		}
		log.ErrorfCtx(ctx, " M (Config): Invalid provider: %s", parts[0])
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
		return err
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return provider.SetObject(ctx, object, values)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if err := provider.SetObject(ctx, object, values); err == nil {
				return nil
			}
		}
	}

	log.ErrorfCtx(ctx, " M (Config): Invalid config object: %s", object)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object: %s", object), v1alpha2.BadRequest)
	return err
}

func (s *ConfigsManager) Delete(ctx context.Context, object string, field string) error {
	ctx, span := observability.StartSpan("Config Manager", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.DebugfCtx(ctx, " M (Config): Delete %v, config provider size %d", object, len(s.ConfigProviders))
	if strings.Index(object, "::") > 0 {
		parts := strings.Split(object, "::")
		if len(parts) != 2 {
			log.ErrorfCtx(ctx, " M (Config): Invalid object: %s", object)
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
			return err
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return provider.Remove(ctx, parts[1], field)
		}
		log.ErrorfCtx(ctx, " M (Config): Invalid provider: %s", parts[0])
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
		return err
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return provider.Remove(ctx, object, field)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if err := provider.Remove(ctx, object, field); err == nil {
				return nil
			}
		}
	}

	log.ErrorfCtx(ctx, " M (Config): Invalid config object or key: %s, %s", object, field)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object or key: %s, %s", object, field), v1alpha2.BadRequest)
	return err
}

func (s *ConfigsManager) DeleteObject(ctx context.Context, object string) error {
	ctx, span := observability.StartSpan("Config Manager", ctx, &map[string]string{
		"method": "DeleteObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.DebugfCtx(ctx, " M (Config): DeleteObject %v, config provider size %d", object, len(s.ConfigProviders))
	if strings.Index(object, "::") > 0 {
		parts := strings.Split(object, "::")
		if len(parts) != 2 {
			log.ErrorfCtx(ctx, " M (Config): Invalid object: %s", object)
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
			return err
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return provider.RemoveObject(ctx, parts[1])
		}
		log.ErrorfCtx(ctx, " M (Config): Invalid provider: %s", parts[0])
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
		return err
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return provider.RemoveObject(ctx, object)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if err := provider.RemoveObject(ctx, object); err == nil {
				return nil
			}
		}
	}

	log.ErrorfCtx(ctx, " M (Config): Invalid config object: %s", object)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object: %s", object), v1alpha2.BadRequest)
	return err
}
