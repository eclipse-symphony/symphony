package edge

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	southbound "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/edge/api/edge_adapter"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/edge/api/system_model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/edge/authprovider"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const loggerName = "providers.target.edge"

var sLog = logger.NewLogger(loggerName)

type EdgeProviderConfig struct {
	Name        string `json:"name"`
	User        string `json:"user"`
	Password    string `json:"password"`
	BaseAddress string `json:"baseAddress"`
}

type EdgeProvider struct {
	Context *contexts.ManagerContext
	Config  EdgeProviderConfig

	AuthService       *authprovider.AuthenticationService
	SystemClient      system_model.SystemModelClient
	EdgeAdapterClient southbound.EdgeAdapterServiceClient
	ApiClient         utils.ApiClient
}

func EdgeProviderConfigFromMap(properties map[string]string) (EdgeProviderConfig, error) {
	config := EdgeProviderConfig{}

	if name, ok := properties["name"]; ok {
		config.Name = name
	}

	if baseAddress, ok := properties["baseAddress"]; ok {
		config.BaseAddress = baseAddress
	} else {
		config.BaseAddress = os.Getenv("ADAPTER_URL") //"https://EAEP25:6201"
	}

	if api_utils.ShouldUseUserCreds() {
		user, err := api_utils.GetString(properties, "user")
		if err != nil {
			return config, err
		}
		config.User = user
		if config.User == "" {
			return config, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		}
		password, err := api_utils.GetString(properties, "password")
		config.Password = password
		if err != nil {
			return config, err
		}
	}

	return config, nil
}

func (h *EdgeProvider) InitWithMap(properties map[string]string) error {
	config, err := EdgeProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return h.Init(config)
}

func (h *EdgeProvider) SetContext(ctx *contexts.ManagerContext) {
	h.Context = ctx
}

func (h *EdgeProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("Edge Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Edge Target): Init()")

	edgeConfig, err := toEdgeProviderConfig(config)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to convert provider config", "error", err)
		return err
	}

	h.Config = edgeConfig
	h.AuthService, err = authprovider.NewAuthenticationService()
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to initialize authentication service", "error", err)
		return err
	}

	h.ApiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
	return nil
}

func toEdgeProviderConfig(config providers.IProviderConfig) (EdgeProviderConfig, error) {
	ret := EdgeProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (h *EdgeProvider) connectToAPI(ctx context.Context, sessionId string, credentials *tls.Config, endpoint string) error {
	var err error
	h.SystemClient, err = NewSystemModelClient(ctx, sessionId, credentials)
	return err
}

func (h *EdgeProvider) establishConnection(ctx context.Context) (context.Context, error) {
	sessionID, _, err := h.AuthService.GetSessionIdAsync(h.Config.BaseAddress)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to get session ID", "error", err)
		return ctx, err
	}

	md := metadata.Pairs(
		"cookie", fmt.Sprintf("sessionId=%s", sessionID),
		"content-type", "application/grpc",
	)
	ctxNew := metadata.NewOutgoingContext(ctx, md)

	if err := h.connectToAPI(ctxNew, sessionID, h.AuthService.Credentials, ""); err != nil {
		sLog.ErrorCtx(ctx, "Failed to connect to API", "error", err)
		return ctxNew, err
	}

	return ctxNew, nil
}

func (h *EdgeProvider) Get(ctx context.Context, reference model.TargetProviderGetReference) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Edge Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Edge Target): getting artifacts: %s - %s", reference.Deployment.Instance.Spec.Scope, reference.Deployment.Instance.ObjectMeta.Name)

	// ctx, cancelFunc := context.WithTimeout(ctx, 10*time.Second)
	requestCtx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()

	// Session ID is taken every time Get() function is called
	requestCtxNew, err := h.establishConnection(requestCtx)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to establish connection", "error", err)
		return nil, err
	}

	ret := []model.ComponentSpec{}

	for _, c := range reference.References {
		uuid := c.Component.Metadata["Uuid"]

		app, err := h.SystemClient.GetAppInstanceById(requestCtxNew, wrapperspb.String(uuid))

		if err != nil {
			sLog.ErrorCtx(ctx, "Failed to probe app by ID", "error", err)
			return nil, err
		}
		if app != nil {
			host := app.Status.RunningHost
			if host != "" && host == reference.TargetName {
				compSpec := appToComponentSpec(app)
				ret = append(ret, compSpec)
			}
		}
	}
	target, err := h.ApiClient.GetTarget(ctx, reference.TargetName, reference.TargetNamespace, h.Config.User, h.Config.Password)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to get target", "error", err)
		return nil, err
	}

	target.Spec.Components = ret

	targetData, _ := json.Marshal(target)
	err = h.ApiClient.CreateTarget(ctx, reference.TargetName, targetData, reference.TargetNamespace, h.Config.User, h.Config.Password)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to create target", "error", err)
		return nil, err
	}
	return ret, nil
}

