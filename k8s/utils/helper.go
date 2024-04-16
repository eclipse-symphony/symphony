/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"regexp"

	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

func IsComponentKey(key string) bool {
	regex := regexp.MustCompile(`^targets\.[^.]+\.[^.]+`)
	return regex.MatchString(key)
}

func IsDeploymentFinished(summary apimodel.SummaryResult) bool {
	return summary.State == apimodel.SummaryStateDone
}
