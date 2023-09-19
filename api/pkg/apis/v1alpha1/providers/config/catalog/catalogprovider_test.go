/*

	MIT License

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
	SOFTWARE

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

	value, err := provider.Read("ai-config", "model")
	assert.Nil(t, err)
	assert.Equal(t, "gpt", value)
	value, err = provider.Read("ai-config-site", "model")
	assert.Nil(t, err)
	assert.Equal(t, "LLaMA", value)
	value, err = provider.Read("ai-config-line", "model")
	assert.Nil(t, err)
	assert.Equal(t, "LLaMA", value)
	value, err = provider.Read("ai-config", "flavor")
	assert.Nil(t, err)
	assert.Equal(t, "cloud", value)
	value, err = provider.Read("ai-config-site", "flavor")
	assert.Nil(t, err)
	assert.Equal(t, "cloud", value)
	value, err = provider.Read("ai-config-line", "flavor")
	assert.Nil(t, err)
	assert.Equal(t, "mobile", value)
	value, err = provider.Read("ai-config", "version")
	assert.Nil(t, err)
	assert.Equal(t, "4.5", value)
	value, err = provider.Read("ai-config-site", "version")
	assert.Nil(t, err)
	assert.Equal(t, "3.3", value)
	value, err = provider.Read("ai-config-line", "version")
	assert.Nil(t, err)
	assert.Equal(t, "3.3", value)
	value, err = provider.Read("combined", "ai-model")
	assert.Nil(t, err)
	assert.Equal(t, "gpt", value)
	value, err = provider.Read("combined", "ai")
	assert.Nil(t, err)
	assert.Equal(t, "{\"flavor\":\"cloud\",\"model\":\"gpt\",\"version\":\"4.5\"}", value)
	value, err = provider.Read("combined", "com")
	assert.Nil(t, err)
	assert.Equal(t, "bar2", value)
	value, err = provider.Read("combined", "less")
	assert.Nil(t, err)
	assert.Equal(t, "<123", value)
	// TODO: needs a good way to detect reference loops
	// value, err = provider.Read("combined", "loop")
	// assert.NotNil(t, err)
}
