#!/bin/bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

deployment=$1 # first parameter file is the deployment object
references=$2 # second parmeter file contains the reference components

# the remove script is called with a list of components to be deleted via
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
# 8001: fialed to update
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
