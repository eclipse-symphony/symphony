/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestMockLedgerProviderInit(t *testing.T) {
	provider := MockLedgerProvider{}
	properties := map[string]string{
		"name": "test",
	}
	err := provider.InitWithMap(properties)
	provider.SetContext(&contexts.ManagerContext{})
	assert.Nil(t, err)
	assert.Equal(t, "test", provider.ID())
}

func TestMockLedgerProviderAppend(t *testing.T) {
	provider := MockLedgerProvider{}
	properties := map[string]string{
		"name": "test",
	}
	err := provider.InitWithMap(properties)
	assert.Nil(t, err)
	trails := []v1alpha2.Trail{
		{
			Origin:  "o1",
			Catalog: "c1",
			Type:    "t1",
		},
		{
			Origin:  "o2",
			Catalog: "c2",
			Type:    "t2",
		},
	}
	err = provider.Append(context.Background(), trails)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(provider.LedgerData))
	assert.Equal(t, trails[0], provider.LedgerData[0])
	assert.Equal(t, trails[1], provider.LedgerData[1])
	for i := 0; i < 52; i++ {
		err = provider.Append(context.Background(), trails)
		assert.Nil(t, err)
	}
	assert.Equal(t, 100, len(provider.LedgerData))
}
