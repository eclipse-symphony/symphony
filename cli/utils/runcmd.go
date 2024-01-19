/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const (
	YesNo  int = 0
	Choice int = 1
)

func GetInput(message string, choices []string, questionType int) int64 {
	for {
		fmt.Printf("  %s ", message)
		if questionType == Choice {
			fmt.Println()
			for i, c := range choices {
				fmt.Printf("\n    %s%d)%s %s", ColorCyan(), i, ColorReset(), c)
			}
			fmt.Printf("\n\n  Your choice (0 - %d): ", len(choices)-1)
		}
		var input string
		fmt.Scanln(&input)
		switch questionType {
		case YesNo:
			str := strings.ToLower(strings.TrimSpace(input))
			if str == "y" || str == "yes" {
				return 1
			}
			return 0
		case Choice:
			index, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
			if err == nil {
				if index <= int64(len(choices)-1) {
					return index
				}
			}
		}
	}
}
func RunCommandNoCapture(message string, successMessage string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	exeErr := cmd.Wait()
	if exeErr == nil {
		fmt.Printf("\r  %s%s%s ...%s%s \n", ColorReset(), message, ColorGreen(), successMessage, ColorReset())
	} else {
		fmt.Printf("\r  %s%s%s ...%s%s \n", ColorReset(), message, ColorYellow(), "failed", ColorReset())
	}
	return "", exeErr
}

func RunCommandWithRetry(message string, successMessage string, showOutput bool, debug bool, name string, args ...string) (string, string, error) {
	var output string
	var errOutput string
	var err error

	retryCount := 0

	b := backoff.NewExponentialBackOff()
	if debug {
		// Customize the backoff parameters.
		b.InitialInterval = 2 * time.Second    // Initial retry interval.
		b.MaxInterval = 30 * time.Second       // Maximum retry interval.
		b.MaxElapsedTime = 1 * time.Nanosecond // Maximum total waiting time. (no retry)
	} else {
		// Customize the backoff parameters.
		b.InitialInterval = 2 * time.Second // Initial retry interval.
		b.MaxInterval = 30 * time.Second    // Maximum retry interval.
		b.MaxElapsedTime = 5 * time.Minute  // Maximum total waiting time.
	}

	retryErr := backoff.Retry(func() error {
		retryCount++
		if retryCount > 1 {
			fmt.Printf("\r  %s%s%s ...%s %d %s%s \n", ColorReset(), message, ColorYellow(), "retrying", retryCount, "round", ColorReset())
		}
		output, errOutput, err = RunCommand(message, successMessage, showOutput, name, args...)
		return err
	}, b)

	if retryErr == nil {
		// success
		return output, "", nil
	} else {
		// failure
		return "", errOutput, fmt.Errorf("failed to run command %s %s", name, strings.Join(args, " "))
	}
}

