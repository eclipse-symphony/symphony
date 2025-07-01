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
	remoteMqtt "github.com/eclipse-symphony/symphony/remote-agent/bindings/mqtt"
	remoteProviders "github.com/eclipse-symphony/symphony/remote-agent/providers"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kardianos/service"
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
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	if err := mainLogic(); err != nil {
		log.Fatalf("Service run error: %v", err)
	}
}

func (p *program) Stop(s service.Service) error {
	log.Println("Service is stopping...")
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "symphony-service",
		DisplayName: "Remote Agent Service",
		Description: "Remote Agent Service",
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	// support command line arguments for install, uninstall, start, stop
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "install", "uninstall", "start", "stop":
			err := service.Control(s, cmd)
			if err != nil {
				log.Fatalf("Service control error: %v", err)
			}
			return
		}
	}
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}

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
	// 读取 CA 路径，可以通过命令行参数、环境变量或配置文件传入
	caPath := "ca.crt" // 默认值，可根据实际情况修改
	flag.StringVar(&caPath, "ca-cert", caPath, "Path to the CA certificate file")
	flag.Parse()

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
	// 加载 MQTT CA 证书
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
	// httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	// compose target providers
	providers := composeTargetProviders(topologyFile)
	if providers == nil {
		return fmt.Errorf("failed to compose target providers")
	}

	// // create HttpBinding
	// h := &remoteHttp.HttpBinding{
	// Agent: agent.Agent{
	// 	Providers: providers,
	// },
	// }
	// h.RequestUrl = symphonyEndpoints.RequestEndpoint
	// h.ResponseUrl = symphonyEndpoints.ResponseEndpoint
	// h.Client = httpClient
	// h.Target = targetName
	// h.Namespace = namespace

	// // start HttpBinding
	// if err := h.Launch(); err != nil {
	// 	return fmt.Errorf("error launching HttpBinding: %v", err)
	// }
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tls://10.172.3.39:8883")
	opts.SetTLSConfig(tlsConfig)
	fmt.Printf("MQTT TLS config: cert=%s, key=%s, ca=%s\n", clientCertPath, clientKeyPath, caPath)
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
		Client:        mqttClient, // 这里要加上
		Target:        targetName,
		RequestTopic:  "symphony/request",
		ResponseTopic: "symphony/response",
		Namespace:     namespace,
	}
	if err := m.Launch(); err != nil {
		return fmt.Errorf("failed to launch MQTT binding: %v", err)
	}
	// keep the main function running
	select {}
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