func appToComponentSpec(app *system_model.AppInstance) model.ComponentSpec {
	metadata := make(map[string]string)
	metadata["Uuid"] = app.Metadata.Uuid
	metadata["OnwerId"] = app.Metadata.OwnerId
	for k, v := range app.Metadata.Labels {
		metadata["labels."+k] = v
	}

	props := map[string]any{}
	if c := app.GetSpec().GetContainer(); c != nil {
		// Convert proto -> map[string]any for clean JSON serialization later
		b, _ := (protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}).Marshal(c)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		props["container"] = m
	}

	compSpec := model.ComponentSpec{
		Name:       app.Metadata.Name,
		Type:       app.Kind,
		Metadata:   metadata,
		Properties: props,
	}
	return compSpec
}

func (h *EdgeProvider) connectToEdgeAdapter(ctx context.Context, sessionId string, credentials *tls.Config) error {
	var err error
	h.EdgeAdapterClient, err = NewEdgeAdapterClient(ctx, sessionId, credentials)
	return err
}

func (h *EdgeProvider) establishEdgeConnection(ctx context.Context) (context.Context, error) {
	sessionID, _, err := h.AuthService.GetSessionIdAsync(h.Config.BaseAddress)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to get session ID", "error", err)
		return ctx, err
	}

	md := metadata.Pairs(
		"cookie", fmt.Sprintf("sessionId=%s", sessionID),
		"content-type", "application/grpc",
	)
	ctxNew := metadata.NewOutgoingContext(ctx, md)

	if err := h.connectToAPI(ctxNew, sessionID, h.AuthService.Credentials, ""); err != nil {
		sLog.ErrorCtx(ctx, "Failed to connect to API", "error", err)
		return ctxNew, err
	}

	os.Setenv("EDGE_ADAPTER_SERVICE_ADDRESS", h.Config.BaseAddress)

	if err := h.connectToEdgeAdapter(ctxNew, sessionID, h.AuthService.Credentials); err != nil {
		sLog.ErrorCtx(ctx, "Failed to connect to EdgeAdapter service", "error", err)
		return ctxNew, err
	}

	return ctxNew, nil
}

func (h *EdgeProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{"app.id", "app.version"},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "app.id", IgnoreCase: false, SkipIfMissing: false},
				{Name: "app.version", IgnoreCase: false, SkipIfMissing: false},
			},
		},
	}
}

func (h *EdgeProvider) Apply(ctx context.Context, reference model.TargetProviderApplyReference) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Edge Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Edge Target): Apply()")

	// validationRule := h.GetValidationRule(ctx)
	// components := make([]model.ComponentSpec, len(reference.Step.Components))
	// for i, componentStep := range reference.Step.Components {
	// 	components[i] = componentStep.Component
	// }
	// if validationErr := validationRule.Validate(components); validationErr != nil {
	// 	sLog.ErrorCtx(ctx, "Component validation failed", "error", validationErr)
	// 	return nil, validationErr
	// }

	if reference.IsDryRun {
		sLog.InfoCtx(ctx, "Dry run mode - skipping actual deployment")
		return make(map[string]model.ComponentResultSpec), nil
	}

	requestCtx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
	defer cancelFunc()

	requestCtxNew, err := h.establishEdgeConnection(requestCtx)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Edge Target): failed to establish edge connection: %+v", err)
		return nil, err
	}

	ret := reference.Step.PrepareResultMap()

	for _, componentStep := range reference.Step.Components {
		if componentStep.Action == model.ComponentUpdate {
			result, err := h.deployEdgeComponent(requestCtxNew, componentStep.Component, reference.TargetName)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Edge Target): failed to deploy component %s: %+v", componentStep.Component.Name, err)
				ret[componentStep.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: fmt.Sprintf("Failed to deploy component: %v", err),
				}
			} else {
				target, err := h.ApiClient.GetTarget(ctx, reference.TargetName, reference.TargetNamespace, h.Config.User, h.Config.Password)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Edge Target): failed to create target: %+v", err)
					return nil, err
				}
				found := false
				for i, component := range target.Spec.Components {
					uuid := component.Metadata["Uuid"]
					if uuid == componentStep.Component.Metadata["Uuid"] {
						found = true
						target.Spec.Components[i] = componentStep.Component // Update existing component
						break
					}
				}
				if !found {
					target.Spec.Components = append(target.Spec.Components, componentStep.Component) // Add new component
				}
				targetData, _ := json.Marshal(target)
				err = h.ApiClient.CreateTarget(ctx, reference.TargetName, targetData, reference.TargetNamespace, h.Config.User, h.Config.Password)
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (Edge Target): failed to create target: %+v", err)
					return nil, err
				}
				ret[componentStep.Component.Name] = result
			}
		}
	}

	return ret, nil
}

