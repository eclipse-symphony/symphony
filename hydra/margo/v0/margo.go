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
	"gopkg.in/yaml.v2"
)

// See https://github.com/margo/margo-specifications/blob/main/system-design/app-interoperability/application-package-definition.md

type ApplicationMetadata struct {
	Name             string `yaml:"name"`
	Version          string `yaml:"version"`
	Description      string `yaml:"description"`
	Icon             string `yaml:"icon,omitempty"`
	Author           string `yaml:"author,omitempty"`
	AuthorEmail      string `yaml:"author-email,omitempty"`
	Organization     string `yaml:"organization"`
	OrganizationSite string `yaml:"organization-site,omitempty"`
	AppTagline       string `yaml:"app-tagline,omitempty"`
	DescriptionLong  string `yaml:"description-long,omitempty"`
	LicenseFile      string `yaml:"license-file,omitempty"`
	AppSite          string `yaml:"app-site,omitempty"`
}
type DeploymentDef struct {
	RepoType string `yaml:"repo-type"`
	Url      string `yaml:"url"`
}
type ApplicationSpec struct {
	HelmChart     DeploymentDef `yaml:"helm-chart"`
	DockerCompose DeploymentDef `yaml:"docker-compose"`
}
type ApplicationDescription struct {
	APIVersion string              `yaml:"apiVersion"`
	Kind       string              `yaml:"kind"`
	Metadata   ApplicationMetadata `yaml:"metadata"`
	Spec       ApplicationSpec     `yaml:"spec"`
}

type MargoSolutionReader struct {
}

// generateMetadataMap generates a map from ApplicationMetadata fields
func generateMetadataMap(metadata ApplicationMetadata) map[string]string {
	metadataMap := make(map[string]string)

	if metadata.Name != "" {
		metadataMap["name"] = metadata.Name
	}
	if metadata.Version != "" {
		metadataMap["version"] = metadata.Version
	}
	if metadata.Description != "" {
		metadataMap["description"] = metadata.Description
	}
	if metadata.Icon != "" {
		metadataMap["icon"] = metadata.Icon
	}
	if metadata.Author != "" {
		metadataMap["author"] = metadata.Author
	}
	if metadata.AuthorEmail != "" {
		metadataMap["author-email"] = metadata.AuthorEmail
	}
	if metadata.Organization != "" {
		metadataMap["organization"] = metadata.Organization
	}
	if metadata.OrganizationSite != "" {
		metadataMap["organization-site"] = metadata.OrganizationSite
	}
	if metadata.AppTagline != "" {
		metadataMap["app-tagline"] = metadata.AppTagline
	}
	if metadata.DescriptionLong != "" {
		metadataMap["description-long"] = metadata.DescriptionLong
	}
	if metadata.LicenseFile != "" {
		metadataMap["license-file"] = metadata.LicenseFile
	}
	if metadata.AppSite != "" {
		metadataMap["app-site"] = metadata.AppSite
	}

	return metadataMap
}

func (m *MargoSolutionReader) Parse(appPackage hydra.AppPackageDescription) (model.SolutionState, error) {
	if appPackage.Type != "margo" {
		return model.SolutionState{}, fmt.Errorf("invalid app package type: %s", appPackage.Type)
	}
	if appPackage.Version != "v0" {
		return model.SolutionState{}, fmt.Errorf("invalid app package version: %s", appPackage.Version)
	}

	// Construct path to margo.yaml
	margoPath := filepath.Join(appPackage.Path, "margo.yaml")

	// Read margo.yaml
	data, err := os.ReadFile(margoPath)
	if err != nil {
		return model.SolutionState{}, fmt.Errorf("error reading margo.yaml: %v", err)
	}

	// Deserialize margo.yaml into ApplicationDescription
	var appDesc ApplicationDescription
	err = yaml.Unmarshal(data, &appDesc)
	if err != nil {
		return model.SolutionState{}, fmt.Errorf("error unmarshalling margo.yaml: %v", err)
	}

	// Construct SolutionState from ApplicationDescription
	solutionState := model.SolutionState{
		ObjectMeta: model.ObjectMeta{
			Name:        appDesc.Metadata.Name,
			Annotations: generateMetadataMap(appDesc.Metadata),
		},
		Spec: &model.SolutionSpec{
			Components: []model.ComponentSpec{},
		},
	}

	if appDesc.Spec.HelmChart.Url != "" {
		solutionState.Spec.Components = append(solutionState.Spec.Components, model.ComponentSpec{
			Type: "helm.v3",
			Name: appDesc.Metadata.Name,
			Properties: map[string]interface{}{
				"chart": map[string]string{
					"repo":    appDesc.Spec.HelmChart.Url,
					"version": appDesc.Metadata.Version,
				},
			},
		})
	}

	return solutionState, nil
}
