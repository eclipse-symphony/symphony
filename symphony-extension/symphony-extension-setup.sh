#!/usr/bin/bash
echo "This script will reset your cluster and install a fresh arc-enabled one with symphony arc extension installed"
read -p "Enter your name or alias: " name
read -p "Enter the version you are working on (please pick a version existing in the symphonycr/arc-extension acr that is not the latest version): " version
echo "*********** Setting up your environment ***********"
minikube delete
minikube start -n 2
ExtensionType="private.projectalicesprings" 
ServiceAccount="private.projectalicesprings"

postFix=$(date +%m%d%H%M%S)
location="westus"
rg="$name-rg-$postFix"
extension="$name-extension-$postFix"
cluster="$name-cluster-$postFix"
customL="$name-location-$postFix"
subscription="77969078-2897-47b0-9143-917252379303"

# Connectedk8s creates a new resource group if one is not already provisioned
az connectedk8s connect --name $cluster -g $rg --location $location

az k8s-extension create -c $cluster -g $rg --name $extension --cluster-type connectedClusters --extension-type $ExtensionType --config Microsoft.CustomLocation.ServiceAccount=$ServiceAccount --release-train dev --auto-upgrade false --version $version
# --config installServiceExt=true

ConnectedClusterResourceId=$(az connectedk8s show -g $rg -n $cluster --query id -o tsv)
ClusterExtensionResourceId=$(az k8s-extension show -g $rg -c $cluster --cluster-type connectedClusters --name $extension --query id -o tsv)

az customlocation create -g $rg -n $customL --namespace "arc" --host-resource-id $ConnectedClusterResourceId --cluster-extension-ids $ClusterExtensionResourceId

# accessToken=$(az account get-access-token --resource https://management.azure.com | jq -r '.accessToken')
# echo "Setting up resource sync rules"
# curl -X POST \
# -H "Authorization: Bearer $accessToken" \
# -H "Content-Type: application/json" \
# -d "{
#     "location": "westus",
#     "properties":{
#         "targetResourceGroup": "/subscriptions/$subscription/resourceGroups/$rg",
#         "priority": 100,
#         "selector": {
#             "matchLabels": {
#                 "management.azure.com/provider-name" : "Private.Symphony"
#             }
#         }
#     }
#     }
# }" \
# "https://management.azure.com/subscriptions/$subscription/resourceGroups/$rg/providers/microsoft.extendedlocation/customlocations/$customL/resourcesyncrules/defaultresourcesyncrule?api-version=2021-08-31-preview"


echo "*********** finished ***********"

echo "*********** Here is your relevant info ***********" 
echo "resource group: $rg"
echo "extension name: $extension"
echo "arc cluster name: $cluster"
echo "custom location name: $customL"
echo "location: $location"
