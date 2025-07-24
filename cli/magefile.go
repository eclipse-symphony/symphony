//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

/*
Use this tool to quickly build symphony api or maestro cli. It can also help generate the release package.
*/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/princjef/mageutil/shellcmd"
)

// Build maestro cli tools for Windows, Mac and Linux.
func BuildCli() error {
	if err := shellcmd.RunAll(
		"CGO_ENABLED=0 go build -o maestro",
		"CGO_ENABLED=0 GOARCH=arm64 go build -o maestro-arm64",
		"CGO_ENABLED=0 GOARCH=arm GOARM=7 go build -o maestro-arm",
		"CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o maestro.exe",
		"CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o maestro-mac",
	); err != nil {
		return err
	}
	return nil
}

// Build Symphony API for Windows, Mac, and Linux.
func BuildApi() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Change directory to Rust project
	err = os.Chdir(filepath.Join(wd, "..", "api/pkg/apis/v1alpha1/providers/target/rust"))
	if err != nil {
		return err
	}

	// Define build commands with environment variables
	cmds := []struct {
		env  []string
		args []string
	}{
		{ // Aarch64
			env:  []string{"CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc", "CC=aarch64-linux-gnu-gcc", "RUSTFLAGS=-C linker=aarch64-linux-gnu-gcc"},
			args: []string{"cargo", "build", "--release", "--target", "aarch64-unknown-linux-gnu"},
		},
		{ // ARMv7
			env:  []string{"CARGO_TARGET_ARM_UNKNOWN_LINUX_GNUEABIHF_LINKER=arm-linux-gnueabihf-gcc", "CC=arm-linux-gnueabihf-gcc", "RUSTFLAGS=-C linker=arm-linux-gnueabihf-gcc"},
			args: []string{"cargo", "build", "--release", "--target", "armv7-unknown-linux-gnueabihf"},
		},
		{ // Standard targets
			env:  []string{"CC=", "RUSTFLAGS="},
			args: []string{"cargo", "build", "--release", "--target", "x86_64-pc-windows-gnu"},
		},
		// {
		// 	env:  []string{"CC=o64-clang", "CXX=o64-clang++", "RUSTFLAGS=-C linker=o64-clang"},
		// 	args: []string{"cargo", "build", "--release", "--target", "x86_64-apple-darwin"},
		// },
		{
			env:  []string{"CC=", "RUSTFLAGS="},
			args: []string{"cargo", "build", "--release", "--target", "x86_64-unknown-linux-gnu"},
		},
	}

	// Run commands
	for _, cmd := range cmds {
		if err := runCommand(cmd.env, cmd.args...); err != nil {
			return err
		}
	}

	// Change back to API directory
	err = os.Chdir(filepath.Join(wd, "..", "api"))
	if err != nil {
		return err
	}

	// Build Symphony API with proper cross-compilation settings
	cmds = []struct {
		env  []string
		args []string
	}{
		{ // Aarch64
			env:  []string{"CC=aarch64-linux-gnu-gcc", "CGO_ENABLED=1", "GOARCH=arm64"},
			args: []string{"go", "build", "-o", "symphony-api-arm64"},
		},
		{ // ARMv7
			env:  []string{"CC=arm-linux-gnueabihf-gcc", "CGO_ENABLED=1", "GOARCH=arm", "GOARM=7"},
			args: []string{"go", "build", "-o", "symphony-api-arm"},
		},
		{ // Linux x86_64
			env: []string{
				"CGO_ENABLED=1",
				"GOARCH=amd64",
				"GOOS=linux",
				"CC=gcc",
				"LD_LIBRARY=./pkg/apis/v1alpha1/providers/target/rust/target/x86_64-unknown-linux-gnu/release",
				"CGO_LDFLAGS=-L./pkg/apis/v1alpha1/providers/target/rust/target/x86_64-unknown-linux-gnu/release"},
			args: []string{"go", "build", "-o", "symphony-api"},
		},
	}

	// Run commands
	for _, cmd := range cmds {
		if err := runCommand(cmd.env, cmd.args...); err != nil {
			return err
		}
	}

	// Change back to the original working directory
	return os.Chdir(wd)
}

// Run a command with optional environment variables
func runCommand(envVars []string, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), envVars...) // Append to existing env vars
	fmt.Printf("Running command: %s\n", cmd.String())
	return cmd.Run()
}

// Run multiple commands
func runCommands(commands [][]string, envVars []string) error {
	for _, cmdArgs := range commands {
		if err := runCommand(envVars, cmdArgs...); err != nil {
			return err
		}
	}
	return nil
}

