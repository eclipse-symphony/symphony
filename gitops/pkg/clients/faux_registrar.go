package clients

import (
	"context"
	"fmt"
	"strings"

	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/models"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/serving"
)

type (
	fauxRegistrar struct {
		repoStore map[string]*RepoStore
	}

	RepoStore struct {
		repoClient  RepoClient
		deployments map[string]*models.CloudGitOpsResource
		edge        map[string]*models.EdgeGitOpsResource
	}
)

var (
	DefaultGitOpsRegistrar GitOpsRegistrar
)

func NewFauxRegistrar() GitOpsRegistrar {
	return &fauxRegistrar{
		repoStore: make(map[string]*RepoStore),
	}
}

func buildRepoId(subscriptionId string, resourceGroupName string, repoName string) string {
	id := serving.RepoURLEndpoint
	id = strings.Replace(id, "{subscriptionId}", subscriptionId, 1)
	id = strings.Replace(id, "{resourceGroup}", resourceGroupName, 1)
	id = strings.Replace(id, "{repoName}", repoName, 1)
	return id
}

func (c *fauxRegistrar) GetRepoClient(subscriptionId string, resourceGroupName string, repoName string) (client RepoClient, err error) {
	if store, ok := c.repoStore[buildRepoId(subscriptionId, resourceGroupName, repoName)]; ok {
		return store.repoClient, nil
	}
	return nil, fmt.Errorf("Repository not found: %s", buildRepoId(subscriptionId, resourceGroupName, repoName))
}

func (c *fauxRegistrar) GetDeploymentUtilization(subscriptionId string, resourceGroupName string, repoName string, utilizationName string) (result *models.CloudGitOpsResource, err error) {
	if store, ok := c.repoStore[buildRepoId(subscriptionId, resourceGroupName, repoName)]; ok {
		if deployment, ok := store.deployments[utilizationName]; ok {
			return deployment, nil
		}
	}
	return nil, fmt.Errorf("Deployment not found: %s", utilizationName)
}

func (c *fauxRegistrar) GetEdgeUtilization(subscriptionId string, resourceGroupName string, repoName string, utilizationName string) (result *models.EdgeGitOpsResource, err error) {
	if store, ok := c.repoStore[buildRepoId(subscriptionId, resourceGroupName, repoName)]; ok {
		if edgeUtil, ok := store.edge[utilizationName]; ok {
			return edgeUtil, nil
		}
	}
	return nil, fmt.Errorf("Edge Utilization not found: %s", utilizationName)
}

func (c *fauxRegistrar) RegisterRepo(repo *models.GitOpsRepoResource) (err error) {
	if _, ok := c.repoStore[repo.Id]; ok {
		c.repoStore[repo.Id].repoClient.SetResource(repo)
	} else {
		c.repoStore[repo.Id] = &RepoStore{
			repoClient:  NewGithubRepoClient(context.Background(), repo),
			deployments: make(map[string]*models.CloudGitOpsResource),
			edge:        make(map[string]*models.EdgeGitOpsResource),
		}
	}

	return nil
}

func (c *fauxRegistrar) RegisterDeploymentUtilization(resource *models.CloudGitOpsResource) (err error) {
	if store, ok := c.repoStore[buildRepoId(resource.GetSubscription(), resource.GetResourceGroup(), resource.GetAzRepoShortName())]; ok {
		store.deployments[resource.Name] = resource
		return nil
	}
	return fmt.Errorf("Repository not found: %s", buildRepoId(resource.GetSubscription(), resource.GetResourceGroup(), resource.GetAzRepoShortName()))
}

func (c *fauxRegistrar) RegisterEdgeUtilization(resource *models.EdgeGitOpsResource) (err error) {
	if store, ok := c.repoStore[buildRepoId(resource.GetSubscription(), resource.GetResourceGroup(), resource.GetAzRepoShortName())]; ok {
		store.edge[resource.Name] = resource
		return nil
	}
	return fmt.Errorf("Repository not found: %s", buildRepoId(resource.GetSubscription(), resource.GetResourceGroup(), resource.GetAzRepoShortName()))
}

func init() {

	DefaultGitOpsRegistrar = NewFauxRegistrar()

}
