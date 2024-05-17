/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package configs

import (
	"fmt"
	"strings"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
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
	if len(s.ConfigProviders) > 0 && len(s.Precedence) < len(s.ConfigProviders) && len(s.ConfigProviders) > 1 {
		log.Error(" M (Config): Not enough precedence values")
		return v1alpha2.NewCOAError(nil, "Not enough precedence values", v1alpha2.BadConfig)
	}
	for _, key := range s.Precedence {
		if _, ok := s.ConfigProviders[key]; !ok {
			log.Error(" M (Config): Invalid precedence value: %s", key)
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid precedence value: %s", key), v1alpha2.BadConfig)
		}
	}
	return nil
}
func (s *ConfigsManager) Get(object string, field string, overlays []string, localContext interface{}) (interface{}, error) {
	log.Infof(" M (Config): Get %v, %d", object, len(s.ConfigProviders))
	if strings.Index(object, ":") > 0 && len(s.ConfigProviders) > 1 {
		parts := strings.Split(object, ":")
		if len(parts) != 2 {
			return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			if field == "" {
				configObj, err := s.getObjectWithOverlay(provider, parts[1], overlays, localContext)
				if err != nil {
					return "", err
				}
				return configObj, nil
			} else {
				return s.getWithOverlay(provider, parts[1], field, overlays, localContext)
			}
		}
		return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			if field == "" {
				configObj, err := s.getObjectWithOverlay(provider, object, overlays, localContext)
				if err != nil {
					return "", err
				}
				return configObj, nil
			} else {
				if value, err := s.getWithOverlay(provider, object, field, overlays, localContext); err == nil {
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
				configObj, err := s.getObjectWithOverlay(provider, object, overlays, localContext)
				if err != nil {
					return "", err
				}
				return configObj, nil
			} else {
				if value, err := s.getWithOverlay(provider, object, field, overlays, localContext); err == nil {
					return value, nil
				}
			}
		}
	}
	return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object or key: %s, %s", object, field), v1alpha2.BadRequest)
}
func (s *ConfigsManager) getWithOverlay(provider config.IConfigProvider, object string, field string, overlays []string, localContext interface{}) (interface{}, error) {
	if len(overlays) > 0 {
		for _, overlay := range overlays {
			if overlayObject, err := provider.Read(overlay, field, localContext); err == nil {
				return overlayObject, nil
			}
		}
	}
	return provider.Read(object, field, localContext)
}

func (s *ConfigsManager) GetObject(object string, overlays []string, localContext interface{}) (map[string]interface{}, error) {
	log.Infof(" M (Config): GetObject %v, %d", object, len(s.ConfigProviders))

	if strings.Index(object, ":") > 0 && len(s.ConfigProviders) > 1 {
		parts := strings.Split(object, ":")
		if len(parts) != 2 {
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return s.getObjectWithOverlay(provider, parts[1], overlays, localContext)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			if value, err := s.getObjectWithOverlay(provider, object, overlays, localContext); err == nil {
				return value, nil
			} else {
				return nil, err
			}
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if value, err := s.getObjectWithOverlay(provider, object, overlays, localContext); err == nil {
				return value, nil
			}
		}
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object: %s", object), v1alpha2.BadRequest)
}
func (s *ConfigsManager) getObjectWithOverlay(provider config.IConfigProvider, object string, overlays []string, localContext interface{}) (map[string]interface{}, error) {
	if len(overlays) > 0 {
		for _, overlay := range overlays {
			if overlayObject, err := provider.ReadObject(overlay, localContext); err == nil {
				return overlayObject, nil
			}
		}
	}
	return provider.ReadObject(object, localContext)
}
func (s *ConfigsManager) Set(object string, field string, value interface{}) error {
	log.Infof(" M (Config): Set %v, %d", object, len(s.ConfigProviders))

	if strings.Index(object, ":") > 0 && len(s.ConfigProviders) > 1 {
		parts := strings.Split(object, ":")
		if len(parts) != 2 {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return provider.Set(parts[1], field, value)
		}
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return provider.Set(object, field, value)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if err := provider.Set(object, field, value); err == nil {
				return nil
			}
		}
	}
	return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object or key: %s, %s", object, field), v1alpha2.BadRequest)
}
func (s *ConfigsManager) SetObject(object string, values map[string]interface{}) error {
	log.Infof(" M (Config): SetObject %v, %d", object, len(s.ConfigProviders))

	if strings.Index(object, ":") > 0 && len(s.ConfigProviders) > 1 {
		parts := strings.Split(object, ":")
		if len(parts) != 2 {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return provider.SetObject(parts[1], values)
		}
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return provider.SetObject(object, values)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if err := provider.SetObject(object, values); err == nil {
				return nil
			}
		}
	}
	return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object: %s", object), v1alpha2.BadRequest)
}
func (s *ConfigsManager) Delete(object string, field string) error {
	log.Infof(" M (Config): Delete %v, %d", object, len(s.ConfigProviders))

	if strings.Index(object, ":") > 0 && len(s.ConfigProviders) > 1 {
		parts := strings.Split(object, ":")
		if len(parts) != 2 {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return provider.Remove(parts[1], field)
		}
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return provider.Remove(object, field)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if err := provider.Remove(object, field); err == nil {
				return nil
			}
		}
	}
	return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object or key: %s, %s", object, field), v1alpha2.BadRequest)
}
func (s *ConfigsManager) DeleteObject(object string) error {
	log.Infof(" M (Config): DeleteObject %v, %d", object, len(s.ConfigProviders))

	if strings.Index(object, ":") > 0 && len(s.ConfigProviders) > 1 {
		parts := strings.Split(object, ":")
		if len(parts) != 2 {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
		}
		if provider, ok := s.ConfigProviders[parts[0]]; ok {
			return provider.RemoveObject(parts[1])
		}
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
	}
	if len(s.ConfigProviders) == 1 {
		for _, provider := range s.ConfigProviders {
			return provider.RemoveObject(object)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.ConfigProviders[key]; ok {
			if err := provider.RemoveObject(object); err == nil {
				return nil
			}
		}
	}
	return v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid config object: %s", object), v1alpha2.BadRequest)
}
