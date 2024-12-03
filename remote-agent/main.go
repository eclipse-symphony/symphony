package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"

	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/script"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
	"github.com/eclipse-symphony/symphony/remote-agent/bindings/http"
	utils "github.com/eclipse-symphony/symphony/remote-agent/common"
)

func main() {
	// Define a command-line flag for the configuration file path
	configPath := flag.String("config", "config.json", "Path to the configuration file")

	// Parse the command-line flags
	flag.Parse()

	// Read the configuration file
	setting, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Println("Error reading configuration file:", err)
		return
	}

	symphonyEndpoints := utils.SymphonyEndpoint{}
	err = json.Unmarshal(setting, &symphonyEndpoints)
	if err != nil {
		fmt.Println("Error unmarshalling configuration file:", err)
		return
	}

	// Compose target providers
	providers := composeTargetProviders()
	// Create the HttpBinding instance
	h := &http.HttpBinding{
		Agent: agent.Agent{
			Providers: providers,
		},
	}

	// Set up the configuration
	config := http.HttpBindingConfig{
		TLS: true,
		CertProvider: http.CertProviderConfig{
			Type:   "certs.localfile",
			Config: map[string]interface{}{"certFile": "path/to/cert.pem", "keyFile": "path/to/key.pem"},
		},
	}

	// Set up the request and response URLs
	requestUrl, err := url.Parse(symphonyEndpoints.RequestEndpoint)
	if err != nil {
		fmt.Println("Error parsing request URL:", err)
		return
	}
	responseUrl, err := url.Parse(symphonyEndpoints.ResponseEndpoint)
	if err != nil {
		fmt.Println("Error parsing response URL:", err)
		return
	}
	h.RequestUrl = requestUrl
	h.ResponseUrl = responseUrl

	// Launch the HttpBinding
	err = h.Launch(config)
	if err != nil {
		fmt.Println("Error launching HttpBinding:", err)
		return
	}

	// Keep the main function running
	select {}
}

func composeTargetProviders() map[string]tgt.ITargetProvider {
	providers := make(map[string]tgt.ITargetProvider)
	// Add the target providers to the map
	// Add the script provider
	mProvider := &script.ScriptProvider{}
	providerConfig := script.ScriptProviderConfig{
		ApplyScript:   "mock-apply.sh",
		GetScript:     "mock-get.sh",
		RemoveScript:  "mock-remove.sh",
		ScriptFolder:  "./script",
		StagingFolder: "./script",
	}
	err := mProvider.Init(providerConfig)
	if err != nil {
		fmt.Println("Error script provider:", err)
	}
	providers["providers.target.script"] = mProvider
	return providers
}
