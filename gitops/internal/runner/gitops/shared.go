package gitops

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

var (
	azClientOptions = policy.ClientOptions{
		Cloud: cloud.Configuration{
			ActiveDirectoryAuthorityHost: os.Getenv("AZURE_AUTHORITY_HOST"),
			Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
				cloud.ResourceManager: {
					Endpoint: os.Getenv("AZURE_RESOURCE_MANAGER_ENDPOINT"),
					Audience: "https://management.azure.com",
				},
			},
		},
	}
	defaultAzCredentialOptions = azidentity.DefaultAzureCredentialOptions{
		ClientOptions: azClientOptions,
	}
	armclientOptions = arm.ClientOptions{
		ClientOptions: azClientOptions,
	}
)
