#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

# Function to print usage
usage() {
    echo -e "\e[31mUsage for HTTP mode:\e[0m"
    echo -e "\e[31m  $0 http <endpoint> <cert_path> <key_path> <target_name> <namespace> <topology> <user> <group>\e[0m"
    echo -e "\e[31mUsage for MQTT mode:\e[0m"
    echo -e "\e[31m  $0 mqtt <broker_address> <broker_port> <cert_path> <key_path> <target_name> <namespace> <topology> <user> <group> [binary_path] [ca_cert_path]\e[0m"
    echo -e "\e[31mNote: binary_path is required when protocol is 'mqtt'\e[0m"
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

# Assign parameters to variables
protocol=$1

if [ "$protocol" = "http" ]; then
    endpoint=$2
    cert_path=$3
    key_path=$4
    target_name=$5
    namespace=$6
    topology=$7
    user=$8
    group=$9
    
    # Validate the endpoint (basic URL validation)
    if ! [[ $endpoint =~ ^https?:// ]]; then
        echo -e "\e[31mError: Invalid endpoint. Must be a valid URL starting with http:// or https://\e[0m"
        usage
    fi
    
    # Create the JSON configuration for HTTP
    echo -e "\e[32mCreating JSON configuration for HTTP mode...\e[0m"
    config_json=$(cat <<EOF
{
    "requestEndpoint": "$endpoint/solution/tasks",
    "responseEndpoint": "$endpoint/solution/task/getResult",
    "baseUrl": "$endpoint"
}
EOF
    )
elif [ "$protocol" = "mqtt" ]; then
    broker_address=$2
    broker_port=$3
    cert_path=$4
    key_path=$5
    target_name=$6
    namespace=$7
    topology=$8
    user=$9
    group=${10}
    binary_path=${11}
    ca_cert_path=${12}
    
    # Validate MQTT broker address
    if [ -z "$broker_address" ]; then
        echo -e "\e[31mError: MQTT broker address cannot be empty\e[0m"
        usage
    fi
    
    # Validate MQTT broker port
    if [ -z "$broker_port" ] || ! [[ "$broker_port" =~ ^[0-9]+$ ]]; then
        echo -e "\e[31mError: MQTT broker port must be a valid number\e[0m"
        usage
    fi
    
    # Create the JSON configuration for MQTT
    echo -e "\e[32mCreating JSON configuration for MQTT mode...\e[0m"
    config_json=$(cat <<EOF
{
    "mqttBroker": "$broker_address",
    "mqttPort": $broker_port
}
EOF
    )
    
    # Check for CA certificate
    if [ -z "$ca_cert_path" ]; then
        echo -e "\e[33mWarning: CA certificate path not provided for MQTT. You may need this for secure MQTT connections.\e[0m"
        read -p "Do you want to provide a CA certificate path? [y/N]: " provide_ca
        if [[ "$provide_ca" =~ ^[Yy]$ ]]; then
            read -p "Please enter the CA certificate path: " ca_cert_path
        fi
    fi
    
    if [ ! -z "$ca_cert_path" ] && [ ! -f "$ca_cert_path" ]; then
        echo -e "\e[31mError: CA certificate file not found at path: $ca_cert_path\e[0m"
        usage
    fi
    
    # 修复MQTT证书格式化代码中的语法错误
    # Check if certificate format needs correction (if it's in single line format)
    if [ -f "$cert_path" ] && grep -q "BEGIN CERTIFICATE" "$cert_path" && ! grep -q "END CERTIFICATE" "$cert_path" ]; then
        echo -e "\e[33mDetected certificate in single line format. Reformatting...\e[0m"
        
        # Extract certificate content
        cert_content=$(cat "$cert_path")
        
        # Extract header and footer
        header=$(echo "$cert_content" | grep -o "-----BEGIN CERTIFICATE-----")
        footer=$(echo "$cert_content" | grep -o "-----END CERTIFICATE-----")
        
        # Extract base64 content between header and footer
        base64_content=$(echo "$cert_content" | sed "s/$header//g" | sed "s/$footer//g")
        
        # Format certificate with proper line breaks
        corrected_cert_content="$header\n$(echo "$base64_content" | fold -w 64)\n$footer"
        
        # Create backup of original certificate
        cp "$cert_path" "${cert_path}.bak"
        
        # Save reformatted certificate
        echo -e "$corrected_cert_content" > "$cert_path"
        echo -e "\e[32mCertificate reformatted successfully. Original saved as ${cert_path}.bak\e[0m"
    fi
    
    # Perform similar check and reformatting for key file if needed
    if [ -f "$key_path" ] && grep -q "BEGIN PRIVATE KEY" "$key_path" && ! grep -q "END PRIVATE KEY" "$key_path" ]; then
        echo -e "\e[33mDetected key file in single line format. Reformatting...\e[0m"
        
        # Extract key content
        key_content=$(cat "$key_path")
        
        # Extract header and footer
        header=$(echo "$key_content" | grep -o "-----BEGIN PRIVATE KEY-----")
        footer=$(echo "$key_content" | grep -o "-----END PRIVATE KEY-----")
        
        # Extract base64 content between header and footer
        base64_content=$(echo "$key_content" | sed "s/$header//g" | sed "s/$footer//g")
        
        # Format key with proper line breaks
        corrected_key_content="$header\n$(echo "$base64_content" | fold -w 64)\n$footer"
        
        # Create backup of original key
        cp "$key_path" "${key_path}.bak"
        
        # Save reformatted key
        echo -e "$corrected_key_content" > "$key_path"
        echo -e "\e[32mKey file reformatted successfully. Original saved as ${key_path}.bak\e[0m"
    fi
    
else
    echo -e "\e[31mError: Protocol must be either 'http' or 'mqtt'.\e[0m"
    usage
fi

# Common validations regardless of protocol
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

# Save the JSON configuration to a file
config_file="config.json"
echo "$config_json" > "$config_file"
echo -e "\e[32mJSON configuration saved to $config_file\e[0m"

# Convert cert_path, key_path, topology_path, config to absolute paths
cert_path=$(realpath $cert_path)
key_path=$(realpath $key_path)
topology=$(realpath $topology)
config=$(realpath $config_file)

# Protocol-specific handling
if [ "$protocol" = "http" ]; then
    # HTTP模式证书处理代码 - 这部分代码保持不变
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
    # MQTT mode: Use original certificates and binary
    echo -e "\e[32mMQTT mode: Using original certificates directly\e[0m"
    echo -e "\e[32mCertificate path: $cert_path\e[0m"
    echo -e "\e[32mKey path: $key_path\e[0m"
    echo -e "\e[32mBroker: $broker_address:$broker_port\e[0m"
    echo -e "\e[32mTopology will be sent via MQTT topic 'symphony/request/$target_name' after agent starts\e[0m"
    
    # Handle binary path for MQTT mode
    if [ -z "$binary_path" ]; then
        read -p "Please input the full path to your remote-agent binary: " agent_path
    else
        agent_path="$binary_path"
        echo -e "\e[32mUsing provided binary path: $agent_path\e[0m"
    fi
    
    agent_path=$(realpath "$agent_path")
    if [ ! -f "$agent_path" ]; then
        echo -e "\e[31mError: remote-agent binary not found at $agent_path. Exiting...\e[0m"
        exit 1
    fi
    chmod +x "$agent_path"
    echo -e "\e[32mUsing remote-agent binary: $agent_path\e[0m"
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
if [ "$protocol" = "http" ]; then
    echo -e "\e[32mpublic.pem\e[0m"
    echo -e "\e[32mprivate.pem\e[0m"
fi
echo -e "\e[32mremote-agent\e[0m"

# Set certificate paths appropriately based on protocol
if [ "$protocol" = "http" ]; then
    # HTTP mode: Use the generated public.pem and private.pem
    public_path=$(realpath "./public.pem")
    private_path=$(realpath "./private.pem")
else
    # MQTT mode: Use the original certificate paths
    public_path=$cert_path
    private_path=$key_path
fi

# Build the service command
service_command="$agent_path -config=$config -client-cert=$public_path -client-key=$private_path -target-name=$target_name -namespace=$namespace -topology=$topology -protocol=$protocol"

# Add CA certificate parameter if available
if [ "$protocol" = "mqtt" ] && [ ! -z "$ca_cert_path" ]; then
    ca_cert_path=$(realpath "$ca_cert_path")
    service_command="$service_command -ca-cert=$ca_cert_path"
    echo -e "\e[32mUsing CA certificate: $ca_cert_path\e[0m"
fi

echo -e "\e[32mService command: $service_command\e[0m"
# Create the remote-agent.service file
echo -e "\e[32mCreating remote-agent.service file...\e[0m"
sudo bash -c "cat <<EOF > /etc/systemd/system/remote-agent.service
[Unit]
Description=Remote Agent Service
After=network.target

[Service]
ExecStart=$service_command
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
sudo systemctl stop remote-agent.service