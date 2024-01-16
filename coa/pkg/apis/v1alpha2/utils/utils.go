/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
)

func UnmarshalDuration(duration string) (time.Duration, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(duration), &v); err != nil {
		return 1 * time.Millisecond, err
	}
	switch value := v.(type) {
	case float64:
		return time.Duration(value), nil
	case string:
		ret, err := time.ParseDuration(value)
		if err != nil {
			return 1 * time.Microsecond, err
		}
		return ret, nil
	default:
		return 1 * time.Microsecond, errors.New("invalid duration format")
	}
}

func ParseProperty(val string) string {
	if strings.HasPrefix(val, "$env:") {
		return os.Getenv(val[5:])
	}
	return val
}

type EvaluationContext struct {
	ConfigProvider config.IExtConfigProvider
	SecretProvider secret.ISecretProvider
	DeploymentSpec interface{}
	Properties     map[string]string
	Inputs         map[string]interface{}
	Outputs        map[string]map[string]interface{}
	Component      string
	Value          interface{}
	Scope          string
}

func (e *EvaluationContext) Clone() *EvaluationContext {
	// The Clone() method shares references to the same ConfigProvider and SecretProvider
	// Other fields are not shared and need to be filled in by the caller
	if e == nil {
		return nil
	}
	return &EvaluationContext{
		ConfigProvider: e.ConfigProvider,
		SecretProvider: e.SecretProvider,
	}
}
