package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/docker"
	targethttp "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/http"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/script"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/win10/sideload"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
	remoteHttp "github.com/eclipse-symphony/symphony/remote-agent/pollers/http"
	remoteMqtt "github.com/eclipse-symphony/symphony/remote-agent/pollers/mqtt"
	remoteProviders "github.com/eclipse-symphony/symphony/remote-agent/providers"
	mqtt "github.com/eclipse/paho.mqtt.golang"
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
	protocol          string
	caPath            string
	useCertSubject    bool
)

// Add this structure to handle MQTT configuration
type SymphonyConfig struct {
	// HTTP fields
	RequestEndpoint  string `json:"requestEndpoint"`
	ResponseEndpoint string `json:"responseEndpoint"`
	BaseUrl          string `json:"baseUrl"`

	// MQTT fields
	MqttBroker string `json:"mqttBroker"`
	MqttPort   int    `json:"mqttPort"`
	TargetName string `json:"targetName"`
	Namespace  string `json:"namespace"`
}

func mainLogic() error {
	rLog = logger.NewLogger("remote-agent")
	rLog.Infof("mainLogic started, args: %v", os.Args)
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
	// log file

	// Log the complete binary execution command
	rLog.Infof("=== Binary Execution Command ===")
	rLog.Infof("Executable: %s", execPath)
	rLog.Infof("Working Directory: %s", execDir)
	rLog.Infof("Command Line: %s %s", execPath, strings.Join(os.Args[1:], " "))
	rLog.Infof("Full Args: %v", os.Args)
	rLog.Infof("================================")
	// extract command line arguments
	flag.StringVar(&configPath, "config", "config.json", "Path to the configuration file")
	flag.StringVar(&clientCertPath, "client-cert", "public.pem", "Path to the client certificate file")
	flag.StringVar(&clientKeyPath, "client-key", "private.pem", "Path to the client key file")
	flag.StringVar(&targetName, "target-name", "remote-target", "remote target name")
	flag.StringVar(&namespace, "namespace", "default", "Namespace to use for the agent")
	flag.StringVar(&topologyFile, "topology", "topology.json", "Path to the topology file")
	flag.StringVar(&caPath, "ca-cert", caPath, "Path to the CA certificate file (for MQTT) or Symphony server CA (for HTTP)")
	flag.StringVar(&protocol, "protocol", "http", "Protocol to use: mqtt or http")
	flag.BoolVar(&useCertSubject, "use-cert-subject", false, "Use certificate subject as topic suffix instead of target name")
	flag.Parse()
	caPath = promptForMqttCaCertIfNeeded(protocol, caPath)

	rLog.Infof("Using client certificate path: %s", clientCertPath)
	// read configuration
	setting, err := os.ReadFile(configPath)
	if err != nil {
		rLog.Errorf("error reading configuration file: %v", err)
		return fmt.Errorf("error reading configuration file: %v", err)
	}

	// Remove UTF-8 BOM if present
	if len(setting) >= 3 && setting[0] == 0xEF && setting[1] == 0xBB && setting[2] == 0xBF {
		rLog.Infof("Removing UTF-8 BOM from config file")
		setting = setting[3:]
	}

	// Use the new config structure
	var config SymphonyConfig
	if err := json.Unmarshal(setting, &config); err != nil {
		rLog.Infof("Error unmarshalling config: %v, Content: %s", err, string(setting))
		return fmt.Errorf("error unmarshalling configuration file: %v", err)
	}

	// load certificates
	rLog.Infof("Loading client certificate from %s and key from %s", clientCertPath, clientKeyPath)
	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load client cert/key: %v", err)
	}
	subjectName := ""
	if len(cert.Certificate) > 0 {
		parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil {
			subjectName = parsedCert.Subject.CommonName
			if subjectName == "" {
				subjectName = parsedCert.Subject.String()
			}
			rLog.Infof("Client certificate subject: %s", subjectName)
		} else {
			rLog.Errorf("Failed to parse client certificate for subject: %v", err)
		}
	}

	// compose target providers
	providers := composeTargetProviders(topologyFile)
	if providers == nil {
		rLog.Errorf("failed to compose target providers")
		return fmt.Errorf("failed to compose target providers")
	}

	// Read topology file content for later updates
	topologyContent, err := os.ReadFile(topologyFile)
	if err != nil {
		rLog.Errorf("Error reading topology file: %v", err)
		return fmt.Errorf("failed to read topology file: %v", err)
	}

	if protocol == "http" {
		if config.RequestEndpoint == "" || config.ResponseEndpoint == "" || config.BaseUrl == "" {
			return fmt.Errorf("RequestEndpoint, ResponseEndpoint, and BaseUrl must be set in the configuration file")
		}
		rLog.Infof("Using HTTP protocol with endpoints: Request=%s, Response=%s", config.RequestEndpoint, config.ResponseEndpoint)
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

		// If HTTP protocol and CA certificate is specified, add it to the trusted roots for Symphony server
		if caPath != "" {
			rLog.Errorf("Loading Symphony server CA certificate from: %s", caPath)
			serverCACert, err := os.ReadFile(caPath)
			if err != nil {
				return fmt.Errorf("failed to read Symphony server CA certificate: %v", err)
			}

			serverCACertPool := x509.NewCertPool()
			if !serverCACertPool.AppendCertsFromPEM(serverCACert) {
				return fmt.Errorf("failed to parse Symphony server CA certificate")
			}

			tlsConfig.RootCAs = serverCACertPool
			rLog.Infof("Successfully loaded Symphony server CA certificate")
		}

		httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
		symphonyEndpoints.RequestEndpoint = config.RequestEndpoint
		// create HttpBinding
		h := &remoteHttp.HttpPoller{
			Agent: agent.Agent{
				Providers: providers,
				RLog:      rLog,
			},
			RLog: rLog,
		}
		h.RequestUrl = config.RequestEndpoint
		h.ResponseUrl = config.ResponseEndpoint
		h.Client = httpClient
		h.Target = targetName
		h.Namespace = namespace

		// send topology update via HTTP
		updateTopologyEndpoint := fmt.Sprintf("%s/targets/updatetopology/%s?namespace=%s",
			config.BaseUrl, targetName, namespace)
		rLog.Infof("Sending topology update via HTTP: %s", updateTopologyEndpoint)

		resp, err := httpClient.Post(updateTopologyEndpoint, "application/json", bytes.NewBuffer(topologyContent))
		if err != nil {
			return fmt.Errorf("failed to update topology: %v", err)
		}
		defer resp.Body.Close()

		// check response status code
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("topology update failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		rLog.Infof("Topology updated successfully via HTTP")

		if err := h.Launch(); err != nil {
			return fmt.Errorf("error launching HttpBinding: %v", err)
		}
		select {}
	} else {
		// Load MQTT CA certificate
		rLog.Infof("Loading CA certificate from %s", caPath)
		caCertPool := x509.NewCertPool()
		caCert, err := os.ReadFile(caPath)
		if err != nil {
			return fmt.Errorf("failed to read MQTT CA file: %v", err)
		}
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return fmt.Errorf("failed to append MQTT CA cert")
		}

		// Use the configuration values from config.json
		rLog.Infof("Config loaded: broker=%s, port=%d, target=%s",
			config.MqttBroker, config.MqttPort, config.TargetName)
		brokerAddr := config.MqttBroker
		brokerPort := config.MqttPort

		// If target name is specified in config and command line is empty, use the config value
		if config.TargetName != "" {
			targetName = config.TargetName
			rLog.Infof("Using target name from config: %s", targetName)
		}

		// If namespace is specified in config and command line is empty, use the config value
		if namespace == "default" && config.Namespace != "" {
			namespace = config.Namespace
			rLog.Infof("Using namespace from config: %s", namespace)
		}
		brokerAddr, brokerPort, err = getBrokerAddressAndPort(brokerAddr, brokerPort)
		if err != nil {
			return err
		}
		brokerUrl := fmt.Sprintf("tls://%s:%d", brokerAddr, brokerPort)
		rLog.Infof("Using MQTT broker: %s", brokerUrl)

		// Determine the correct ServerName for TLS verification
		// If connecting to 127.0.0.1 or localhost, use "localhost" as ServerName
		serverName := brokerAddr
		if brokerAddr == "127.0.0.1" || brokerAddr == "::1" {
			serverName = "localhost"
		}
		rLog.Infof("Using ServerName '%s' for TLS verification (connecting to %s)", serverName, brokerAddr)

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ServerName:   serverName, // Use appropriate ServerName for certificate verification
			MinVersion:   tls.VersionTLS12,
			MaxVersion:   tls.VersionTLS13,
		}
		opts := mqtt.NewClientOptions()
		opts.AddBroker(brokerUrl)
		opts.SetTLSConfig(tlsConfig)
		opts.SetClientID(strings.ToLower(targetName)) // Ensure lowercase is used
		// Set client ID
		rLog.Infof("MQTT TLS config: cert=%s, key=%s, ca=%s, clientID=%s",
			clientCertPath, clientKeyPath, caPath, strings.ToLower(targetName))
		rLog.Infof("begin to connect to MQTT broker %s", brokerUrl)
		mqttClient := mqtt.NewClient(opts)
		// Ensure lowercase is used
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			rLog.Errorf("failed to connect to MQTT broker: %v", token.Error())
			return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
		} else {
			rLog.Infof("Connected to MQTT broker")
		}
		// choose topic suffix
		topicSuffix := getTopicSuffix(targetName, subjectName, useCertSubject)
		m := &remoteMqtt.MqttPoller{
			Agent: agent.Agent{
				Providers: providers,
				RLog:      rLog,
			},
			Client:        mqttClient,
			Target:        targetName,
			RequestTopic:  fmt.Sprintf("symphony/request/%s", topicSuffix),
			ResponseTopic: fmt.Sprintf("symphony/response/%s", topicSuffix),
			Namespace:     namespace,
			RLog:          rLog,
		}

		// First establish MQTT subscription for responses
		rLog.Infof("Setting up MQTT subscription for responses...")
		if err := m.Subscribe(); err != nil {
			return fmt.Errorf("failed to setup MQTT subscription: %v", err)
		}

		// Keep retrying until the topology update is confirmed.
		retryInterval := 2 * time.Minute
		for {
			if err := m.UpdateTopology(topologyContent); err != nil {
				rLog.Errorf("Topology update failed: %v. Retrying in %s", err, retryInterval)
				time.Sleep(retryInterval)
				continue
			}
			rLog.Infof("Topology update confirmed successful")
			break
		}

		rLog.Infof("Topology update confirmed successful")
		if err := m.Launch(); err != nil {
			return fmt.Errorf("failed to launch MQTT binding: %v", err)
		}

		select {}
	}
}

