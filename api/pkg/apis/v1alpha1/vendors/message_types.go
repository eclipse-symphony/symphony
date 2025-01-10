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
	ID                   string `json:"opeateionId"`
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
	PreviousDesiredState *solution.SolutionManagerDeploymentState `json:"previous`
	Status               string                                   `json:"status"`
	TargetResult         map[string]int
	Namespace            string `json:"namespace"`
	Remove               bool
	StepStates           []StepState
}

// for step
type StepResult struct {
	Step             model.DeploymentStep                 `json:"step"`
	TargetResultSpec model.TargetResultSpec               `json:"targetResult"`
	PlanId           string                               `json:"planId"`
	StepId           int                                  `json:"stepId"`
	Timestamp        time.Time                            `json:"timestamp"`
	GetResult        []model.ComponentSpec                // for get result
	ApplyResult      map[string]model.ComponentResultSpec `json:"components"` // for apply result
	Error            string                               `json:"string,omitempty"`
	Target           string
}
type StepEnvelope struct {
	Step      model.DeploymentStep `json:"step"`
	Remove    bool                 `json:"remove"`
	StepId    int                  `json:"stepId"`
	PlanState *PlanState           `json:"planState"`
}

type OperationBody struct {
	StepId    int
	PlanId    string
	Target    string
	Action    JobPhase
	NameSpace string
	Remove    bool
	MessageId string
}

type StepState struct {
	Index      int
	Target     string
	Role       string
	Components []model.ComponentStep
	State      string
	GetResult  []model.ComponentSpec
	Error      string
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
