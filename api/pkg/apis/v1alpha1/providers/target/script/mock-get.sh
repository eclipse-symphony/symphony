#!/bin/bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

deployment=$1 # first parameter file is the deployment object
references=$2 # second parmeter file contains the reference components

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
echo "$output_components" > ${deployment%.*}-get-output.${deployment##*.}