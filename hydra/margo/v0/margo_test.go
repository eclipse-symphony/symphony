package margo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse-symphony/symphony/hydra"
	"gopkg.in/yaml.v2"
)

func TestParse(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Define the content of margo.yaml
	yamlContent := `
apiVersion: v1
kind: Application
metadata:
  name: ExampleApp
  version: "1.0.0"
  description: This is an example application.
  organization: ExampleOrg
`

	// Write the content to a file named margo.yaml in the temporary directory
	margoPath := filepath.Join(tempDir, "margo.yaml")
	err := os.WriteFile(margoPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write margo.yaml: %v", err)
	}

	// Create a MockAppPackageDescription
	appPackage := hydra.AppPackageDescription{
		Type:    "margo",
		Path:    tempDir,
		Version: "v0",
	}

	// Create an instance of MargoSolutionReader
	reader := MargoSolutionReader{}

	// Call the Parse method
	_, err = reader.Parse(appPackage)
	if err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}

	// Add further checks to validate the parsed data if needed
	// For example, you can unmarshal the data and check if the fields match the expected values
	var appDesc ApplicationDescription
	data, err := os.ReadFile(margoPath)
	if err != nil {
		t.Fatalf("failed to read margo.yaml: %v", err)
	}
	err = yaml.Unmarshal(data, &appDesc)
	if err != nil {
		t.Fatalf("failed to unmarshal margo.yaml: %v", err)
	}

	if appDesc.APIVersion != "v1" || appDesc.Kind != "Application" || appDesc.Metadata.Name != "ExampleApp" {
		t.Fatalf("parsed data does not match expected values: %+v", appDesc)
	}
}
