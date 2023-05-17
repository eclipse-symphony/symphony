package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
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
func RunCommand(message string, successMessage string, showOutput bool, name string, args ...string) (string, error) {
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
	return strings.Join(output, " "), exeErr
}

func CheckDocker(verbose bool) bool {
	_, err := RunCommand("Checking Docker", "found", verbose, "docker", "info")
	return err == nil
}

func CheckKubectl(verbose bool) bool {
	_, err := RunCommand("Checking kubectl", "found", verbose, "kubectl")
	return err == nil
}

func CheckK8sConnection(verbose bool) (string, bool) {
	str, err := RunCommand("Checking kubectl context", "OK", verbose, "kubectl", "config", "current-context")
	if err != nil {
		return str, false
	}
	_, err = RunCommand("Checking Kubernetes connection", "OK", verbose, "kubectl", "cluster-info")
	return str, err == nil
}

func CheckHelm(verbose bool) bool {
	info, err := RunCommand("Checking Helm", "found", verbose, "helm", "version")
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
		_, err := RunCommand("Downloading Docker Desktop Engine", "done", verbose, "curl", "-Lo", "docker-msi.exe", "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe")
		if err != nil {
			fmt.Printf("\n%s  Failed to download Docker Desktop Engine.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, err = RunCommand("Installing Docker Desktop Engine", "done", verbose, "start", "/w", "", "docker-msi.exe", "install", "--quiet", "--accept-license")
		if err != nil {
			fmt.Printf("\n%s  Failed to install Docker Desktop Engine.%s\n\n", ColorRed(), ColorReset())
			return false
		}
	case "darwin":
		//TODO: This downloads only for Intel chips
		_, err := RunCommand("Downloading Docker Desktop Engine", "done", verbose, "curl", "-Lo", "Docker.dmg", "https://desktop.docker.com/mac/main/amd64/Docker.dmg")
		if err != nil {
			fmt.Printf("\n%s  Failed to download Docker Desktop Engine.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, err = RunCommand("Attaching Docker Desktop Engine installer", "done", verbose, "sudo", "hdiutil", "attach", "Docker.dmg")
		if err != nil {
			fmt.Printf("\n%s  Failed to attach Docker Desktop Engine installer.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, err = RunCommand("Installing Docker Desktop Engine", "done", verbose, "sudo", "/Volumes/Docker/Docker.app/Contents/MacOS/install")
		if err != nil {
			fmt.Printf("\n%s  Failed to intall Docker Desktop Engine.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, err = RunCommand("Detaching Docker Desktop Engine installer", "done", verbose, "sudo", "hdiutil", "detach", "/Volumes/Docker")
		if err != nil {
			fmt.Printf("\n%s  Failed to detach Docker Desktop Engine installer.%s\n\n", ColorRed(), ColorReset())
			return false
		}
	case "linux":
		_, err := RunCommand("Updating package index", "done", verbose, "sudo", "apt-get", "update")
		if err != nil {
			fmt.Printf("\n%s  Failed to update package index.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, err = RunCommand("Installing Docker", "done", verbose, "sudo", "apt-get", "install", "-y", "docker.io")
		if err != nil {
			fmt.Printf("\n%s  Failed to install Docker.%s\n\n", ColorRed(), ColorReset())
			return false
		}
		// err, _ = utils.RunCommand("Adding Docker user group", "done", verbose, "sudo", "groupadd", "docker")
		// if err != nil {
		// 	fmt.Printf("\n%s  Failed to add Docker user group./%s\n\n", utils.ColorRed(), utils.ColorReset())
		// 	return false
		// }
		user := os.Getenv("USER")
		_, err = RunCommand("Adding current user to Docker user group", "done", verbose, "sudo", "usermod", "-aG", "docker", user)
		if err != nil {
			fmt.Printf("\n%s  Failed to add current user to Docker user group./%s\n\n", ColorRed(), ColorReset())
			return false
		}
		_, err = RunCommand("Activating group membership", "done", verbose, "newgrp", "docker")
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
