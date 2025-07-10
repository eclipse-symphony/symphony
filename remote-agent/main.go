package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	tgt "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/docker"
	targethttp "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/http"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/script"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/win10/sideload"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
	remoteHttp "github.com/eclipse-symphony/symphony/remote-agent/bindings/http"
	remoteMqtt "github.com/eclipse-symphony/symphony/remote-agent/bindings/mqtt"
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
	log.Printf("mainLogic started, args: %v", os.Args)
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
	logFile, err := os.OpenFile(filepath.Join(execDir, "transcript.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
	// extract command line arguments
	flag.StringVar(&configPath, "config", "config.json", "Path to the configuration file")
	flag.StringVar(&clientCertPath, "client-cert", "public.pem", "Path to the client certificate file")
	flag.StringVar(&clientKeyPath, "client-key", "private.pem", "Path to the client key file")
	flag.StringVar(&targetName, "target-name", "remote-target", "remote target name")
	flag.StringVar(&namespace, "namespace", "default", "Namespace to use for the agent")
	flag.StringVar(&topologyFile, "topology", "topology.json", "Path to the topology file")
	flag.StringVar(&caPath, "ca-cert", caPath, "Path to the CA certificate file")
	flag.StringVar(&protocol, "protocol", "http", "Protocol to use: mqtt or http")
	flag.BoolVar(&useCertSubject, "use-cert-subject", false, "Use certificate subject as topic suffix instead of target name")
	flag.Parse()
	if protocol == "mqtt" && (caPath == "") {
		fmt.Print("please enter the MQTT CA certificate path (e.g., ca.pem): ")
		var input string
		fmt.Scanln(&input)
		if input != "" {
			caPath = input
		}
		fmt.Printf("Using CA certificate path: %s\n", caPath)
	}
	fmt.Printf("Using client certificate path: %s\n", clientCertPath)
	// read configuration
	setting, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading configuration file: %v", err)
	}

	// Remove UTF-8 BOM if present
	if len(setting) >= 3 && setting[0] == 0xEF && setting[1] == 0xBB && setting[2] == 0xBF {
		fmt.Println("Removing UTF-8 BOM from config file")
		setting = setting[3:]
	}

	// Use the new config structure
	var config SymphonyConfig
	if err := json.Unmarshal(setting, &config); err != nil {
		fmt.Printf("Error unmarshalling config: %v\nContent: %s\n", err, string(setting))
		return fmt.Errorf("error unmarshalling configuration file: %v", err)
	}

	// load certificates
	log.Printf("Loading client certificate from %s and key from %s", clientCertPath, clientKeyPath)
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
			fmt.Printf("Client certificate subject: %s\n", subjectName)
		} else {
			fmt.Printf("Failed to parse client certificate for subject: %v\n", err)
		}
	}

	// compose target providers
	providers := composeTargetProviders(topologyFile)
	if providers == nil {
		return fmt.Errorf("failed to compose target providers")
	}

	// Read topology file content for later updates
	topologyContent, err := os.ReadFile(topologyFile)
	if err != nil {
		log.Printf("Error reading topology file: %v", err)
		return fmt.Errorf("failed to read topology file: %v", err)
	}

	if protocol == "http" {
		if config.RequestEndpoint == "" || config.ResponseEndpoint == "" || config.BaseUrl == "" {
			fmt.Errorf("RequestEndpoint, ResponseEndpoint, and BaseUrl must be set in the configuration file")
			return fmt.Errorf("RequestEndpoint, ResponseEndpoint, and BaseUrl must be set in the configuration file")
		}
		fmt.Printf("Using HTTP protocol with endpoints: Request=%s, Response=%s\n", config.RequestEndpoint, config.ResponseEndpoint)
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
		symphonyEndpoints.RequestEndpoint = config.RequestEndpoint
		// create HttpBinding
		h := &remoteHttp.HttpBinding{
			Agent: agent.Agent{
				Providers: providers,
			},
		}
		h.RequestUrl = config.RequestEndpoint
		h.ResponseUrl = config.ResponseEndpoint
		h.Client = httpClient
		h.Target = targetName
		h.Namespace = namespace

		// send topology update via HTTP
		updateTopologyEndpoint := fmt.Sprintf("%s/targets/updatetopology/%s?namespace=%s",
			config.BaseUrl, targetName, namespace)
		log.Printf("Sending topology update via HTTP: %s", updateTopologyEndpoint)

		resp, err := httpClient.Post(updateTopologyEndpoint, "application/json", bytes.NewBuffer(topologyContent))
		if err != nil {
			return fmt.Errorf("failed to update topology: %v", err)
		}
		defer resp.Body.Close()

		// check response status code
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("topology update failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		log.Printf("Topology updated successfully via HTTP")

		if err := h.Launch(); err != nil {
			return fmt.Errorf("error launching HttpBinding: %v", err)
		}
		select {}
	} else {
		// Load MQTT CA certificate
		fmt.Printf("Loading CA certificate from %s\n", caPath)
		caCertPool := x509.NewCertPool()
		caCert, err := os.ReadFile(caPath)
		if err != nil {
			return fmt.Errorf("failed to read MQTT CA file: %v", err)
		}
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return fmt.Errorf("failed to append MQTT CA cert")
		}

		// Use the configuration values from config.json
		fmt.Printf("Config loaded: broker=%s, port=%d, target=%s\n",
			config.MqttBroker, config.MqttPort, config.TargetName)
		brokerAddr := config.MqttBroker
		brokerPort := config.MqttPort

		// If target name is specified in config and command line is empty, use the config value
		if config.TargetName != "" {
			targetName = config.TargetName
			fmt.Printf("Using target name from config: %s\n", targetName)
		}

		// If namespace is specified in config and command line is empty, use the config value
		if namespace == "default" && config.Namespace != "" {
			namespace = config.Namespace
			fmt.Printf("Using namespace from config: %s\n", namespace)
		}

		if brokerAddr == "" {
			fmt.Print("MQTT broker address not found in config. Please enter MQTT broker address: ")
			fmt.Scanln(&brokerAddr)
			if brokerAddr == "" {
				return fmt.Errorf("MQTT broker address is required")
			}
		}

		if brokerPort == 0 {
			fmt.Print("MQTT broker port not found in config. Please enter MQTT broker port: ")
			var portStr string
			fmt.Scanln(&portStr)
			if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
				brokerPort = p
			} else {
				fmt.Println("Using default MQTT TLS port 8883")
				brokerPort = 8883 // Default MQTT TLS port
			}
		}

		brokerUrl := fmt.Sprintf("tls://%s:%d", brokerAddr, brokerPort)
		fmt.Printf("Using MQTT broker: %s\n", brokerUrl)

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ServerName:   brokerAddr, // Use broker address from config instead of hardcoded value
			MinVersion:   tls.VersionTLS12,
			MaxVersion:   tls.VersionTLS13,
		}
		opts := mqtt.NewClientOptions()
		opts.AddBroker(brokerUrl)
		opts.SetTLSConfig(tlsConfig)
		opts.SetClientID(strings.ToLower(targetName)) // Ensure lowercase is used
		// Set client ID
		fmt.Printf("MQTT TLS config: cert=%s, key=%s, ca=%s, clientID=%s\n",
			clientCertPath, clientKeyPath, caPath, strings.ToLower(targetName))
		fmt.Printf("begin to connect to MQTT broker %s\n", brokerUrl)
		mqttClient := mqtt.NewClient(opts)
		// Ensure lowercase is used
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			fmt.Printf("failed to connect to MQTT broker: %v\n", token.Error())
			return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
		} else {
			fmt.Println("Connected to MQTT broker")
		}
		fmt.Println("Begin to request topic", "ddc")
		// 选择 topic 后缀
		topicSuffix := strings.ToLower(targetName)
		if useCertSubject && subjectName != "" {
			topicSuffix = strings.ToLower(subjectName)
			fmt.Printf("Using certificate subject as topic suffix: %s\n", topicSuffix)
		} else {
			fmt.Printf("Using target name as topic suffix: %s\n", topicSuffix)
		}
		m := &remoteMqtt.MqttBinding{
			Agent: agent.Agent{
				Providers: providers,
			},
			Client:        mqttClient,
			Target:        targetName,
			RequestTopic:  fmt.Sprintf("symphony/request/%s", topicSuffix),
			ResponseTopic: fmt.Sprintf("symphony/response/%s", topicSuffix),
			Namespace:     namespace,
		}

		// Update topology configuration - this operation will first subscribe to the response topic
		log.Printf("Sending topology update via MQTT and waiting for confirmation...")
		if err := m.UpdateTopology(topologyContent); err != nil {
			return fmt.Errorf("topology update failed: %v", err)
		}
		log.Printf("Topology update confirmed successful")
		if err := m.Launch(); err != nil {
			return fmt.Errorf("failed to launch MQTT binding: %v", err)
		}

		select {}
	}
}

func composeTargetProviders(topologyPath string) map[string]tgt.ITargetProvider {
	topologyContent, err := os.ReadFile(topologyPath)
	if err != nil {
		log.Printf("Error reading topology file: %v", err)
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
				log.Printf("Error initializing script provider: %v", err)
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
				log.Printf("Error remote agent provider: %v", err)
			}
			providers[binding.Role] = rProvider
		case "providers.target.win10.sideload":
			mProvider := &sideload.Win10SideLoadProvider{}
			if err := mProvider.Init(binding.Config); err != nil {
				log.Printf("Error initializing win10.sideload provider: %v", err)
			}
			providers[binding.Role] = mProvider
		case "providers.target.docker":
			mProvider := &docker.DockerTargetProvider{}
			if err := mProvider.Init(binding.Config); err != nil {
				log.Printf("Error initializing docker provider: %v", err)
			}
			providers[binding.Role] = mProvider
		case "providers.target.http":
			mProvider := &targethttp.HttpTargetProvider{}
			if err := mProvider.Init(binding.Config); err != nil {
				log.Printf("Error initializing http provider: %v", err)
			}
			providers[binding.Role] = mProvider
		default:
			log.Printf("Unknown provider type: %s", binding.Role)
		}
	}
	return providers
}
