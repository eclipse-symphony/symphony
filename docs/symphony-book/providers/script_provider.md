# providers.target.script
The script provider enables you to extend Symphony by scripts like Bash scripts and PowerShell scripts. Because Symphony built-in providers are compiled into the Symphony binary, using a new provider needs a new Symphony version. Symphony [proxy provider](../providers/proxy_provider.md), on the other hand, allows a provider to be externalized as a sidecar. However, this requires you to host a web server that implements the provider REST API. The script provider offers the most flexibility without needing an extra sidecar. 

Symphony interacts with the script provider through a staging folder. For example, when Symphony deploys a solution instance, it writes the deployment spec to a file and passes the file name to the script as a parameter. The script is expected to pick up and process the file.

## Provider Configuration
| Field | Comment |
|--------|--------|
| ```applyScript``` | Script for the ```Apply()``` method on the provider interface |
| ```getScript``` | Script for the ```Get()``` method on the provider interface |
| ```needsRemove``` | (optional) Script for the ```NeedsRemove()``` method on the provider interface |
| ```needsUpdate``` | (optional) Script for the ```NeedsUpdate()``` method on the provider interface |
| ```removeScript``` | Script for the ```Remove()``` method on the provider interface |
| ```scriptFolder``` | (optional)  The folder where the scripts are stored<sup>1</sup>. | 
| ```stagingFolder``` | (optional) Where download scripts, and input/out files are stored | 

<sup>1</sup>: If the ```scriptFolder``` is a URL, the provider attempts to download scripts from ```scriptFolder/<script name>``` during initialization. For example, if ```scriptFolder``` is set to ```http://localhost/scripts``` and ```applyScript``` is set to ```apply.sh```, the provider will try to download from ```http://localhost/scripts/apply.sh``` and save the result to the ```stagingFolder```.

## Writing Shell Scripts

### Get script
The get script takes an input file, which contains a deployment spec JSON document, and generates an output file contains a list of component spec. In the simpliest 
```bash
#!/bin/bash
a=$1
echo $(echo $(<$1) | jq '.solution.components') > ${a%.*}-output.${a##*.}
```
This generally works because how Symphony decides if an update is needed, or a removal needs to be carried out. Technically, you should only return components that are assigned and present on the current target. To check if an update is needed, Symphony looks at the desired state and compares it with the desired state. If all components in the desired state are found in the current state, an update is skipped. Hence, returning more components doesn’t affect the logic. The following JSON segment shows an example of how components are assigned to targets (see ```samples/script-provider/deployment.json``` for a complete sample of deployment input file).
```json
 "assignments": {
    "target-1": "{component-1,component-2}"
},
```

### Apply script
The apply script reads component assignments and carries out actual deployment actions. The following Shell script loops over components and display their container image names. 

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

### NeedsUpdate script
This script is optional. It takes two parameters, which are two JSON files that contain the current component list and the desired component list, respectively. The script is supposed to compare the two list and return ```true``` or ```1``` if an update is needed. If this script is omitted, the default behavior is used: if all the desired components exist (with the same content) in the current component list, then an update is ignored.

### NeedsRemove script
This script is optional. It takes two parameters, which are two JSON files that contain the current component list and the desired component list, respectively. The script is supposed to compare the two list and return ```true``` or ```1``` if  removal is needed. If this script is omitted, the default behavior is used: if any of the desired components exist in the current component list (regardless of content differences), then removal is needed.

## Related Topics

* [Provider Interface](./provider_interface.md)