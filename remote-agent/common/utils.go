package utils

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

type AgentRequest struct {
	OperationID string `json:"operationID"`
	Provider    string `json:"provider"`
	Action      string `json:"action"`
}

type ProviderGetRequest struct {
	AgentRequest
	Deployment model.DeploymentSpec  `json:"deployment"`
	References []model.ComponentStep `json:"references"`
}

type ProviderApplyRequest struct {
	AgentRequest
	Deployment model.DeploymentSpec `json:"deployment"`
	Step       model.DeploymentStep `json:"step"`
	IsDryRun   bool                 `json:"isDryRun,omitempty"`
}

type ProviderGetValidationRuleRequest struct {
	AgentRequest
}

type AsyncResult struct {
	OperationID string `json:"operationID"`
	Error       error  `json:"error,omitempty"`
	Body        []byte `json:"body"`
}

type SymphonyEndpoint struct {
	RequestEndpoint  string `json:"requestEndpoint,omitempty"`
	ResponseEndpoint string `json:"responseEndpoint,omitempty"`
}
