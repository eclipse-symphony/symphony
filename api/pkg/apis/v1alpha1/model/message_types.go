/*
* Copyright (c) Microsoft Corporation.
* Licensed under the MIT license.
* SPDX-License-Identifier: MIT
 */

package model

import (
	"time"
)

type JobPhase string

const (
	PhaseGet               JobPhase = "get"
	PhaseApply             JobPhase = "apply"
	DeploymentPlanTopic             = "deployment-plan"
	DeploymentStepTopic             = "deployment-step"
	CollectStepResultTopic          = "step-result"
)

type PlanResult struct {
	PlanState PlanState `json:"planstate"`
	EndTime   time.Time `json:"endTime"`
	Error     string    `json:"error,omitempty"`
}

type PlanEnvelope struct {
	Plan                 DeploymentPlan                 `json:"plan"`
	Deployment           DeploymentSpec                 `json:"deployment"`
	MergedState          DeploymentState                `json:"mergedState"`
	PreviousDesiredState SolutionManagerDeploymentState `json:"previousDesiredState"`
	CurrentState         DeploymentState                `json:"currentState"`
	Remove               bool                           `json:"remove"`
	Namespace            string                         `json:"namespace"`
	PlanId               string                         `json:"planId"`
	PlanName             string                         `json:"planName"`
	Generation           string                         `json:"generation"` // deployment version
	Hash                 string                         `json:"hash"`
	Phase                JobPhase                       `json:"phase"`
}

// for step
type StepResult struct {
	Step             DeploymentStep                 `json:"step"`
	TargetResultSpec TargetResultSpec               `json:"targetResult"`
	PlanId           string                         `json:"planId"`
	PlanName         string                         `json:"planName"`
	StepId           int                            `json:"stepId"`
	Timestamp        time.Time                      `json:"timestamp"`
	GetResult        []ComponentSpec                `json:"getResult"`  // for get result
	ApplyResult      map[string]ComponentResultSpec `json:"components"` // for apply result
	Error            string                         `json:"string,omitempty"`
	Target           string                         `json:"Target"`
	NameSpace        string                         `json:"Namespace"`
}
type StepEnvelope struct {
	Step      DeploymentStep `json:"step"`
	Remove    bool           `json:"remove"`
	StepId    int            `json:"stepId"`
	PlanState PlanState      `json:"planState"`
}

type OperationBody struct {
	StepId    int      `json:"stepId"`
	PlanId    string   `json:"planId"`
	PlanName  string   `json:"planName"`
	Target    string   `json:"Target"`
	Action    JobPhase `json:"action"`
	NameSpace string   `json:"Namespace"`
	Remove    bool     `json:"remove"`
	MessageId string   `json:"messageId"`
}

type StepState struct {
	Index      int             `json:"Index"`
	Target     string          `json:"Target"`
	Role       string          `json:"Role"`
	Components []ComponentStep `json:"Components"`
	State      string          `json:"State"`
	GetResult  []ComponentSpec `json:"GetResult"`
	Error      string          `json:"Error"`
}

type AgentRequest struct {
	OperationID string `json:"operationID"`
	Provider    string `json:"provider"`
	Action      string `json:"action"`
}

type ProviderGetRequest struct {
	AgentRequest
	Deployment DeploymentSpec  `json:"deployment"`
	References []ComponentStep `json:"references"`
}

type ProviderPagingRequest struct {
	RequestList   []map[string]interface{} `json:"requestList"`
	LastMessageID string                   `json:"lastMessageID"`
	CorrelationID string                   `json:"X-Activity-correlationId,omitempty"`
}
type ProviderApplyRequest struct {
	AgentRequest
	Deployment DeploymentSpec `json:"deployment"`
	Step       DeploymentStep `json:"step"`
	IsDryRun   bool           `json:"isDryRun,omitempty"`
}

type PlanState struct {
	PlanId               string `json:"planId"`
	PlanName             string `json:"planName"`
	Phase                JobPhase
	CompletedSteps       int                            `json:"completedSteps"`
	Status               string                         `json:"status"`
	MergedState          DeploymentState                `json:"mergedState"`
	Deployment           DeploymentSpec                 `json:"deployment"`
	CurrentState         DeploymentState                `json:"currentState"`
	PreviousDesiredState SolutionManagerDeploymentState `json:"previousDesiredState"`
	TargetResult         map[string]int                 `json:"targetResult"`
	Namespace            string                         `json:"namespace"`
	TotalSteps           int                            `json:"totalSteps"`
	StepStates           []StepState                    `json:"stepStates"`
	Steps                []DeploymentStep               `json:"steps"`
	DefaultScope         string                         `json:"defaultScope,omitempty"`
}

type ProviderGetValidationRuleRequest struct {
	AgentRequest
}

type AsyncResult struct {
	OperationID string `json:"operationID"`
	Namespace   string `json:"namespace"`
	Error       error  `json:"error,omitempty"`
	Body        []byte `json:"body"`
}

type SymphonyEndpoint struct {
	BaseUrl          string `json:"baseUrl"`
	RequestEndpoint  string `json:"requestEndpoint,omitempty"`
	ResponseEndpoint string `json:"responseEndpoint,omitempty"`
}
