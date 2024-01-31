package mage

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/princjef/mageutil/bintool"
	"github.com/princjef/mageutil/shellcmd"
)

//go:embed .golangci.yml
var golangci string

//go:embed .gomarkdoc.yml
var gomarkdoc string

var (
	golines = bintool.Must(bintool.NewGo(
		"github.com/segmentio/golines",
		"v0.11.0",
	))
	linter = bintool.Must(bintool.New(
		"golangci-lint{{.BinExt}}",
		"1.51.1",
		"https://github.com/golangci/golangci-lint/releases/download/v{{.Version}}/golangci-lint-{{.Version}}-{{.GOOS}}-{{.GOARCH}}{{.ArchiveExt}}",
	))
	documenter = bintool.Must(bintool.New(
		"gomarkdoc{{.BinExt}}",
		"0.4.1",
		"https://github.com/princjef/gomarkdoc/releases/download/v{{.Version}}/gomarkdoc_{{.Version}}_{{.GOOS}}_{{.GOARCH}}{{.ArchiveExt}}",
	))
)

func ensureFormatter() error {
	return golines.Ensure()
}

func ensureLinter() error {
	return linter.Ensure()
}

func ensureDocumenter() error {
	return documenter.Ensure()
}

// EnsureAllTools checks to see if a valid version of the needed tools are
// installed, and downloads/installs them if not.
func EnsureAllTools() error {
	if err := ensureFormatter(); err != nil {
		return err
	}

	if err := ensureLinter(); err != nil {
		return err
	}

	if err := ensureDocumenter(); err != nil {
		return err
	}

	return nil
}

// Format formats the code.
func Format() error {
	if err := ensureFormatter(); err != nil {
		return err
	}

	return golines.Command("-m 80 --no-reformat-tags --base-formatter gofmt -w .").
		Run()
}

// Lint lints the code.
func Lint() error {
	if err := ensureLinter(); err != nil {
		return err
	}

	close, err := TmpFile(".golangci.yml", golangci)
	if err != nil {
		return err
	}
	defer close()

	return linter.Command("run").Run()
}

// Doc generates documents for the code.
func Doc() error {
	if err := ensureDocumenter(); err != nil {
		return err
	}

	close, err := docCfg()
	if err != nil {
		return err
	}
	defer close()

	return shellcmd.RunAll(
		documenter.Command("./..."),
		// Remove internal READMEs to prevent unnecessary thrashing.
		// TODO: See if this can be built into gomarkdoc.
		`find . -path '*/internal/*README.md' -exec rm {} +`,
		`find . -path '*/proto/*README.md' -exec rm {} +`,
	)
}

// Create a temporary gomarkdoc config with the current path.
func docCfg() (func(), error) {
	path, err := Gitwd()
	if err != nil {
		return nil, err
	}

	t, err := template.New("gomarkdoc").Delims("<<", ">>").Parse(gomarkdoc)
	if err != nil {
		return nil, err
	}

	var data strings.Builder
	if err := t.Execute(&data, path); err != nil {
		return nil, err
	}

	return TmpFile(".gomarkdoc.yml", data.String())
}

// Test runs the unit tests.
func Test() error {
	return shellcmd.Command(`go test -race -timeout 35s -cover ./...`).Run()
}

// TestRace runs unit tests without the test cache.
// TODO: remove once integration tests no longer reference this.
func TestRace() error {
	return shellcmd.RunAll(
		`go clean -testcache`,
		`go test -race -timeout 35s -cover ./...`,
	)
}

// CleanTest runs unit tests without the test cache.
func CleanTest() error {
	return shellcmd.RunAll(
		`go clean -testcache`,
		`go test -race -timeout 35s -cover ./...`,
	)
}

// Cover checks code coverage from unit tests.
func Cover(file string) error {
	return shellcmd.RunAll(
		`go test -coverprofile=coverage.out -coverpkg="./..." ./...`,
		shellcmd.Command(
			fmt.Sprintf(`go tool cover -html=coverage.out -o="%s"`, file),
		),
	)
}

// CI runs format, lint, doc and test.
func CI() error {
	if err := Format(); err != nil {
		return err
	}

	if err := Lint(); err != nil {
		return err
	}

	// if err := Doc(); err != nil {
	// 	return err
	// }

	if err := Test(); err != nil {
		return err
	}

	return nil
}

// CIVerify checks if format, lint, doc and test were ran.
func CIVerify() error {
	if err := Format(); err != nil {
		return err
	}

	if err := Lint(); err != nil {
		return err
	}

	// TODO: DocVerify does not work with manual internal removal.

	if err := Test(); err != nil {
		return err
	}

	return nil
}

// Build docker image with docker-compose.
func DockerBuild() error {
	return shellcmd.Command("docker-compose -f docker-compose.yaml build").Run()
}
