#!/bin/bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

inputs_file=$1

# Read the key-value pairs from the input file
inputs=$(jq -r 'to_entries[] | "\(.key): \(.value)"' "$inputs_file")

# Initialize the updated_inputs variable as an empty object
updated_inputs={}

# Loop through the key-value pairs and print them out
while IFS= read -r line; do
    key=$(echo "$line" | cut -d ':' -f 1)
    value=$(echo "$line" | cut -d ':' -f 2 | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' | tr '[:lower:]' '[:upper:]')
    echo "KEY: $key Value: $value"
    updated_inputs=$(echo "$updated_inputs" | jq --arg key "$key" --arg value "$value" '.[$key] = $value')
done <<< "$inputs"

# Generate the output file name and write the updated key-value pairs to the output file
output_file="${inputs_file%.*}-output.${inputs_file##*.}"
echo "$updated_inputs" | jq -c '.' > "$output_file"