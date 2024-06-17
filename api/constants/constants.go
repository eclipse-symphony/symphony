//go:build !azure
// +build !azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package constants

import _ "embed"

// Eula Message
var (
	//go:embed eula.txt
	EulaMessage string
)

const (
	DefaultScope = "default"
	SATokenPath  = "/var/run/secrets/tokens/symphony-api-token"
	// These constants need to be in a shared package.
	GroupPrefix     = "symphony"
	ManagerMetaKey  = GroupPrefix + "/managed-by"
	InstanceMetaKey = GroupPrefix + "/instance"
	NotFound        = "Not Found"
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
