package models

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
)

type EdgeGitOpsResource struct {
	azResource
	Properties EdgeGitOpsProperties `json:"properties,omitempty"`
}

type EdgeGitOpsProperties struct {
	BaseGitOpsProperties
	ExtendedLocationId string                      `json:"extendedLocationId,omitempty"`
	DeploymentScheme   DeploymentScheme            `json:"deploymentScheme,omitempty"`
	Mode               armresources.DeploymentMode `json:"mode,omitempty"`
}

type DeploymentScheme struct {
	Stages []GitOpsStage `json:"stages,omitempty"`
	Scope  string        `json:"scope,omitempty"`
}

type GitOpsStage struct {
	Name       string                         `json:"name,omitempty"`
	Template   GitOpsEdgeTemplateProperties   `json:"template,omitempty"`
	Parameters GitOpsEdgeParametersProperties `json:"parameters,omitempty"`
	TargetRef  model.TargetRefSpec            `json:"targetRef,omitempty"`
}

type GitOpsEdgeTemplateProperties struct {
	Path string `json:"path,omitempty"`
	Name string `json:"name,omitempty"`
}

type GitOpsEdgeTemplate struct {
	Template map[string]interface{} `json:"template,omitempty"`
	Name     string                 `json:"name,omitempty"`
}

type GitOpsEdgeParametersProperties struct {
	Path string `json:"path,omitempty"`
	Name string `json:"name,omitempty"`
}

type GitOpsEdgeParameters struct {
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Name       string                 `json:"name,omitempty"`
}

func (g *EdgeGitOpsResource) GetAzRepoShortName() string {
	// the id is in the form of
	// /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.GitOps/repo/{repoName}/deploymentUtilization/{utilizationName}
	// so we need to split the id and get the {repoName} element
	// TODO: Find a better way to do this
	split := strings.Split(g.Id, "/")
	return split[len(split)-3]
}
