/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/spf13/cobra"
)

// The version is auto updated by the release pipeline, do not change it manually
const SymphonyAPIVersion = "0.48.45"
const KANPortalVersion = "0.39.0-main-603f4b9-amd64"

var (
	symphonyVersion string
	portalVersion   string
	//portalType          string
	//useWizard           bool
	noK8s               bool
	noRestApi           bool
	storageRP           string
	storageAccount      string
	storageContainer    string
	azureSubscription   string
	tenantId            string
	clientId            string
	clientSecret        string
	customVisionRP      string
	customVisionAccount string
	namespace           string
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Install Symphony on a Kubernetes cluster, or run Symphony locally",
	Run: func(cmd *cobra.Command, args []string) {
		// if portalType == "list" {
		// 	fmt.Println("NAME\t\tDESCRIPTION")
		// 	fmt.Println("OSS\t\tPercept Open Source Portal")
		// 	fmt.Println("Samsung\t\tSamsung Management Portal")
		// 	fmt.Println("Playground\tSymphony Playground")
		// 	return
		// }
		u, err := user.Current()
		if err != nil {
			fmt.Printf("\n%s  Failed: %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if noK8s {
			if !updateSymphonyContext("no-k8s", "localhost") {
				return
			}
			os.Setenv("SYMPHONY_API_URL", "http://localhost:8082/v1alpha2/")
			os.Setenv("USE_SERVICE_ACCOUNT_TOKENS", "false")
			_, err := utils.RunCommandNoCapture("Launching Symphony in standalone mode", "done", filepath.Join(u.HomeDir, ".symphony/symphony-api"), "-c", filepath.Join(u.HomeDir, ".symphony/symphony-api-no-k8s.json"), "-l", "Debug")
			if err != nil {
				fmt.Printf("\n%s  Failed: %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
				return
			}
		} else {
			// we don't need to check for Docker, as we are not using it
			// if !handleDocker() {
			//	return
			// }
			if !handleKubectl() {
				return
			}
			k8sContext, ret := handleK8sConnection()
			if !ret {
				return
			}
			_, errOutput, err := utils.RunCommand("Installing cert manager", "done", verbose, "kubectl", "apply", "-f", "https://github.com/jetstack/cert-manager/releases/download/v1.4.0/cert-manager.yaml")
			if err != nil {
				fmt.Printf("\n%s  Failed.%s\n\n", utils.ColorRed(), utils.ColorReset())
				fmt.Printf("\n%s  Detailed Messages: %s%s\n\n", utils.ColorRed(), errOutput, utils.ColorReset())
				return
			}
			if !handleHelm() {
				return
			}
			if !handleSymphony(noRestApi) {
				return
			}
			var tunnelCMD *exec.Cmd
			if !noRestApi {
				// only start tunnel for minikube
				if k8sContext == "minikube" {
					tunnelCMD, err = handleMinikubeTunnel()
					if err != nil {
						return
					}
				}

				ret, apiAddress := checkSymphonyAddress()
				if !ret {
					return
				}

				if !updateSymphonyContext(k8sContext, apiAddress) {
					return
				}

				fmt.Printf("  %sSymphony API:%s %s%s\n", utils.ColorGreen(), utils.ColorWhite(), "http://"+apiAddress+":8080/v1alpha2/greetings", utils.ColorReset())
				fmt.Println()

				if k8sContext == "minikube" {
					fmt.Printf("  %sKeeping minikube tunnel for API use. Press CTRL + C to stop the tunnel and quit.%s\n", utils.ColorGreen(), utils.ColorReset())
					fmt.Println()
					tunnelCMD.Wait()
				}
			}

			// if portalType != "" {
			// 	if !handlePortal(apiAddress) {
			// 		return
			// 	}
			// }

			// portalAddress := ""
			// if portalType != "" {
			// 	ret, portalAddress = checkPortalAddress()
			// 	if !ret {
			// 		return
			// 	}
			// }

			fmt.Printf("\n%s  Done!%s\n\n", utils.ColorCyan(), utils.ColorReset())
			// if portalType != "" {
			// 	fmt.Printf("  %sSymphony portal:%s %s%s\n", utils.ColorGreen(), utils.ColorWhite(), "http://"+portalAddress+"/", utils.ColorReset())
			// }

			if noRestApi {
				fmt.Printf("  %sREST API is turned off in no-rest-api Mode and you can interact with Symphony using kubectl.%s\n", utils.ColorYellow(), utils.ColorReset())
			}
			fmt.Println()

		}
	},
}

func init() {
	UpCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	UpCmd.Flags().StringVarP(&symphonyVersion, "symphony-version", "s", SymphonyAPIVersion, "Symphony API version")
	UpCmd.Flags().BoolVar(&noK8s, "no-k8s", false, "Launch in standalone mode (no Kubernetes)")
	UpCmd.Flags().BoolVar(&noRestApi, "no-rest-api", false, "Doesn't expose Symphony API, interact with k8s.")
	//UpCmd.Flags().StringVarP(&portalVersion, "portal-version", "p", KANPortalVersion, "Symphony Portal version")
	//UpCmd.Flags().StringVarP(&portalType, "with-portal", "", "", "Install Symphony Portal")
	// UpCmd.Flags().StringVarP(&storageRP, "storage-resource-group", "", "", "Azure Storage account resource group")
	// UpCmd.Flags().StringVarP(&storageAccount, "storage-account", "", "", "Azure Storage account")
	// UpCmd.Flags().StringVarP(&storageContainer, "storage-container", "", "", "Azure Storage container")
	// UpCmd.Flags().StringVarP(&azureSubscription, "azure-subscription", "", "", "Azure subscription")
	// UpCmd.Flags().StringVarP(&tenantId, "sp-tenant-id", "", "", "AAD tenant id")
	// UpCmd.Flags().StringVarP(&clientId, "sp-client-id", "", "", "AAD client id")
	// UpCmd.Flags().StringVarP(&clientSecret, "sp-client-secret", "", "", "AAD client secret")
	// UpCmd.Flags().StringVarP(&customVisionRP, "custom-vision-resource-group", "", "", "Azure Custom Vision resource group")
	// UpCmd.Flags().StringVarP(&customVisionAccount, "custom-vision-account", "", "", "Azure Custom Vision account")
	//UpCmd.MarkFlagsRequiredTogether("with-portal", "storage-resource-group", "storage-account", "storage-container", "azure-subscription", "sp-tenant-id", "sp-client-id", "sp-client-secret", "custom-vision-resource-group", "custom-vision-account")
	//iUpCmd.MarkFlagsMutuallyExclusive("with-portal", "no-k8s")
	RootCmd.AddCommand(UpCmd)
}

// func checkPortalAddress() (bool, string) {
// 	switch strings.ToLower(portalType) {
// 	case "oss":
// 		count := 0
// 		for {
// 			str, _, err := utils.RunCommand("Checking Symphony Portal address", "OK", verbose, "kubectl", "get", "svc", "ingress-nginx-controller", "-n", "ingress-nginx", "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}")
// 			if err != nil {
// 				fmt.Printf("\n%s  Failed to check Symphony Portal address.%s\n\n", utils.ColorRed(), utils.ColorReset())
// 				return false, ""
// 			}
// 			if str != "" {
// 				return true, str
// 			}
// 			count += 1
// 			if count > 5 {
// 				fmt.Printf("\n%s  Failed to check public Symphony Portal address. You may need to set up port forwarding to access the portal locally. %s\n\n", utils.ColorYellow(), utils.ColorReset())
// 				return true, ""
// 			}
// 			time.Sleep(5 * time.Second)
// 		}
// 	case "samsung":
// 		return true, "localhost:3000"
// 	}
// 	return false, ""
// }

func checkSymphonyAddress() (bool, string) {
	count := 0
	for {
		str, _, err := utils.RunCommand("Checking public Symphony API address", "", verbose, "kubectl", "get", "svc", "symphony-service-ext", "-n", namespace, "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}")
		if err == nil && str != "" {
			return true, str
		}
		count += 1
		if count > 5 {
			fmt.Printf("\n%s  Failed to check public Symphony API address. You can still access Symphony API through kubectl locally. %s\n\n", utils.ColorYellow(), utils.ColorReset())
			return true, ""
		}
		time.Sleep(5 * time.Second)
	}
}

// func handlePortal(apiAddress string) bool {
// 	switch strings.ToLower(portalType) {
// 	case "oss":
// 		str, _, _ := utils.RunCommand("Checking OSS portal", "done", verbose, "helm", "list", "-q", "-l", "name=voe")

// 		if str != "voe" {
// 			_, _, err := utils.RunCommand("Deploying OSS portal", "done", verbose, "helm", "upgrade", "--install", "voe", "oci://p4etest.azurecr.io/helm/voe", "--version", portalVersion,
// 				"--set", "storage.storageResourceGroup="+storageRP,
// 				"--set", "storage.storageAccount="+storageAccount,
// 				"--set", "storage.storageContainer="+storageContainer,
// 				"--set", "storage.subscriptionId="+azureSubscription,
// 				"--set", "customvision.endpoint=$(az cognitiveservices account show -n "+customVisionAccount+" -g "+customVisionRP+" | jq -r .properties.endpoint)",
// 				"--set", `customvision.trainingKey=$(az cognitiveservices account keys list -n `+customVisionAccount+` -g `+customVisionRP+` | jq -r ".key1")`,
// 				"--set", "servicePrincipal.tenantId="+tenantId,
// 				"--set", "servicePrincipal.clientId="+clientId,
// 				"--set", "servicePrincipal.clientSecret="+clientSecret)
// 			if err != nil {
// 				fmt.Printf("\n%s  Failed to deploy OSS Portal.%s\n\n", utils.ColorRed(), utils.ColorReset())
// 				return false
// 			}
// 		}
// 		if verbose {
// 			fmt.Printf("\n%s  WARNING: existing OSS portal deployment is found. To install new version, use p4ectl remove to remove it first.%s\n\n", utils.ColorYellow(), utils.ColorReset())
// 		}
// 		return true
// 	case "samsung":
// 		_, _, err := utils.RunCommand("Launching Samsung portal", "done", verbose, "docker", "run", "-dit", "--rm", "-p", "3000:3000", "-e", "NEXT_PUBLIC_BACKEND="+apiAddress, "dcp-symphony:1.0.2")
// 		if err != nil {
// 			fmt.Printf("\n%s  Failed to launch Samsung Portal.%s\n\n", utils.ColorRed(), utils.ColorReset())
// 			return false
// 		}
// 	}
// 	return true
// }

func handleSymphony(norest bool) bool {
	str, _, _ := utils.RunCommand("Checking Symphony API (Symphony)", "done", verbose, "helm", "list", "-q", "-l", "name=symphony")

	if str != "symphony" {

		cmd := exec.Command("kubectl", "get", "target", "--no-headers=true", "-o", "custom-columns=Name:.metadata.name")
		stdout, _ := cmd.Output()
		targets := strings.Fields(string(stdout))
		for _, t := range targets {
			c := exec.Command("kubectl", "patch", "target.fabric.symphony", t, "-p", `'{"metadata":{"finalizers":null}}'`, "--type=merge")
			c.Run()
		}

		cmd = exec.Command("kubectl", "get", "instance", "--no-headers=true", "-o", "custom-columns=Name:.metadata.name")
		stdout, _ = cmd.Output()
		instances := strings.Fields(string(stdout))
		for _, t := range instances {
			c := exec.Command("kubectl", "patch", "instance.solution.symphony", t, "-p", `'{"metadata":{"finalizers":null}}'`, "--type=merge")
			c.Run()
		}
	}

	fmt.Printf("  Deploying Symphony API (Symphony), installServiceExt: %t\n", !norest)
	installServiceExt := fmt.Sprintf("installServiceExt=%t", !norest)
	_, errOutput, err := utils.RunCommandWithRetry("Deploying Symphony API (Symphony)", "done", verbose, debug, "helm", "upgrade", "--install", "symphony", "oci://ghcr.io/eclipse-symphony/helm/symphony", "--version", symphonyVersion, "--set", "CUSTOM_VISION_KEY=dummy", "--set", "symphonyImage.pullPolicy=Always", "--set", "paiImage.pullPolicy=Always", "--set", installServiceExt, "--namespace", namespace)
	if err != nil {
		fmt.Printf("\n%s  Failed.%s\n\n", utils.ColorRed(), utils.ColorReset())
		fmt.Printf("\n%s  Detailed Messages: %s%s\n\n", utils.ColorRed(), errOutput, utils.ColorReset())
		return false
	}
	return true
}

func handleDocker() bool {
	if !utils.CheckDocker(verbose) {
		return utils.InstallDocker(verbose)
	}
	return true
}
func updateSymphonyContext(context string, apiAddress string) bool {
	err := config.UpdateMaestroConfig(context, apiAddress)
	if err != nil {
		fmt.Printf("\n%s  Failed to update maestro config file.%s\n\n", utils.ColorRed(), utils.ColorReset())
	}
	return true
}
func handleHelm() bool {
	if !utils.CheckHelm(verbose) {
		return installHelm()
	}
	return true
}

func installHelm() bool {
	ios := runtime.GOOS
	switch ios {
	case "darwin", "linux":
		_, _, err := utils.RunCommand("Downloading Helm 3", "done", verbose, "curl", "-fsSL", "-o", "get_helm.sh", "https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3")
		if err != nil {
			fmt.Printf("\n%s  Failed to download Helm 3.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		_, _, err = utils.RunCommand("Updating Helm 3 access", "done", verbose, "chmod", "+x", "./get_helm.sh")
		if err != nil {
			fmt.Printf("\n%s  Failed to update Helm 3 access.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		_, _, err = utils.RunCommand("Installing Helm 3", "done", verbose, "./get_helm.sh")
		if err != nil {
			fmt.Printf("\n%s  Failed to install Helm 3.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
	case "windows":
		_, _, err := utils.RunCommand("Downloading Helm 3", "done", verbose, "curl", "-fsSL", "-o", "helm.zip", "https://get.helm.sh/helm-v3.13.2-windows-amd64.zip")
		if err != nil {
			fmt.Printf("\n%s  Failed to download Helm 3.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		var des = filepath.Join(os.Getenv("programfiles"), "maestro", "helm")
		if err = os.MkdirAll(des, os.ModePerm); err != nil {
			fmt.Printf("\n%s  Failed to mkdir for Helm 3.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}

		if err = utils.Unzip("helm.zip", des, false); err != nil {
			fmt.Printf("\n%s  Failed to unzip for Helm 3.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}

		path := utils.AddtoPath(des)
		var command = fmt.Sprintf("[Environment]::SetEnvironmentVariable(\"Path\", \"%s\", \"Machine\")", path)
		_, _, err = utils.RunCommand("Setting path for Helm 3", "done", verbose, "powershell", "-NoProfile", command)
		if err != nil {
			fmt.Printf("\n%s  Failed to setting path for Helm 3.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}

		if err = os.Setenv("path", path); err != nil {
			fmt.Printf("\n%s  Failed to setting path for Helm 3.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
	default:
		fmt.Printf("\n%s  Doesn't know how to install Docker on %s%s\n\n", utils.ColorRed(), ios, utils.ColorReset())
		return false
	}
	return true
}

func handleKubectl() bool {
	if !utils.CheckKubectl(verbose) {
		input := utils.GetInput("kubectl is not found. Do you want to install it? [Yes/No]", nil, utils.YesNo)
		if input == 0 {
			fmt.Printf("\n%s  Kubectl is not found. Please install kubectl first. See: https://kubernetes.io/docs/tasks/tools/%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		} else {
			return installKubectl()
		}
	}
	return true
}
func handleK8sConnection() (string, bool) {
	address, ok := utils.CheckK8sConnection(verbose)
	if !ok {
		input := utils.GetInput("kubectl is not connected to a Kubernetes cluster, what do you want to do?", []string{"Install a local cluster (Minukube)", "Connect to a remote cluster (AKS)"}, utils.Choice)
		switch input {
		case 0:
			ok := utils.CheckMinikube(false)
			if ok {
				_, errOutput, err := utils.RunCommand("Creating Kubernetes cluster", "done", verbose, "minikube", "start")
				if err != nil {
					fmt.Printf("\n%s  Failed to create Kubernetes cluster.%s\n\n", utils.ColorRed(), utils.ColorReset())
					fmt.Printf("\n%s  Detailed Messages: %s%s\n\n", utils.ColorRed(), errOutput, utils.ColorReset())
					return "", false
				}
				return "minikube", true
			} else {
				if !installMinikube() {
					return "", false
				}
				return createMinikubeCluster()
			}
		case 1:
		default:
			fmt.Printf("\n%s  Can't connect to a Kubernetes cluster. Please configure your kubectl context to a valid Kubernetes cluster.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return "", false
		}
	}
	return address, true
}
func setupK8sConnection() bool {
	return true
}
func handleMinikubeTunnel() (*exec.Cmd, error) {
	// ensure we can run minikube tunnel given users have a connected k8s context which is minikube K8S but not prepared by maestro
	ok := utils.CheckMinikube(false)
	if !ok {
		installMinikube()
	}
	cmd := exec.Command("minikube", "tunnel")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("\n%s  Failed to create minikube tunnel.%s\n\n", utils.ColorRed(), utils.ColorReset())
		return nil, err
	}
	return cmd, nil
}
func createMinikubeCluster() (string, bool) {
	_, errOutput, err := utils.RunCommand("Creating Kubernetes cluster", "done", verbose, "minikube", "start")
	if err != nil {
		fmt.Printf("\n%s  Failed to create Kubernetes cluster.%s\n\n", utils.ColorRed(), utils.ColorReset())
		fmt.Printf("\n%s  Detailed Messages: %s%s\n\n", utils.ColorRed(), errOutput, utils.ColorReset())
		return "", false
	}
	return "minikube", true
}
func installMinikube() bool {
	osName := runtime.GOOS
	cpu := runtime.GOARCH
	switch osName {
	case "windows":
		switch cpu {
		case "amd64":
			_, _, err := utils.RunCommand("Downloading minikube", "done", verbose, "curl", "-Lo", "minikube.exe", "https://github.com/kubernetes/minikube/releases/download/v1.32.0/minikube-windows-amd64.exe")
			if err != nil {
				fmt.Printf("\n%s  Failed to download minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
				return false
			}
		default:
			fmt.Printf("\n%s  Minikube doesn't support %s on %s%s\n\n", cpu, utils.ColorRed(), osName, utils.ColorReset())
			return false
		}
		var des = filepath.Join(os.Getenv("programfiles"), "maestro", "minikube")
		if err := os.MkdirAll(des, os.ModePerm); err != nil {
			fmt.Printf("\n%s  Failed to mkdir for Minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}

		var path = utils.AddtoPath(des)
		var command = fmt.Sprintf("Copy-Item minikube.exe \"%s\\minikube.exe\";[Environment]::SetEnvironmentVariable(\"Path\", \"%s\", \"Machine\")", des, path)
		_, _, err := utils.RunCommand("Setting path for minikube", "done", verbose, "powershell", "-NoProfile", command)
		if err != nil {
			fmt.Printf("\n%s  Failed to setting path for minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		if err := os.Remove("minikube.exe"); err != nil {
			fmt.Printf("\n%s  Failed to remove tmp minikube.exe. This would cause minikube command to fail.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
	case "darwin":
		switch cpu {
		case "amd64":
			_, _, err := utils.RunCommand("Downloading minikube", "done", verbose, "curl", "-Lo", "./minikube", "https://storage.googleapis.com/minikube/releases/v1.32.0/minikube-darwin-amd64")
			if err != nil {
				fmt.Printf("\n%s  Failed to download minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
				return false
			}
		case "arm64":
			_, _, err := utils.RunCommand("Downloading minikube", "done", verbose, "curl", "-Lo", "./minikube", "https://storage.googleapis.com/minikube/releases/v1.32.0/minikube-darwin-arm64")
			if err != nil {
				fmt.Printf("\n%s  Failed to download minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
				return false
			}
		default:
			fmt.Printf("\n%s  Minikube doesn't support %s on %s%s\n\n", cpu, utils.ColorRed(), osName, utils.ColorReset())
			return false
		}
		_, _, err := utils.RunCommand("Updating minikube access", "done", verbose, "chmod", "+x", "./minikube")
		if err != nil {
			fmt.Printf("\n%s  Failed to update minikube access.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		_, _, err = utils.RunCommand("Installing minikube", "done", verbose, "sudo", "install", "./minikube", "/usr/local/bin/minikube")
		if err != nil {
			fmt.Printf("\n%s  Failed to install minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
	case "linux":
		switch cpu {
		case "amd64":
			_, _, err := utils.RunCommand("Downloading minikube", "done", verbose, "curl", "-Lo", "./minikube", "https://storage.googleapis.com/minikube/releases/v1.32.0/minikube-linux-amd64")
			if err != nil {
				fmt.Printf("\n%s  Failed to download minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
				return false
			}
		case "arm64":
			_, _, err := utils.RunCommand("Downloading minikube", "done", verbose, "curl", "-Lo", "./minikube", "https://storage.googleapis.com/minikube/releases/v1.32.0/minikube-linux-arm64")
			if err != nil {
				fmt.Printf("\n%s  Failed to download minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
				return false
			}
		case "arm":
			_, _, err := utils.RunCommand("Downloading minikube", "done", verbose, "curl", "-Lo", "./minikube", "https://storage.googleapis.com/minikube/releases/v1.32.0/minikube-linux-arm")
			if err != nil {
				fmt.Printf("\n%s  Failed to download minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
				return false
			}
		case "PPC64":
			_, _, err := utils.RunCommand("Downloading minikube", "done", verbose, "curl", "-Lo", "./minikube", "https://storage.googleapis.com/minikube/releases/v1.32.0/minikube-linux-ppc64le")
			if err != nil {
				fmt.Printf("\n%s  Failed to download minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
				return false
			}
		case "S390X":
			_, _, err := utils.RunCommand("Downloading minikube", "done", verbose, "curl", "-Lo", "./minikube", "https://storage.googleapis.com/minikube/releases/v1.32.0/minikube-linux-s390x")
			if err != nil {
				fmt.Printf("\n%s  Failed to download minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
				return false
			}
		default:
			fmt.Printf("\n%s  Minikube doesn't support %s on %s%s\n\n", cpu, utils.ColorRed(), osName, utils.ColorReset())
			return false
		}
		_, _, err := utils.RunCommand("Updating minikube access", "done", verbose, "chmod", "+x", "./minikube")
		if err != nil {
			fmt.Printf("\n%s  Failed to update minikube access.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		_, _, err = utils.RunCommand("Installing minikube", "done", verbose, "sudo", "install", "./minikube", "/usr/local/bin/minikube")
		if err != nil {
			fmt.Printf("\n%s  Failed to install minikube.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
	default:
		fmt.Printf("\n%s  Doesn't know how to install kubectl on %s%s\n\n", utils.ColorRed(), osName, utils.ColorReset())
		return false
	}
	return true
}
func installKubectl() bool {
	os := runtime.GOOS
	switch os {
	case "windows":
		_, _, err := utils.RunCommand("Downloading kubectl", "done", verbose, "curl", "-LO", "https://dl.k8s.io/release/v1.24.0/bin/windows/amd64/kubectl.exe")
		if err != nil {
			fmt.Printf("\n%s  Failed to download kubectl.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
	case "darwin":
		_, _, err := utils.RunCommand("Downloading kubectl", "done", verbose, "curl", "-LO", "https://dl.k8s.io/release/v1.25.2/bin/darwin/amd64/kubectl")
		if err != nil {
			fmt.Printf("\n%s  Failed to download kubectl.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		_, _, err = utils.RunCommand("Updating kubectl access", "done", verbose, "chmod", "+x", "./kubectl")
		if err != nil {
			fmt.Printf("\n%s  Failed to update kubectl access.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		_, _, err = utils.RunCommand("Moving kubectl", "done", verbose, "sudo", "mv", "./kubectl", "/usr/local/bin/kubectl")
		if err != nil {
			fmt.Printf("\n%s  Failed to move kubectl.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		_, _, err = utils.RunCommand("Updating kubectl access", "done", verbose, "sudo", "chown", "root:", "/usr/local/bin/kubectl")
		if err != nil {
			fmt.Printf("\n%s  Failed to update kubectl access.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
	case "linux":
		_, _, err := utils.RunCommand("Downloading kubectl", "done", verbose, "curl", "-LO", "https://dl.k8s.io/release/v1.25.2/bin/linux/amd64/kubectl")
		if err != nil {
			fmt.Printf("\n%s  Failed to download kubectl.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
		_, _, err = utils.RunCommand("Installing kubectl", "done", verbose, "sudo", "install", "-o", "root", "-g", "root", "-m", "0755", "kubectl", "/usr/local/bin/kubectl")
		if err != nil {
			fmt.Printf("\n%s  Failed to install kubectl.%s\n\n", utils.ColorRed(), utils.ColorReset())
			return false
		}
	default:
		fmt.Printf("\n%s  Doesn't know how to install kubectl on %s%s\n\n", utils.ColorRed(), os, utils.ColorReset())
		return false
	}
	return true
}
