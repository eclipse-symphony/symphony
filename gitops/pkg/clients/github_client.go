package clients

import (
	"context"
	"errors"
	"fmt"

	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/logger"
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/models"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type (
	githubRepoClient struct {
		rs  *models.GitOpsRepoResource
		ghc *github.Client
		log logger.Logger
		ctx context.Context
	}
)

var GithubFileNotFoundError error = errors.New("Not Found")

func NewGithubRepoClient(ctx context.Context, rs *models.GitOpsRepoResource) RepoClient {

	return &githubRepoClient{
		rs:  rs,
		ctx: ctx,
		ghc: newGithubClient(ctx, rs.Properties.ConfigurationProtectedSettings.Token),
		log: logger.NewLogger(ctx, "github_client"),
	}
}

func newGithubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func (g *githubRepoClient) CommitFiles(ctx context.Context, files []models.File, commitMessage string) (err error) {
	repoName := g.rs.Properties.Name
	owner := g.rs.Properties.Owner
	branch := g.rs.Properties.Branch

	currentTree, _, err := g.ghc.Git.GetTree(ctx, owner, repoName, branch, false)

	if err != nil {
		g.log.Error(err)
		return
	}

	ghFiles := make([]github.TreeEntry, len(files))
	for i, file := range files {
		ghFiles[i] = *file.GetTreeEntry()
	}

	tree, _, err := g.ghc.Git.CreateTree(ctx, owner, repoName, *currentTree.SHA, ghFiles)
	if err != nil {
		g.log.Error(err)
		return
	}

	parent, _, err := g.ghc.Repositories.GetCommit(ctx, owner, repoName, branch)
	if err != nil {
		g.log.Error(err)
		return
	}
	// Hack to get around the fact that the github API doesn't return the SHA of the parent commit
	if parent.Commit.SHA == nil {
		parent.Commit.SHA = parent.SHA
	}

	newCommit, _, err := g.ghc.Git.CreateCommit(ctx, owner, repoName, &github.Commit{
		Message: &commitMessage,
		Tree:    tree,
		Parents: []github.Commit{*parent.Commit},
	})

	if err != nil {
		g.log.Error(err)
		return
	}
	_, _, err = g.ghc.Git.UpdateRef(ctx, owner, repoName, &github.Reference{
		Ref:    github.String(fmt.Sprintf("refs/heads/%s", branch)),
		Object: &github.GitObject{SHA: newCommit.SHA},
	}, true)
	if err != nil {
		g.log.Error(err)
		return
	}
	return nil
}

func (g *githubRepoClient) GetContent(ctx context.Context, path string) (*github.RepositoryContent, error) {
	repoContent, _, _, err := g.ghc.Repositories.GetContents(
		g.ctx,
		g.rs.GetOwner(),
		g.rs.GetRepo(),
		path,
		&github.RepositoryContentGetOptions{
			Ref: g.rs.GetBranch(),
		},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == 404 {
				return nil, GithubFileNotFoundError
			}
		}
		return nil, err
	}

	if repoContent == nil {
		return nil, fmt.Errorf("no content found at path %s", path)
	}

	if repoContent.Type == nil || *repoContent.Type != "file" {
		return nil, fmt.Errorf("content at path %s is not a file", path)
	}

	return repoContent, nil
}

func (g *githubRepoClient) GetResource() models.GitOpsRepoResource {
	return *g.rs
}

func (g *githubRepoClient) SetResource(resource *models.GitOpsRepoResource) (err error) {
	if resource == nil {
		return fmt.Errorf("resource is nil")
	}
	g.rs = resource
	g.ghc = newGithubClient(g.ctx, g.rs.Properties.ConfigurationProtectedSettings.Token)
	return nil
}
