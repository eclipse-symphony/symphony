package edge

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/contexts"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/edge/api/system_model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/edge/authprovider"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const loggerName = "providers.target.edge"

var sLog = logger.NewLogger(loggerName)

var (
	BaseAddress = "https://10.228.234.29:6201"
)

type EdgeProviderConfig struct {
	Name string `json:"name"`
}

type EdgeProvider struct {
	Context *contexts.ManagerContext
	Config  EdgeProviderConfig

	AuthService  *authprovider.AuthenticationService
	SystemClient system_model.SystemModelClient
}

func EdgeProviderConfigFromMap(properties map[string]string) (EdgeProviderConfig, error) {
	config := EdgeProviderConfig{}

	if name, ok := properties["name"]; ok {
		config.Name = name
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

func (h *EdgeProvider) establishConnection(ctx context.Context) error {
	sessionID, _, err := h.AuthService.GetSessionIdAsync(BaseAddress)
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to get session ID", "error", err)
		return err
	}

	md := metadata.Pairs(
		"cookie", fmt.Sprintf("sessionId=%s", sessionID),
		"content-type", "application/grpc",
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	if err := h.connectToAPI(ctx, sessionID, h.AuthService.Credentials, ""); err != nil {
		sLog.ErrorCtx(ctx, "Failed to connect to API", "error", err)
		return err
	}

	return nil
}

func (h *EdgeProvider) Get(ctx context.Context, reference model.TargetProviderGetReference) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Edge Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Edge Target): getting artifacts: %s - %s", reference.Deployment.Instance.Spec.Scope, reference.Deployment.Instance.ObjectMeta.Name)

	ctx, cancelFunc := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFunc()

	// Session ID is taken every time Get() function is called
	if err := h.establishConnection(ctx); err != nil {
		sLog.ErrorCtx(ctx, "Failed to establish connection", "error", err)
		return nil, err
	}

	app, err := h.SystemClient.GetAppInstanceById(ctx, wrapperspb.String(reference.Deployment.Instance.ObjectMeta.Name))
	if err != nil {
		sLog.ErrorCtx(ctx, "Failed to get app by ID", "error", err)
		return nil, err
	}

	if app == nil {
		sLog.ErrorCtx(ctx, "App not found", "deviceName", reference.Deployment.Instance.ObjectMeta.Name)
		return nil, fmt.Errorf("app %s not found", reference.Deployment.Instance.ObjectMeta.Name)
	}

	compSpec := appToComponentSpec(app)

	return []model.ComponentSpec{compSpec}, nil
}

func appToComponentSpec(app *system_model.AppInstance) model.ComponentSpec {
	// TODO: Implement conversion logic from system_model.Device to model.ComponentSpec
	metadata := make(map[string]string)
	metadata["Uuid"] = app.Metadata.Uuid
	metadata["OnwerId"] = app.Metadata.OwnerId
	for k, v := range app.Metadata.Labels {
		metadata["labels."+k] = v
	}

	metadata["namespace"] = "default"
	metadata["etag"] = "1"

	compSpec := model.ComponentSpec{
		Name:     app.Metadata.Name,
		Type:     app.Kind,
		Metadata: metadata,
		Properties: map[string]interface{}{
			"container": app.Spec.Data.(*system_model.AppInstanceSpec_Container).Container,
		},
	}
	return compSpec
}
