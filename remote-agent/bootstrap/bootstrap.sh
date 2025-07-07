#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

# Function to print usage
usage() {
    echo -e "\e[31mUsage: $0 <endpoint> <cert_path> <key_path> <target_name> <namespace> <topology> <user> <group>\e[0m"
    exit 1
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "\e[33mjq is not installed. Installing jq...\e[0m"
    sudo apt-get update && sudo apt-get install -y jq
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to install jq. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mjq installed successfully.\e[0m"
    fi
fi

# Check if the correct number of parameters are provided
if [ "$#" -ne 9 ]; then
    echo -e "\e[31mError: Invalid number of parameters.\e[0m"
    usage
fi

# Assign parameters to variables
endpoint=$1
cert_path=$2
key_path=$3
target_name=$4
namespace=$5
topology=$6
protocol=$7
user=$8
group=$9

# Validate the endpoint (basic URL validation)
if [ "$protocol" = "http" ]; then
    if ! [[ $endpoint =~ ^https?:// ]]; then
        echo -e "\e[31mError: Invalid endpoint. Must be a valid URL starting with http:// or https://\e[0m"
        usage
    fi
fi

# Validate the certificate path (check if the file exists)
if [ ! -f "$cert_path" ]; then
    echo -e "\e[31mError: Certificate file not found at path: $cert_path\e[0m"
    usage
fi

# Validate the key path (check if the file exists)
if [ ! -f "$key_path" ]; then
    echo -e "\e[31mError: Key file not found at path: $key_path\e[0m"
    usage
fi

# Validate the target name (non-empty string)
if [ -z "$target_name" ]; then
    echo -e "\e[31mError: Target name must be a non-empty string.\e[0m"
    usage
fi

# Validate the namespace (default if not provided)
if [ -z "$namespace" ]; then
    echo -e "\e[31mError: Namespace must be a non-empty string.\e[0m"
    namespace="default"
fi

# Validate the topology file (non-empty string)
if [ -z "$topology" ]; then
    echo -e "\e[31mError: Topology file must be a non-empty string.\e[0m"
    usage
fi

# Validate the user (non-empty string)
if [ -z "$user" ]; then
    echo -e "\e[31mError: User must be a non-empty string.\e[0m"
    usage
fi

# Validate the group (non-empty string)
if [ -z "$group" ]; then
    echo -e "\e[31mError: Group must be a non-empty string.\e[0m"
    usage
fi

# Validate the protocol (default to http if not provided)
if [ -z "$protocol" ]; then
    echo -e "\e[31mError: Protocol must be either 'http' or 'mqtt'.\e[0m"
    usage
fi

# Create the JSON configuration
echo -e "\e[32mCreating JSON configuration...\e[0m"
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
echo -e "\e[32mJSON configuration saved to $config_file\e[0m"

# Convert cert_path, key_path, topology_path, config to absolute paths
cert_path=$(realpath $cert_path)
key_path=$(realpath $key_path)
topology=$(realpath $topology)
config=$(realpath $config_file)

# Call the endpoint with targetName and cert
if [ "$protocol" = "http" ]; then
    # HTTP mode: Call the certificate endpoint to get the public and private keys
    bootstarpCertEndpoint="$endpoint/targets/getcert/$target_name?namespace=$namespace&osPlatform=linux"
    echo -e "\e[32mCalling certificate endpoint: $bootstarpCertEndpoint\e[0m"
    # Read the topology file content
    TOPOLOGY_DATA=$(cat "$topology")

    # Get certificate
    result=$(curl --cert "$cert_path" --key "$key_path" -X POST "$bootstarpCertEndpoint" \
            -H "Content-Type: application/json" )

    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to call certificate endpoint. Please check the endpoint and try again.\e[0m"
        exit 1
    else
        echo -e "\e[32mCertificate endpoint response received\e[0m"
    fi

    # Parse JSON response and extract fields
    public=$(echo $result | jq -r '.public')
    # Extract header and footer
    header=$(echo "$public" | awk '{print $1, $2}')
    footer=$(echo "$public" | awk '{print $(NF-1), $NF}')

    # Extract base64 content and replace spaces with newlines
    base64_content=$(echo "$public" | awk '{for (i=3; i<=NF-2; i++) printf "%s\n", $i}')

    # Combine header, base64 content and footer
    corrected_public_content="$header\n$base64_content\n$footer"

    private=$(echo $result | jq -r '.private')
    # Extract header and footer
    header=$(echo "$private" | awk '{print $1, $2, $3, $4}')
    footer=$(echo "$private" | awk '{print $(NF-3), $(NF-2), $(NF-1), $NF}')

    # Extract base64 content and replace spaces with newlines
    base64_content=$(echo "$private" | awk '{for (i=5; i<=NF-4; i++) printf "%s\n", $i}')

    # Combine header, base64 content and footer
    corrected_private_content="$header\n$base64_content\n$footer"

    # Save public key certificate to public.pem
    echo -e "$corrected_public_content" > public.pem
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to save public certificate to public.pem. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mPublic certificate saved to public.pem\e[0m"
    fi

    # Save private key to private.pem
    echo -e "$corrected_private_content" > private.pem
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to save private key to private.pem. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mPrivate key saved to private.pem\e[0m"
    fi

    # No longer update the topology here, it's handled when remote-agent starts
    echo -e "\e[32mCertificates prepared. Topology will be updated when remote-agent starts.\e[0m"
    
else
    # MQTT mode: Only need to prepare certificate files, topology will be sent through MQTT communication when agent starts
    echo -e "\e[32mMQTT mode: Topology will be sent via MQTT topic 'symphony/request/$target_name' after agent starts\e[0m"
    
    # For MQTT mode, directly use existing public and private key files
    if [ ! -f "public.pem" ]; then
        cp "$cert_path" "public.pem"
        echo -e "\e[32mCopied certificate to public.pem\e[0m"
    fi
    
    if [ ! -f "private.pem" ]; then
        cp "$key_path" "private.pem" 
        echo -e "\e[32mCopied key to private.pem\e[0m"
    fi

    # Remind user to provide CA certificate
    echo -e "\e[33mNote: When starting the remote agent, you need to provide the MQTT server's CA certificate\e[0m"
    echo -e "\e[33m      You can use the '-ca-cert' parameter to specify the CA certificate path\e[0m"
fi

# Download the remote-agent binary
if [ "$protocol" = "mqtt" ]; then
    read -p "Please input the full path to your remote-agent binary: " agent_path
    agent_path=$(realpath "$agent_path")
    if [ ! -f "$agent_path" ]; then
        echo -e "\e[31mError: remote-agent binary not found at $agent_path. Exiting...\e[0m"
        exit 1
    fi
    chmod +x "$agent_path"
    echo -e "\e[32mUsing user-supplied remote-agent binary: $agent_path\e[0m"
else
    echo -e "\e[32mDownloading remote-agent binary...\e[0m"
    curl --cert $cert_path --key $key_path -X GET "$endpoint/files/remote-agent" -o remote-agent
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to download remote-agent binary. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mRemote-agent binary downloaded\e[0m"
    fi
    chmod +x remote-agent
    agent_path=$(realpath ./remote-agent)
fi

# Make the remote-agent binary executable
chmod +x remote-agent
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to make remote-agent binary executable. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mRemote-agent binary made executable\e[0m"
fi

echo -e "\e[32mFiles created successfully:\e[0m"
echo -e "\e[32mpublic.pem\e[0m"
echo -e "\e[32mprivate.pem\e[0m"
echo -e "\e[32mremote-agent\e[0m"

# Convert public.pem, private.pem, remote-agent to absolute paths
public_path=$(realpath "./public.pem")
private_path=$(realpath "./private.pem")
# Elegantly set agent_path
if [ "$protocol" = "mqtt" ]; then
    # In MQTT mode, agent_path is already entered and verified by the user, no change needed
    :
else
    # In HTTP mode, use the downloaded remote-agent
    agent_path=$(realpath "./remote-agent")
fi

# Create the remote-agent.service file
echo -e "\e[32mCreating remote-agent.service file...\e[0m"
sudo bash -c "cat <<EOF > /etc/systemd/system/remote-agent.service
[Unit]
Description=Remote Agent Service
After=network.target

[Service]
ExecStart=$agent_path -config=$config -client-cert=$public_path -client-key=$private_path -target-name=$target_name -namespace=$namespace -topology=$topology -protocol=$protocol
Restart=always
User=$user
Group=$group

[Install]
WantedBy=multi-user.target
EOF"
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to create remote-agent.service file. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mremote-agent.service file created\e[0m"
fi

# Reload systemd to recognize the new service
sudo systemctl daemon-reload
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to reload systemd daemon. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mSystemd daemon reloaded\e[0m"
fi

# Enable the service to start on boot
sudo systemctl enable remote-agent.service
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to enable remote-agent.service. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mremote-agent.service enabled to start on boot\e[0m"
fi

# Start the service
sudo systemctl start remote-agent.service
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to start remote-agent.service. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mremote-agent.service started\e[0m"
fi

# Check the status of the service
sudo systemctl status remote-agent.service