func (h *EdgeProvider) deployEdgeComponent(ctx context.Context, component model.ComponentSpec, deviceUUID string) (model.ComponentResultSpec, error) {
	containerInfo, ok := component.Properties["container"]
	if !ok {
		return model.ComponentResultSpec{}, fmt.Errorf("container information is required")
	}
	containerMap, ok := containerInfo.(map[string]interface{})
	if !ok {
		return model.ComponentResultSpec{}, fmt.Errorf("invalid container information format")
	}

	containerImage, ok := containerMap["Image"].(string)
	if !ok || containerImage == "" {
		return model.ComponentResultSpec{}, fmt.Errorf("container image is required")
	}

	containerName, ok := containerMap["Name"].(string)
	if !ok || containerName == "" {
		containerName = component.Name
	}

	var containerNetworks []*southbound.ContainerNetwork
	if networks, ok := containerMap["Networks"].([]interface{}); ok {
		for _, network := range networks {
			if netMap, ok := network.(map[string]interface{}); ok {
				containerNetwork := &southbound.ContainerNetwork{
					NetworkId: "",
					Ipv4:      "",
					Ipv6:      "",
				}
				if networkId, ok := netMap["NetworkId"].(string); ok {
					containerNetwork.NetworkId = networkId
				}
				if ipv4, ok := netMap["Ipv4"].(string); ok {
					containerNetwork.Ipv4 = ipv4
				}
				if ipv6, ok := netMap["Ipv6"].(string); ok {
					containerNetwork.Ipv6 = ipv6
				}
				containerNetworks = append(containerNetworks, containerNetwork)
			}
		}
	}

	resourceLimits := make(map[string]string)
	if resources, ok := containerMap["Resources"].(map[string]interface{}); ok {
		if limits, ok := resources["Limits"].(map[string]interface{}); ok {
			for key, value := range limits {
				if strValue, ok := value.(string); ok {
					resourceLimits[strings.ToLower(key)] = strValue
				}
			}
		}
	}

	componentName := component.Name
	if componentName == "" {
		componentName = containerName
	}
	componentType := component.Type
	if componentType == "" {
		componentType = "container"
	}

	request := &southbound.EdgeAdapterGrpcRequest{
		Name:   componentName,
		Kind:   componentType,
		Labels: make(map[string]string),
		AppSpec: &southbound.EdgeAppSpec{
			Name:     containerName,
			Image:    containerImage,
			Networks: containerNetworks,
			Resources: &southbound.Resource{
				Limits: resourceLimits,
			},
		},
	}

	for key, value := range component.Metadata {
		if strings.HasPrefix(key, "labels.") {
			labelKey := strings.TrimPrefix(key, "labels.")
			request.Labels[labelKey] = fmt.Sprintf("%v", value)
		} else {
			request.Labels[key] = fmt.Sprintf("%v", value)
		}
	}

	deviceID := deviceUUID
	ownerID := "default-owner"
	if partnerOwner, exists := component.Metadata["labels.partnerOwnerId"]; exists {
		ownerID = fmt.Sprintf("%v", partnerOwner)
	}

	request.Node = &southbound.Node{
		DeviceId: deviceID,
		OwnerId:  ownerID,
	}

	deviceSpec, err := h.getDeviceSpec(ctx, deviceID)
	if err != nil {
		return model.ComponentResultSpec{}, fmt.Errorf("failed to get device specification: %w", err)
	}
	request.NodeSpec = deviceSpec

	if err := h.addReservedInterlinkIP(ctx, deviceID, request); err != nil {
		sLog.WarnfCtx(ctx, "Failed to add reserved interlink IP: %v", err)
	}

	requestData, _ := json.MarshalIndent(request, "", "  ")
	sLog.InfoCtx(ctx, "Deploying component to Edge device: %s\nRequest: %s", deviceID, string(requestData))

	response, err := h.EdgeAdapterClient.DeployAsync(ctx, request)
	if err != nil {
		return model.ComponentResultSpec{}, fmt.Errorf("failed to deploy via EdgeAdapter: %w", err)
	}

	if response.HttpCode != 200 {
		return model.ComponentResultSpec{
			Status:  v1alpha2.UpdateFailed,
			Message: fmt.Sprintf("EdgeAdapter deployment failed: HTTP %d, Error %d, %s", response.HttpCode, response.ErrorCode, response.Message),
		}, nil
	}

	return model.ComponentResultSpec{
		Status:  v1alpha2.Updated,
		Message: "Component deployed successfully via EdgeAdapter",
	}, nil
}

