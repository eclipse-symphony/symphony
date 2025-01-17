//go:build !azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package helper

func GetInstanceSolutionName(name string) string {
	return name
}

func GetInstanceRootResource(name string) string {
	return ""
}
