package models

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type CloudGitOpsResource struct {
	azResource
	Properties CloudGitOpsResourceProperties `json:"properties,omitempty"`
}
type CloudGitOpsResourceProperties struct {
	BaseGitOpsProperties
	TemplatePath       string                      `json:"templatePath,omitempty"`
	ParametersPath     string                      `json:"parametersPath,omitempty"`
	Mode               armresources.DeploymentMode `json:"mode,omitempty"`
	TargetSubscription string                      `json:"targetSubscription,omitempty"`
}

func (g *CloudGitOpsResource) GetAzRepoShortName() string {
	// the id is in the form of
	// /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.GitOps/repo/{repoName}/cloudResource/{utilizationName}
	// so we need to split the id and get the {repoName} element
	// TODO: Find a better way to do this
	split := strings.Split(g.Id, "/")
	return split[len(split)-3]
}
