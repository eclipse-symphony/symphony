/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package margo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/hydra"
)

type MargoAppMetadata struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}
type MargoHelmSpec struct {
	RepoType string `yaml:"repoType"`
	URL      string `yaml:"url"`
}
type MargoDockerComposeSpec struct {
	RepoType string `yaml:"repoType"`
	URL      string `yaml:"url"`
}
type MargoAppSpec struct {
}
type MargoAppDefinition struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	Metadata   MargoAppMetadata `yaml:"metadata"`
	Spec       MargoAppSpec     `yaml:"spec"`
}

type MargoSolutionReader struct {
}

func (m *MargoSolutionReader) Parse(appPackage hydra.AppPackageDescription) (model.SolutionState, error) {
	// Construct path to margo.yaml
	margoPath := filepath.Join(appPackage.Path, "margo.yaml")

	// Read margo.yaml
	data, err := os.ReadFile(margoPath)
	if err != nil {
		return model.SolutionState{}, fmt.Errorf("error reading margo.yaml: %v", err)
	}

	return model.SolutionState{}, nil
}
