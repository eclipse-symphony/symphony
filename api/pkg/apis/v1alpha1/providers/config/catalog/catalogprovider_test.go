/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalog

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	// To make this test work, you'll need these configurations:
	// ai-config: flavor=cloud, model=gpt, version=4.5
	// ai-config-site: model=LLaMA, version=3.3
	// ai-config-line: flavor=mobile
	// combined: ai=<ai-config>, ai-mode=<ai-config>.model, com:<combined-1>.foo, less=<123, e4k=<e4k-config>, influxdb=<influx-db-config>, loop=<combined-1>.loop
	// combined-1: foo=<combined-2>.foo, loop=<combined-2>.loop
	// combined-2: foo=bar2, loop=<combined>.loop
	// os.Setenv("CATALOG_API_URL", "http://localhost:8080/v1alpha2/")
	// os.Setenv("CATALOG_API_USER", "admin")
	// os.Setenv("CATALOG_API_PASSWORD", "")
	catalogAPIUrl := os.Getenv("CATALOG_API_URL")
	if catalogAPIUrl == "" {
		t.Skip("Skipping becasue CATALOG_API_URL is missing or not set to 'yes'")
	}
	catalogAPIUser := os.Getenv("CATALOG_API_USER")
	catalogAPIPassword := os.Getenv("CATALOG_API_PASSWORD")

	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{BaseUrl: catalogAPIUrl, User: catalogAPIUser, Password: catalogAPIPassword})
	assert.Nil(t, err)

	value, err := provider.Read("ai-config", "model", nil)
	assert.Nil(t, err)
	assert.Equal(t, "gpt", value)
	value, err = provider.Read("ai-config-site", "model", nil)
	assert.Nil(t, err)
	assert.Equal(t, "LLaMA", value)
	value, err = provider.Read("ai-config-line", "model", nil)
	assert.Nil(t, err)
	assert.Equal(t, "LLaMA", value)
	value, err = provider.Read("ai-config", "flavor", nil)
	assert.Nil(t, err)
	assert.Equal(t, "cloud", value)
	value, err = provider.Read("ai-config-site", "flavor", nil)
	assert.Nil(t, err)
	assert.Equal(t, "cloud", value)
	value, err = provider.Read("ai-config-line", "flavor", nil)
	assert.Nil(t, err)
	assert.Equal(t, "mobile", value)
	value, err = provider.Read("ai-config", "version", nil)
	assert.Nil(t, err)
	assert.Equal(t, "4.5", value)
	value, err = provider.Read("ai-config-site", "version", nil)
	assert.Nil(t, err)
	assert.Equal(t, "3.3", value)
	value, err = provider.Read("ai-config-line", "version", nil)
	assert.Nil(t, err)
	assert.Equal(t, "3.3", value)
	value, err = provider.Read("combined", "ai-model", nil)
	assert.Nil(t, err)
	assert.Equal(t, "gpt", value)
	value, err = provider.Read("combined", "ai", nil)
	assert.Nil(t, err)
	assert.Equal(t, "{\"flavor\":\"cloud\",\"model\":\"gpt\",\"version\":\"4.5\"}", value)
	value, err = provider.Read("combined", "com", nil)
	assert.Nil(t, err)
	assert.Equal(t, "bar2", value)
	value, err = provider.Read("combined", "less", nil)
	assert.Nil(t, err)
	assert.Equal(t, "<123", value)
	// TODO: needs a good way to detect reference loops
	// value, err = provider.Read("combined", "loop")
	// assert.NotNil(t, err)
}
