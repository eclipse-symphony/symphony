//go:build azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package helper

import (
	"fmt"
	"strings"
)

func GetInstanceSolutionName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return ""
	}
	version := parts[len(parts)-1]
	solution := parts[len(parts)-3]
	return fmt.Sprintf("%s:%s", solution, version)
}

func GetInstanceRootResource(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return ""
	}
	return parts[len(parts)-3]
}
