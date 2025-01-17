/*
* Copyright (c) Microsoft Corporation.
* Licensed under the MIT license.
* SPDX-License-Identifier: MIT
 */

package vendors

import (
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

type JobPhase string

const (
	SYMPHONY_AGENT string   = "/symphony-agent:"
	ENV_NAME       string   = "SYMPHONY_AGENT_ADDRESS"
	PhaseGet       JobPhase = "get"
	PhaseApply     JobPhase = "apply"
	// DeploymentType_Update indicates the type of deployment is Update. This is
	// to give a deployment status on Symphony Target deployment.
	DeploymentType_Update string = "Target Update"
	// DeploymentType_Delete indicates the type of deployment is Delete. This is
	// to give a deployment status on Symphony Target deployment.
	DeploymentType_Delete string = "Target Delete"

	Summary                    = "Summary"
	DeploymentState            = "DeployState"
	DeploymentPlanTopic        = "deployment-plan"
	DeploymentStepTopic        = "deployment-step"
	CollectStepResultTopic     = "step-result"
	MaxRetries             int = 3               // Maximum retry attempts
	RetryDelay                 = time.Second * 2 // Delay between retries
)

// for plan storage
type PlanManager struct {
	Plans sync.Map // map[string] *Planstate
}

type PlanResult struct {
	PlanState *PlanState `json:"planstate"`
	EndTime   time.Time  `json:"endTime"`
	Error     string     `json:"error,omitempty"`
}

type PlanEnvelope struct {
	Plan                 model.DeploymentPlan                     `json:"plan"`
	Deployment           model.DeploymentSpec                     `json:"deployment"`
	MergedState          model.DeploymentState                    `json:"mergedState"`
	PreviousDesiredState *solution.SolutionManagerDeploymentState `json:"previousDesiredState"`
	CurrentState         model.DeploymentState                    `json:"currentState"`
	Remove               bool                                     `json:"remove"`
	Namespace            string                                   `json:"namespace"`
	PlanId               string                                   `json:"planId"`
	Generation           string                                   `json:"generation"` // deployment version
	Hash                 string                                   `json:"hash"`
	Phase                JobPhase
}

type PlanState struct {
	ID                   string `json:"operationId"`
	PlanId               string `json:"planId"`
	Phase                JobPhase
	StartTime            time.Time                                `json:"startTime"`
	ExpireTime           time.Time                                `json:"expireTime"`
	TotalSteps           int                                      `json:"totalSteps"`
	CompletedSteps       int                                      `json:"completedSteps"`
	Summary              model.SummarySpec                        `json:"summary"`
	MergedState          model.DeploymentState                    `json:"mergedState"`
	Deployment           model.DeploymentSpec                     `json:"deployment"`
	CurrentState         model.DeploymentState                    `json:"currentState"`
	PreviousDesiredState *solution.SolutionManagerDeploymentState `json:"previousDesiredState"`
	Status               string                                   `json:"status"`
	TargetResult         map[string]int                           `json:"targetResult"`
	Namespace            string                                   `json:"namespace"`
	Remove               bool                                     `json:"remove"`
	StepStates           []StepState                              `json:"stepStates"`
	Steps                []model.DeploymentStep                   `json:"steps"`
}

type StepResult struct {
	Step             model.DeploymentStep                 `json:"step"`
	TargetResultSpec model.TargetResultSpec               `json:"targetResult"`
	PlanId           string                               `json:"planId"`
	StepId           int                                  `json:"stepId"`
	Timestamp        time.Time                            `json:"timestamp"`
	GetResult        []model.ComponentSpec                `json:"getResult"`
	ApplyResult      map[string]model.ComponentResultSpec `json:"components"`
	Error            string                               `json:"error,omitempty"`
	Target           string                               `json:"target"`
}

type StepEnvelope struct {
	Step      model.DeploymentStep `json:"step"`
	Remove    bool                 `json:"remove"`
	StepId    int                  `json:"stepId"`
	PlanState *PlanState           `json:"planState"`
}

type OperationBody struct {
	StepId    int      `json:"stepId"`
	PlanId    string   `json:"planId"`
	Target    string   `json:"target"`
	Action    JobPhase `json:"action"`
	NameSpace string   `json:"namespace"`
	Remove    bool     `json:"remove"`
	MessageId string   `json:"messageId"`
}

type StepState struct {
	Index      int                   `json:"index"`
	Target     string                `json:"target"`
	Role       string                `json:"role"`
	Components []model.ComponentStep `json:"components"`
	State      string                `json:"state"`
	GetResult  []model.ComponentSpec `json:"getResult"`
	Error      string                `json:"error"`
}

var deploymentTypeMap = map[bool]string{
	true:  DeploymentType_Delete,
	false: DeploymentType_Update,
}

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

type ProviderPagingRequest struct {
	RequestList   []AgentRequest `json:"requestList"`
	LastMessageID string         `json:"lastMessageID"`
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
