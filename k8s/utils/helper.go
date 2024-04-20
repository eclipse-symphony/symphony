/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"regexp"
)

func IsComponentKey(key string) bool {
	regex := regexp.MustCompile(`^targets\.[^.]+\.[^.]+`)
	return regex.MatchString(key)
}
