# providers.target.script
The script provider enables you to extend Symphony by scripts like Bash scripts and PowerShell scripts. Because Symphony built-in providers are compiled into the Symphony binary, using a new provider needs a new Symphony version. Symphony [proxy provider](../providers/proxy_provider.md), on the other hand, allows a provider to be externalized as a sidecar. However, this requires you to host a web server that implements the provider REST API. The script provider offers the most flexibility without needing an extra sidecar. 

Symphony interacts with the script provider through a staging folder. For example, when Symphony deploys a solution instance, it writes the deployment spec to a file and passes the file name to the script as a parameter. The script is expected to pick up and process the file.

## Provider Configuration
| Field | Comment |
|--------|--------|
| ```applyScript``` | Script for the ```Apply()``` method on the provider interface |
| ```getScript``` | Script for the ```Get()``` method on the provider interface |
| ```needsRemove``` | Script for the ```NeedsRemove()``` method on the provider interface |
| ```needsUpdate``` | Script for the ```NeedsUpdate()``` method on the provider interface |
| ```removeScript``` | Script for the ```Remove()``` method on the provider interface |
| ```scriptFolder``` | (optional)  The folder where the scripts are stored<sup>1</sup>. | 
| ```scriptEngine```| (optional) Script engine to use, default is ```bash```, can be either ```bash``` or ```powershell```. |
| ```stagingFolder``` | (optional) Where download scripts, and input/out files are stored | 

<sup>1</sup>: If the ```scriptFolder``` is a URL, the provider attempts to download scripts from ```scriptFolder/<script name>``` during initialization. For example, if ```scriptFolder``` is set to ```http://localhost/scripts``` and ```applyScript``` is set to ```apply.sh```, the provider will try to download from ```http://localhost/scripts/apply.sh``` and save the result to the ```stagingFolder```.

## Writing Shell Scripts

### Get script
The get script takes an input file, which contains a deployment spec JSON document, and generates an output file contains a list of component spec. In the simplest form, your script can simply play back the component list.

> **Why Symphony passes the deployment spec during GET**: Symphony assumes providers to get stateless. When it calls a provider, it provides all the required contexts for the provider to carry out specific actions. By passing the deployment spec, Symphony tells the provider what it needs to care about. For example, a Windows machine may have 100s of apps installed. The deployment spec informs the provider that only specific apps (passed in as components) need to be considered.

When the get script is called, Symphony writes a deployment spec to a temporary file and passes in the full file path as a parameter to the script. It's expected that the script will generate an output file under the same folder with a ```-output.json``` post fix. For example, if the input is ```abc.json```, then the output should be ```abc-output.json```.

```bash
#!/bin/bash
a=$1
echo $(echo $(<$1) | jq '.solution.components') > ${a%.*}-output.${a##*.}
```
Technically, the above component list may contain more components that need to be handled. In the deployment spec object, there's an ```assignment``` property that further constraints which target should care about which components, for example (see ```samples/script-provider/deployment.json``` for a complete sample of deployment input file).

```json
 "assignments": {
    "target-1": "{component-1,component-2}"
},
```

This means although the component list may contain more components, ```target-1``` only needs to care about ```component-1``` and ```component-2```.

Regardless, the above script generally works because when Symphony decides if a deployment needs to be updated or removed, it compares only components that are defined in the desired state of the deployment. If you returned extra components in the ```GET``` request, they will be ignored. 

You should check the target system to see if the request components are at the desired state. For example, the following PowerShell script checks if a "notepad" process is running as specified in the deployment spec:
```json
...
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
...
```
The script iterates through the component list, and checks if "notepad" is running. If not, it removes the component from the list and writes the current state as the output file:
```powershell
param(
    [String]$InputFile
)

# Load the JSON from the input file
$json = Get-Content -Encoding UTF8 $InputFile | ConvertFrom-Json

# Loop through the components and remove those with app.package equals to "notepad" if notepad process is not running
foreach ($component in $json.Solution.Components) {
    if ($component.Properties."app.package" -eq "notepad") {
        if ((Get-Process -Name "notepad" -ErrorAction SilentlyContinue) -eq $null) {
            # Remove the component from the Components list
            $json.Solution.Components = $json.Solution.Components | Where-Object { $_ -ne $component }
        }
    }
}

# Write the updated JSON to an output file
"[" + ($json.Solution.Components | ConvertTo-Json -Compress) + "]" | Out-File -Encoding ASCII $InputFile.Replace(".json", "-output.json")
```
### Apply script
The apply script reads component assignments and carries out actual deployment actions. The following Shell script loops over components and displays their container image names. 

The script should return a ```true``` string on a successful application.

```bash
#!/bin/bash

# loop through components from the deployment spec. The following example displays container.image property
components=$(echo $(<$1) | jq -c -r '.solution.components[]')
for component in ${components[@]}; do
    echo $component | jq '.properties."container.image"'
    # do you application logic here
done  
# return "true" or "1" to indicate a successful deployment
echo "true"
```

### Remove script
The remove script reads component assignments and removes the components from the target.

The script should return a ```true``` string on a successful removal.

### NeedsUpdate script
This script is optional. It takes two parameters, which are two JSON files that contain the current component list and the desired component list, respectively. The script is supposed to compare the two lists and return ```true``` or ```1``` if an update is needed. If this script is omitted, the default behavior is used: if all the desired components exist (with the same content) in the current component list, then an update is ignored.

The following PowerShell example shows a typical NeedsUpdate implemenation:

```powershell
param (
    [Parameter(Mandatory=$true)]
    [string]$file1,

    [Parameter(Mandatory=$true)]
    [string]$file2
)

# Load the JSON from the input files
$json1 = Get-Content -Path $file1 -Raw | ConvertFrom-Json
$json2 = Get-Content -Path $file2 -Raw | ConvertFrom-Json

# Convert the JSON objects to strings and compare them for equivalence
if ($json1 | ConvertTo-Json -Compress -Depth 100 -ErrorAction SilentlyContinue -WarningAction SilentlyContinue -InformationAction SilentlyContinue -OutVariable str1 | Out-Null) {
    $str1 = $json1 | Out-String
}
if ($json2 | ConvertTo-Json -Compress -Depth 100 -ErrorAction SilentlyContinue -WarningAction SilentlyContinue -InformationAction SilentlyContinue -OutVariable str2 | Out-Null) {
    $str2 = $json2 | Out-String
}

if ($str1 -eq $str2) {
    Write-Output "false"
}
else {
    Write-Output "true"
}

```

### NeedsRemove script
This script is optional. It takes two parameters, which are two JSON files that contain the current component list and the desired component list, respectively. The script is supposed to compare the two lists and return ```true``` or ```1``` if  removal is needed. If this script is omitted, the default behavior is used: if any of the desired components exist in the current component list (regardless of content differences), then removal is needed.

## Related Topics

* [Provider Interface](./provider_interface.md)
