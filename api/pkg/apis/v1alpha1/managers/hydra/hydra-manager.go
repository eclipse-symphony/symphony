/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package hydra

import (
	"context"
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/hydra"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type HydraManager struct {
	managers.Manager
	HydraProviders map[string]hydra.IHydraProvider
	apiClient      utils.ApiClient
	user           string
	password       string
}

func (s *HydraManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	s.HydraProviders = make(map[string]hydra.IHydraProvider)
	for k, p := range providers {
		if c, ok := p.(hydra.IHydraProvider); ok {
			s.HydraProviders[k] = c

		}
	}
	if len(s.HydraProviders) == 0 {
		return v1alpha2.NewCOAError(nil, "Hydra providers are not supplied", v1alpha2.MissingConfig)
	}
	if utils.ShouldUseUserCreds() {
		user, err := utils.GetString(s.Manager.Config.Properties, "user")
		if err != nil {
			return err
		}
		s.user = user
		if s.user == "" {
			return v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		}
		password, err := utils.GetString(s.Manager.Config.Properties, "password")
		if err != nil {
			return err
		}
		s.password = password
	}
	s.apiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
	return nil
}

func (s *HydraManager) GetArtifacts(ctx context.Context, system string, objType string, key string) ([]byte, error) {
	ctx, span := observability.StartSpan("Hydra Manager", ctx, &map[string]string{
		"method": "GetArtifacts",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	if provider, ok := s.HydraProviders[system]; ok {
		artifactPack, err := s.CollectArtifactPack(ctx, system, objType, key)
		if err != nil {
			log.ErrorfCtx(ctx, "  M (Hydra): GetArtifacts failed - %s", err.Error())
			return nil, err
		}
		ret, err := provider.GetArtifact(system, artifactPack)
		if err != nil {
			log.ErrorfCtx(ctx, "  M (Hydra): GetArtifacts failed - %s", err.Error())
			return nil, err
		}
		return ret, nil
	}

	err = v1alpha2.NewCOAError(nil, "Hydra provider not found for system: "+system, v1alpha2.NotFound)
	log.ErrorfCtx(ctx, "  M (Hydra): GetArtifacts failed - %s", err.Error())
	return nil, err
}

func (s *HydraManager) SetArtifacts(system string, artifacts []byte) error {
	ctx, span := observability.StartSpan("Hydra Manager", context.TODO(), &map[string]string{
		"method": "SetArtifacts",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	if provider, ok := s.HydraProviders[system]; ok {
		artifactPack, err := provider.SetArtifact(system, artifacts)
		if err != nil {
			log.ErrorfCtx(ctx, "  M (Hydra): SetArtifacts failed - %s", err.Error())
			return err
		}
		return s.ApplyArtifactPack(ctx, system, artifactPack)
	}
	err = v1alpha2.NewCOAError(nil, "Hydra provider not found for system: "+system, v1alpha2.NotFound)
	log.ErrorfCtx(ctx, "  M (Hydra): SetArtifacts failed - %s", err.Error())
	return err
}

func (s *HydraManager) CollectArtifactPack(ctx context.Context, system string, objType string, key string) (model.ArtifactPack, error) {
	ctx, span := observability.StartSpan("Hydra Manager", ctx, &map[string]string{
		"method": "CollectArtifactPack",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	ret := model.ArtifactPack{
		Targets: []model.TargetState{},
	}
	targets, err := s.apiClient.GetHydraTargetsForAllNamespaces(ctx, s.user, s.password, system, objType, key)
	if err != nil {
		log.ErrorfCtx(ctx, "  M (Hydra): CollectArtifactPack failed - %s", err.Error())
		return model.ArtifactPack{}, err
	}
	ret.Targets = append(ret.Targets, targets...)
	return ret, nil
}

func (s *HydraManager) ApplyArtifactPack(ctx context.Context, system string, artifactPack model.ArtifactPack) error {
	ctx, span := observability.StartSpan("Hydra Manager", ctx, &map[string]string{
		"method": "ApplyArtifactPack",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	for _, solution := range artifactPack.Solutions {
		payload, _ := json.Marshal(solution)
		err = s.apiClient.CreateSolution(ctx, solution.ObjectMeta.Name, payload, solution.ObjectMeta.Namespace, s.user, s.password)
		if err != nil {
			log.ErrorfCtx(ctx, "  M (Hydra): ApplyArtifactPack failed - %s", err.Error())
			return err
		}
	}
	for _, target := range artifactPack.Targets {
		payload, _ := json.Marshal(target)
		err = s.apiClient.CreateTarget(ctx, target.ObjectMeta.Name, payload, target.ObjectMeta.Namespace, s.user, s.password)
		if err != nil {
			log.ErrorfCtx(ctx, "  M (Hydra): ApplyArtifactPack failed - %s", err.Error())
			return err
		}
	}

	for _, instance := range artifactPack.Instances {
		payload, _ := json.Marshal(instance)
		err = s.apiClient.CreateInstance(ctx, instance.ObjectMeta.Name, payload, instance.ObjectMeta.Namespace, s.user, s.password)
		if err != nil {
			log.ErrorfCtx(ctx, "  M (Hydra): ApplyArtifactPack failed - %s", err.Error())
			return err
		}
	}

	return nil
}