func getBrokerAddressAndPort(brokerAddr string, brokerPort int) (string, int, error) {
	if brokerAddr == "" {
		rLog.Infof("MQTT broker address not found in config. Please enter MQTT broker address: ")
		fmt.Scanln(&brokerAddr)
		if brokerAddr == "" {
			return "", 0, fmt.Errorf("MQTT broker address is required")
		}
	}

	if brokerPort == 0 {
		rLog.Infof("MQTT broker port not found in config. Please enter MQTT broker port: ")
		var portStr string
		fmt.Scanln(&portStr)
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			brokerPort = p
		} else {
			rLog.Infof("Using default MQTT TLS port 8883")
			brokerPort = 8883 // Default MQTT TLS port
		}
	}
	return brokerAddr, brokerPort, nil
}

func promptForMqttCaCertIfNeeded(protocol string, caPath string) string {
	if protocol == "mqtt" && caPath == "" {
		fmt.Print("please enter the MQTT CA certificate path (e.g., ca.pem): ")
		var input string
		fmt.Scanln(&input)
		if input != "" {
			caPath = input
		}
		rLog.Infof("Using CA certificate path: %s", caPath)
	}
	return caPath
}

func getTopicSuffix(targetName, subjectName string, useCertSubject bool) string {
	if useCertSubject && subjectName != "" {
		s := strings.ToLower(subjectName)
		rLog.Infof("Using certificate subject as topic suffix: %s", s)
		return s
	}
	s := strings.ToLower(targetName)
	rLog.Infof("Using target name as topic suffix: %s", s)
	return s
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
