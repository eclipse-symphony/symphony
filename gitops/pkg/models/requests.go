package models

type GitCommitRequest struct {
	CommitMessage string `json:"commitMessage"`
	Files         []File `json:"files"`
}

type RepoRequest = GitOpsRepoResource
type DeploymentUtilationRequest = CloudGitOpsResource
type EdgeUtilizationRequest = EdgeGitOpsResource
