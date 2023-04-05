# Resource Sync Rules Configuration

The following endpoint was used for the Dogfood environment

https://api-dogfood.resources.windows-int.net/subscriptions/34d077e3-2c1a-48f5-8629-ffc0e657e0ed/resourceGroups/{rgName}/providers/microsoft.extendedlocation/customlocations/{customLocationName}/resourcesyncrules/defaultresourcesyncrule?api-version=2021-08-31-preview

```json
{ 
    "location": "westus",
    "properties": {
        "targetResourceGroup": "/subscriptions/34d077e3-2c1a-48f5-8629-ffc0e657e0ed/resourceGroups/{rgName}",
        "priority": 100,
        "selector": {
            "matchLabels": {
                "management.azure.com/provider-name": "Microsoft.Symphony"
            }
        }
    }
}
```
# Resource Sync Rules Output
```json
{
    "id": "/subscriptions/34d077e3-2c1a-48f5-8629-ffc0e657e0ed/resourcegroups/{rgName}/providers/microsoft.extendedlocation/customlocations/{customLocationName}/resourcesyncrules/defaultresourcesyncrule",
    "name": "defaultresourcesyncrule",
    "location": "westus",
    "type": "Microsoft.ExtendedLocation/customLocations/resourceSyncRules",
    "systemData": {
        "createdBy": "{usersEmail}",
        "createdByType": "User",
        "createdAt": "2023-02-15T16:54:47.4122631Z",
        "lastModifiedBy": "{usersEmail}",
        "lastModifiedByType": "User",
        "lastModifiedAt": "2023-02-15T17:01:10.4225547Z"
    },
    "properties": {
        "priority": 100,
        "provisioningState": "Updating",
        "targetResourceGroup": "/subscriptions/34d077e3-2c1a-48f5-8629-ffc0e657e0ed/resourceGroups/{rgName}",
        "selector": {
            "matchLabels": {
                "management.azure.com/provider-name": "Microsoft.Symphony"
            }
        }
    }
}
```