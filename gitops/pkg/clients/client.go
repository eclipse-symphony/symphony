package clients

import (
	"context"

	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/models"
	"github.com/google/go-github/github"
)

type (
	GitOpsRegistrar interface {
		RegisterRepo(resource *models.GitOpsRepoResource) (err error)
		RegisterDeploymentUtilization(resource *models.CloudGitOpsResource) (err error)
		RegisterEdgeUtilization(resource *models.EdgeGitOpsResource) (err error)
		GetRepoClient(subscriptionId string, resourceGroupName string, repoName string) (client RepoClient, err error)
		GetDeploymentUtilization(subscriptionId string, resourceGroupName string, repoName string, utilizationName string) (result *models.CloudGitOpsResource, err error)
		GetEdgeUtilization(subscriptionId string, resourceGroupName string, repoName string, utilizationName string) (result *models.EdgeGitOpsResource, err error)
	}

	RepoClient interface {
		GetContent(ctx context.Context, path string) (*github.RepositoryContent, error)
		CommitFiles(ctx context.Context, files []models.File, commitMessage string) (err error)
		SetResource(resource *models.GitOpsRepoResource) error
		GetResource() models.GitOpsRepoResource
	}
)
