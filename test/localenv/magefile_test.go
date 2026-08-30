//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"
)

func TestHasStuckRelease(t *testing.T) {
	cases := []struct {
		name    string
		json    string
		want    bool
		wantErr bool
	}{
		{
			name: "no releases is clean",
			json: `[]`,
			want: false,
		},
		{
			name: "deployed release is left alone",
			json: `[{"name":"symphony","namespace":"default","revision":"3","status":"deployed","chart":"symphony-0.1.0"}]`,
			want: false,
		},
		{
			name: "uninstalling release needs cleanup",
			json: `[{"name":"symphony","namespace":"default","revision":"3","status":"uninstalling","chart":"symphony-0.1.0"}]`,
			want: true,
		},
		{
			name: "failed release needs cleanup",
			json: `[{"name":"symphony","namespace":"default","revision":"1","status":"failed","chart":"symphony-0.1.0"}]`,
			want: true,
		},
		{
			name: "pending-install release needs cleanup",
			json: `[{"name":"symphony","namespace":"default","revision":"1","status":"pending-install","chart":"symphony-0.1.0"}]`,
			want: true,
		},
		{
			name:    "invalid json is an error",
			json:    `not-json`,
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := hasStuckRelease([]byte(c.json))
			if (err != nil) != c.wantErr {
				t.Fatalf("hasStuckRelease(%s) error = %v, wantErr %v", c.json, err, c.wantErr)
			}
			if got != c.want {
				t.Errorf("hasStuckRelease(%s) = %v, want %v", c.json, got, c.want)
			}
		})
	}
}

func TestIsNamespaceNotFound(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "helm namespace not found on stderr",
			err:  &exec.ExitError{Stderr: []byte(`Error: namespaces "custom-ns" not found`)},
			want: true,
		},
		{
			name: "wrapped exit error",
			err:  fmt.Errorf("helm list: %w", &exec.ExitError{Stderr: []byte(`Error: namespaces "custom-ns" not found`)}),
			want: true,
		},
		{
			name: "exit error with other stderr",
			err:  &exec.ExitError{Stderr: []byte("Error: Kubernetes cluster unreachable")},
			want: false,
		},
		{
			name: "unrelated not found stderr",
			err:  &exec.ExitError{Stderr: []byte(`error: context was not found for specified context: foo`)},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("connection refused"),
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isNamespaceNotFound(c.err); got != c.want {
				t.Errorf("isNamespaceNotFound(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}