func RunCommand(message string, successMessage string, showOutput bool, name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	output := []string{}
	errOutput := []string{}
	cmdReader, _ := cmd.StdoutPipe()
	outScanner := bufio.NewScanner(cmdReader)
	errReader, _ := cmd.StderrPipe()
	errScanner := bufio.NewScanner(errReader)
	done := make(chan bool)
	running := false
	var exeErr error
	hideCursor()
	go func() {
		cmd.Start()
		running = true
		exeErr = cmd.Wait()
		running = false
		done <- true
	}()
	go func() {
		for errScanner.Scan() {
			errOutput = append(errOutput, errScanner.Text())
		}
		<-done
	}()
	go func() {
		for outScanner.Scan() {
			output = append(output, outScanner.Text())
		}
		<-done
	}()
	go func() {
		for {
			for _, r := range `⣾⣽⣻⢿⡿⣟⣯⣷` {
				if running {
					fmt.Printf("\r  %s%s%s %c%s", ColorReset(), message, ColorYellow(), r, ColorReset())
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
	<-done
	running = false
	time.Sleep(200 * time.Millisecond)
	if exeErr == nil {
		fmt.Printf("\r  %s%s%s ...%s%s \n", ColorReset(), message, ColorGreen(), successMessage, ColorReset())
	} else {
		fmt.Printf("\r  %s%s%s ...%s%s \n", ColorReset(), message, ColorYellow(), "failed", ColorReset())
	}

	if showOutput {
		for _, o := range output {
			fmt.Printf("    %s\n", o)
		}
		for _, o := range errOutput {
			fmt.Printf("    %s%s\n%s", ColorRed(), o, ColorReset())
		}
	}
	showCursor()
	return strings.Join(output, " "), strings.Join(errOutput, "\n"), exeErr
}

func AddtoPath(des string) string {
	path := os.Getenv("PATH")
	pathlist := strings.Split(path, ";")
	for _, p := range pathlist {
		if strings.EqualFold(p, des) {
			return path
		}
	}
	return path + ";" + des
}

func CheckDocker(verbose bool) bool {
	_, _, err := RunCommand("Checking Docker", "found", verbose, "docker", "info")
	return err == nil
}

func CheckKubectl(verbose bool) bool {
	_, _, err := RunCommand("Checking kubectl", "found", verbose, "kubectl")
	return err == nil
}

func CheckMinikube(verbose bool) bool {
	// on Windows, add the known Windows installation path to env:PATH to avoid minikube checking error
	osName := runtime.GOOS
	if strings.EqualFold(osName, "windows") {
		var des = filepath.Join(os.Getenv("programfiles"), "maestro", "minikube")
		path := AddtoPath(des)
		if err := os.Setenv("path", path); err != nil {
			fmt.Printf("\n%s  Failed to setting path for minikube.%s\n\n", ColorRed(), ColorReset())
		}
	}
	_, _, err := RunCommand("Checking minikube", "found", verbose, "minikube", "version")
	return err == nil
}

func CheckK8sConnection(verbose bool) (string, bool) {
	str, _, err := RunCommand("Checking kubectl context", "OK", verbose, "kubectl", "config", "current-context")
	if err != nil {
		return str, false
	}
	_, _, err = RunCommand("Checking Kubernetes connection", "OK", verbose, "kubectl", "cluster-info")
	return str, err == nil
}

func CheckHelm(verbose bool) bool {
	info, _, err := RunCommand("Checking Helm", "found", verbose, "helm", "version")
	if err != nil {
		return false
	}
	re := regexp.MustCompile(`v[0-9]+\.[0-9]+\.[0-9]+`)
	versionStr := re.FindString(info)
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Split the version string into its major, minor, and patch components
	versionParts := strings.Split(versionStr, ".")
	if len(versionParts) < 2 {
		fmt.Printf("\n%s Error parsing Helm version: %s %s\n\n", ColorRed(), versionStr, ColorReset())
		return false
	}
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		fmt.Printf("\n%s Error parsing Helm version: %s %s\n\n", ColorRed(), versionStr, ColorReset())
		return false
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		fmt.Printf("\n%s Error parsing Helm version: %s %s\n\n", ColorRed(), versionStr, ColorReset())
		return false
	}

	// Check if the Helm version is at least 3.8
	if major < 3 || (major == 3 && minor < 8) {
		fmt.Printf("\n%s  Helm version 3.8 or above is required but %s is found%s\n\n", ColorRed(), versionStr, ColorReset())
		return false
	}
	return true
}

func InstallDocker(verbose bool) bool {
	ios := runtime.GOOS
	switch ios {
	case "windows":
		_, _, err := RunCommand("Downloading Docker Desktop Engine", "done", verbose, "curl", "-Lo", "docker-msi.exe", "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe")
		if err != nil {
			fmt.Printf("\n%s  Failed to download Docker Desktop Engine.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, _, err = RunCommand("Installing Docker Desktop Engine", "done", verbose, "start", "/w", "", "docker-msi.exe", "install", "--quiet", "--accept-license")
		if err != nil {
			fmt.Printf("\n%s  Failed to install Docker Desktop Engine.%s\n\n", ColorRed(), ColorReset())
			return false
		}
	case "darwin":
		//TODO: This downloads only for Intel chips
		_, _, err := RunCommand("Downloading Docker Desktop Engine", "done", verbose, "curl", "-Lo", "Docker.dmg", "https://desktop.docker.com/mac/main/amd64/Docker.dmg")
		if err != nil {
			fmt.Printf("\n%s  Failed to download Docker Desktop Engine.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, _, err = RunCommand("Attaching Docker Desktop Engine installer", "done", verbose, "sudo", "hdiutil", "attach", "Docker.dmg")
		if err != nil {
			fmt.Printf("\n%s  Failed to attach Docker Desktop Engine installer.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, _, err = RunCommand("Installing Docker Desktop Engine", "done", verbose, "sudo", "/Volumes/Docker/Docker.app/Contents/MacOS/install")
		if err != nil {
			fmt.Printf("\n%s  Failed to intall Docker Desktop Engine.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, _, err = RunCommand("Detaching Docker Desktop Engine installer", "done", verbose, "sudo", "hdiutil", "detach", "/Volumes/Docker")
		if err != nil {
			fmt.Printf("\n%s  Failed to detach Docker Desktop Engine installer.%s\n\n", ColorRed(), ColorReset())
			return false
		}
	case "linux":
		_, _, err := RunCommand("Updating package index", "done", verbose, "sudo", "apt-get", "update")
		if err != nil {
			fmt.Printf("\n%s  Failed to update package index.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, _, err = RunCommand("Installing Docker", "done", verbose, "sudo", "apt-get", "install", "-y", "docker.io")
		if err != nil {
			fmt.Printf("\n%s  Failed to install Docker.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		// _, _, err = utils.RunCommand("Adding Docker user group", "done", verbose, "sudo", "groupadd", "docker")
		// if err != nil {
		// 	fmt.Printf("\n%s  Failed to add Docker user group./%s\n\n", utils.ColorRed(), utils.ColorReset())
		// 	return false
		// }
		user := os.Getenv("USER")
		_, _, err = RunCommand("Adding current user to Docker user group", "done", verbose, "sudo", "usermod", "-aG", "docker", user)
		if err != nil {
			fmt.Printf("\n%s  Failed to add current user to Docker user group./%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, _, err = RunCommand("Activating group membership", "done", verbose, "newgrp", "docker")
		if err != nil {
			fmt.Printf("\n%s  Failed to activate user group.%s\n\n", ColorRed(), ColorReset())
			return false
		}
	default:
		fmt.Printf("\n%s  Doesn't know how to install Docker on %s%s\n\n", ColorRed(), ios, ColorReset())
		return false
	}
	return true
}

func Unzip(src, dest string, isStructured bool) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		var path string
		if isStructured {
			path = filepath.Join(dest, f.Name)
		} else {
			path = filepath.Join(dest, f.FileInfo().Name())
		}
		if f.FileInfo().IsDir() {
			if isStructured {
				os.MkdirAll(path, f.Mode())
			}
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var Esc = "\x1b"

func escape(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s", Esc, fmt.Sprintf(format, args...))
}

func showCursor() {
	fmt.Print(escape("[?25h"))
}

// Hide returns ANSI escape sequence to hide the cursor
func hideCursor() {
	fmt.Print(escape("[?25l"))
}
func ColorCyan() string {
	return escape("[36m")
}
func ColorReset() string {
	return escape("[0m")
}
func ColorGreen() string {
	return escape("[32m")
}
func ColorRed() string {
	return escape("[31m")
}
func ColorYellow() string {
	return escape("[33m")
}
func ColorBlue() string {
	return escape("[34m")
}
func ColorPurple() string {
	return escape("[35m")
}
func ColorWhite() string {
	return escape("[37m")
}