func (h *EdgeProvider) getDeviceSpec(ctx context.Context, deviceUUID string) (*southbound.NodeSpec, error) {
	deviceIDValue := wrapperspb.String(deviceUUID)

	device, err := h.SystemClient.GetDeviceById(ctx, deviceIDValue)
	if err != nil {
		return nil, fmt.Errorf("failed to get device by ID %s: %w", deviceUUID, err)
	}

	if device == nil || device.Spec == nil {
		return nil, fmt.Errorf("device spec is nil for device %s", deviceUUID)
	}

	nodeSpec := &southbound.NodeSpec{
		Addresses:         device.Spec.Addresses,
		Networks:          []*southbound.HostNetwork{},
		ContainerNetworks: []*southbound.DockerNetwork{},
	}

	for _, network := range device.Spec.Networks {
		hostNetwork := &southbound.HostNetwork{
			NetName:        network.NetName,
			NicName:        network.NicName,
			RedundancyMode: network.RedundancyMode,
			NicList:        network.NicList,
			Ipv4:           network.Ipv4,
			Gateway:        network.Gateway,
		}
		nodeSpec.Networks = append(nodeSpec.Networks, hostNetwork)
	}

	for _, containerNetwork := range device.Spec.ContainerNetworks {
		dockerNetwork := &southbound.DockerNetwork{
			NetworkId: containerNetwork.NetworkId,
			Subnet:    containerNetwork.Subnet,
			Gateway:   containerNetwork.Gateway,
			NicName:   containerNetwork.NicName,
			Type:      containerNetwork.Type,
		}
		nodeSpec.ContainerNetworks = append(nodeSpec.ContainerNetworks, dockerNetwork)
	}

	return nodeSpec, nil
}

func (h *EdgeProvider) addReservedInterlinkIP(ctx context.Context, deviceUUID string, request *southbound.EdgeAdapterGrpcRequest) error {
	deviceIDValue := wrapperspb.String(deviceUUID)
	device, err := h.SystemClient.GetDeviceById(ctx, deviceIDValue)
	if err != nil {
		return fmt.Errorf("failed to get device for reserved IP: %w", err)
	}

	if device == nil || device.Spec == nil {
		return fmt.Errorf("device spec is nil for reserved IP extraction")
	}

	if device.Spec.ReservedAppInterlinkIp != "" {
		reservedNetwork := &southbound.ContainerNetwork{
			Ipv4:      device.Spec.ReservedAppInterlinkIp,
			NetworkId: "softdpacInterlinkNet",
			Ipv6:      "",
		}
		request.AppSpec.Networks = append(request.AppSpec.Networks, reservedNetwork)
		sLog.InfoCtx(ctx, "Added reserved interlink IP to container networks", "ip", device.Spec.ReservedAppInterlinkIp, "networkId", "softdpacInterlinkNet")
	}

	return nil
}
