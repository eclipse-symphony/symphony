/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package activations

import (
	"context"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

const (
	// DefaultRetentionDuration is the default time to cleanup completed activations
	DefaultRetentionDuration = 180 * time.Hour * 24
)

type ActivationsCleanupManager struct {
	ActivationsManager
	RetentionDuration time.Duration
}

func (s *ActivationsCleanupManager) Init(ctx *vendorCtx.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	s.ActivationsManager = ActivationsManager{}
	err := s.ActivationsManager.Init(ctx, config, providers)
	if err != nil {
		return err
	}

	// Set activation cleanup interval after they are done. If not set, use default 180 days.
	if val, ok := config.Properties["RetentionDuration"]; ok {
		s.RetentionDuration, err = time.ParseDuration(val)
		if err != nil {
			return v1alpha2.NewCOAError(nil, "RetentionDuration cannot be parsed, please enter a valid duration", v1alpha2.BadConfig)
		} else if s.RetentionDuration < 0 {
			return v1alpha2.NewCOAError(nil, "RetentionDuration cannot be negative", v1alpha2.BadConfig)
		}
	} else {
		s.RetentionDuration = DefaultRetentionDuration
	}

	log.Info("M (Activation Cleanup): Initialize RetentionDuration as " + s.RetentionDuration.String())
	return nil
}

func (s *ActivationsCleanupManager) Enabled() bool {
	return true
}

func (s *ActivationsCleanupManager) Poll() []error {
	// TODO: initialize the context with id correctly
	ctx, span := observability.StartSpan("Activations Cleanup Manager", context.Background(), &map[string]string{
		"method": "Poll",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.InfoCtx(ctx, "M (Activation Cleanup): Polling activations")
	activations, err := s.ActivationsManager.ListState(ctx, "")
	if err != nil {
		return []error{err}
	}
	ret := []error{}
	for _, activation := range activations {
		if activation.Status.Status != v1alpha2.Done {
			continue
		}
		if activation.Status.UpdateTime == "" {
			// Ugrade scenario: update time is not set for activations created before. Set it to now and the activation will be deleted later.
			// UpdateTime will be set in ReportStatus function
			err = s.ActivationsManager.ReportStatus(ctx, activation.ObjectMeta.Name, activation.ObjectMeta.Namespace, *activation.Status)
			if err != nil {
				// Delete activation immediately if update time cannot be set? Cx may be confused why activations disappeared
				// Just leave those activations as it is and let Cx delete them manually
				log.ErrorfCtx(ctx, "M (Activation Cleanup): Cannot set update time for activation %s since update time cannot be set: %+v", activation.ObjectMeta.Name, err)
				ret = append(ret, err)
			}
			continue
		}

		// Check update time of completed activations.
		updateTime, err := time.Parse(time.RFC3339, activation.Status.UpdateTime)
		if err != nil {
			// TODO: should not happen, force update time to Time.Now() ?
			log.InfofCtx(ctx, "M (Activation Cleanup): Cannot parse update time of %s", activation.ObjectMeta.Name)
			ret = append(ret, err)
		}
		duration := time.Since(updateTime)
		if duration > s.RetentionDuration {
			log.InfofCtx(ctx, "M (Activation Cleanup): Deleting activation %s since it has completed for %s", activation.ObjectMeta.Name, duration.String())
			err = s.ActivationsManager.DeleteState(ctx, activation.ObjectMeta.Name, activation.ObjectMeta.Namespace)
			if err != nil {
				ret = append(ret, err)
			}
		}
	}
	return ret
}

func (s *ActivationsCleanupManager) Reconcil() []error {
	return nil
}
