package serving

// At the moment these routes are assumed to live under the Symphony RP namespace
// const ProviderNamespace = "Microsoft.GitOps"
const ProviderNamespace = "private.symphony"
const ApiVersion = "2020-01-01-preview"

const HealthzEndpoint = "/healthz"
const ReadyzEndpoint = "/readyz"

const RepoURLEndpoint = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroup}/providers/" + ProviderNamespace + "/gitRepos/{repoName}"
const ArmGitOpsEndpoint = RepoURLEndpoint + "/armDeploymentGitOps/{deploymentUtilization}"
const EdgeGitOpsEndpoint = RepoURLEndpoint + "/edgeDeploymentGitOps/{edgeUtilization}"

// Create endpoints
const resourceCreationValidate = "/ResourceCreationValidate"
const RepoCreationValidateEndpoint = RepoURLEndpoint + resourceCreationValidate
const ArmDeploymentCreationValidateEndpoint = ArmGitOpsEndpoint + resourceCreationValidate
const EdgeDeploymentCreationValidateEndpoint = EdgeGitOpsEndpoint + resourceCreationValidate

const RepoCreationBegin = RepoURLEndpoint
const ArmDeploymentCreationBegin = ArmGitOpsEndpoint
const EdgeDeploymentCreationBegin = EdgeGitOpsEndpoint

// Read endpoints
const resourceReadValidate = "/ResourceReadValidate"
const resourceReadBegin = "/ResourceReadBegin"

const RepoReadValidateEndpoint = RepoURLEndpoint + resourceReadValidate
const ArmDeploymentReadValidateEndpoint = ArmGitOpsEndpoint + resourceReadValidate
const EdgeDeploymentReadValidateEndpoint = EdgeGitOpsEndpoint + resourceReadValidate

const RepoReadBegin = RepoURLEndpoint + resourceReadBegin
const ArmDeploymentReadBegin = ArmGitOpsEndpoint + resourceReadBegin
const EdgeDeploymentReadBegin = EdgeGitOpsEndpoint + resourceReadBegin

// Update endpoints
const resourcePatchValidate = "/ResourcePatchValidate"
const RepoPatchValidateEndpoint = RepoURLEndpoint + resourcePatchValidate
const ArmDeploymentPatchValidateEndpoint = ArmGitOpsEndpoint + resourcePatchValidate
const EdgeDeploymentPatchValidateEndpoint = EdgeGitOpsEndpoint + resourcePatchValidate

const RepoPatchBegin = RepoURLEndpoint
const ArmDeploymentPatchBegin = ArmGitOpsEndpoint
const EdgeDeploymentPatchBegin = EdgeGitOpsEndpoint

// Delete endpoints
const resourceDeletionValidate = "/ResourceDeletionValidate"

const RepoDeletionValidateEndpoint = RepoURLEndpoint + resourceDeletionValidate
const ArmDeploymentDeletionValidateEndpoint = ArmGitOpsEndpoint + resourceDeletionValidate
const EdgeDeploymentDeletionValidateEndpoint = EdgeGitOpsEndpoint + resourceDeletionValidate

const RepoDeletionBegin = RepoURLEndpoint
const ArmDeploymentDeletionBegin = ArmGitOpsEndpoint
const EdgeDeploymentDeletionBegin = EdgeGitOpsEndpoint

const CommitEndpoint = RepoURLEndpoint + "/gitRepos/{repoName}/commit"
