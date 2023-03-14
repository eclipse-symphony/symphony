

## Build Pipelines

[Symphony - Buddy Build](https://dev.azure.com/msazure/One/_build?definitionId=311591)
- Type: Buddy
- Trigger: Manual
- ACR: symphonycr.azurcr.io

[Symphony - PR Build](https://dev.azure.com/msazure/One/_build?definitionId=<ADD_DEF_ID>)
- Type: PullRequest
- Trigger: PR into 'main' branch
- ACR: nopush

