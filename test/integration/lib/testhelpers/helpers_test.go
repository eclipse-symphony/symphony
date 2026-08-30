/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package testhelpers

import (
	"errors"
	"testing"
)

func TestWebhookReady(t *testing.T) {
	cases := []struct {
		name   string
		output string
		err    error
		want   bool
	}{
		{
			name:   "dry-run apply succeeded",
			output: "target.fabric.symphony/webhook-readiness-probe configured (server dry run)",
			err:    nil,
			want:   true,
		},
		{
			name:   "admission denial proves the webhook answered",
			output: `Error from server: error when creating "probe.yaml": admission webhook "vtarget.kb.io" denied the request: missing properties`,
			err:    errors.New("exit status 1"),
			want:   true,
		},
		{
			name:   "webhook call failure keeps polling",
			output: `Error from server (InternalError): error when creating "probe.yaml": Internal error occurred: failed calling webhook "mtarget.kb.io"`,
			err:    errors.New("exit status 1"),
			want:   false,
		},
		{
			name:   "API server down keeps polling",
			output: "The connection to the server localhost:8080 was refused - did you specify the right host or port?",
			err:    errors.New("exit status 1"),
			want:   false,
		},
		{
			name:   "CRDs not installed keeps polling",
			output: `error: resource mapping not found for name: "webhook-readiness-probe" namespace: "default" from "probe.yaml": no matches for kind "Target" in version "fabric.symphony/v1"`,
			err:    errors.New("exit status 1"),
			want:   false,
		},
		{
			name:   "empty output with error keeps polling",
			output: "",
			err:    errors.New("exit status 1"),
			want:   false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := webhookReady(c.output, c.err); got != c.want {
				t.Errorf("webhookReady(%q, %v) = %v, want %v", c.output, c.err, got, c.want)
			}
		})
	}
}
