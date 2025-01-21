/*
* Copyright (c) Microsoft Corporation.
* Licensed under the MIT license.
* SPDX-License-Identifier: MIT
 */

package solution

import (
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

type JobPhase string

const (
	PhaseGet               JobPhase = "get"
	PhaseApply             JobPhase = "apply"
	DeploymentPlanTopic             = "deployment-plan"
	DeploymentStepTopic             = "deployment-step"
	CollectStepResultTopic          = "step-result"
	MaxRetries             int      = 3               // Maximum retry attempts
	RetryDelay                      = time.Second * 2 // Delay between retries
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
	Plan                 model.DeploymentPlan            `json:"plan"`
	Deployment           model.DeploymentSpec            `json:"deployment"`
	MergedState          model.DeploymentState           `json:"mergedState"`
	PreviousDesiredState *SolutionManagerDeploymentState `json:"previousDesiredState"`
	CurrentState         model.DeploymentState           `json:"currentState"`
	Remove               bool                            `json:"remove"`
	Namespace            string                          `json:"namespace"`
	PlanId               string                          `json:"planId"`
	Generation           string                          `json:"generation"` // deployment version
	Hash                 string                          `json:"hash"`
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
	Steps                []model.DeploymentStep
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
