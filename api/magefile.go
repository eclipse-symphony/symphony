//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"fmt"
	"os"

	//mage:import
	_ "github.com/eclipse-symphony/symphony/packages/mage"
	"github.com/princjef/mageutil/shellcmd"
)

func Build() error {
	return shellcmd.Command("go build -o bin/symphony-api").Run()
}

// Runs both api unit tests as well as coa unit tests.
func TestWithCoa() error {
	// Unit tests for api
	testHelper()

	// Change directory to coa
	os.Chdir("../coa")

	// Unit tests for coa
	testHelper()
	return nil
}

func raceCheckSkipped() bool {
	return os.Getenv("SKIP_RACE_CHECK") == "true"
}

func raceOpt() string {
	if raceCheckSkipped() {
		return ""
	}
	return "-race"
}

func testHelper() error {
	if err := shellcmd.RunAll(
		"go clean -testcache",
		fmt.Sprintf("go test %s -timeout 5m -cover -coverprofile=coverage.out ./...", raceOpt()),
	); err != nil {
		return err
	}
	return nil
}

func DockerBuildTargetAgent() error {
	return shellcmd.Command("docker-compose -f docker-compose-target-agent.yaml build").Run()
}
func DockerBuildPollAgent() error {
	return shellcmd.Command("docker-compose -f docker-compose-poll-agent.yaml build").Run()
}
