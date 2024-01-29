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

func TestEdgeMatchOneEmpty(t *testing.T) {
	edge1 := EdgeSpec{
		Source: ConnectionSpec{
			Node:  "node1",
			Route: "route1",
		},
		Target: ConnectionSpec{
			Node:  "node2",
			Route: "route1",
		},
	}
	res, err := edge1.DeepEquals(nil)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConnectionSpecMatchOneEmpty(t *testing.T) {
	conn1 := ConnectionSpec{
		Node:  "node1",
		Route: "route1",
	}
	res, err := conn1.DeepEquals(nil)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConnectionSpecNodeNotMatch(t *testing.T) {
	conn1 := &ConnectionSpec{
		Node: "node1",
	}
	conn2 := &ConnectionSpec{
		Node: "node2",
	}
	res, err := conn1.DeepEquals(conn2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConnectionSpecRouteNotMatch(t *testing.T) {
	conn1 := &ConnectionSpec{
		Route: "route1",
	}
	conn2 := &ConnectionSpec{
		Route: "route2",
	}
	res, err := conn1.DeepEquals(conn2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestConnectionSpecEqual(t *testing.T) {
	conn1 := &ConnectionSpec{
		Node:  "node",
		Route: "route",
	}
	conn2 := &ConnectionSpec{
		Node:  "node",
		Route: "route",
	}
	res, err := conn1.DeepEquals(conn2)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestEdgeSpecSourceNotMatch(t *testing.T) {
	edge1 := &EdgeSpec{
		Source: ConnectionSpec{
			Node: "node1",
		},
	}
	edge2 := &EdgeSpec{
		Source: ConnectionSpec{
			Node: "node2",
		},
	}
	res, err := edge1.DeepEquals(edge2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestEdgeSpecTargetNotMatch(t *testing.T) {
	edge1 := &EdgeSpec{
		Target: ConnectionSpec{
			Node: "node1",
		},
	}
	edge2 := &EdgeSpec{
		Target: ConnectionSpec{
			Node: "node2",
		},
	}
	res, err := edge1.DeepEquals(edge2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestEdgeSpecEqual(t *testing.T) {
	edge1 := &EdgeSpec{
		Source: ConnectionSpec{
			Node:  "node1",
			Route: "route1",
		},
		Target: ConnectionSpec{
			Node:  "node2",
			Route: "route1",
		},
	}
	edge2 := &EdgeSpec{
		Source: ConnectionSpec{
			Node:  "node1",
			Route: "route1",
		},
		Target: ConnectionSpec{
			Node:  "node2",
			Route: "route1",
		},
	}
	res, err := edge1.DeepEquals(edge2)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestEdgeArrayEqual(t *testing.T) {
	e1 := EdgeSpec{
		Source: ConnectionSpec{
			Node:  "node1",
			Route: "route",
		},
		Target: ConnectionSpec{
			Node:  "node2",
			Route: "route",
		},
	}
	e2 := EdgeSpec{
		Source: ConnectionSpec{
			Node:  "node1",
			Route: "route",
		},
		Target: ConnectionSpec{
			Node:  "node2",
			Route: "route",
		},
	}
	res, err := e1.DeepEquals(&e2)
	assert.Nil(t, err)
	assert.True(t, res)

	es1 := make([]*EdgeSpec, 0)
	es1 = append(es1, &e1)
	es2 := make([]*EdgeSpec, 0)
	es2 = append(es2, &e2)
	res = SlicesEqual(es1, es2)
	assert.True(t, res)
}
