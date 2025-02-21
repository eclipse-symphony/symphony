#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

# Function to print usage
usage() {
    echo "Usage: $0 <endpoint> <cert_path> <key_path> <target_name> <namespace> <topology>"
    exit 1
}

# Check if the correct number of parameters are provided
if [ "$#" -ne 8 ]; then
    echo "Error: Invalid number of parameters."
    usage
fi

# Assign parameters to variables
endpoint=$1
cert_path=$2
key_path=$3
target_name=$4
namespace=$5
topology=$6
user=$7
group=$8
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

# Validate the namespace (default if not provided)
if [ -z "$namespace" ]; then
    echo "Error: namespace must be a non-empty string."
    $namespace = "default"
fi

# Validate the topology file (non-empty string)
if [ -z "$topology" ]; then
    echo "Error: Topology file must be a non-empty string."
    usage
fi

# Validate the user (non-empty string)
if [ -z "$user" ]; then
    echo "Error: User must be a non-empty string."
    usage
fi  

# Validate the group (non-empty string)
if [ -z "$group" ]; then
    echo "Error: Group must be a non-empty string."
    usage
fi

# Create the JSON configuration
config_json=$(cat <<EOF
{
    "requestEndpoint": "$endpoint/solution/tasks",
    "responseEndpoint": "$endpoint/solution/task/getResult",
    "baseUrl": "$endpoint"
}
EOF
)

# Save the JSON configuration to a file
config_file="config.json"
echo "$config_json" > "$config_file"


# cert_path, key_path, topology_path, config to abosolute path
cert_path=$(realpath $cert_path)
key_path=$(realpath $key_path)
topology=$(realpath $topology)
config=$(realpath $config_file)


# call the endpoint with targetName and cert
bootstarpEndpoint="$endpoint/targets/bootstrap/$target_name?namespace=$namespace&osPlatform=linux"
# read the topology file and POST as the body
TOPOLOGY_DATA=$(cat "$topology")

result=$(curl --cert "$cert_path" --key "$key_path" -X POST "$bootstarpEndpoint" \
        -H "Content-Type: application/json" \
        -d "$TOPOLOGY_DATA")
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

# Save the public certificate to public.pem
echo -e "$corrected_public_content" > public.pem

# Save the private key to private.pem
echo -e "$corrected_private_content" > private.pem

# Download the remote-agent binary
curl --cert $cert_path --key $key_path -X GET "$endpoint/files/remote-agent" -o remote-agent

# Make the remote-agent binary executable
chmod +x remote-agent

echo "Files created successfully:"
echo "public.pem"
echo "private.pem"
echo "remote-agent"

# public.pem, private.pem, remote-agent to abosolute path
public_path=$(realpath "./public.pem")
private_path=$(realpath "./private.pem")
agent_path=$(realpath "./remote-agent")

# Create the remote-agent.service file
sudo bash -c "cat <<EOF > /etc/systemd/system/remote-agent.service
[Unit]
Description=Remote Agent Service
After=network.target

[Service]
ExecStart=$agent_path -config=$config -client-cert=$public_path -client-key=$private_path -target-name=$target_name -namespace=$namespace -topology=$topology
Restart=always
User=$user
Group=$group

[Install]
WantedBy=multi-user.target
EOF"

# Reload systemd to recognize the new service
sudo systemctl daemon-reload

# Enable the service to start on boot
sudo systemctl enable remote-agent.service

# Start the service
sudo systemctl start remote-agent.service

# Check the status of the service
sudo systemctl status remote-agent.service
