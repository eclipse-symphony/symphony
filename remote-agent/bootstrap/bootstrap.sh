#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

# Function to print usage
usage() {
    echo "Usage: $0 <endpoint> <cert_path>"
    exit 1
}

# Check if the correct number of parameters are provided
if [ "$#" -ne 3 ]; then
    echo "Error: Invalid number of parameters."
    usage
fi

# Assign parameters to variables
endpoint=$1
cert_path=$2
key_path=$3
target_name=$4
namespace=$5

# Validate the endpoint (basic URL validation)
if ! [[ $endpoint =~ ^https?:// ]]; then
    echo "Error: Invalid endpoint. Must be a valid URL starting with http:// or https://"
    usage
fi

# Validate the certificate path (check if the file exists)
if [ ! -f "$cert_path" ]; then
    echo "Error: Certificate file not found at path: $cert_path"
    usage
fi

# Validate the certificate path (check if the file exists)
if [ ! -f "$key_path" ]; then
    echo "Error: key file not found at path: $key_path"
    usage
fi

# Validate the target name (non-empty string)
if [ -z "$target_name" ]; then
    echo "Error: Target name must be a non-empty string."
    usage
fi

# Validate the target name (non-empty string)
if [ -z "$namespace" ]; then
    echo "Error: namespace must be a non-empty string."
    usage
fi


# call the endpoint with targetName and cert
bootstarpEndpoint="$endpoint/targets/bootstrap/$endpoint?namespace=$namespace&osPlatform=linux"
result=$(curl --cert "$cert_path" --key "$key_path" -X POST "$endpoint")
# Parse the JSON response and extract the fields
public=$(echo $result | jq -r '.public')
# Extract the header and footer
header=$(echo "$public" | awk '{print $1, $2}')
footer=$(echo "$public" | awk '{print $(NF-1), $NF}')

# Extract the base64 content and replace spaces with newlines
base64_content=$(echo "$public" | awk '{for (i=3; i<=NF-2; i++) printf "%s\n", $i}')

# Combine the header, base64 content, and footer
corrected_public_content="$header\n$base64_content\n$footer"

private=$(echo $result | jq -r '.private')
# Extract the header and footer
header=$(echo "$private" | awk '{print $1, $2, $3, $4}')
footer=$(echo "$private" | awk '{print $(NF-3), $(NF-2), $(NF-1), $NF}')

# Extract the base64 content and replace spaces with newlines
base64_content=$(echo "$private" | awk '{for (i=5; i<=NF-4; i++) printf "%s\n", $i}')

# Combine the header, base64 content, and footer
corrected_private_content="$header\n$base64_content\n$footer"
file=$(echo $result | jq -r '.file')

# Save the public certificate to public.pem
echo -e "$corrected_public_content" > public.pem

# Save the private key to private.pem
echo -e "$corrected_private_content" > private.pem

# Decode the base64-encoded binary data and save it to remote-agent
echo "$file" | base64 --decode > remote-agent

# Make the remote-agent binary executable
chmod +x remote-agent

echo "Files created successfully:"
echo "public.pem"
echo "private.pem"
echo "remote-agent"

./remote-agent -config=./config.json -client-cert=./public.pem -client-key=./private.pem
