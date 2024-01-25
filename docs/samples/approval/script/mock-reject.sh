#!/bin/bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

inputs_file=$1

output_file="${inputs_file%.*}-output.${inputs_file##*.}"

# reject by returning no errors and a JSON object with status 403
echo "{\"status\":403}" | jq -c '.' > "$output_file" 

# you can also reject by returning an error code
# exit 1