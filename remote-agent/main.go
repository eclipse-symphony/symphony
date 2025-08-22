package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/docker"
	targethttp "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/http"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/script"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/win10/sideload"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
	remoteHttp "github.com/eclipse-symphony/symphony/remote-agent/bindings/http"
	remoteProviders "github.com/eclipse-symphony/symphony/remote-agent/providers"
)

const version = "0.0.0.1"

var (
	symphonyEndpoints model.SymphonyEndpoint
	clientCertPath    string
	configPath        string
	clientKeyPath     string
	namespace         string
	targetName        string
	topologyFile      string
	httpClient        *http.Client
	execDir           string
	rLog              logger.Logger
)

func mainLogic() error {
	// get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	execDir = filepath.Dir(execPath)
	os.Setenv("SYMPHONY_REMOTE_AGENT", "true")
	// log file
	rLog = logger.NewLogger("remote-agent")
	// extract command line arguments
	flag.StringVar(&configPath, "config", "config.json", "Path to the configuration file")
	flag.StringVar(&clientCertPath, "client-cert", "public.pem", "Path to the client certificate file")
	flag.StringVar(&clientKeyPath, "client-key", "private.pem", "Path to the client key file")
	flag.StringVar(&targetName, "target-name", "remote-target", "remote target name")
	flag.StringVar(&namespace, "namespace", "default", "Namespace to use for the agent")
	flag.StringVar(&topologyFile, "topology", "topology.json", "Path to the topology file")
	flag.Parse()

	// read configuration
	setting, err := os.ReadFile(configPath)
	if err != nil {
		rLog.Errorf("error reading configuration file: %v", err)
		return fmt.Errorf("error reading configuration file: %v", err)
	}
	if err := json.Unmarshal(setting, &symphonyEndpoints); err != nil {
		rLog.Errorf("error unmarshalling configuration file: %v", err)
		return fmt.Errorf("error unmarshalling configuration file: %v", err)
	}

	// load certificates
	rLog.Infof("Loading client certificate from %s and key from %s", clientCertPath, clientKeyPath)
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		rLog.Errorf("error loading client certificate and key: %v", err)
		return fmt.Errorf("error loading client certificate and key: %v", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{clientCert}}
	httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	// compose target providers
	providers := composeTargetProviders(topologyFile)
	if providers == nil {
		rLog.Errorf("failed to compose target providers")
		return fmt.Errorf("failed to compose target providers")
	}

	// create HttpBinding
	h := &remoteHttp.HttpBinding{
		Agent: agent.Agent{
			Providers: providers,
			RLog:      rLog,
		},
		RLog: rLog,
	}
	h.RequestUrl = symphonyEndpoints.RequestEndpoint
	h.ResponseUrl = symphonyEndpoints.ResponseEndpoint
	h.Client = httpClient
	h.Target = targetName
	h.Namespace = namespace

	// start HttpBinding
	if err := h.Launch(); err != nil {
		rLog.Errorf("error launching HttpBinding: %v", err)
		return fmt.Errorf("error launching HttpBinding: %v", err)
	}

	// keep the main function running
	select {}
}

func composeTargetProviders(topologyPath string) map[string]tgt.ITargetProvider {
	topologyContent, err := os.ReadFile(topologyPath)
	if err != nil {
		rLog.Errorf("Error reading topology file: %v", err)
		return nil
	}
	var topology model.TopologySpec
	json.Unmarshal(topologyContent, &topology)
	providers := make(map[string]tgt.ITargetProvider)
	for _, binding := range topology.Bindings {
		switch binding.Provider {
		case "providers.target.script":
			provider := &script.ScriptProvider{}
			if err := provider.Init(binding.Config); err != nil {
				rLog.Errorf("Error initializing script provider: %v", err)
			}
			providers[binding.Role] = provider
		case "providers.target.remote-agent":
			rProvider := &remoteProviders.RemoteAgentProvider{}
			rProvider.Client = httpClient
			rProviderConfig := remoteProviders.RemoteAgentProviderConfig{
				PublicCertPath: clientCertPath,
				PrivateKeyPath: clientKeyPath,
				ConfigPath:     configPath,
				BaseUrl:        symphonyEndpoints.BaseUrl,
				Version:        version,
				Namespace:      namespace,
				TargetName:     targetName,
				TopologyPath:   topologyPath,
				ExecDir:        execDir,
			}
			if err := rProvider.Init(rProviderConfig); err != nil {
				rLog.Errorf("Error remote agent provider: %v", err)
			}
			providers[binding.Role] = rProvider
		case "providers.target.win10.sideload":
			mProvider := &sideload.Win10SideLoadProvider{}
			if err := mProvider.Init(binding.Config); err != nil {
				rLog.Errorf("Error initializing win10.sideload provider: %v", err)
			}
			providers[binding.Role] = mProvider
		case "providers.target.docker":
			mProvider := &docker.DockerTargetProvider{}
			if err := mProvider.Init(binding.Config); err != nil {
				rLog.Errorf("Error initializing docker provider: %v", err)
			}
			providers[binding.Role] = mProvider
		case "providers.target.http":
			mProvider := &targethttp.HttpTargetProvider{}
			if err := mProvider.Init(binding.Config); err != nil {
				rLog.Errorf("Error initializing http provider: %v", err)
			}
			providers[binding.Role] = mProvider
		default:
			rLog.Errorf("Unknown provider type: %s", binding.Role)
		}
	}
	return providers
}