// Generate packages with Symphony api, maestro cli and samples for Windows, Mac and Linux.
func GeneratePackages(des string) error {
	des = filepath.Clean(des)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	symphonyPath := filepath.Dir(wd)
	fmt.Println(symphonyPath)
	if err := BuildCli(); err != nil {
		return err
	}
	if err := BuildApi(); err != nil {
		return err
	}
	err = os.MkdirAll(des, os.ModePerm)
	if err != nil {
		return err
	}

	// remove previous packages, if any
	if err := removePakcageIfExist(fmt.Sprintf("%s/maestro_linux_amd64.tar.gz", des)); err != nil {
		return err
	}
	if err := removePakcageIfExist(fmt.Sprintf("%s/maestro_linux_arm64.tar.gz", des)); err != nil {
		return err
	}
	if err := removePakcageIfExist(fmt.Sprintf("%s/maestro_linux_arm.tar.gz", des)); err != nil {
		return err
	}
	if err := removePakcageIfExist(fmt.Sprintf("%s/maestro_windows_amd64.zip", des)); err != nil {
		return err
	}
	if err := removePakcageIfExist(fmt.Sprintf("%s/maestro_darwin_amd64.tar.gz", des)); err != nil {
		return err
	}

	// copy new binary files, configuration files and scripts
	if err := shellcmd.RunAll(
		shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-api %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-api-arm64 %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-api-arm %s", symphonyPath, des)),
		// TODO: Re-enable Mac and Windows cross build
		// shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-api.exe %s", symphonyPath, des)),
		// shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-api-mac %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-agent.json %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-api-no-k8s.json %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/cli/maestro %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/cli/maestro-arm64 %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/cli/maestro-arm %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/cli/maestro.exe %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s//cli/maestro-mac %s", symphonyPath, des)),
	); err != nil {
		return err
	}

	// copy over samples
	err = os.MkdirAll(filepath.Join(des, "k8s"), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(des, "iot-edge"), os.ModePerm)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err := shellcmd.RunAll(
		shellcmd.Command(fmt.Sprintf("cp %s/docs/samples/samples.json %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp -r %s/docs/samples/k8s/hello-world/ %s", symphonyPath, filepath.Join(des, "k8s"))),
		shellcmd.Command(fmt.Sprintf("cp -r %s/docs/samples/k8s/staged/ %s", symphonyPath, filepath.Join(des, "k8s"))),
		shellcmd.Command(fmt.Sprintf("cp -r %s/docs/samples/iot-edge/simulated-temperature-sensor/ %s", symphonyPath, filepath.Join(des, "iot-edge"))),
	); err != nil {
		return err
	}

	// change working directory to des folder
	err = os.Chdir(des)
	if err != nil {
		return err
	}

	// package Linux
	linuxCommand := fmt.Sprintf("tar -czvf maestro_linux_amd64.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json symphony-agent.json k8s iot-edge")
	if err := shellcmd.RunAll(
		shellcmd.Command(linuxCommand),
	); err != nil {
		return err
	}

	// package windows
	// windowsCommand := fmt.Sprintf("zip -r maestro_windows_amd64.zip maestro.exe symphony-api.exe symphony-api-no-k8s.json samples.json k8s iot-edge")
	// TODO: re-enable windows package
	windowsCommand := fmt.Sprintf("zip -r maestro_windows_amd64.zip maestro.exe symphony-api-no-k8s.json samples.json k8s iot-edge")
	if err := shellcmd.RunAll(
		shellcmd.Command(windowsCommand),
	); err != nil {
		return err
	}

	// package mac
	//macComomand := fmt.Sprintf("tar -czvf maestro_darwin_amd64.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json k8s iot-edge")
	// TODO: re-enable mac package
	macComomand := fmt.Sprintf("tar -czvf maestro_darwin_amd64.tar.gz maestro symphony-api-no-k8s.json samples.json k8s iot-edge")
	if err := shellcmd.RunAll(
		shellcmd.Command(fmt.Sprintf("rm maestro")),
		// shellcmd.Command(fmt.Sprintf("rm symphony-api")),
		shellcmd.Command(fmt.Sprintf("mv maestro-mac maestro")),
		// TODO: re-enable mac package
		// shellcmd.Command(fmt.Sprintf("mv symphony-api-mac symphony-api")),
		shellcmd.Command(macComomand),
	); err != nil {
		return err
	}

	// package arm64
	arm64Comomand := fmt.Sprintf("tar -czvf maestro_linux_arm64.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json symphony-agent.json k8s iot-edge")
	if err := shellcmd.RunAll(
		shellcmd.Command(fmt.Sprintf("rm maestro")),
		shellcmd.Command(fmt.Sprintf("rm symphony-api")),
		shellcmd.Command(fmt.Sprintf("mv maestro-arm64 maestro")),
		shellcmd.Command(fmt.Sprintf("mv symphony-api-arm64 symphony-api")),
		shellcmd.Command(arm64Comomand),
	); err != nil {
		return err
	}

	// package arm64
	arm7Command := fmt.Sprintf("tar -czvf maestro_linux_arm.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json symphony-agent.json k8s iot-edge")
	if err := shellcmd.RunAll(
		shellcmd.Command(fmt.Sprintf("rm maestro")),
		shellcmd.Command(fmt.Sprintf("rm symphony-api")),
		shellcmd.Command(fmt.Sprintf("mv maestro-arm maestro")),
		shellcmd.Command(fmt.Sprintf("mv symphony-api-arm symphony-api")),
		shellcmd.Command(arm7Command),
	); err != nil {
		return err
	}

	return nil
}

func removePakcageIfExist(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := shellcmd.Command(
			fmt.Sprintf("rm %s", path),
		).Run(); err != nil {
			return err
		}
	}
	return nil
}
