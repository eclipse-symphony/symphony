#!/bin/bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

inputs_file=$1

output_file="${inputs_file%.*}-output.${inputs_file##*.}"

# approve by returning no errors and a JSON object with status
echo "{\"status\":200}" | jq -c '.' > "$output_file" 
