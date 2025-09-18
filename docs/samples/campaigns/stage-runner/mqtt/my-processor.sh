#!/bin/bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

inputs_file=$1

output_file="${inputs_file%.*}-output.${inputs_file##*.}"

# returns 200 status with outputs
echo "{\"status\":200}" | jq -c '.' > "$output_file" 

# return non-zero exit code to indicate failure
# return 1