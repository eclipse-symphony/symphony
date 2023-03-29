package serving

import "fmt"

const ProviderNamespace = "Microsoft.GitOps"

const HealthzEndpoint = "/healthz"
const ReadyzEndpoint = "/readyz"

var RepoURLEndpoint = fmt.Sprintf("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroup}/providers/%s/repo/{repoName}", ProviderNamespace)
var CloudGitOpsEndpoint = fmt.Sprintf("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroup}/providers/%s/repo/{repoName}/cloudResource/{deploymentUtilization}", ProviderNamespace)
var EdgeGitOpsEndpoint = fmt.Sprintf("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroup}/providers/%s/repo/{repoName}/edgeResource/{edgeUtilization}", ProviderNamespace)
var CommitEndpoint = fmt.Sprintf("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroup}/providers/%s/repo/{repoName}/commit", ProviderNamespace)
