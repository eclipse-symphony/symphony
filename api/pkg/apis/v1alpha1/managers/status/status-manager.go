/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package status

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	apiutils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
)

const ReportedAtProperty = "agent.reportedAt"

type StatusManager struct {
	managers.Manager
	Reporter  reporter.IReporter
	target    string
	namespace string
	bag       map[string]string
}

func (s *StatusManager) Init(ctx *contexts.VendorContext, config managers.ManagerConfig, ps map[string]providers.IProvider) error {
	if err := s.Manager.Init(ctx, config, ps); err != nil {
		return err
	}
	var err error
	s.Reporter, err = managers.GetReporter(config, ps)
	if err != nil {
		return err
	}
	s.target, err = localValue(config.Properties["target.name"])
	if err != nil || strings.TrimSpace(s.target) == "" {
		return v1alpha2.NewCOAError(err, "status manager requires target.name", v1alpha2.BadConfig)
	}
	s.namespace = constants.DefaultScope
	if raw, ok := config.Properties["target.namespace"]; ok {
		s.namespace, err = localValue(raw)
		if err != nil || strings.TrimSpace(s.namespace) == "" {
			return v1alpha2.NewCOAError(err, "status manager requires a nonempty target.namespace", v1alpha2.BadConfig)
		}
	}
	var bag map[string]interface{}
	if err := json.Unmarshal([]byte(config.Properties["report.properties"]), &bag); err != nil {
		return v1alpha2.NewCOAError(err, "report.properties must be a JSON object of strings", v1alpha2.BadConfig)
	}
	if bag == nil {
		return v1alpha2.NewCOAError(nil, "report.properties must be a JSON object of strings", v1alpha2.BadConfig)
	}
	s.bag = make(map[string]string, len(bag))
	for key, raw := range bag {
		if strings.TrimSpace(key) == "" || key == ReportedAtProperty {
			return v1alpha2.NewCOAError(nil, "report.properties contains an empty or reserved property name", v1alpha2.BadConfig)
		}
		value, ok := raw.(string)
		if !ok {
			return v1alpha2.NewCOAError(nil, "report.properties values must be strings", v1alpha2.BadConfig)
		}
		s.bag[key] = value
	}
	if enabled := config.Properties["poll.enabled"]; enabled != "" && enabled != "true" && enabled != "false" {
		return v1alpha2.NewCOAError(nil, "poll.enabled must be true or false", v1alpha2.BadConfig)
	}
	return nil
}

func localValue(value string) (string, error) {
	if strings.HasPrefix(value, "$env:") {
		name := strings.TrimPrefix(value, "$env:")
		resolved, ok := os.LookupEnv(name)
		if !ok {
			return "", fmt.Errorf("environment variable %q is not set", name)
		}
		return resolved, nil
	}
	return value, nil
}

func (s *StatusManager) Enabled() bool {
	return s.Config.Properties["poll.enabled"] == "true"
}

func (s *StatusManager) Poll() []error {
	if err := s.Report(context.Background()); err != nil {
		s.Context.Logger.Errorf(" M (Status): failed to report status for target %s/%s: %v", s.namespace, s.target, err)
		return []error{err}
	}
	return nil
}

func (s *StatusManager) Reconcil() []error {
	return nil
}

func (s *StatusManager) Report(ctx context.Context) error {
	eval := utils.EvaluationContext{}
	if s.VendorContext != nil && s.VendorContext.EvaluationContext != nil {
		eval = *s.VendorContext.EvaluationContext.Clone()
	}
	eval.Context = ctx
	eval.Namespace = s.namespace
	eval.Properties = make(map[string]string, len(s.bag))
	for key, value := range s.bag {
		resolved, err := localValue(value)
		if err != nil {
			return fmt.Errorf("evaluate status property %q: %w", key, err)
		}
		eval.Properties[key] = resolved
	}
	properties := make(map[string]string, len(s.bag)+1)
	for key, expression := range s.bag {
		value := eval.Properties[key]
		// Environment values are data, not expressions to evaluate a second time.
		if !strings.HasPrefix(expression, "$env:") && strings.Contains(expression, "${{") {
			result, err := apiutils.NewParser(expression).Eval(eval)
			if err != nil {
				return fmt.Errorf("evaluate status property %q: %w", key, err)
			}
			if result == nil {
				return fmt.Errorf("status property %q evaluated to null", key)
			}
			value = apiutils.FormatAsString(result)
		}
		properties[key] = value
	}
	properties[ReportedAtProperty] = time.Now().UTC().Format(time.RFC3339Nano)
	return s.Reporter.Report(s.target, s.namespace, model.FabricGroup, "targets", "v1", properties, false)
}

func (s *StatusManager) Shutdown(ctx context.Context) error {
	if terminable, ok := s.Reporter.(interface{ Shutdown(context.Context) error }); ok {
		return terminable.Shutdown(ctx)
	}
	return nil
}
