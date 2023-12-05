/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package observability

import (
	"testing"

	v1alpha2 "github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestConsolePipeline(t *testing.T) {
	ob := Observability{}
	err := ob.Init(ObservabilityConfig{
		Pipelines: []PipelineConfig{
			{
				Exporter: ExporterConfig{
					Type:       v1alpha2.TracingExporterConsole,
					BackendUrl: "",
					Sampler: SamplerConfig{
						SampleRate: "always",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
func TestZipkinPipeline(t *testing.T) {
	ob := Observability{}
	err := ob.Init(ObservabilityConfig{
		Pipelines: []PipelineConfig{
			{
				Exporter: ExporterConfig{
					Type:       v1alpha2.TracingExporterZipkin,
					BackendUrl: "http://localhost:9411/api/v2/spans",
					Sampler: SamplerConfig{
						SampleRate: "always",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
