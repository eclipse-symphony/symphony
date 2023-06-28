package mage

import (
	"os"
	"strings"

	"github.com/princjef/mageutil/shellcmd"
)

// Gitwd is a utility to get the relative working directory within the Git root.
func Gitwd() (string, error) {
	prefix, err := shellcmd.Command(`git rev-parse --show-toplevel`).Output()
	if err != nil {
		return "", err
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(wd, strings.TrimSpace(string(prefix))), nil
}

// TmpFile writes the contents to the given temporary file and returns a
// function to clean up that file.
func TmpFile(name, contents string) (func(), error) {
	if err := os.WriteFile(name, []byte(contents), 0o600); err != nil {
		return nil, err
	}
	return func() { os.Remove(name) }, nil
}
