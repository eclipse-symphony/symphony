package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"net/http"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/script"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
	remoteHttp "github.com/eclipse-symphony/symphony/remote-agent/bindings/http"
	utils "github.com/eclipse-symphony/symphony/remote-agent/common"
	remoteProviders "github.com/eclipse-symphony/symphony/remote-agent/providers"
)

// The version should be hardcoded in the build process
const version = "0.0.0.2"

var (
	symphonyEndpoints utils.SymphonyEndpoint
	clientCertPath    *string
	configPath        *string
	clientKeyPath     *string
	namespace         *string
	targetName        *string
	httpClient        *http.Client
)

func main() {
	// Allocate memory for shouldEnd
	// Define a command-line flag for the configuration file path
	configPath = flag.String("config", "config.json", "Path to the configuration file")
	clientCertPath = flag.String("client-cert", "client-cert.pem", "Path to the client certificate file")
	clientKeyPath = flag.String("client-key", "client-key.pem", "Path to the client key file")
	targetName = flag.String("target-name", "remote-target", "remote target name")
	namespace = flag.String("namespace", "default", "Namespace to use for the agent")
	topologyFile := flag.String("topology", "topology.json", "Path to the topology file")

	// Parse the command-line flags
	flag.Parse()

	// Read the configuration file
	setting, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Println("Error reading configuration file:", err)
		return
	}

	// Load client cert
	clientCert, err := tls.LoadX509KeyPair(*clientCertPath, *clientKeyPath)
	if err != nil {
		fmt.Println("Error loading client certificate and key:", err)
		return
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
	}

	// Create HTTP client with TLS configuration
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	print(httpClient)
	err = json.Unmarshal(setting, &symphonyEndpoints)
	if err != nil {
		fmt.Println("Error unmarshalling configuration file:", err)
		return
	}

	// Compose target providers
	providers := composeTargetProviders(*topologyFile)
	// Create the HttpBinding instance
	h := &remoteHttp.HttpBinding{
		Agent: agent.Agent{
			Providers: providers,
		},
	}

	// Set up the request and response URLs
	h.RequestUrl = symphonyEndpoints.RequestEndpoint
	h.ResponseUrl = symphonyEndpoints.ResponseEndpoint
	h.Client = httpClient
	h.Target = *targetName
	h.Namespace = *namespace

	// Launch the HttpBinding
	err = h.Launch()
	if err != nil {
		fmt.Println("Error launching HttpBinding:", err)
		return
	}

	// Keep the main function running
	select {}
}

func composeTargetProviders(topologyPath string) map[string]tgt.ITargetProvider {
	// read the topology file
	topologyContent, err := os.ReadFile(topologyPath)
	if err != nil {
		fmt.Println("Error reading topology file:", err)
		return nil
	}

	var topology model.TopologySpec
	json.Unmarshal(topologyContent, &topology)

	providers := make(map[string]tgt.ITargetProvider)
	// Add the target providers to the map
	// Add the script provider
	for _, binding := range topology.Bindings {
		switch binding.Role {
		case "script":
			provider := &script.ScriptProvider{}
			err := provider.Init(binding.Config)
			if err != nil {
				fmt.Println("Error initializing script provider:", err)
			}
			providers["script"] = provider
		case "remote-agent":
			rProvider := &remoteProviders.RemoteAgentProvider{}
			rProvider.Client = httpClient
			rProviderConfig := remoteProviders.RemoteAgentProviderConfig{
				PublicCertPath: *clientCertPath,
				PrivateKeyPath: *clientKeyPath,
				ConfigPath:     *configPath,
				BaseUrl:        symphonyEndpoints.BaseUrl,
				Version:        version,
				Namespace:      *namespace,
				TargetName:     *targetName,
				TopologyPath:   topologyPath,
			}
			err = rProvider.Init(rProviderConfig)
			if err != nil {
				fmt.Println("Error remote agent provider:", err)
			}
			providers["remote-agent"] = rProvider
		default:
			fmt.Println("Unknown provider type:", binding.Role)
		}

	}
	return providers
}
