/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package agent

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

type Agent struct {
	Providers map[string]tgt.ITargetProvider
	RLog      logger.Logger
}

func (s Agent) Handle(req []byte, ctx context.Context) model.AsyncResult {
	request := model.AgentRequest{}
	err := json.Unmarshal(req, &request)
	if err != nil {
		return model.AsyncResult{OperationID: request.OperationID, Error: err}
	}

	body := new([]byte)

	provider := s.Providers[request.Provider]
	if provider == nil {
		return model.AsyncResult{OperationID: request.OperationID, Error: errors.New("provider not found")}
	}

	var result model.AsyncResult

	switch request.Action {
	case "get":
		var getRequest model.ProviderGetRequest
		if err := json.Unmarshal(req, &getRequest); err != nil {
			result = model.AsyncResult{OperationID: request.OperationID, Error: err}
		} else {
			specs, err := provider.Get(ctx, getRequest.Deployment, getRequest.References)
			if err != nil {
				s.RLog.ErrorfCtx(ctx, "error getting specs: %v", err)
			}
			*body, err = json.Marshal(specs)
			if err != nil {
				s.RLog.ErrorfCtx(ctx, "error marshalling specs: %v", err)
			}
			result = model.AsyncResult{OperationID: request.OperationID, Body: *body, Error: err}
		}

	case "apply":
		var applyRequest model.ProviderApplyRequest
		if err := json.Unmarshal(req, &applyRequest); err != nil {
			result = model.AsyncResult{OperationID: request.OperationID, Error: err}
		} else {
			specs, err := provider.Apply(ctx, applyRequest.Deployment, applyRequest.Step, applyRequest.Deployment.IsDryRun)
			if err != nil {
				s.RLog.ErrorfCtx(ctx, "error applying specs: %v", err)
			}
			*body, err = json.Marshal(specs)
			if err != nil {
				s.RLog.ErrorfCtx(ctx, "error marshalling specs: %v", err)
			}
			result = model.AsyncResult{OperationID: request.OperationID, Body: *body, Error: err}
		}

	case "getValidationRule":
		var getValidationRuleRequest model.ProviderGetValidationRuleRequest
		if err := json.Unmarshal(req, &getValidationRuleRequest); err != nil {
			result = model.AsyncResult{OperationID: request.OperationID, Error: err}
		} else {
			rule := provider.GetValidationRule(ctx)
			*body, err = json.Marshal(rule)
			if err != nil {
				s.RLog.ErrorfCtx(ctx, "error marshalling validation rule: %v", err)
			}
			result = model.AsyncResult{OperationID: request.OperationID, Body: *body, Error: err}
		}
	default:
		result = model.AsyncResult{OperationID: request.OperationID, Error: errors.New("action not found")}
	}

	return result
}
