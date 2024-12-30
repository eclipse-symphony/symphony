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

	Summary         = "Summary"
	DeploymentState = "DeployState"
)

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
type PlanResult struct {
	PlanState *PlanState `json:"planstate"`
	EndTime   time.Time  `json:"endTime"`
	Error     string     `json:"error,omitempty"`
}
type StepEnvelope struct {
	Step       model.DeploymentStep `json:"step"`
	Deployment model.DeploymentSpec `json:"deployment"`
	Remove     bool                 `json:"remove"`
	Namespace  string               `json:"Namespace"`
	PlanId     string               `json:"planId"`
	StepId     int                  `json:"stepId"`
	PlanState  *PlanState           `json:"planState"`
	// Provider   providers.IProvider  `json:"provider"`
	Phase           JobPhase
	DeploymentState model.DeploymentState
}

type PlanManager struct {
	Plans   sync.Map // map[string] *Planstate
	Timeout time.Duration
}

type StepResult struct {
	Step             model.DeploymentStep                 `json:"step"`
	Success          bool                                 `json:"success"`
	TargetResultSpec model.TargetResultSpec               `json:"targetResult"`
	Components       map[string]model.ComponentResultSpec `json:"components"`
	PlanId           string                               `json:"planId"`
	StepId           int                                  `json:"stepId"`
	Remove           bool                                 `json:"remove"`
	Timestamp        time.Time                            `json:"timestamp"`
	ApplyResult      interface{}                          `json:"applyresult"`
	GetResult        interface{}                          `json:"getresult"`
	Phase            JobPhase
	retComoponents   []model.ComponentSpec
	Error            error `json:"error,omitempty"`
	Target           string
	Namespace        string `json:"namespace"`
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

type OperationBody struct {
	StepId    int
	PlanId    string
	Target    string
	Action    JobPhase
	NameSpace string
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
	Delete               bool
	StepStates           []StepState
}
type JobState string

const (
	JobStateQueued    JobState = "queued"
	JobStateRunning   JobState = "running"
	JobStateCompleted JobState = "completed"
	JobStateFailed    JobState = "Failed"
)

type Job struct {
	ID                   string
	Phase                JobPhase
	PlanID               string
	StepIndex            int
	Target               string
	Role                 string
	Deployment           model.DeploymentSpec
	Components           []model.ComponentStep
	State                JobState
	Result               model.DeploymentState
	Error                string
	PreviousDesiredState *solution.SolutionManagerDeploymentState `json:"previous`
	CreateTime           time.Time
	UpdateTime           time.Time
	PlanState            PlanState
}
type StepState struct {
	Index       int
	Target      string
	Role        string
	Components  []model.ComponentStep
	State       string
	GetResult   []model.ComponentSpec
	ApplyResult model.DeploymentState
	Error       string
}

var deploymentTypeMap = map[bool]string{
	true:  DeploymentType_Delete,
	false: DeploymentType_Update,
}
