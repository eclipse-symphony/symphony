/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package remoteAgent

import (
	"context"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/targets"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	// DefaultRetentionDuration is the default time to cleanup completed activations
	DefaultRetentionDuration = 30 * time.Minute
)

type RemoteTargetSchedulerManager struct {
	managers.Manager
	RetentionDuration time.Duration
	TargetsManager    *targets.TargetsManager
}

var log = logger.NewLogger("RemoteTargetSchedulerManager")
var desiredVersion = os.Getenv("AGENT_VERSION")

func (s *RemoteTargetSchedulerManager) Init(ctx *vendorCtx.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	// initialize the target manager
	s.TargetsManager = &targets.TargetsManager{}
	err := s.TargetsManager.Init(ctx, config, providers)
	if err != nil {
		return err
	}
	// Set scheduler interval after they are done. If not set, use default 30 miniutes.
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

	log.Info("M (RemoteTarget Scheduler): Initialize RetentionDuration as " + s.RetentionDuration.String())
	return nil
}

func (s *RemoteTargetSchedulerManager) Enabled() bool {
	return true
}

func (s *RemoteTargetSchedulerManager) Poll() []error {
	// TODO: initialize the context with id correctly
	ctx, span := observability.StartSpan("RemoteTarget Scheduler Manager", context.Background(), &map[string]string{
		"method": "Poll",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.InfoCtx(ctx, "M (RemoteTarget Scheduler): Polling targets")
	targets, err := s.TargetsManager.ListState(ctx, "")
	if err != nil {
		return []error{err}
	}
	ret := []error{}
	for _, target := range targets {
		isRemote := false
		componentName := ""
		components := target.Spec.Components
		for _, component := range components {
			if component.Type == "remote-agent" {
				componentName = component.Name
				isRemote = true
			} else {
				continue
			}
		}
		if !isRemote {
			continue
		}
		remoteAgentStatus, ok := target.Status.Properties[fmt.Sprintf("targets.%s.%s", target.ObjectMeta.Name, componentName)]
		if !ok {
			log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Remote agent status not found for target %s", target.ObjectMeta.Name)
			continue
		}
		var remoteAgentStatusMap map[string]string
		err = json.Unmarshal([]byte(remoteAgentStatus), &remoteAgentStatusMap)
		if err != nil {
			log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Cannot unmarshal remote agent status for target %s", target.ObjectMeta.Name)
			continue
		}

		currentVersion, ok := remoteAgentStatusMap["version"]
		if !ok {
			log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Version not found for target %s", target.ObjectMeta.Name)
			continue
		}

		certificateExpiration, ok := remoteAgentStatusMap["certificateExpiration"]
		if !ok {
			log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): certificateExpiration not found for target %s", target.ObjectMeta.Name)
			continue
		}

		if currentVersion != os.Getenv("AGENT_VERSION") {
			err = s.updateTargetToIssueUpgradeJob(ctx, target, componentName)
			if err != nil {
				log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Cannot issue upgrade job for target %s", target.ObjectMeta.Name)
				ret = append(ret, err)
			}
		}

		secretName := fmt.Sprintf("%s-tls", target.ObjectMeta.Name)
		cert, err := s.TargetsManager.SecretProvider.Read(ctx, secretName, "tls.crt", coa_utils.EvaluationContext{Namespace: target.ObjectMeta.Namespace})
		if err != nil {
			log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Cannot read certificate secret expiration for target %s", target.ObjectMeta.Name)
			continue
		}

		// decode cert and get the expiration date
		certSecretExpiration, err := s.getCertificateExpirationOrThumbPrint(cert, "expiration")
		if err != nil {
			log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Cannot get certificate expiration for target %s", target.ObjectMeta.Name)
			continue
		}

		// use the same format and timezone for both dates
		certificateExpirationTime, err := time.Parse(time.RFC3339, certificateExpiration)
		if err != nil {
			log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Cannot parse certificate expiration for target %s", target.ObjectMeta.Name)
			continue
		}
		certSecretExpirationTime, err := time.Parse(time.RFC3339, certSecretExpiration)
		if err != nil {
			log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Cannot parse certificate secret expiration for target %s", target.ObjectMeta.Name)
			continue
		}
		if certificateExpirationTime.Before(certSecretExpirationTime) {
			thumbprint, err := s.getCertificateExpirationOrThumbPrint(cert, "thumbprint")
			if err != nil {
				log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Cannot get certificate thumbprint for target %s", target.ObjectMeta.Name)
				continue
			}
			err = s.updateTargetToIssueSRJob(ctx, target, componentName, thumbprint)
			if err != nil {
				log.WarnfCtx(ctx, "M (RemoteTarget Scheduler): Cannot issue SR job for target %s", target.ObjectMeta.Name)
				ret = append(ret, err)
			}
		}
	}
	return ret
}

func (s *RemoteTargetSchedulerManager) Reconcil() []error {
	return nil
}

func (s *RemoteTargetSchedulerManager) getCertificateExpirationOrThumbPrint(certPEM string, kind string) (string, error) {
	certBytes := []byte(certPEM)
	block, _ := pem.Decode(certBytes)
	if block == nil {
		return "", fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	if kind == "thumbprint" {
		thumbprint := sha1.Sum(cert.Raw)
		return hex.EncodeToString(thumbprint[:]), nil
	} else {
		return cert.NotAfter.Format(time.RFC3339), nil
	}
}

func (s *RemoteTargetSchedulerManager) updateTargetToIssueUpgradeJob(ctx context.Context, target model.TargetState, componentName string) error {
	log.InfofCtx(ctx, "M (RemoteTarget Scheduler): Issuing upgrade job for target %s", target.ObjectMeta.Name)
	// update the target spec component to issue upgrade job
	var newComponents []model.ComponentSpec
	components := target.Spec.Components
	for _, component := range components {
		if component.Type == "remote-agent" && component.Name == componentName {
			newComponents = append(newComponents, model.ComponentSpec{
				Name: component.Name,
				Type: component.Type,
				Properties: map[string]interface{}{
					"action":  "upgrade",
					"version": os.Getenv("AGENT_VERSION"),
				},
			})
		} else {
			newComponents = append(newComponents, component)
		}
	}
	target.Spec.Components = newComponents
	err := s.TargetsManager.UpsertState(ctx, target.ObjectMeta.Name, target)
	return err
}

func (s *RemoteTargetSchedulerManager) updateTargetToIssueSRJob(ctx context.Context, target model.TargetState, componentName string, thumbprint string) error {
	log.InfofCtx(ctx, "M (RemoteTarget Scheduler): Issuing SR job for target %s", target.ObjectMeta.Name)
	// update the target spec component to issue upgrade job
	var newComponents []model.ComponentSpec
	components := target.Spec.Components
	for _, component := range components {
		if component.Type == "remote-agent" && component.Name == componentName {
			newComponents = append(newComponents, model.ComponentSpec{
				Name: component.Name,
				Type: component.Type,
				Properties: map[string]interface{}{
					"action":     "secretrotation",
					"thumbprint": thumbprint,
				},
			})
		} else {
			newComponents = append(newComponents, component)
		}
	}
	target.Spec.Components = newComponents
	err := s.TargetsManager.UpsertState(ctx, target.ObjectMeta.Name, target)
	return err
}
