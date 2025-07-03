package main

import (
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
	protocol          string = "http" // default protocol
	caPath            string
)

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
	log.Printf("Using protocol: %s", protocol)
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

	// read configuration
	setting, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading configuration file: %v", err)
	}
	if err := json.Unmarshal(setting, &symphonyEndpoints); err != nil {
		return fmt.Errorf("error unmarshalling configuration file: %v", err)
	}

	// load certificates
	log.Printf("Loading client certificate from %s and key from %s", clientCertPath, clientKeyPath)
	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load client cert/key: %v", err)
	}
	// 打印 client cert 的 subject name，并用作 topic 前缀
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

	// httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	// compose target providers
	providers := composeTargetProviders(topologyFile)
	if providers == nil {
		return fmt.Errorf("failed to compose target providers")
	}

	if protocol == "http" {
		fmt.Printf("Using HTTP protocol with endpoints: Request=%s, Response=%s\n", symphonyEndpoints.RequestEndpoint, symphonyEndpoints.ResponseEndpoint)
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

		// create HttpBinding
		h := &remoteHttp.HttpBinding{
			Agent: agent.Agent{
				Providers: providers,
			},
		}
		h.RequestUrl = symphonyEndpoints.RequestEndpoint
		h.ResponseUrl = symphonyEndpoints.ResponseEndpoint
		h.Client = httpClient
		h.Target = targetName
		h.Namespace = namespace
		if err := h.Launch(); err != nil {
			return fmt.Errorf("error launching HttpBinding: %v", err)
		}
		select {}
	} else {
		// 加载 MQTT CA 证书
		fmt.Printf("Loading CA certificate from %s\n", caPath)
		caCertPool := x509.NewCertPool()
		caCert, err := os.ReadFile(caPath)
		if err != nil {
			return fmt.Errorf("failed to read MQTT CA file: %v", err)
		}
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return fmt.Errorf("failed to append MQTT CA cert")
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ServerName:   "10.172.3.39", // 必须和证书CN/SAN一致
			MinVersion:   tls.VersionTLS12,
			MaxVersion:   tls.VersionTLS13,
		}
		for i, c := range tlsConfig.Certificates {
			for _, cert := range c.Certificate {
				parsed, err := x509.ParseCertificate(cert)
				if err == nil {
					fmt.Printf("Loaded client cert[%d]: Subject=%s, Issuer=%s, NotAfter=%s\n", i, parsed.Subject, parsed.Issuer, parsed.NotAfter)
				}
			}
		}
		fmt.Printf("TLS MinVersion: %v, MaxVersion: %v\n", tlsConfig.MinVersion, tlsConfig.MaxVersion)
		opts := mqtt.NewClientOptions()
		opts.AddBroker("tls://10.172.3.39:8883")
		opts.SetTLSConfig(tlsConfig)
		// 设置 client id
		fmt.Printf("MQTT TLS config: cert=%s, key=%s, ca=%s, clientID=%s\n", clientCertPath, clientKeyPath, caPath)
		fmt.Printf("begin to connect to MQTT broker %s\n", "tls://10.172.3.39:8883")
		mqttClient := mqtt.NewClient(opts)
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			fmt.Printf("failed to connect to MQTT broker: %v\n", token.Error())
			return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
		} else {
			fmt.Println("Connected to MQTT broker")
		}
		fmt.Println("Begin to request topic", "ddc")
		m := &remoteMqtt.MqttBinding{
			Agent: agent.Agent{
				Providers: providers,
			},
			Client:        mqttClient,
			Target:        targetName,
			RequestTopic:  fmt.Sprintf("symphony/request/%s", subjectName),
			ResponseTopic: fmt.Sprintf("symphony/response/%s", subjectName),
			Namespace:     namespace,
		}
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
