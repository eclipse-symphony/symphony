/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSiteMatch(t *testing.T) {
	s1 := SiteSpec{
		Name:      "site",
		PublicKey: "publicKey",
		Properties: map[string]string{
			"foo": "bar",
		},
	}
	s2 := SiteSpec{
		Name:      "site",
		PublicKey: "publicKey",
		Properties: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestSiteNameNotMatch(t *testing.T) {
	s1 := SiteSpec{
		Name: "site",
		Properties: map[string]string{
			"foo": "bar",
		},
	}
	s2 := SiteSpec{
		Name: "site2",
		Properties: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestSitePublicKeyNotMatch(t *testing.T) {
	s1 := SiteSpec{
		Name:      "site",
		PublicKey: "publicKey",
	}
	s2 := SiteSpec{
		Name:      "site",
		PublicKey: "publicKey2",
	}
	equal, err := s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestSiteEqualNil(t *testing.T) {
	s1 := SiteSpec{
		Name: "site",
		Properties: map[string]string{
			"foo": "bar",
		},
	}
	res, err := s1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a SiteSpec type")
	assert.False(t, res)
}
