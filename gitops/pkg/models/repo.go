package models

type GitOpsRepoResource struct {
	azResource
	Properties GitOpsRepoProperties `json:"properties,omitempty"`
}

type GitOpsRepoProperties struct {
	Branch                         string                         `json:"branch,omitempty"`
	Name                           string                         `json:"name,omitempty"`
	Owner                          string                         `json:"owner,omitempty"`
	Provider                       string                         `json:"provider,omitempty"`
	ConfigurationProtectedSettings ConfigurationProtectedSettings `json:"configurationProtectedSettings,omitempty"`
}

type ConfigurationProtectedSettings struct {
	Token string `json:"token"`
}

func (rs *GitOpsRepoResource) GetBranch() (branch string) {
	return rs.Properties.Branch
}

func (rs *GitOpsRepoResource) GetOwner() (owner string) {
	return rs.Properties.Owner
}

func (rs *GitOpsRepoResource) GetRepo() (repo string) {
	return rs.Properties.Name
}
