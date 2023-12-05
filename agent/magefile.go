//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

/*
Use this tool to quickly build Piccolo.
*/
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/princjef/mageutil/shellcmd"
)

// Build Symphony agent Piccolo with mode release or debug.
func BuildPiccolo(mode string) error {
	mode = strings.ToLower(mode)
	var command = ""
	if mode != "release" && mode != "debug" {
		return errors.New(fmt.Sprintf("Mode not allowed: %s", mode))
	}
	if mode == "release" {
		command = "cargo build --release"
	} else {
		command = "cargo build"
	}
	if err := shellcmd.RunAll(
		shellcmd.Command(command),
	); err != nil {
		return err
	}
	return nil
}
