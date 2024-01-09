/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package activations

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

const (
	// DefaultRetentionInMinutes is the default time to cleanup completed activations
	DefaultRetentionInMinutes = 1440
)

type ActivationsCleanupManager struct {
	ActivationsManager
	RetentionInMinutes int
}

func (s *ActivationsCleanupManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.ActivationsManager.Init(context, config, providers)
	if err != nil {
		return err
	}

	// Set activation cleanup interval after they are done. If not set, use default 60 minutes.
	if val, ok := config.Properties["RetentionInMinutes"]; ok {
		s.RetentionInMinutes, err = strconv.Atoi(val)
		if err != nil {
			s.RetentionInMinutes = DefaultRetentionInMinutes
		}
	} else {
		s.RetentionInMinutes = DefaultRetentionInMinutes
	}
	log.Info("M (Activation Cleanup): Initialize RetentionInMinutes as " + fmt.Sprint(s.RetentionInMinutes))
	return nil
}

func (s *ActivationsCleanupManager) Enabled() bool {
	return true
}

func (s *ActivationsCleanupManager) Poll() []error {
	log.Info("M (Activation Cleanup): Polling activations")
	activations, err := s.ActivationsManager.ListSpec(context.Background())
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
			err = s.ActivationsManager.ReportStatus(context.Background(), activation.Id, *activation.Status)
			if err != nil {
				// Delete activation immediately if update time cannot be set? Cx may be confused why activations disappeared
				// Just leave those activations as it is and let Cx delete them manually
				log.Error("M (Activation Cleanup): Cannot set update time for activation "+activation.Id+" since update time cannot be set: %+v", err)
				ret = append(ret, err)
			}
			continue
		}

		// Check update time of completed activations.
		updateTime, err := time.Parse(time.RFC3339, activation.Status.UpdateTime)
		if err != nil {
			// TODO: should not happen, force update time to Time.Now() ?
			log.Info("M (Activation Cleanup): Cannot parse update time of " + activation.Id)
			ret = append(ret, err)
		}
		duration := time.Since(updateTime)
		if duration > time.Duration(s.RetentionInMinutes)*time.Minute {
			log.Info("M (Activation Cleanup): Deleting activation " + activation.Id + " since it has completed for " + duration.String())
			err = s.ActivationsManager.DeleteSpec(context.Background(), activation.Id)
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
