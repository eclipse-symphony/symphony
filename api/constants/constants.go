//go:build !azure
// +build !azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package constants

const (
	EulaMessage = `MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE`
	DefaultScope = "default"
	SATokenPath  = "/var/run/secrets/tokens/symphony-api-token"
	// These constants need to be in a shared package.
	GroupPrefix     = "iotoperations.azure.com"
	ManagerMetaKey  = GroupPrefix + "/managed-by"
	InstanceMetaKey = GroupPrefix + "/instance"
)

// Environment variables keys
const (
	SymphonyCertEnvName           = "SYMPHONY_ROOT_CA"
	SATokenPathName               = "SA_TOKEN_PATH"
	ApiCertEnvName                = "API_SERVING_CA"
	UseServiceAccountTokenEnvName = "USE_SERVICE_ACCOUNT_TOKENS"
	SymphonyAPIUrlEnvName         = "SYMPHONY_API_URL"
	API                           = "symphony-api"
)
