# providers.target.script

_(last edit: 6/26/2023)_

The script provider enables you to extend Symphony with scripts like Bash scripts and PowerShell scripts. Because Symphony's built-in providers are compiled into the Symphony binary, using a new provider needs a new Symphony version. Symphony [HTTP proxy provider](../http_proxy_provider.md) or Symphony [MQTT proxy provider](../mqtt_proxy_provider.md), on the other hand, allows a provider to be externalized as a sidecar. However, this requires you to host a web server that implements the provider REST API. The script provider offers the most flexibility without needing an extra sidecar.

Symphony interacts with the script provider through a staging folder. For example, when Symphony deploys a solution instance, it writes the deployment spec to a file and passes the file name to the script as a parameter. The script is expected to pick up and process the file.

## Provider configuration

| Field | Comment |
|--------|--------|
| `applyScript` | Script for applying updated components from the `Apply()` method on the provider interface |
| `getScript` | Script for the `Get()` method on the provider interface |
| `removeScript` | Script for removing deleted components from the `Apply()` method on the provider interface |
| `scriptFolder` | (optional)  The folder where the scripts are stored<sup>1</sup>. | 
| `scriptEngine`| (optional) Script engine to use, default is `bash`, can be either `bash` or `powershell`. |
| `stagingFolder` | (optional) Where download scripts, and input/out files are stored | 

1: If the `scriptFolder` is a URL, the provider attempts to download scripts from `scriptFolder/<script name>` during initialization. For example, if `scriptFolder` is set to `http://localhost/scripts` and `applyScript` is set to `apply.sh`, the provider will try to download from `http://localhost/scripts/apply.sh` and save the result to the `stagingFolder`.

## Write shell scripts

### Get script

The get script takes two input files. The first file contains a deployment spec JSON document, and the second file contains a list of interested component specs. The get script generates an output file containing a list of component specs that reflect what are installed on the target system.

Symphony passes the deployment spec and a reference component list during GET because Symphony assumes that providers are stateless. When it calls a provider, it provides all the required contexts for the provider to carry out specific actions. By passing the reference component list, Symphony tells the provider what it needs to care about. For example, a Windows machine may have hundreds of apps installed. The reference component list informs the provider that only specific apps need to be considered.

When the get script is called, Symphony writes a deployment spec to a temporary file and passes in the full file path as a parameter to the script. It's expected that the script will generate an output file under the same folder with a `-output.json` suffix. For example, if the input is `abc.json`, then the output should be `abc-output.json`.

```bash
#!/bin/bash

deployment=$1 # first parameter file is the deployment object
references=$2 # second parameter file contains the reference components

# to get the list components you need to return during this Get() call, you can 
# read from the references parameter file. This file gives you a list of components and 
# their associated actions, which can be either "update" or "delete". Your script is 
# supposed to use this list as a reference (regardless of the action flag) to collect
# the current state of the corresponding components, and return the list. If a component
# doesn't exist, simply skip the component. 

components=$(jq -c '.[]' "$references")

while IFS= read -r line; do
    # Extract the name and age fields from each JSON object
    action=$(echo "$line" | jq -r '.action')
    component=$(echo "$line" | jq -r '.component')
    echo "ACTION: $action"
    echo "COMPONENT: $component"
done <<< "$components"

# optionally, you can use the deployment parameter to get additional contextual information as needed.
# for example, you can the following query to get the instance scope. 

scope=$(jq '.instance.scope' "$deployment")
echo "SCOPE: $scope"

# the following is an example of generating required output file. The example simply extracts
# all reference components and writes them into the output JSON file.

output_components=$(jq -r '[.[] | .component]' "$references")
echo "$output_components" > ${deployment%.*}-output.${deployment##*.}
```

You should check the target system to see if the request components are at the desired state. For example, the following snippet checks if a notepad process is running as specified in the deployment spec:

```json
"components": [
   {
    "name": "notepad",
    "type": "notepad",
    "properties": {
     "app.package": "notepad",
     "state": "running"
    }
   }
]
```

The script iterates through the component list, and checks if notepad is running. If not, it removes the component from the list and writes the current state as the output file:

```powershell
param(
    [String]$DeploymentFile,
    [String]$ComponentListFile
)

# Load the JSON from the input file
$json = Get-Content -Encoding UTF8 $ComponentListFile | ConvertFrom-Json

# Loop through the components and remove those with app.package equals to "notepad" if notepad process is not running
foreach ($component in $json) {
    if ($component.Component.Properties."app.package" -eq "notepad") {
        if ((Get-Process -Name "notepad" -ErrorAction SilentlyContinue) -eq $null) {
            # Remove the component from the Components list
            $json= $json | Where-Object { $_ -ne $component }
        }
    }
}

# Write the updated JSON to an output file
"[" + ($json | ForEach-Object {$_.Component} | ConvertTo-Json -Compress) + "]" | Out-File -Encoding ASCII $DeploymentFile.Replace(".json", "-output.json")
```

### Apply script

The apply script reads reference component list (the second file parameter) and carries out the actual deployment actions. The script is supposed to generate an output JSON file that contains the deployment status of each component.

The following is a sample script that reads the component list, displays component properties, and then generates a hardcoded output file:

```bash
#!/bin/bash

deployment=$1 # first parameter file is the deployment object
references=$2 # second parmeter file contains the reference components

# the apply script is called with a list of components to be updated via
# the references parameter
components=$(jq -c '.[]' "$references")

echo "COMPONENTS: $components"

while IFS= read -r line; do
    # Extract the name and age fields from each JSON object
    name=$(echo "$line" | jq -r '.name')
    properties=$(echo "$line" | jq -r '.properties')
    echo "NAME: $name"
    echo "PROPERTIES: $properties"
done <<< "$components"

# optionally, you can use the deployment parameter to get additional contextual information as needed.
# for example, you can the following query to get the instance scope. 

scope=$(jq '.instance.scope' "$deployment")
echo "SCOPE: $scope"


# your script needs to generate an output file that contains a map of component results. For each
# component result, the status code should be one of
# 8001: failed to update
# 8002: failed to delete
# 8003: failed to validate component artifact
# 8004: updated (success)
# 8005: deleted (success)
# 9998: untouched - no actions are taken/necessary

output_results='{
    "com1": {
        "status": 8004,
        "message": ""
    },
    "com2": {
        "status": 8001,
        "message": "update error message" 
    }
}'

echo "$output_results" > ${deployment%.*}-output.${deployment##*.}

```


Find full scenarios at [this location](../../../samples/script-provider/solution.yaml)

### Remove script

The remove script reads component assignments and removes the components from the target.

## Related topics

* [Provider Interface](./provider_interface.md)