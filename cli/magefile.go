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

// Build Symphony api for Windows, Mac and Linux.
func BuildApi() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(filepath.Join(wd, "..", "api/pkg/apis/v1alpha1/target/rust"))
	if err != nil {
		return err
	}
	if err := shellcmd.RunAll(
		// Build the Cargo project for each target
		shellcmd.Command("cargo build --release --target aarch64-unknown-linux-gnu"),
		shellcmd.Command("cargo build --release --target armv7-unknown-linux-gnueabihf"),
		shellcmd.Command("cargo build --release --target x86_64-pc-windows-gnu"),
		shellcmd.Command("cargo build --release --target x86_64-apple-darwin"),
		shellcmd.Command("cargo build --release --target x86_64-unknown-linux-gnu"),
		shellcmd.Command("cargo build --release"),
	); err != nil {
		return err
	}

	err = os.Chdir(filepath.Join(wd, "..", "api"))
	if err != nil {
		return err
	}
	if err := shellcmd.RunAll(
		shellcmd.Command("CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOARCH=arm64 go build -o symphony-api-arm64"),
		shellcmd.Command("CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOARCH=arm GOARM=7 go build -o symphony-api-arm"),
		shellcmd.Command("CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o symphony-api.exe"),
		shellcmd.Command("CC=o64-clang CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o symphony-api-mac"),
		shellcmd.Command("CC=gcc CGO_ENABLED=1 go build -o symphony-api"),
	); err != nil {
		return err
	}
	err = os.Chdir(filepath.Join(wd))
	if err != nil {
		return err
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
		shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-api.exe %s", symphonyPath, des)),
		shellcmd.Command(fmt.Sprintf("cp %s/api/symphony-api-mac %s", symphonyPath, des)),
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
	linuxCommand := fmt.Sprintf("tar -czvf maestro_linux_amd64.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json k8s iot-edge")
	if err := shellcmd.RunAll(
		shellcmd.Command(linuxCommand),
	); err != nil {
		return err
	}

	// package windows
	windowsCommand := fmt.Sprintf("zip -r maestro_windows_amd64.zip maestro.exe symphony-api.exe symphony-api-no-k8s.json samples.json k8s iot-edge")
	if err := shellcmd.RunAll(
		shellcmd.Command(windowsCommand),
	); err != nil {
		return err
	}

	// package mac
	macComomand := fmt.Sprintf("tar -czvf maestro_darwin_amd64.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json k8s iot-edge")
	if err := shellcmd.RunAll(
		shellcmd.Command(fmt.Sprintf("rm maestro")),
		shellcmd.Command(fmt.Sprintf("rm symphony-api")),
		shellcmd.Command(fmt.Sprintf("mv maestro-mac maestro")),
		shellcmd.Command(fmt.Sprintf("mv symphony-api-mac symphony-api")),
		shellcmd.Command(macComomand),
	); err != nil {
		return err
	}

	// package arm64
	arm64Comomand := fmt.Sprintf("tar -czvf maestro_linux_arm64.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json k8s iot-edge")
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
	arm7Command := fmt.Sprintf("tar -czvf maestro_linux_arm.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json k8s iot-edge")
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
